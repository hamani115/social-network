package server

import "log"

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
			if oldClient, exists := h.clients[client.userID]; exists {
				if oldClient != client {
					close(oldClient.send)
					oldClient.conn.Close()
				}
			}

			h.clients[client.userID] = client

			log.Printf(
				"websocket connected: user %d",
				client.userID,
			)

		case client := <-h.unregister:
			currentClient, exists := h.clients[client.userID]

			if exists && currentClient == client {
				delete(h.clients, client.userID)
				close(client.send)

				log.Printf(
					"websocket disconnected: user %d",
					client.userID,
				)
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

				delete(
					h.clients,
					delivery.UserID,
				)
			}
		}
	}
}
