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
	userID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			posts.id,
			posts.user_id,
			users.first_name || ' ' || users.last_name AS author_name,
			COALESCE(users.nickname, '') AS author_nickname,
			posts.content,
			COALESCE(posts.image_path, '') AS image_path,
			posts.privacy,
			posts.created_at
		FROM posts
		JOIN users ON users.id = posts.user_id
		WHERE posts.privacy = 'public'
		   OR posts.user_id = ?
		ORDER BY posts.created_at DESC
	`, userID)

	if err != nil {
		errorJSON(w, "could not load posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := []PostResponse{}

	for rows.Next() {
		var post PostResponse

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorNickname,
			&post.Content,
			&post.ImagePath,
			&post.Privacy,
			&post.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read post data", http.StatusInternalServerError)
			return
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading posts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	var content string
	var privacy string
	var imagePath string

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			errorJSON(w, "could not read form data", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(r.FormValue("content"))
		privacy = strings.TrimSpace(r.FormValue("privacy"))

		imagePath, err = saveUploadedImage(r, "image", "uploads/posts")
		if err != nil {
			errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else { // ! WILL REMOVE LATER
		var req CreatePostRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorJSON(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(req.Content)
		privacy = strings.TrimSpace(req.Privacy)
	}

	if content == "" {
		errorJSON(w, "post content is required", http.StatusBadRequest)
		return
	}

	if privacy == "" {
		privacy = "public"
	}

	if privacy != "public" && privacy != "followers" && privacy != "private" {
		errorJSON(w, "invalid privacy value", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO posts (
			user_id,
			content,
			image_path,
			privacy
		)
		VALUES (?, ?, ?, ?)
	`,
		userID,
		content,
		imagePath,
		privacy,
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
