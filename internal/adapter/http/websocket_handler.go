package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/hillscheck/internal/adapter/websocket"
)

type WebSocketHandler struct {
	hub *websocket.Hub
	log *zap.Logger
}

func NewWebSocketHandler(hub *websocket.Hub, log *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, log: log}
}

// ServeWS upgrades the connection to WebSocket.
// The userID is taken from the "token" query parameter for now.
// TODO: validate the token properly (JWT verification).
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("token")
	if userID == "" {
		http.Error(w, "token query param required", http.StatusUnauthorized)
		return
	}
	h.hub.ServeWS(userID, w, r)
}
