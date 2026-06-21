package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func groupExists(groupID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func requireGroupMember(w http.ResponseWriter, r *http.Request, groupID int) (int, bool) {
	currentUserID := r.Context().Value(userIDKey).(int)

	exists, err := groupExists(groupID)
	if err != nil {
		errorJSON(w, "could not check group", http.StatusInternalServerError)
		return 0, false
	}

	if !exists {
		errorJSON(w, "group not found", http.StatusNotFound)
		return 0, false
	}

	member, err := isGroupMember(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group membership", http.StatusInternalServerError)
		return 0, false
	}

	if !member {
		errorJSON(w, "only group members can access this", http.StatusForbidden)
		return 0, false
	}

	return currentUserID, true
}

func listGroupPostsHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	_, ok := requireGroupMember(w, r, groupID)
	if !ok {
		return
	}

	rows, err := db.Query(`
		SELECT
			group_posts.id,
			group_posts.group_id,
			group_posts.user_id,
			users.first_name || ' ' || users.last_name AS author_name,
			COALESCE(users.nickname, '') AS author_nickname,
			group_posts.content,
			COALESCE(group_posts.image_path, '') AS image_path,
			group_posts.created_at
		FROM group_posts
		JOIN users ON users.id = group_posts.user_id
		WHERE group_posts.group_id = ?
		ORDER BY group_posts.created_at DESC
	`, groupID)

	if err != nil {
		errorJSON(w, "could not load group posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := []GroupPostResponse{}

	for rows.Next() {
		var post GroupPostResponse

		err := rows.Scan(
			&post.ID,
			&post.GroupID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorNickname,
			&post.Content,
			&post.ImagePath,
			&post.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read group post data", http.StatusInternalServerError)
			return
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading group posts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func createGroupPostHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID, ok := requireGroupMember(w, r, groupID)
	if !ok {
		return
	}

	var content string
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

		imagePath, err = saveUploadedImage(r, "image", "uploads/group-posts")
		if err != nil {
			errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		var req CreateGroupPostRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorJSON(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(req.Content)
	}

	if content == "" {
		errorJSON(w, "group post content is required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO group_posts (
			group_id,
			user_id,
			content,
			image_path
		)
		VALUES (?, ?, ?, ?)
	`, groupID, currentUserID, content, imagePath)

	if err != nil {
		errorJSON(w, "could not create group post", http.StatusInternalServerError)
		return
	}

	groupPostID, err := result.LastInsertId()
	if err != nil {
		errorJSON(w, "group post created but could not read id", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":       "group post created successfully",
		"group_post_id": groupPostID,
	})
}

func canViewGroupPost(userID int, groupID int, groupPostID int) (bool, error) {
	member, err := isGroupMember(userID, groupID)
	if err != nil {
		return false, err
	}

	if !member {
		return false, nil
	}

	var count int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM group_posts
		WHERE id = ?
		  AND group_id = ?
	`, groupPostID, groupID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func listGroupPostCommentsHandler(w http.ResponseWriter, r *http.Request, groupID int, groupPostID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	canView, err := canViewGroupPost(currentUserID, groupID, groupPostID)
	if err != nil {
		errorJSON(w, "could not check group post access", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "group post not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query(`
		SELECT
			group_comments.id,
			group_comments.group_post_id,
			group_comments.user_id,
			users.first_name || ' ' || users.last_name AS author_name,
			COALESCE(users.nickname, '') AS author_nickname,
			group_comments.content,
			COALESCE(group_comments.image_path, '') AS image_path,
			group_comments.created_at
		FROM group_comments
		JOIN users ON users.id = group_comments.user_id
		WHERE group_comments.group_post_id = ?
		ORDER BY group_comments.created_at ASC
	`, groupPostID)

	if err != nil {
		errorJSON(w, "could not load group comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []GroupCommentResponse{}

	for rows.Next() {
		var comment GroupCommentResponse

		err := rows.Scan(
			&comment.ID,
			&comment.GroupPostID,
			&comment.UserID,
			&comment.AuthorName,
			&comment.AuthorNickname,
			&comment.Content,
			&comment.ImagePath,
			&comment.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read group comment data", http.StatusInternalServerError)
			return
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading group comments", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func createGroupPostCommentHandler(w http.ResponseWriter, r *http.Request, groupID int, groupPostID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	canView, err := canViewGroupPost(currentUserID, groupID, groupPostID)
	if err != nil {
		errorJSON(w, "could not check group post access", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "group post not found", http.StatusNotFound)
		return
	}

	var content string
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

		imagePath, err = saveUploadedImage(r, "image", "uploads/group-comments")
		if err != nil {
			errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		var req CreateGroupCommentRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorJSON(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(req.Content)
	}

	if content == "" {
		errorJSON(w, "group comment content is required", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT INTO group_comments (
			group_post_id,
			user_id,
			content,
			image_path
		)
		VALUES (?, ?, ?, ?)
	`, groupPostID, currentUserID, content, imagePath)

	if err != nil {
		errorJSON(w, "could not create group comment", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "group comment created successfully",
	})
}
