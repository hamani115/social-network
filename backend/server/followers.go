package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

func userSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	targetUserID, err := strconv.Atoi(parts[0])
	if err != nil || targetUserID <= 0 {
		errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	action := parts[1]

	if action == "follow" {
		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		followUserHandler(w, r, targetUserID)
		return
	}

	if action == "unfollow" {
		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		unfollowUserHandler(w, r, targetUserID)
		return
	}

	if action == "followers" {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listFollowersHandler(w, r, targetUserID)
		return
	}

	if action == "following" {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listFollowingHandler(w, r, targetUserID)
		return
	}

	errorJSON(w, "route not found", http.StatusNotFound)
}

func followUserHandler(w http.ResponseWriter, r *http.Request, targetUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	if currentUserID == targetUserID {
		errorJSON(w, "you cannot follow yourself", http.StatusBadRequest)
		return
	}

	var targetIsPublic int

	err := db.QueryRow(`
		SELECT is_public
		FROM users
		WHERE id = ?
	`, targetUserID).Scan(&targetIsPublic)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "user not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not check user", http.StatusInternalServerError)
		return
	}

	alreadyFollowing, err := isFollowing(currentUserID, targetUserID)
	if err != nil {
		errorJSON(w, "could not check follow status", http.StatusInternalServerError)
		return
	}

	if alreadyFollowing {
		writeJSON(w, http.StatusOK, FollowUserResponse{
			Message: "you are already following this user",
			Status:  "following",
		})
		return
	}

	if targetIsPublic == 1 {
		_, err = db.Exec(`
			INSERT INTO followers (follower_id, following_id)
			VALUES (?, ?)
		`, currentUserID, targetUserID)

		if err != nil {
			errorJSON(w, "could not follow user", http.StatusInternalServerError)
			return
		}

		_, _ = db.Exec(`
			DELETE FROM follow_requests
			WHERE requester_id = ?
			  AND target_id = ?
		`, currentUserID, targetUserID)

		writeJSON(w, http.StatusOK, FollowUserResponse{
			Message: "you are now following this user",
			Status:  "following",
		})
		return
	}

	var existingStatus string

	err = db.QueryRow(`
		SELECT status
		FROM follow_requests
		WHERE requester_id = ?
		  AND target_id = ?
	`, currentUserID, targetUserID).Scan(&existingStatus)

	if err == nil {
		if existingStatus == "pending" {
			writeJSON(w, http.StatusOK, FollowUserResponse{
				Message: "follow request is already pending",
				Status:  "pending",
			})
			return
		}

		_, err = db.Exec(`
			UPDATE follow_requests
			SET status = 'pending',
			    updated_at = CURRENT_TIMESTAMP
			WHERE requester_id = ?
			  AND target_id = ?
		`, currentUserID, targetUserID)

		if err != nil {
			errorJSON(w, "could not update follow request", http.StatusInternalServerError)
			return
		}

		err = notifyFollowRequestReceived(targetUserID, currentUserID)
		if err != nil {
			errorJSON(w, "follow request updated but notification failed", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, FollowUserResponse{
			Message: "follow request sent",
			Status:  "pending",
		})
		return
	}

	if err != sql.ErrNoRows {
		errorJSON(w, "could not check follow request", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		INSERT INTO follow_requests (requester_id, target_id, status)
		VALUES (?, ?, 'pending')
	`, currentUserID, targetUserID)

	if err != nil {
		errorJSON(w, "could not create follow request", http.StatusInternalServerError)
		return
	}

	err = notifyFollowRequestReceived(targetUserID, currentUserID)
	if err != nil {
		errorJSON(w, "follow request created but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, FollowUserResponse{
		Message: "follow request sent",
		Status:  "pending",
	})
}

func isFollowing(followerID int, followingID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ?
		  AND following_id = ?
	`, followerID, followingID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func unfollowUserHandler(w http.ResponseWriter, r *http.Request, targetUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	if currentUserID == targetUserID {
		errorJSON(w, "you cannot unfollow yourself", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
		DELETE FROM followers
		WHERE follower_id = ?
		  AND following_id = ?
	`, currentUserID, targetUserID)

	if err != nil {
		errorJSON(w, "could not unfollow user", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		DELETE FROM follow_requests
		WHERE requester_id = ?
		  AND target_id = ?
		  AND status = 'pending'
	`, currentUserID, targetUserID)

	if err != nil {
		errorJSON(w, "could not cancel follow request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "follow removed or request cancelled",
	})
}

func listFollowersHandler(w http.ResponseWriter, r *http.Request, targetUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	canView, err := canViewUserProfile(currentUserID, targetUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "user not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not check profile visibility", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "this profile is private", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT
			users.id,
			users.email,
			users.first_name,
			users.last_name,
			COALESCE(users.nickname, '')
		FROM followers
		JOIN users ON users.id = followers.follower_id
		WHERE followers.following_id = ?
		ORDER BY users.first_name, users.last_name
	`, targetUserID)

	if err != nil {
		errorJSON(w, "could not load followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []UserListItem{}

	for rows.Next() {
		var user UserListItem

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Nickname,
		)

		if err != nil {
			errorJSON(w, "could not read follower data", http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading followers", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func listFollowingHandler(w http.ResponseWriter, r *http.Request, targetUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	canView, err := canViewUserProfile(currentUserID, targetUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "user not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not check profile visibility", http.StatusInternalServerError)
		return
	}

	if !canView {
		errorJSON(w, "this profile is private", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT
			users.id,
			users.email,
			users.first_name,
			users.last_name,
			COALESCE(users.nickname, '')
		FROM followers
		JOIN users ON users.id = followers.following_id
		WHERE followers.follower_id = ?
		ORDER BY users.first_name, users.last_name
	`, targetUserID)

	if err != nil {
		errorJSON(w, "could not load following users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []UserListItem{}

	for rows.Next() {
		var user UserListItem

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Nickname,
		)

		if err != nil {
			errorJSON(w, "could not read following data", http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading following users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func followRequestsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/follow-requests")
	path = strings.Trim(path, "/")

	if path == "" {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listFollowRequestsHandler(w, r)
		return
	}

	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	requestID, err := strconv.Atoi(parts[0])
	if err != nil || requestID <= 0 {
		errorJSON(w, "invalid follow request id", http.StatusBadRequest)
		return
	}

	action := parts[1]

	if r.Method != http.MethodPost {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if action == "accept" {
		acceptFollowRequestHandler(w, r, requestID)
		return
	}

	if action == "decline" {
		declineFollowRequestHandler(w, r, requestID)
		return
	}

	errorJSON(w, "route not found", http.StatusNotFound)
}

func listFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			follow_requests.id,
			follow_requests.requester_id,
			users.first_name || ' ' || users.last_name AS requester_name,
			COALESCE(users.nickname, '') AS requester_nickname,
			follow_requests.target_id,
			follow_requests.status,
			follow_requests.created_at
		FROM follow_requests
		JOIN users ON users.id = follow_requests.requester_id
		WHERE follow_requests.target_id = ?
		  AND follow_requests.status = 'pending'
		ORDER BY follow_requests.created_at DESC
	`, currentUserID)

	if err != nil {
		errorJSON(w, "could not load follow requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	requests := []FollowRequestResponse{}

	for rows.Next() {
		var request FollowRequestResponse

		err := rows.Scan(
			&request.ID,
			&request.RequesterID,
			&request.RequesterName,
			&request.RequesterNick,
			&request.TargetID,
			&request.Status,
			&request.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read follow request data", http.StatusInternalServerError)
			return
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading follow requests", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func acceptFollowRequestHandler(w http.ResponseWriter, r *http.Request, requestID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var requesterID int
	var targetID int
	var status string

	err := db.QueryRow(`
		SELECT requester_id, target_id, status
		FROM follow_requests
		WHERE id = ?
	`, requestID).Scan(&requesterID, &targetID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "follow request not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load follow request", http.StatusInternalServerError)
		return
	}

	if targetID != currentUserID {
		errorJSON(w, "you cannot accept this request", http.StatusForbidden)
		return
	}

	if status != "pending" {
		errorJSON(w, "follow request is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT INTO followers (follower_id, following_id)
		VALUES (?, ?)
	`, requesterID, targetID)

	if err != nil {
		errorJSON(w, "could not create follower relationship", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		UPDATE follow_requests
		SET status = 'accepted',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, requestID)

	if err != nil {
		errorJSON(w, "could not update follow request", http.StatusInternalServerError)
		return
	}

	err = notifyFollowRequestAccepted(requesterID, currentUserID)
	if err != nil {
		errorJSON(w, "follow request accepted but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "follow request accepted",
	})
}

func declineFollowRequestHandler(w http.ResponseWriter, r *http.Request, requestID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var targetID int
	var status string

	err := db.QueryRow(`
		SELECT target_id, status
		FROM follow_requests
		WHERE id = ?
	`, requestID).Scan(&targetID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "follow request not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load follow request", http.StatusInternalServerError)
		return
	}

	if targetID != currentUserID {
		errorJSON(w, "you cannot decline this request", http.StatusForbidden)
		return
	}

	if status != "pending" {
		errorJSON(w, "follow request is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		UPDATE follow_requests
		SET status = 'declined',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, requestID)

	if err != nil {
		errorJSON(w, "could not decline follow request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "follow request declined",
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listUsersHandler(w, r)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			users.id,
			users.email,
			users.first_name,
			users.last_name,
			COALESCE(users.nickname, ''),
			users.is_public,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM followers
					WHERE followers.follower_id = ?
					  AND followers.following_id = users.id
				) THEN 'following'

				WHEN EXISTS (
					SELECT 1
					FROM follow_requests
					WHERE follow_requests.requester_id = ?
					  AND follow_requests.target_id = users.id
					  AND follow_requests.status = 'pending'
				) THEN 'pending'

				ELSE 'none'
			END AS follow_status
		FROM users
		WHERE users.id != ?
		ORDER BY users.first_name, users.last_name
	`, currentUserID, currentUserID, currentUserID)

	if err != nil {
		errorJSON(w, "could not load users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []UserWithFollowStatus{}

	for rows.Next() {
		var user UserWithFollowStatus
		var isPublicInt int

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Nickname,
			&isPublicInt,
			&user.FollowStatus,
		)

		if err != nil {
			errorJSON(w, "could not read user data", http.StatusInternalServerError)
			return
		}

		user.IsPublic = isPublicInt == 1

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}
