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

	"github.com/hillscheck/internal/infrastructure/auth"
	mw "github.com/hillscheck/internal/adapter/http/middleware"
)

type Handlers struct {
	Auth        *AuthHandler
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
	jwt *auth.JWTService,
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

	// Auth routes — no JWT required, tighter rate limit (10 req/min per IP)
	r.Route("/auth", func(r chi.Router) {
		r.Use(mw.RateLimit(rdb, 10, time.Minute))
		r.Post("/register", h.Auth.Register)
		r.Post("/login", h.Auth.Login)
		r.Post("/refresh", h.Auth.Refresh)
		r.Post("/logout", h.Auth.Logout)
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.NewAuth(jwt))

		r.Get("/accounts", h.Transaction.ListAccounts)
		r.Post("/transactions/import", h.Transaction.Import)
		r.Get("/transactions", h.Transaction.List)

		r.Post("/receipts/upload", h.Receipt.Upload)

		r.Get("/profile", h.Profile.Get)
		r.Get("/quests", h.Quests.List)
	})

	// WebSocket — JWT validated inside handler via ?token= query param
	r.Get("/ws", h.WebSocket.ServeWS)

	// Health check — no auth required
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve embedded React SPA — must be last.
	// embed.FS rejects paths with a leading slash (fs.ValidPath rule),
	// so strip it before checking whether the asset exists.
	if staticFiles != nil {
		fileServer := http.FileServer(http.FS(staticFiles))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			clean := strings.TrimPrefix(r.URL.Path, "/")
			if _, err := staticFiles.Open(clean); err != nil {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	_ = log
	return r
}
