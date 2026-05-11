package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func postSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[1] != "comments" {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	postID, err := strconv.Atoi(parts[0])
	if err != nil || postID <= 0 {
		errorJSON(w, "invalid post id", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		listCommentsHandler(w, r, postID)
		return
	}

	if r.Method == http.MethodPost {
		createCommentHandler(w, r, postID)
		return
	}

	errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
}

func canViewPost(userID int, postID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM posts
		WHERE id = ?
		  AND (
		    privacy = 'public'
		    OR user_id = ?
		  )
	`, postID, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func createCommentHandler(w http.ResponseWriter, r *http.Request, postID int) {
	userID := r.Context().Value(userIDKey).(int)

	canView, err := canViewPost(userID, postID)
	if err != nil {
		errorJSON(w, "could not check post", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "post not found", http.StatusNotFound)
		return
	}

	var req CreateCommentRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	req.Content = strings.TrimSpace(req.Content)

	if req.Content == "" {
		errorJSON(w, "comment content is required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO comments (
			post_id,
			user_id,
			content
		)
		VALUES (?, ?, ?)
	`,
		postID,
		userID,
		req.Content,
	)

	if err != nil {
		errorJSON(w, "could not create comment", http.StatusInternalServerError)
		return
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		errorJSON(w, "comment created but could not read comment id", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "comment created successfully",
		"comment_id": commentID,
	})
}

func listCommentsHandler(w http.ResponseWriter, r *http.Request, postID int) {
	userID := r.Context().Value(userIDKey).(int)

	canView, err := canViewPost(userID, postID)
	if err != nil {
		errorJSON(w, "could not check post", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "post not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query(`
		SELECT
			comments.id,
			comments.post_id,
			comments.user_id,
			users.first_name || ' ' || users.last_name AS author_name,
			COALESCE(users.nickname, '') AS author_nickname,
			comments.content,
			COALESCE(comments.image_path, '') AS image_path,
			comments.created_at
		FROM comments
		JOIN users ON users.id = comments.user_id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)

	if err != nil {
		errorJSON(w, "could not load comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []CommentResponse{}

	for rows.Next() {
		var comment CommentResponse

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.AuthorName,
			&comment.AuthorNickname,
			&comment.Content,
			&comment.ImagePath,
			&comment.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read comment data", http.StatusInternalServerError)
			return
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading comments", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, comments)
}
