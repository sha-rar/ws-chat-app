package websocket

type BroadcastMessage struct {
	RoomID string
	Data   []byte
}

// DirectMessage is delivered to a single client rather than its whole room
type DirectMessage struct {
	Client *Client
	Data   []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan BroadcastMessage
	direct     chan DirectMessage
	register   chan *Client
	unregister chan *Client
	revoke     chan string
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		direct:     make(chan DirectMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		revoke:     make(chan string),
	}
}

// DisconnectSession closes every socket opened with the given session, e.g. after logout
func (h *Hub) DisconnectSession(sessionID string) {
	if sessionID == "" {
		return
	}
	h.revoke <- sessionID
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

		case sessionID := <-h.revoke:
			// Closing send makes writePump close the socket
			for roomID, roomClients := range h.rooms {
				for client := range roomClients {
					if client.sessionID == sessionID {
						close(client.send)
						delete(roomClients, client)
					}
				}
				if len(roomClients) == 0 {
					delete(h.rooms, roomID)
				}
			}

		case msg := <-h.direct:
			// The client may have been dropped already; only send if still registered
			if roomClients, ok := h.rooms[msg.Client.roomID]; ok && roomClients[msg.Client] {
				select {
				case msg.Client.send <- msg.Data:
				default:
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
