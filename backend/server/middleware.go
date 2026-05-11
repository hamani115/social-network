package server

import (
	"net/http"
	"time"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			errorJSON(w, "not authenticated", http.StatusUnauthorized)
			return
		}

		var userID int
		var expiresAt time.Time

		err = db.QueryRow(`
			SELECT user_id, expires_at
			FROM sessions
			WHERE id = ?
		`, cookie.Value).Scan(&userID, &expiresAt)

		if err != nil {
			errorJSON(w, "invalid session", http.StatusUnauthorized)
			return
		}

		if time.Now().After(expiresAt) {
			db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
			errorJSON(w, "session expired", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = contextWithUserID(ctx, userID)

		next(w, r.WithContext(ctx))
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
