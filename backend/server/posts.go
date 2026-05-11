package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listPostsHandler(w, r)
		return
	}

	if r.Method == http.MethodPost {
		createPostHandler(w, r)
		return
	}

	errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
}

func listPostsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT 
			posts.id,
			posts.content,
			posts.privacy,
			posts.created_at,
			users.first_name,
			users.last_name
		FROM posts
		JOIN users ON users.id = posts.user_id
		WHERE posts.privacy = 'public'
		ORDER BY posts.created_at DESC
	`)

	if err != nil {
		errorJSON(w, "could not load posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}

	for rows.Next() {
		var id int
		var content, privacy, createdAt, firstName, lastName string

		err := rows.Scan(&id, &content, &privacy, &createdAt, &firstName, &lastName)
		if err != nil {
			errorJSON(w, "could not read posts", http.StatusInternalServerError)
			return
		}

		posts = append(posts, map[string]interface{}{
			"id":         id,
			"content":    content,
			"privacy":    privacy,
			"created_at": createdAt,
			"author":     firstName + " " + lastName,
		})
	}

	writeJSON(w, http.StatusOK, posts)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	var req CreatePostRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	req.Privacy = strings.TrimSpace(req.Privacy)

	if req.Content == "" {
		errorJSON(w, "post content is required", http.StatusBadRequest)
		return
	}

	if req.Privacy == "" {
		req.Privacy = "public"
	}

	if req.Privacy != "public" && req.Privacy != "followers" && req.Privacy != "private" {
		errorJSON(w, "invalid privacy value", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO posts (
			user_id,
			content,
			privacy
		)
		VALUES (?, ?, ?)
	`,
		userID,
		req.Content,
		req.Privacy,
	)

	if err != nil {
		errorJSON(w, "could not create post", http.StatusInternalServerError)
		return
	}

	postID, err := result.LastInsertId()
	if err != nil {
		errorJSON(w, "post created but could not read post id", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "post created successfully",
		"post_id": postID,
	})
}
