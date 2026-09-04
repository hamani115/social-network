package server

import (
	"net/http"
	"strconv"
)

func listGroupMessagesHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	if _, ok := requireGroupMember(w, r, groupID); !ok {
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
			group_messages.id,
			group_messages.group_id,
			group_messages.sender_id,
			users.first_name || ' ' || users.last_name AS sender_name,
			COALESCE(users.avatar_path, '') AS sender_avatar_path,
			group_messages.content,
			group_messages.created_at
		FROM group_messages
		JOIN users
			ON users.id = group_messages.sender_id
		WHERE
			group_messages.group_id = ?
			AND (
				? = 0
				OR group_messages.id < ?
			)
		ORDER BY
			group_messages.id DESC
		LIMIT ?
	`, groupID, beforeID, beforeID, limit+1)

	if err != nil {
		errorJSON(w, "could not load group messages", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	messages := []GroupMessageResponse{}

	for rows.Next() {
		var message GroupMessageResponse

		err := rows.Scan(
			&message.ID,
			&message.GroupID,
			&message.SenderID,
			&message.SenderName,
			&message.SenderAvatarPath,
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read group message", http.StatusInternalServerError)
			return
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading group messages", http.StatusInternalServerError)
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

func saveGroupMessage(groupID int, senderID int, content string) (GroupMessageResponse, error) {

	result, err := db.Exec(`
		INSERT INTO group_messages (
			group_id,
			sender_id,
			content
		)
		VALUES (?, ?, ?)
	`, groupID, senderID, content)

	if err != nil {
		return GroupMessageResponse{}, err
	}

	messageID, err := result.LastInsertId()

	if err != nil {
		return GroupMessageResponse{}, err
	}

	var message GroupMessageResponse

	err = db.QueryRow(`
		SELECT
			group_messages.id,
			group_messages.group_id,
			group_messages.sender_id,
			users.first_name || ' ' ||
			users.last_name AS sender_name,
			COALESCE(users.avatar_path, '') AS sender_avatar_path,
			group_messages.content,
			group_messages.created_at
		FROM group_messages
		JOIN users
		  ON users.id = group_messages.sender_id
		WHERE group_messages.id = ?
	`, messageID).Scan(
		&message.ID,
		&message.GroupID,
		&message.SenderID,
		&message.SenderName,
		&message.SenderAvatarPath,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		return GroupMessageResponse{}, err
	}

	return message, nil
}

func getGroupMemberIDs(groupID int) ([]int, error) {

	rows, err := db.Query(`
		SELECT user_id
		FROM group_members
		WHERE group_id = ?
	`, groupID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	memberIDs := []int{}

	for rows.Next() {
		var userID int

		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}

		memberIDs = append(
			memberIDs,
			userID,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return memberIDs, nil
}
