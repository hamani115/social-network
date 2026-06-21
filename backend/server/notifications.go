package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listNotificationsHandler(w, r)
}

func notificationSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	path = strings.Trim(path, "/")

	if path == "read-all" {
		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		markAllNotificationsReadHandler(w, r)
		return
	}

	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[1] != "read" {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	notificationID, err := strconv.Atoi(parts[0])
	if err != nil || notificationID <= 0 {
		errorJSON(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	markNotificationReadHandler(w, r, notificationID)
}

func listNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			COALESCE(requester_id, 0),
			type,
			message,
			link_path,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, currentUserID)

	if err != nil {
		errorJSON(w, "could not load notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notifications := []NotificationResponse{}

	for rows.Next() {
		var notification NotificationResponse
		var isReadInt int

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.RequesterID,
			&notification.Type,
			&notification.Message,
			&notification.LinkPath,
			&isReadInt,
			&notification.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read notification data", http.StatusInternalServerError)
			return
		}

		notification.IsRead = isReadInt == 1

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading notifications", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, notifications)
}

func markNotificationReadHandler(w http.ResponseWriter, r *http.Request, notificationID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	result, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE id = ?
		  AND user_id = ?
	`, notificationID, currentUserID)

	if err != nil {
		errorJSON(w, "could not update notification", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errorJSON(w, "could not check notification update", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		errorJSON(w, "notification not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "notification marked as read",
	})
}

func markAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	_, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ?
	`, currentUserID)

	if err != nil {
		errorJSON(w, "could not update notifications", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "all notifications marked as read",
	})
}

func createNotification(userID int, requesterID int, notificationType string, message string, linkPath string) error {
	_, err := db.Exec(`
		INSERT INTO notifications (
			user_id,
			requester_id,
			type,
			message,
			link_path
		)
		VALUES (?, ?, ?, ?, ?)
	`, userID, requesterID, notificationType, message, linkPath)

	return err
}

func getUserDisplayName(userID int) string {
	var firstName string
	var lastName string
	var nickname string

	err := db.QueryRow(`
		SELECT
			first_name,
			last_name,
			COALESCE(nickname, '')
		FROM users
		WHERE id = ?
	`, userID).Scan(&firstName, &lastName, &nickname)

	if err != nil {
		return "Someone"
	}

	fullName := strings.TrimSpace(firstName + " " + lastName)

	if nickname != "" {
		return fmt.Sprintf("%s (%s)", fullName, nickname)
	}

	return fullName
}

func notifyFollowRequestReceived(targetUserID int, requesterID int) error {
	requesterName := getUserDisplayName(requesterID)

	return createNotification(
		targetUserID,
		requesterID,
		"follow_request",
		requesterName+" wants to follow you.",
		fmt.Sprintf("/profiles/%d", requesterID),
	)
}

func notifyFollowRequestAccepted(requesterID int, accepterID int) error {
	accepterName := getUserDisplayName(accepterID)

	return createNotification(
		requesterID,
		accepterID,
		"follow_accepted",
		accepterName+" accepted your follow request.",
		fmt.Sprintf("/profiles/%d", accepterID),
	)
}

func getGroupTitle(groupID int) string {
	var title string

	err := db.QueryRow(`
		SELECT title
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&title)

	if err != nil {
		return "a group"
	}

	return title
}

func getGroupOwnerID(groupID int) (int, error) {
	var ownerID int

	err := db.QueryRow(`
		SELECT user_id
		FROM group_members
		WHERE group_id = ?
		  AND role = 'owner'
	`, groupID).Scan(&ownerID)

	return ownerID, err
}

func notifyGroupJoinRequestReceived(groupID int, requesterID int) error {
	ownerID, err := getGroupOwnerID(groupID)
	if err != nil {
		return err
	}

	requesterName := getUserDisplayName(requesterID)
	groupTitle := getGroupTitle(groupID)

	return createNotification(
		ownerID,
		requesterID,
		"group_join_request",
		requesterName+" requested to join your group: "+groupTitle+".",
		fmt.Sprintf("/groups/%d", groupID),
	)
}

func notifyGroupJoinRequestAccepted(requesterID int, ownerID int, groupID int) error {
	ownerName := getUserDisplayName(ownerID)
	groupTitle := getGroupTitle(groupID)

	return createNotification(
		requesterID,
		ownerID,
		"group_join_accepted",
		ownerName+" accepted your request to join: "+groupTitle+".",
		fmt.Sprintf("/groups/%d", groupID),
	)
}

func notifyGroupJoinRequestDeclined(requesterID int, ownerID int, groupID int) error {
	ownerName := getUserDisplayName(ownerID)
	groupTitle := getGroupTitle(groupID)

	return createNotification(
		requesterID,
		ownerID,
		"group_join_declined",
		ownerName+" declined your request to join: "+groupTitle+".",
		fmt.Sprintf("/groups/%d", groupID),
	)
}

func notifyGroupInvitationReceived(inviteeID int, inviterID int, groupID int) error {
	inviterName := getUserDisplayName(inviterID)
	groupTitle := getGroupTitle(groupID)

	return createNotification(
		inviteeID,
		inviterID,
		"group_invitation",
		inviterName+" invited you to join: "+groupTitle+".",
		fmt.Sprintf("/groups/%d", groupID),
	)
}

func notifyGroupInvitationAccepted(inviterID int, inviteeID int, groupID int) error {
	inviteeName := getUserDisplayName(inviteeID)
	groupTitle := getGroupTitle(groupID)

	return createNotification(
		inviterID,
		inviteeID,
		"group_invitation_accepted",
		inviteeName+" accepted your invitation to join: "+groupTitle+".",
		fmt.Sprintf("/groups/%d", groupID),
	)
}
