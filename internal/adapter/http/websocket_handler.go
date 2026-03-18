package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/hillscheck/internal/adapter/websocket"
	"github.com/hillscheck/internal/infrastructure/auth"
)

type WebSocketHandler struct {
	hub *websocket.Hub
	jwt *auth.JWTService
	log *zap.Logger
}

func NewWebSocketHandler(hub *websocket.Hub, jwt *auth.JWTService, log *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, jwt: jwt, log: log}
}

// ServeWS upgrades the connection to WebSocket.
// Validates the JWT from the "token" query param and registers the client by real userID.
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token query param required", http.StatusUnauthorized)
		return
	}
	claims, err := h.jwt.Validate(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	h.hub.ServeWS(claims.UserID, w, r)
}
