package ws

import "log"

type Broadcast[T Client] struct {
	ClientIds []string
	Payload   []byte
}

type Hub[T ComparableClient] struct {
	Register   chan T
	Unregister chan T
	broadcast  chan Broadcast[T]
	clients    map[T]bool
}

func NewHub[T ComparableClient]() *Hub[T] {
	return &Hub[T]{
		Register:   make(chan T),
		Unregister: make(chan T),
		broadcast:  make(chan Broadcast[T]),
		clients:    make(map[T]bool),
	}
}

func (h *Hub[T]) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Println("Client registered: ", client.Id())
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send())
				log.Println("Client unregistered: ", client.Id())
			}
		case message := <-h.broadcast:
			targets := make(map[string]bool, len(message.ClientIds))
			for _, id := range message.ClientIds {
				targets[id] = true
			}

			for client := range h.clients {
				if !targets[client.Id()] {
					continue
				}
				select {
				case client.Send() <- message.Payload:
				default:
					close(client.Send())
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub[T]) SendTo(clientIDs []string, payload []byte) {
	h.broadcast <- Broadcast[T]{
		ClientIds: clientIDs,
		Payload:   payload,
	}
}
