package port

import "github.com/hillscheck/internal/domain"

// Notifier pushes real-time updates to connected clients via WebSocket.
type Notifier interface {
	PushProfileUpdate(userID string, profile domain.GameProfile)
	PushTransactionProcessed(userID string, tx domain.Transaction)
}
