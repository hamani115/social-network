package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func myProfileHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	if r.Method == http.MethodGet {
		getProfileHandler(w, r, currentUserID)
		return
	}

	if r.Method == http.MethodPut {
		updateMyProfileHandler(w, r, currentUserID)
		return
	}

	errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
}

func profileSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	path = strings.Trim(path, "/")

	if path == "" {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	parts := strings.Split(path, "/")

	profileUserID, err := strconv.Atoi(parts[0])
	if err != nil || profileUserID <= 0 {
		errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		getProfileHandler(w, r, profileUserID)
		return
	}

	if len(parts) == 2 && parts[1] == "posts" {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listProfilePostsHandler(w, r, profileUserID)
		return
	}

	errorJSON(w, "route not found", http.StatusNotFound)
}

func getProfileHandler(w http.ResponseWriter, r *http.Request, profileUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var profile ProfileResponse
	var isPublicInt int

	err := db.QueryRow(`
		SELECT
			id,
			email,
			first_name,
			last_name,
			date_of_birth,
			COALESCE(avatar_path, ''),
			COALESCE(nickname, ''),
			COALESCE(about_me, ''),
			is_public
		FROM users
		WHERE id = ?
	`, profileUserID).Scan(
		&profile.ID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.DateOfBirth,
		&profile.AvatarPath,
		&profile.Nickname,
		&profile.AboutMe,
		&isPublicInt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "profile not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load profile", http.StatusInternalServerError)
		return
	}

	profile.IsPublic = isPublicInt == 1
	profile.IsOwner = currentUserID == profileUserID

	followStatus, err := getFollowStatus(currentUserID, profileUserID)
	if err != nil {
		errorJSON(w, "could not check follow status", http.StatusInternalServerError)
		return
	}

	profile.FollowStatus = followStatus

	profile.CanViewProfile = profile.IsOwner || profile.IsPublic || profile.FollowStatus == "following"

	profile.FollowersCount, err = countFollowers(profileUserID)
	if err != nil {
		errorJSON(w, "could not count followers", http.StatusInternalServerError)
		return
	}

	profile.FollowingCount, err = countFollowing(profileUserID)
	if err != nil {
		errorJSON(w, "could not count following", http.StatusInternalServerError)
		return
	}

	if !profile.IsOwner {
		profile.Email = ""
		profile.DateOfBirth = ""
	}

	if !profile.CanViewProfile {
		profile.AboutMe = ""
	}

	writeJSON(w, http.StatusOK, profile)
}

func getFollowStatus(currentUserID int, targetUserID int) (string, error) {
	if currentUserID == targetUserID {
		return "self", nil
	}

	following, err := isFollowing(currentUserID, targetUserID)
	if err != nil {
		return "", err
	}

	if following {
		return "following", nil
	}

	var count int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM follow_requests
		WHERE requester_id = ?
		  AND target_id = ?
		  AND status = 'pending'
	`, currentUserID, targetUserID).Scan(&count)

	if err != nil {
		return "", err
	}

	if count > 0 {
		return "pending", nil
	}

	return "none", nil
}

func countFollowers(userID int) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE following_id = ?
	`, userID).Scan(&count)

	return count, err
}

func countFollowing(userID int) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ?
	`, userID).Scan(&count)

	return count, err
}

func listProfilePostsHandler(w http.ResponseWriter, r *http.Request, profileUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

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
		WHERE posts.user_id = ?
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
		ORDER BY posts.created_at DESC
	`, profileUserID, currentUserID, currentUserID, currentUserID)

	if err != nil {
		errorJSON(w, "could not load profile posts", http.StatusInternalServerError)
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
			errorJSON(w, "could not read profile post data", http.StatusInternalServerError)
			return
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading profile posts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func updateMyProfileHandler(w http.ResponseWriter, r *http.Request, currentUserID int) {
	var req UpdateProfileRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	nickname := strings.TrimSpace(req.Nickname)
	aboutMe := strings.TrimSpace(req.AboutMe)

	var isPublicInt int

	if req.IsPublic == nil {
		err = db.QueryRow(`
			SELECT is_public
			FROM users
			WHERE id = ?
		`, currentUserID).Scan(&isPublicInt)

		if err != nil {
			errorJSON(w, "could not read current profile status", http.StatusInternalServerError)
			return
		}
	} else if *req.IsPublic {
		isPublicInt = 1
	} else {
		isPublicInt = 0
	}

	_, err = db.Exec(`
		UPDATE users
		SET nickname = ?,
		    about_me = ?,
		    is_public = ?
		WHERE id = ?
	`, nickname, aboutMe, isPublicInt, currentUserID)

	if err != nil {
		errorJSON(w, "could not update profile", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "profile updated successfully",
	})
}

func canViewUserProfile(currentUserID int, targetUserID int) (bool, error) {
	if currentUserID == targetUserID {
		return true, nil
	}

	var isPublicInt int

	err := db.QueryRow(`
		SELECT is_public
		FROM users
		WHERE id = ?
	`, targetUserID).Scan(&isPublicInt)

	if err != nil {
		return false, err
	}

	if isPublicInt == 1 {
		return true, nil
	}

	return isFollowing(currentUserID, targetUserID)
}
