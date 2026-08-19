package server

import "net/http"

func listGroupMessagesHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	_, ok := requireGroupMember(w, r, groupID)

	if !ok {
		return
	}

	rows, err := db.Query(`
		SELECT
			group_messages.id,
			group_messages.group_id,
			group_messages.sender_id,

			users.first_name || ' ' ||
			users.last_name AS sender_name,

			group_messages.content,
			group_messages.created_at

		FROM group_messages

		JOIN users
		  ON users.id = group_messages.sender_id

		WHERE group_messages.group_id = ?

		ORDER BY
			group_messages.created_at ASC,
			group_messages.id ASC
	`, groupID)

	if err != nil {
		errorJSON(
			w,
			"could not load group messages",
			http.StatusInternalServerError,
		)
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
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			errorJSON(
				w,
				"could not read group message",
				http.StatusInternalServerError,
			)
			return
		}

		messages = append(
			messages,
			message,
		)
	}

	if err := rows.Err(); err != nil {
		errorJSON(
			w,
			"error while reading group messages",
			http.StatusInternalServerError,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		messages,
	)
}

func saveGroupMessage(groupID int, senderID int, content string) (GroupMessageResponse, error) {

	result, err := db.Exec(`
		INSERT INTO group_messages (
			group_id,
			sender_id,
			content
		)
		VALUES (?, ?, ?)
	`,
		groupID,
		senderID,
		content,
	)

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
