package server

import (
	"net/http"
	"strconv"
	"strings"
)

func canPrivateChat(userAID, userBID int) (bool, error) {
	if userAID == userBID {
		return false, nil
	}

	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE (follower_id = ? AND following_id = ?)
			OR (follower_id = ? AND following_id = ?)
	`, userAID, userBID, userBID, userAID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func chatUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			users.id,
			users.first_name,
			users.last_name,
			COALESCE(users.nickname, ''),
			COALESCE(users.avatar_path, '')
		FROM users
		WHERE users.id != ?
		  AND EXISTS (
			  SELECT 1
			  FROM followers
			  WHERE (followers.follower_id = ? AND followers.following_id = users.id)
				  OR (followers.follower_id = users.id AND followers.following_id = ?)
		  )
		ORDER BY users.first_name, users.last_name
	`, currentUserID, currentUserID, currentUserID)

	if err != nil {
		errorJSON(w, "could not load chat users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []ChatUserResponse{}

	for rows.Next() {
		var user ChatUserResponse

		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Nickname, &user.AvatarPath)

		if err != nil {
			errorJSON(w, "could not read chat user", http.StatusInternalServerError)
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading chat users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func chatSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/chat/")

	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 2 || parts[1] != "messages" {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	otherUserID, err := strconv.Atoi(parts[0])
	if err != nil || otherUserID <= 0 {
		errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listPrivateMessagesHandler(w, r, otherUserID)
}

func listPrivateMessagesHandler(w http.ResponseWriter, r *http.Request, otherUserID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	if currentUserID == otherUserID {
		errorJSON(w, "cannot open a private chat with yourself", http.StatusBadRequest)
		return
	}

	var userCount int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE id = ?
	`, otherUserID).Scan(&userCount)

	if err != nil {
		errorJSON(w, "could not check user", http.StatusInternalServerError)
		return
	}

	if userCount == 0 {
		errorJSON(w, "user not found", http.StatusNotFound)
		return
	}

	allowed, err := canPrivateChat(currentUserID, otherUserID)

	if err != nil {
		errorJSON(w, "could not check chat permission", http.StatusInternalServerError)
		return
	}

	if !allowed {
		errorJSON(w, "you cannot chat with this user", http.StatusForbidden)
		return
	}

	limit := 30

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {

		parsedLimit, err := strconv.Atoi(rawLimit)

		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			errorJSON(w, "invalid message limit", http.StatusBadRequest)
			return
		}

		limit = parsedLimit
	}

	beforeID := 0

	if rawBefore := r.URL.Query().Get("before_id"); rawBefore != "" {

		parsedBefore, err := strconv.Atoi(rawBefore)

		if err != nil || parsedBefore <= 0 {
			errorJSON(w, "invalid message cursor", http.StatusBadRequest)
			return
		}

		beforeID = parsedBefore
	}

	rows, err := db.Query(`
		SELECT
			private_messages.id,
			private_messages.sender_id,
			private_messages.receiver_id,
			users.first_name || ' ' || users.last_name AS sender_name,
			COALESCE(users.avatar_path, '') AS sender_avatar_path,
			private_messages.content,
			private_messages.created_at
		FROM private_messages
		JOIN users
			ON users.id = private_messages.sender_id
		WHERE
			(
				(private_messages.sender_id = ? AND private_messages.receiver_id = ?)
				OR (private_messages.sender_id = ? AND private_messages.receiver_id = ?)
			)
			AND (? = 0 OR private_messages.id < ?)
		ORDER BY private_messages.id DESC
		LIMIT ?
	`, currentUserID, otherUserID, otherUserID, currentUserID, beforeID, beforeID, limit+1)

	if err != nil {
		errorJSON(w, "could not load private messages", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	messages := []PrivateMessageResponse{}

	for rows.Next() {
		var message PrivateMessageResponse

		err := rows.Scan(&message.ID, &message.SenderID, &message.ReceiverID, &message.SenderName, &message.SenderAvatarPath, &message.Content, &message.CreatedAt)

		if err != nil {
			errorJSON(w, "could not read private message", http.StatusInternalServerError)
			return
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading private messages", http.StatusInternalServerError)
		return
	}

	hasMore := len(messages) > limit

	if hasMore {
		messages = messages[:limit]
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {

		messages[i], messages[j] = messages[j], messages[i]
	}

	nextBeforeID := 0

	if len(messages) > 0 {
		nextBeforeID = messages[0].ID
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"messages":       messages,
		"has_more":       hasMore,
		"next_before_id": nextBeforeID,
	})
}

func savePrivateMessage(senderID int, receiverID int, content string) (PrivateMessageResponse, error) {

	result, err := db.Exec(`
		INSERT INTO private_messages (
			sender_id,
			receiver_id,
			content
		)
		VALUES (?, ?, ?)
	`, senderID, receiverID, content)

	if err != nil {
		return PrivateMessageResponse{}, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return PrivateMessageResponse{}, err
	}

	var message PrivateMessageResponse

	err = db.QueryRow(`
		SELECT
			private_messages.id,
			private_messages.sender_id,
			private_messages.receiver_id,
			users.first_name || ' ' || users.last_name AS sender_name,
			COALESCE(users.avatar_path, '') AS sender_avatar_path,
			private_messages.content,
			private_messages.created_at
		FROM private_messages
		JOIN users
		  ON users.id = private_messages.sender_id
		WHERE private_messages.id = ?
	`, messageID).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.SenderName,
		&message.SenderAvatarPath,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		return PrivateMessageResponse{}, err
	}

	return message, nil
}

func chatUserExists(userID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE id = ?
	`, userID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
