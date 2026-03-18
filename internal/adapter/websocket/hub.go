package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // relax for dev; tighten in prod
}

type client struct {
	userID string
	send   chan []byte
	conn   *websocket.Conn
}

// Hub manages all active WebSocket connections, keyed by userID.
// It implements port.Notifier.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string][]*client
	register   chan *client
	unregister chan *client
	log        *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[string][]*client),
		register:   make(chan *client, 64),
		unregister: make(chan *client, 64),
		log:        log,
	}
}

// Run starts the hub's event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.userID] = append(h.clients[c.userID], c)
			h.mu.Unlock()
			h.log.Info("ws client connected", zap.String("user_id", c.userID))

		case c := <-h.unregister:
			h.mu.Lock()
			list := h.clients[c.userID]
			for i, existing := range list {
				if existing == c {
					h.clients[c.userID] = append(list[:i], list[i+1:]...)
					break
				}
			}
			if len(h.clients[c.userID]) == 0 {
				delete(h.clients, c.userID)
			}
			h.mu.Unlock()
			close(c.send)
			h.log.Info("ws client disconnected", zap.String("user_id", c.userID))
		}
	}
}

// ServeWS upgrades the HTTP connection and registers the client.
func (h *Hub) ServeWS(userID string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", zap.Error(err))
		return
	}

	c := &client{userID: userID, send: make(chan []byte, 256), conn: conn}
	h.register <- c

	go c.writePump(h)
	go c.readPump(h)
}

func (h *Hub) PushProfileUpdate(userID string, profile domain.GameProfile) {
	h.broadcast(userID, "profile_update", profile)
}

func (h *Hub) PushTransactionProcessed(userID string, tx domain.Transaction) {
	h.broadcast(userID, "transaction_processed", tx)
}

func (h *Hub) broadcast(userID, msgType string, payload any) {
	data, err := json.Marshal(map[string]any{"type": msgType, "payload": payload})
	if err != nil {
		h.log.Error("marshal ws message", zap.Error(err))
		return
	}

	h.mu.RLock()
	clients := make([]*client, len(h.clients[userID]))
	copy(clients, h.clients[userID])
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.send <- data:
		default:
			// Slow client — drop message rather than block the caller.
			h.log.Warn("ws send buffer full, dropping message", zap.String("user_id", userID))
		}
	}
}

func (c *client) writePump(h *Hub) {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *client) readPump(h *Hub) {
	defer func() { h.unregister <- c }()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
		// Inbound messages from clients are ignored for now.
	}
}
