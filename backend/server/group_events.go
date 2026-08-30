package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	err = notifyGroupEventCreated(
		groupID,
		currentUserID,
		title,
	)

	if err != nil {
		log.Printf(
			"group event %d created, but notifications failed: %v",
			eventID,
			err,
		)
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

	if err == sql.ErrNoRows {
		return "none", nil
	}

	if err != nil {
		return "", err
	}

	return response, nil
}

func listGroupEventsHandler(
	w http.ResponseWriter,
	r *http.Request,
	groupID int,
) {
	currentUserID, ok :=
		requireGroupMember(
			w,
			r,
			groupID,
		)

	if !ok {
		return
	}

	// PAGINATION
	limit := 10

	if rawLimit :=
		r.URL.Query().Get("limit"); rawLimit != "" {

		parsedLimit, err :=
			strconv.Atoi(rawLimit)

		if err != nil ||
			parsedLimit <= 0 ||
			parsedLimit > 50 {

			errorJSON(
				w,
				"invalid event limit",
				http.StatusBadRequest,
			)
			return
		}

		limit = parsedLimit
	}

	offset := 0

	if rawOffset :=
		r.URL.Query().Get("offset"); rawOffset != "" {

		parsedOffset, err :=
			strconv.Atoi(rawOffset)

		if err != nil ||
			parsedOffset < 0 {

			errorJSON(
				w,
				"invalid event offset",
				http.StatusBadRequest,
			)
			return
		}

		offset = parsedOffset
	}

	// UPCOMING + PAST
	scope :=
		strings.ToLower(
			strings.TrimSpace(
				r.URL.Query().Get("scope"),
			),
		)

	if scope == "" {
		scope = "upcoming"
	}

	var timeCondition string
	var orderDirection string

	switch scope {

	case "upcoming":
		timeCondition =
			"group_events.event_time >= ?"

		orderDirection = "ASC"

	case "past":
		timeCondition =
			"group_events.event_time < ?"

		orderDirection = "DESC"

	default:
		errorJSON(
			w,
			"invalid event scope",
			http.StatusBadRequest,
		)
		return
	}

	nowValue :=
		strings.TrimSpace(
			r.URL.Query().Get("now"),
		)

	if nowValue == "" {
		nowValue =
			time.Now().Format(
				"2006-01-02 15:04:05",
			)
	}

	_, err :=
		time.Parse(
			"2006-01-02 15:04:05",
			nowValue,
		)

	if err != nil {
		errorJSON(
			w,
			"invalid current time",
			http.StatusBadRequest,
		)
		return
	}

	query := fmt.Sprintf(`
		SELECT
			group_events.id,
			group_events.group_id,
			group_events.creator_id,

			users.first_name || ' ' ||
				users.last_name
				AS creator_name,
			
			COALESCE(
				users.avatar_path,
				''
			) AS creator_avatar_path,

			group_events.title,
			group_events.description,
			group_events.event_time,

			(
				SELECT COUNT(*)
				FROM group_event_responses
				WHERE
					group_event_responses.event_id =
						group_events.id
					AND
					group_event_responses.response =
						'going'
			) AS going_count,

			(
				SELECT COUNT(*)
				FROM group_event_responses
				WHERE
					group_event_responses.event_id =
						group_events.id
					AND
					group_event_responses.response =
						'not_going'
			) AS not_going_count,

			COALESCE(
				(
					SELECT
						group_event_responses.response

					FROM group_event_responses

					WHERE
						group_event_responses.event_id =
							group_events.id
						AND
						group_event_responses.user_id = ?

					LIMIT 1
				),
				'none'
			) AS my_response,

			group_events.created_at

		FROM group_events

		JOIN users
			ON users.id =
				group_events.creator_id

		WHERE
			group_events.group_id = ?

			AND %s

		ORDER BY
			group_events.event_time %s,
			group_events.id %s

		LIMIT ?
		OFFSET ?
	`,
		timeCondition,
		orderDirection,
		orderDirection,
	)

	rows, err :=
		db.Query(
			query,
			currentUserID,
			groupID,
			nowValue,
			limit+1,
			offset,
		)

	if err != nil {
		errorJSON(
			w,
			"could not load group events",
			http.StatusInternalServerError,
		)
		return
	}

	defer rows.Close()

	events :=
		[]GroupEventResponse{}

	for rows.Next() {
		var event GroupEventResponse

		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.CreatorName,
			&event.CreatorAvatarPath,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.GoingCount,
			&event.NotGoingCount,
			&event.MyResponse,
			&event.CreatedAt,
		)

		if err != nil {
			errorJSON(
				w,
				"could not read group event data",
				http.StatusInternalServerError,
			)
			return
		}

		events =
			append(events, event)
	}

	if err := rows.Err(); err != nil {
		errorJSON(
			w,
			"error while reading group events",
			http.StatusInternalServerError,
		)
		return
	}

	hasMore :=
		len(events) > limit

	if hasMore {
		events =
			events[:limit]
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"events": events,

			"has_more": hasMore,

			"next_offset": offset + len(events),
		},
	)
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
