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
		WHERE posts.id = ?
		  AND (
			posts.privacy = 'public'
			OR posts.user_id = ?
			OR (
				posts.privacy = 'followers'
				AND EXISTS (
					SELECT 1
					FROM followers
					WHERE followers.follower_id = ?
					  AND followers.following_id = posts.user_id
				)
			)
			OR (
				posts.privacy = 'private'
				AND EXISTS (
					SELECT 1
					FROM post_allowed_users
					WHERE post_allowed_users.post_id = posts.id
					  AND post_allowed_users.user_id = ?
				)
			)
		  )
	`, postID, userID, userID, userID).Scan(&count)

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

	var content string
	var imagePath string

	keepUploadedImage := false

	defer func() {
		if !keepUploadedImage {
			removeUploadedImage(imagePath)
		}
	}()

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			errorJSON(w, "could not read form data", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(r.FormValue("content"))

		imagePath, err = saveUploadedImage(r, "image", "uploads/comments")
		if err != nil {
			errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		var req CreateCommentRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorJSON(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(req.Content)
	}

	if content == "" {
		errorJSON(w, "comment content is required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO comments (
			post_id,
			user_id,
			content,
			image_path
		)
		VALUES (?, ?, ?, ?)
	`,
		postID,
		userID,
		content,
		imagePath,
	)

	if err != nil {
		errorJSON(w, "could not create comment", http.StatusInternalServerError)
		return
	}

	keepUploadedImage = true

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

	// Default amount of comments to load.
	limit := 5

	if rawLimit :=
		r.URL.Query().Get("limit"); rawLimit != "" {

		parsedLimit, err :=
			strconv.Atoi(rawLimit)

		if err != nil ||
			parsedLimit <= 0 ||
			parsedLimit > 20 {

			errorJSON(
				w,
				"invalid comment limit",
				http.StatusBadRequest,
			)
			return
		}

		limit = parsedLimit
	}

	beforeID := 0

	if rawBefore :=
		r.URL.Query().Get("before_id"); rawBefore != "" {

		parsedBefore, err :=
			strconv.Atoi(rawBefore)

		if err != nil ||
			parsedBefore <= 0 {

			errorJSON(
				w,
				"invalid comment cursor",
				http.StatusBadRequest,
			)
			return
		}

		beforeID = parsedBefore
	}

	rows, err := db.Query(`
		SELECT
			comments.id,
			comments.post_id,
			comments.user_id,
			users.first_name || ' ' ||
				users.last_name
				AS author_name,
			COALESCE(
				users.nickname,
				''
			) AS author_nickname,
			COALESCE(
				users.avatar_path,
				''
			) AS author_avatar_path,
			comments.content,
			COALESCE(
				comments.image_path,
				''
			) AS image_path,
			comments.created_at
		FROM comments
		JOIN users
			ON users.id =
				comments.user_id
		WHERE comments.post_id = ?
		  AND (
			? = 0
			OR comments.id < ?
		  )
		ORDER BY comments.id DESC
		LIMIT ?
	`,
		postID,
		beforeID,
		beforeID,
		limit+1,
	)

	if err != nil {
		errorJSON(
			w,
			"could not load comments",
			http.StatusInternalServerError,
		)
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
			&comment.AuthorAvatarPath,
			&comment.Content,
			&comment.ImagePath,
			&comment.CreatedAt,
		)

		if err != nil {
			errorJSON(
				w,
				"could not read comment data",
				http.StatusInternalServerError,
			)
			return
		}

		comments =
			append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		errorJSON(
			w,
			"error while reading comments",
			http.StatusInternalServerError,
		)
		return
	}

	// We requested one extra record so we
	// can determine whether earlier
	// comments still exist.
	hasMore :=
		len(comments) > limit

	if hasMore {
		comments =
			comments[:limit]
	}

	// SQL returned newest -> oldest.
	// The UI should display comments
	// oldest -> newest.
	for i, j :=
		0, len(comments)-1; i < j; i, j = i+1, j-1 {

		comments[i],
			comments[j] =
			comments[j],
			comments[i]
	}

	nextBeforeID := 0

	if len(comments) > 0 {
		nextBeforeID =
			comments[0].ID
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"comments":       comments,
			"has_more":       hasMore,
			"next_before_id": nextBeforeID,
		},
	)
}
