package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func createGroupEventHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID, ok := requireGroupMember(w, r, groupID)
	if !ok {
		return
	}

	var req CreateGroupEventRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorJSON(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	eventTime := strings.TrimSpace(req.EventTime)

	if title == "" {
		errorJSON(w, "event title is required", http.StatusBadRequest)
		return
	}

	if eventTime == "" {
		errorJSON(w, "event time is required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		INSERT INTO group_events (
			group_id,
			creator_id,
			title,
			description,
			event_time
		)
		VALUES (?, ?, ?, ?, ?)
	`, groupID, currentUserID, title, description, eventTime)

	if err != nil {
		errorJSON(w, "could not create group event", http.StatusInternalServerError)
		return
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		errorJSON(w, "event created but could not read event id", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "group event created successfully",
		"event_id": eventID,
	})
}

func countGroupEventResponses(eventID int, response string) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM group_event_responses
		WHERE event_id = ?
		  AND response = ?
	`, eventID, response).Scan(&count)

	return count, err
}

func getMyGroupEventResponse(eventID int, userID int) (string, error) {
	var response string

	err := db.QueryRow(`
		SELECT response
		FROM group_event_responses
		WHERE event_id = ?
		  AND user_id = ?
	`, eventID, userID).Scan(&response)

	if err != nil {
		return "none", nil
	}

	return response, nil
}

func listGroupEventsHandler(w http.ResponseWriter, r *http.Request, groupID int) {
	currentUserID, ok := requireGroupMember(w, r, groupID)
	if !ok {
		return
	}

	rows, err := db.Query(`
		SELECT
			group_events.id,
			group_events.group_id,
			group_events.creator_id,
			users.first_name || ' ' || users.last_name AS creator_name,
			group_events.title,
			group_events.description,
			group_events.event_time,
			group_events.created_at
		FROM group_events
		JOIN users ON users.id = group_events.creator_id
		WHERE group_events.group_id = ?
		ORDER BY group_events.event_time ASC
	`, groupID)

	if err != nil {
		errorJSON(w, "could not load group events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []GroupEventResponse{}

	for rows.Next() {
		var event GroupEventResponse

		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.CreatorName,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
		)

		if err != nil {
			errorJSON(w, "could not read group event data", http.StatusInternalServerError)
			return
		}

		event.GoingCount, err = countGroupEventResponses(event.ID, "going")
		if err != nil {
			errorJSON(w, "could not count going responses", http.StatusInternalServerError)
			return
		}

		event.NotGoingCount, err = countGroupEventResponses(event.ID, "not_going")
		if err != nil {
			errorJSON(w, "could not count not-going responses", http.StatusInternalServerError)
			return
		}

		event.MyResponse, err = getMyGroupEventResponse(event.ID, currentUserID)
		if err != nil {
			errorJSON(w, "could not read your event response", http.StatusInternalServerError)
			return
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		errorJSON(w, "error while reading group events", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, events)
}

func groupEventBelongsToGroup(eventID int, groupID int) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM group_events
		WHERE id = ?
		  AND group_id = ?
	`, eventID, groupID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func respondGroupEventHandler(w http.ResponseWriter, r *http.Request, groupID int, eventID int, response string) {
	currentUserID, ok := requireGroupMember(w, r, groupID)
	if !ok {
		return
	}

	belongs, err := groupEventBelongsToGroup(eventID, groupID)
	if err != nil {
		errorJSON(w, "could not check event", http.StatusInternalServerError)
		return
	}

	if !belongs {
		errorJSON(w, "event not found", http.StatusNotFound)
		return
	}

	_, err = db.Exec(`
		INSERT INTO group_event_responses (
			event_id,
			user_id,
			response
		)
		VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id)
		DO UPDATE SET
			response = excluded.response,
			updated_at = CURRENT_TIMESTAMP
	`, eventID, currentUserID, response)

	if err != nil {
		errorJSON(w, "could not save event response", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "event response saved",
		"response": response,
	})
}
