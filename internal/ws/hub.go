package ws

import "log"

type Broadcast struct {
	ClientIds []string
	Payload   []byte
}

type Hub struct {
	Register   chan *Client
	unregister chan *Client
	broadcast  chan Broadcast
	clients    map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Broadcast),
		clients:    make(map[*Client]bool),
	}
}

func RunHub(h *Hub) {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Println("Client registered: ", client.Id)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Println("Client unregistered: ", client.Id)
			}
		case message := <-h.broadcast:
			targets := make(map[string]bool, len(message.ClientIds))
			for _, id := range message.ClientIds {
				targets[id] = true
			}

			for client := range h.clients {
				if !targets[client.Id] {
					continue
				}
				select {
				case client.Send <- message.Payload:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) SendTo(clientIDs []string, payload []byte) {
	h.broadcast <- Broadcast{
		ClientIds: clientIDs,
		Payload:   payload,
	}
}
