package http

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	mw "github.com/hillscheck/internal/adapter/http/middleware"
)

type Handlers struct {
	Transaction *TransactionHandler
	Receipt     *ReceiptHandler
	Profile     *ProfileHandler
	Quests      *QuestHandler
	WebSocket   *WebSocketHandler
}

// NewRouter builds and returns the main chi router.
// staticFiles is the embedded web/dist FS (pass nil to skip static serving).
func NewRouter(
	h Handlers,
	rdb *redis.Client,
	log *zap.Logger,
	staticFiles fs.FS,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(mw.RequestID)
	r.Use(mw.RateLimit(rdb, 200, time.Minute))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/transactions/import", h.Transaction.Import)
		r.Get("/transactions", h.Transaction.List)

		r.Post("/receipts/upload", h.Receipt.Upload)

		r.Get("/profile", h.Profile.Get)
		r.Get("/quests", h.Quests.List)
	})

	// WebSocket — auth handled inside the handler via query param token.
	r.Get("/ws", h.WebSocket.ServeWS)

	// Health check — no auth required.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve embedded React SPA — must be last.
	if staticFiles != nil {
		fileServer := http.FileServer(http.FS(staticFiles))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// fs.FS paths must not start with '/'.
			fsPath := strings.TrimPrefix(r.URL.Path, "/")
			if _, err := staticFiles.Open(fsPath); err != nil {
				// SPA fallback: unknown path → serve index.html for client-side routing.
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	_ = log
	return r
}
