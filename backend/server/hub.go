package server

import (
	"log"
	"sort"
)

func newHub() *Hub {
	return &Hub{
		clients: make(map[int]*Client),

		register:   make(chan *Client),
		unregister: make(chan *Client),

		deliver: make(chan HubDelivery),
	}
}

var chatHub = newHub()

func (h *Hub) run() {
	for {
		select {

		case client := <-h.register:
			oldClient, wasOnline := h.clients[client.userID]

			if wasOnline &&
				oldClient != client {

				close(oldClient.send)
				oldClient.conn.Close()
			}

			h.clients[client.userID] = client

			h.sendPresenceSnapshot(client)

			if !wasOnline {
				h.broadcastPresence(
					client.userID,
					true,
				)
			}

			log.Printf("websocket connected: user %d", client.userID)

		case client := <-h.unregister:
			currentClient, exists := h.clients[client.userID]

			if exists &&
				currentClient == client {

				delete(h.clients, client.userID)

				close(client.send)

				h.broadcastPresence(client.userID, false)

				log.Printf("websocket disconnected: user %d", client.userID)
			}

		case delivery := <-h.deliver:
			client, exists := h.clients[delivery.UserID]

			if !exists {
				continue
			}

			select {
			case client.send <- delivery.Event:

			default:
				close(client.send)
				client.conn.Close()

				delete(h.clients, delivery.UserID)

				h.broadcastPresence(delivery.UserID, false)

				log.Printf("websocket disconnected: user %d", delivery.UserID)
			}
		}
	}
}

func (h *Hub) sendPresenceSnapshot(client *Client) {
	userIDs := make([]int, 0, len(h.clients))

	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}

	sort.Ints(userIDs)

	event := WebSocketEvent{
		Type: "presence_snapshot",
		Data: PresenceSnapshot{
			UserIDs: userIDs,
		},
	}

	select {
	case client.send <- event:
	default:
	}
}

func (h *Hub) broadcastPresence(userID int, online bool) {
	event := WebSocketEvent{
		Type: "presence",
		Data: PresenceUpdate{
			UserID: userID,
			Online: online,
		},
	}

	for connectedUserID, client := range h.clients {

		if connectedUserID == userID {
			continue
		}

		select {
		case client.send <- event:
		default:
		}
	}
}
