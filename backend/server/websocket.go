package server

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		if origin == "" {
			return true
		}

		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		hostname := u.Hostname()

		return hostname == "localhost" ||
			hostname == "127.0.0.1"
	},
}

func websocketHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		errorJSON(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	currentUserID := r.Context().Value(userIDKey).(int)

	conn, err := websocketUpgrader.Upgrade(
		w,
		r,
		nil,
	)

	if err != nil {
		log.Printf(
			"websocket upgrade error for user %d: %v",
			currentUserID,
			err,
		)
		return
	}

	client := &Client{
		hub:    chatHub,
		userID: currentUserID,
		conn:   conn,

		send: make(chan WebSocketEvent, 32),
	}

	chatHub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for event := range c.send {
		err := c.conn.WriteJSON(event)

		if err != nil {
			log.Printf(
				"websocket write error for user %d: %v",
				c.userID,
				err,
			)

			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var event IncomingWebSocketEvent

		err := c.conn.ReadJSON(&event)

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf(
					"websocket read error for user %d: %v",
					c.userID,
					err,
				)
			}

			return
		}

		switch event.Type {

		case "private_message":
			c.handlePrivateMessage(event)

		case "private_typing":
			c.handlePrivateTyping(event)

		case "group_message":
			c.handleGroupMessage(event)

		default:
			c.sendError("unknown websocket event type")
		}
	}
}

func (c *Client) sendError(message string) {
	select {
	case c.send <- WebSocketEvent{
		Type:  "error",
		Error: message,
	}:

	default:
	}
}

func (c *Client) handlePrivateMessage(event IncomingWebSocketEvent) {
	content := strings.TrimSpace(event.Content)

	if event.ReceiverID <= 0 {
		c.sendError("invalid receiver id")
		return
	}

	if event.ReceiverID == c.userID {
		c.sendError("cannot send a private message to yourself")
		return
	}

	if content == "" {
		c.sendError("message content is required")
		return
	}

	exists, err := chatUserExists(event.ReceiverID)
	if err != nil {
		log.Printf(
			"could not check receiver %d: %v",
			event.ReceiverID,
			err,
		)

		c.sendError("could not check receiver")
		return
	}

	if !exists {
		c.sendError("receiver not found")
		return
	}

	allowed, err := canPrivateChat(
		c.userID,
		event.ReceiverID,
	)

	if err != nil {
		log.Printf(
			"could not check private chat permission: %v",
			err,
		)

		c.sendError("could not check chat permission")
		return
	}

	if !allowed {
		c.sendError("you cannot chat with this user")
		return
	}

	message, err := savePrivateMessage(
		c.userID,
		event.ReceiverID,
		content,
	)

	if err != nil {
		log.Printf(
			"could not save private message: %v",
			err,
		)

		c.sendError("could not save private message")
		return
	}

	messageEvent := WebSocketEvent{
		Type: "private_message",
		Data: message,
	}

	c.hub.deliver <- HubDelivery{
		UserID: event.ReceiverID,
		Event:  messageEvent,
	}

	c.hub.deliver <- HubDelivery{
		UserID: c.userID,
		Event:  messageEvent,
	}
}

func (c *Client) handlePrivateTyping(
	event IncomingWebSocketEvent,
) {
	if event.ReceiverID <= 0 {
		c.sendError("invalid receiver id")
		return
	}

	if event.ReceiverID == c.userID {
		c.sendError(
			"cannot send typing status to yourself",
		)
		return
	}

	exists, err :=
		chatUserExists(event.ReceiverID)

	if err != nil {
		log.Printf(
			"could not check typing receiver %d: %v",
			event.ReceiverID,
			err,
		)

		c.sendError("could not check receiver")
		return
	}

	if !exists {
		c.sendError("receiver not found")
		return
	}

	allowed, err := canPrivateChat(
		c.userID,
		event.ReceiverID,
	)

	if err != nil {
		log.Printf(
			"could not check private chat permission: %v",
			err,
		)

		c.sendError(
			"could not check chat permission",
		)
		return
	}

	if !allowed {
		c.sendError(
			"you cannot chat with this user",
		)
		return
	}

	typingEvent := WebSocketEvent{
		Type: "private_typing",
		Data: PrivateTypingUpdate{
			SenderID: c.userID,

			ReceiverID: event.ReceiverID,

			Typing: event.Typing,
		},
	}

	c.hub.deliver <- HubDelivery{
		UserID: event.ReceiverID,
		Event:  typingEvent,
	}
}

func (c *Client) handleGroupMessage(event IncomingWebSocketEvent) {
	content := strings.TrimSpace(
		event.Content,
	)

	if event.GroupID <= 0 {
		c.sendError(
			"invalid group id",
		)
		return
	}

	if content == "" {
		c.sendError(
			"message content is required",
		)
		return
	}

	exists, err := groupExists(
		event.GroupID,
	)

	if err != nil {
		log.Printf(
			"could not check group %d: %v",
			event.GroupID,
			err,
		)

		c.sendError(
			"could not check group",
		)

		return
	}

	if !exists {
		c.sendError(
			"group not found",
		)
		return
	}

	member, err := isGroupMember(
		c.userID,
		event.GroupID,
	)

	if err != nil {
		log.Printf(
			"could not check group membership: %v",
			err,
		)

		c.sendError(
			"could not check group membership",
		)

		return
	}

	if !member {
		c.sendError(
			"only group members can send group messages",
		)

		return
	}

	message, err := saveGroupMessage(
		event.GroupID,
		c.userID,
		content,
	)

	if err != nil {
		log.Printf(
			"could not save group message: %v",
			err,
		)

		c.sendError(
			"could not save group message",
		)

		return
	}

	memberIDs, err := getGroupMemberIDs(
		event.GroupID,
	)

	if err != nil {
		log.Printf(
			"could not load group members: %v",
			err,
		)

		c.sendError(
			"message saved but could not deliver it",
		)

		return
	}

	messageEvent := WebSocketEvent{
		Type: "group_message",
		Data: message,
	}

	for _, memberID := range memberIDs {
		c.hub.deliver <- HubDelivery{
			UserID: memberID,
			Event:  messageEvent,
		}
	}
}
