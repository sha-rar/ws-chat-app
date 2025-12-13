package websocket

type BroadcastMessage struct {
	RoomID string
	Data   []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run listens for register/unregister/broadcast events
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			roomID := client.roomID
			if roomID == "" {
				roomID = "general"
				client.roomID = roomID
			}

			if h.rooms[roomID] == nil {
				h.rooms[roomID] = make(map[*Client]bool)
			}
			h.rooms[roomID][client] = true

		case client := <-h.unregister:
			roomID := client.roomID
			if roomClients, ok := h.rooms[roomID]; ok {
				if _, exists := roomClients[client]; exists {
					delete(roomClients, client)
					close(client.send)
					if len(roomClients) == 0 {
						delete(h.rooms, roomID)
					}
				}
			}

		case msg := <-h.broadcast:
			roomClients, ok := h.rooms[msg.RoomID]
			if !ok {
				continue
			}

			for client := range roomClients {
				select {
				case client.send <- msg.Data:
				default:
					close(client.send)
					delete(roomClients, client)
					if len(roomClients) == 0 {
						delete(h.rooms, msg.RoomID)
					}
				}
			}
		}
	}
}
