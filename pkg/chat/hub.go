package chat

// Hub manages the WebSocket clients and facilitates message broadcasting
type Hub struct {
	clients    map[*Client]bool // A map to store connected clients
	broadcast  chan []byte   // Channel for broadcasting messages to clients
	register   chan *Client  // Channel for registering clients
	unregister chan *Client  // Channel for unregistering clients
}

// NewHub creates a new instance of Hub with initialized channels and map
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the Hub's main loop for handling client registration, unregistration, and message broadcasting
func (h *Hub) Run() {
	for {
		select {
			// Register a new client
		case client := <-h.register:
			h.clients[client] = true
			// Unregister an existing client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			// Broadcast a message to all connected clients
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
