package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		errorJSON(w, "missing required fields", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errorJSON(w, "could not hash password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		INSERT INTO users 
		(email, password_hash, first_name, last_name, date_of_birth, nickname, about_me)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, req.Email, string(hash), req.FirstName, req.LastName, req.DateOfBirth, req.Nickname, req.AboutMe)

	if err != nil {
		errorJSON(w, "could not create user, email may already exist", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "user registered successfully",
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	var userID int
	var passwordHash string

	err := db.QueryRow(`
		SELECT id, password_hash
		FROM users
		WHERE email = ?
	`, req.Email).Scan(&userID, &passwordHash)

	if err != nil {
		errorJSON(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		errorJSON(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	sessionID, err := generateSessionID()
	if err != nil {
		errorJSON(w, "could not create session", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`, sessionID, userID, expiresAt)

	if err != nil {
		errorJSON(w, "could not save session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "login successful",
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	var user User

	err := db.QueryRow(`
		SELECT id, email, first_name, last_name, COALESCE(nickname, '')
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Nickname)

	if err != nil {
		errorJSON(w, "user not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
