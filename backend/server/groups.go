package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listGroupsHandler(w, r)
		return
	}

	if r.Method == http.MethodPost {
		createGroupHandler(w, r)
		return
	}

	errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
}

func groupSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	path = strings.Trim(path, "/")

	if path == "" {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	parts := strings.Split(path, "/")

	groupID, err := strconv.Atoi(parts[0])
	if err != nil || groupID <= 0 {
		errorJSON(w, "invalid group id", http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		getGroupHandler(w, r, groupID)
		return
	}

	if len(parts) == 2 && parts[1] == "join-request" {
		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		requestJoinGroupHandler(w, r, groupID)
		return
	}

	if len(parts) == 2 && parts[1] == "join-requests" {
		if r.Method != http.MethodGet {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listGroupJoinRequestsHandler(w, r, groupID)
		return
	}

	if len(parts) == 4 && parts[1] == "join-requests" {
		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		requestID, err := strconv.Atoi(parts[2])
		if err != nil || requestID <= 0 {
			errorJSON(w, "invalid join request id", http.StatusBadRequest)
			return
		}

		action := parts[3]

		if action == "accept" {
			acceptGroupJoinRequestHandler(w, r, groupID, requestID)
			return
		}

		if action == "decline" {
			declineGroupJoinRequestHandler(w, r, groupID, requestID)
			return
		}
	}

	if len(parts) == 2 && parts[1] == "invitations" {
		if r.Method == http.MethodGet {
			listGroupInvitationsHandler(w, r, groupID)
			return
		}

		if r.Method == http.MethodPost {
			createGroupInvitationHandler(w, r, groupID)
			return
		}

		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 2 && parts[1] == "posts" {
		if r.Method == http.MethodGet {
			listGroupPostsHandler(w, r, groupID)
			return
		}

		if r.Method == http.MethodPost {
			createGroupPostHandler(w, r, groupID)
			return
		}

		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 4 && parts[1] == "posts" && parts[3] == "comments" {
		groupPostID, err := strconv.Atoi(parts[2])
		if err != nil || groupPostID <= 0 {
			errorJSON(w, "invalid group post id", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			listGroupPostCommentsHandler(w, r, groupID, groupPostID)
			return
		}

		if r.Method == http.MethodPost {
			createGroupPostCommentHandler(w, r, groupID, groupPostID)
			return
		}

		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 2 && parts[1] == "events" {
		if r.Method == http.MethodGet {
			listGroupEventsHandler(w, r, groupID)
			return
		}

		if r.Method == http.MethodPost {
			createGroupEventHandler(w, r, groupID)
			return
		}

		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 4 && parts[1] == "events" {
		eventID, err := strconv.Atoi(parts[2])
		if err != nil || eventID <= 0 {
			errorJSON(w, "invalid event id", http.StatusBadRequest)
			return
		}

		if r.Method != http.MethodPost {
			errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		action := parts[3]

		if action == "going" {
			respondGroupEventHandler(w, r, groupID, eventID, "going")
			return
		}

		if action == "not-going" {
			respondGroupEventHandler(w, r, groupID, eventID, "not_going")
			return
		}
	}

	errorJSON(w, "route not found", http.StatusNotFound)
}

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var req CreateGroupRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)

	if title == "" {
		errorJSON(w, "group title is required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO groups (creator_id, title, description)
		VALUES (?, ?, ?)
	`, currentUserID, title, description)

	if err != nil {
		errorJSON(w, "could not create group", http.StatusInternalServerError)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		errorJSON(w, "group created but could not read group id", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'owner')
	`, groupID, currentUserID)

	if err != nil {
		errorJSON(w, "group created but could not add owner as member", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "group created successfully",
		"group_id": groupID,
	})
}

func getGroupMembershipStatus(userID int, groupID int) (string, error) {
	var role string

	err := db.QueryRow(`
		SELECT role
		FROM group_members
		WHERE user_id = ?
		  AND group_id = ?
	`, userID, groupID).Scan(&role)

	if err == nil {
		return role, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	var pendingRequestCount int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM group_join_requests
		WHERE requester_id = ?
		  AND group_id = ?
		  AND status = 'pending'
	`, userID, groupID).Scan(&pendingRequestCount)

	if err != nil {
		return "", err
	}

	if pendingRequestCount > 0 {
		return "pending", nil
	}

	var pendingInvitationCount int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM group_invitations
		WHERE invitee_id = ?
		  AND group_id = ?
		  AND status = 'pending'
	`, userID, groupID).Scan(&pendingInvitationCount)

	if err != nil {
		return "", err
	}

	if pendingInvitationCount > 0 {
		return "invited", nil
	}

	return "none", nil
}

func countGroupMembers(groupID int) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ?
	`, groupID).Scan(&count)

	return count, err
}

func isGroupMember(userID int, groupID int) (bool, error) {
	status, err := getGroupMembershipStatus(userID, groupID)
	if err != nil {
		return false, err
	}

	return status == "owner" || status == "member", nil
}

func isGroupOwner(userID int, groupID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE user_id = ?
		  AND group_id = ?
		  AND role = 'owner'
	`, userID, groupID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func listGroupsHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			groups.id,
			groups.creator_id,
			users.first_name || ' ' || users.last_name AS creator_name,
			groups.title,
			groups.description,
			groups.created_at
		FROM groups
		JOIN users ON users.id = groups.creator_id
		ORDER BY groups.created_at DESC
	`)

	if err != nil {
		errorJSON(w, "could not load groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []GroupResponse{}

	for rows.Next() {
		var group GroupResponse

		err := rows.Scan(
			&group.ID,
			&group.CreatorID,
			&group.CreatorName,
			&group.Title,
			&group.Description,
			&group.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read group data", http.StatusInternalServerError)
			return
		}

		group.MemberCount, err = countGroupMembers(group.ID)
		if err != nil {
			errorJSON(w, "could not count group members", http.StatusInternalServerError)
			return
		}

		group.MembershipStatus, err = getGroupMembershipStatus(currentUserID, group.ID)
		if err != nil {
			errorJSON(w, "could not check group membership", http.StatusInternalServerError)
			return
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading groups", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func getGroupHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var group GroupResponse

	err := db.QueryRow(`
		SELECT
			groups.id,
			groups.creator_id,
			users.first_name || ' ' || users.last_name AS creator_name,
			groups.title,
			groups.description,
			groups.created_at
		FROM groups
		JOIN users ON users.id = groups.creator_id
		WHERE groups.id = ?
	`, groupID).Scan(
		&group.ID,
		&group.CreatorID,
		&group.CreatorName,
		&group.Title,
		&group.Description,
		&group.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "group not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load group", http.StatusInternalServerError)
		return
	}

	group.MemberCount, err = countGroupMembers(group.ID)
	if err != nil {
		errorJSON(w, "could not count group members", http.StatusInternalServerError)
		return
	}

	group.MembershipStatus, err = getGroupMembershipStatus(currentUserID, group.ID)
	if err != nil {
		errorJSON(w, "could not check group membership", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func requestJoinGroupHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var groupExists int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&groupExists)

	if err != nil {
		errorJSON(w, "could not check group", http.StatusInternalServerError)
		return
	}

	if groupExists == 0 {
		errorJSON(w, "group not found", http.StatusNotFound)
		return
	}

	member, err := isGroupMember(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group membership", http.StatusInternalServerError)
		return
	}

	if member {
		errorJSON(w, "you are already a member of this group", http.StatusBadRequest)
		return
	}

	var pendingCount int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM group_join_requests
		WHERE group_id = ?
		  AND requester_id = ?
		  AND status = 'pending'
	`, groupID, currentUserID).Scan(&pendingCount)

	if err != nil {
		errorJSON(w, "could not check join request", http.StatusInternalServerError)
		return
	}

	if pendingCount > 0 {
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "join request is already pending",
			"status":  "pending",
		})
		return
	}

	_, err = db.Exec(`
		INSERT INTO group_join_requests (group_id, requester_id, status)
		VALUES (?, ?, 'pending')
		ON CONFLICT(group_id, requester_id)
		DO UPDATE SET
			status = 'pending',
			updated_at = CURRENT_TIMESTAMP
	`, groupID, currentUserID)

	if err != nil {
		errorJSON(w, "could not create join request", http.StatusInternalServerError)
		return
	}

	err = notifyGroupJoinRequestReceived(groupID, currentUserID)
	if err != nil {
		errorJSON(w, "join request created but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "join request sent",
		"status":  "pending",
	})
}

func listGroupJoinRequestsHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	owner, err := isGroupOwner(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group ownership", http.StatusInternalServerError)
		return
	}

	if !owner {
		errorJSON(w, "only the group owner can view join requests", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT
			group_join_requests.id,
			group_join_requests.group_id,
			group_join_requests.requester_id,
			users.first_name || ' ' || users.last_name AS requester_name,
			COALESCE(users.nickname, '') AS requester_nickname,
			group_join_requests.status,
			group_join_requests.created_at
		FROM group_join_requests
		JOIN users ON users.id = group_join_requests.requester_id
		WHERE group_join_requests.group_id = ?
		  AND group_join_requests.status = 'pending'
		ORDER BY group_join_requests.created_at DESC
	`, groupID)

	if err != nil {
		errorJSON(w, "could not load join requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	requests := []GroupJoinRequestResponse{}

	for rows.Next() {
		var request GroupJoinRequestResponse

		err := rows.Scan(
			&request.ID,
			&request.GroupID,
			&request.RequesterID,
			&request.RequesterName,
			&request.RequesterNickname,
			&request.Status,
			&request.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read join request data", http.StatusInternalServerError)
			return
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading join requests", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func acceptGroupJoinRequestHandler(w http.ResponseWriter, r *http.Request, groupID int, requestID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	owner, err := isGroupOwner(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group ownership", http.StatusInternalServerError)
		return
	}

	if !owner {
		errorJSON(w, "only the group owner can accept join requests", http.StatusForbidden)
		return
	}

	var requesterID int
	var status string

	err = db.QueryRow(`
		SELECT requester_id, status
		FROM group_join_requests
		WHERE id = ?
		  AND group_id = ?
	`, requestID, groupID).Scan(&requesterID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "join request not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load join request", http.StatusInternalServerError)
		return
	}

	if status != "pending" {
		errorJSON(w, "join request is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'member')
	`, groupID, requesterID)

	if err != nil {
		errorJSON(w, "could not add user to group", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		UPDATE group_join_requests
		SET status = 'accepted',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND group_id = ?
	`, requestID, groupID)

	if err != nil {
		errorJSON(w, "could not update join request", http.StatusInternalServerError)
		return
	}

	err = notifyGroupJoinRequestAccepted(requesterID, currentUserID, groupID)
	if err != nil {
		errorJSON(w, "join request accepted but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "join request accepted",
	})
}

func declineGroupJoinRequestHandler(w http.ResponseWriter, r *http.Request, groupID int, requestID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	owner, err := isGroupOwner(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group ownership", http.StatusInternalServerError)
		return
	}

	if !owner {
		errorJSON(w, "only the group owner can decline join requests", http.StatusForbidden)
		return
	}

	var requesterID int
	var status string

	err = db.QueryRow(`
		SELECT requester_id, status
		FROM group_join_requests
		WHERE id = ?
		AND group_id = ?
	`, requestID, groupID).Scan(&requesterID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "join request not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load join request", http.StatusInternalServerError)
		return
	}

	if status != "pending" {
		errorJSON(w, "join request is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		UPDATE group_join_requests
		SET status = 'declined',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND group_id = ?
	`, requestID, groupID)

	if err != nil {
		errorJSON(w, "could not decline join request", http.StatusInternalServerError)
		return
	}

	err = notifyGroupJoinRequestDeclined(requesterID, currentUserID, groupID)
	if err != nil {
		errorJSON(w, "join request declined but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "join request declined",
	})
}

func createGroupInvitationHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	owner, err := isGroupOwner(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group ownership", http.StatusInternalServerError)
		return
	}

	if !owner {
		errorJSON(w, "only the group owner can invite users", http.StatusForbidden)
		return
	}

	var req CreateGroupInvitationRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.InviteeID <= 0 {
		errorJSON(w, "invalid invitee id", http.StatusBadRequest)
		return
	}

	if req.InviteeID == currentUserID {
		errorJSON(w, "you cannot invite yourself", http.StatusBadRequest)
		return
	}

	var inviteeExists int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE id = ?
	`, req.InviteeID).Scan(&inviteeExists)

	if err != nil {
		errorJSON(w, "could not check invitee", http.StatusInternalServerError)
		return
	}

	if inviteeExists == 0 {
		errorJSON(w, "invitee not found", http.StatusNotFound)
		return
	}

	member, err := isGroupMember(req.InviteeID, groupID)
	if err != nil {
		errorJSON(w, "could not check group membership", http.StatusInternalServerError)
		return
	}

	if member {
		errorJSON(w, "user is already a group member", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT INTO group_invitations (
			group_id,
			inviter_id,
			invitee_id,
			status
		)
		VALUES (?, ?, ?, 'pending')
		ON CONFLICT(group_id, invitee_id)
		DO UPDATE SET
			inviter_id = excluded.inviter_id,
			status = 'pending',
			updated_at = CURRENT_TIMESTAMP
	`, groupID, currentUserID, req.InviteeID)

	if err != nil {
		errorJSON(w, "could not create group invitation", http.StatusInternalServerError)
		return
	}

	err = notifyGroupInvitationReceived(req.InviteeID, currentUserID, groupID)
	if err != nil {
		errorJSON(w, "invitation created but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "group invitation sent",
		"status":  "pending",
	})
}

func listGroupInvitationsHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	owner, err := isGroupOwner(currentUserID, groupID)
	if err != nil {
		errorJSON(w, "could not check group ownership", http.StatusInternalServerError)
		return
	}

	if !owner {
		errorJSON(w, "only the group owner can view invitations", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT
			group_invitations.id,
			group_invitations.group_id,
			groups.title,
			group_invitations.inviter_id,
			inviter.first_name || ' ' || inviter.last_name AS inviter_name,
			group_invitations.invitee_id,
			invitee.first_name || ' ' || invitee.last_name AS invitee_name,
			COALESCE(invitee.nickname, '') AS invitee_nickname,
			group_invitations.status,
			group_invitations.created_at
		FROM group_invitations
		JOIN groups ON groups.id = group_invitations.group_id
		JOIN users AS inviter ON inviter.id = group_invitations.inviter_id
		JOIN users AS invitee ON invitee.id = group_invitations.invitee_id
		WHERE group_invitations.group_id = ?
		ORDER BY group_invitations.created_at DESC
	`, groupID)

	if err != nil {
		errorJSON(w, "could not load group invitations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	invitations := []GroupInvitationResponse{}

	for rows.Next() {
		var invitation GroupInvitationResponse

		err := rows.Scan(
			&invitation.ID,
			&invitation.GroupID,
			&invitation.GroupTitle,
			&invitation.InviterID,
			&invitation.InviterName,
			&invitation.InviteeID,
			&invitation.InviteeName,
			&invitation.InviteeNickname,
			&invitation.Status,
			&invitation.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read invitation data", http.StatusInternalServerError)
			return
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading invitations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, invitations)
}

func groupInvitationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listMyGroupInvitationsHandler(w, r)
}

func groupInvitationsSubroutesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/group-invitations/")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		errorJSON(w, "route not found", http.StatusNotFound)
		return
	}

	invitationID, err := strconv.Atoi(parts[0])
	if err != nil || invitationID <= 0 {
		errorJSON(w, "invalid invitation id", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := parts[1]

	if action == "accept" {
		acceptGroupInvitationHandler(w, r, invitationID)
		return
	}

	if action == "decline" {
		declineGroupInvitationHandler(w, r, invitationID)
		return
	}

	errorJSON(w, "route not found", http.StatusNotFound)
}

func listMyGroupInvitationsHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value(userIDKey).(int)

	rows, err := db.Query(`
		SELECT
			group_invitations.id,
			group_invitations.group_id,
			groups.title,
			group_invitations.inviter_id,
			inviter.first_name || ' ' || inviter.last_name AS inviter_name,
			group_invitations.invitee_id,
			invitee.first_name || ' ' || invitee.last_name AS invitee_name,
			COALESCE(invitee.nickname, '') AS invitee_nickname,
			group_invitations.status,
			group_invitations.created_at
		FROM group_invitations
		JOIN groups ON groups.id = group_invitations.group_id
		JOIN users AS inviter ON inviter.id = group_invitations.inviter_id
		JOIN users AS invitee ON invitee.id = group_invitations.invitee_id
		WHERE group_invitations.invitee_id = ?
		  AND group_invitations.status = 'pending'
		ORDER BY group_invitations.created_at DESC
	`, currentUserID)

	if err != nil {
		errorJSON(w, "could not load my group invitations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	invitations := []GroupInvitationResponse{}

	for rows.Next() {
		var invitation GroupInvitationResponse

		err := rows.Scan(
			&invitation.ID,
			&invitation.GroupID,
			&invitation.GroupTitle,
			&invitation.InviterID,
			&invitation.InviterName,
			&invitation.InviteeID,
			&invitation.InviteeName,
			&invitation.InviteeNickname,
			&invitation.Status,
			&invitation.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read invitation data", http.StatusInternalServerError)
			return
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading invitations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, invitations)
}

func acceptGroupInvitationHandler(w http.ResponseWriter, r *http.Request, invitationID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var groupID int
	var inviterID int
	var inviteeID int
	var status string

	err := db.QueryRow(`
		SELECT group_id, inviter_id, invitee_id, status
		FROM group_invitations
		WHERE id = ?
	`, invitationID).Scan(&groupID, &inviterID, &inviteeID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "invitation not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load invitation", http.StatusInternalServerError)
		return
	}

	if inviteeID != currentUserID {
		errorJSON(w, "you cannot accept this invitation", http.StatusForbidden)
		return
	}

	if status != "pending" {
		errorJSON(w, "invitation is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'member')
	`, groupID, currentUserID)

	if err != nil {
		errorJSON(w, "could not add you to group", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
		UPDATE group_invitations
		SET status = 'accepted',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, invitationID)

	if err != nil {
		errorJSON(w, "could not update invitation", http.StatusInternalServerError)
		return
	}

	err = notifyGroupInvitationAccepted(inviterID, currentUserID, groupID)
	if err != nil {
		errorJSON(w, "invitation accepted but notification failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "group invitation accepted",
	})
}

func declineGroupInvitationHandler(w http.ResponseWriter, r *http.Request, invitationID int) {
	currentUserID := r.Context().Value(userIDKey).(int)

	var inviteeID int
	var status string

	err := db.QueryRow(`
		SELECT invitee_id, status
		FROM group_invitations
		WHERE id = ?
	`, invitationID).Scan(&inviteeID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, "invitation not found", http.StatusNotFound)
			return
		}

		errorJSON(w, "could not load invitation", http.StatusInternalServerError)
		return
	}

	if inviteeID != currentUserID {
		errorJSON(w, "you cannot decline this invitation", http.StatusForbidden)
		return
	}

	if status != "pending" {
		errorJSON(w, "invitation is not pending", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		UPDATE group_invitations
		SET status = 'declined',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, invitationID)

	if err != nil {
		errorJSON(w, "could not decline invitation", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "group invitation declined",
	})
}

