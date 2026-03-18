package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	hillscheck "github.com/hillscheck"
	adapthttp "github.com/hillscheck/internal/adapter/http"
	"github.com/hillscheck/internal/adapter/ai"
	"github.com/hillscheck/internal/adapter/postgres"
	adapws "github.com/hillscheck/internal/adapter/websocket"
	"github.com/hillscheck/internal/infrastructure/auth"
	"github.com/hillscheck/internal/infrastructure/cache"
	"github.com/hillscheck/internal/infrastructure/config"
	"github.com/hillscheck/internal/infrastructure/db"
	"github.com/hillscheck/internal/infrastructure/worker"
	"github.com/hillscheck/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	migrateFlag := flag.String("migrate", "", "run migrations: 'up' or 'down'")
	flag.Parse()

	log, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	cfg := config.Load()

	// ── Infrastructure ────────────────────────────────────────────────────────
	pgPool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("connect postgres", zap.Error(err))
	}
	defer pgPool.Close()

	rdb, err := cache.New(cfg.RedisURL)
	if err != nil {
		log.Fatal("connect redis", zap.Error(err))
	}
	defer rdb.Close()

	pool := worker.New(cfg.WorkerCount, cfg.QueueSize)

	// ── Migrations ────────────────────────────────────────────────────────────
	if err := runMigrations(cfg.DatabaseURL, *migrateFlag, log); err != nil {
		log.Fatal("migrate", zap.Error(err))
	}
	if *migrateFlag != "" {
		return // migration-only mode; exit after running
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo    := postgres.NewUserRepo(pgPool)
	accountRepo := postgres.NewAccountRepo(pgPool)
	txRepo      := postgres.NewTransactionRepo(pgPool)
	gameRepo    := postgres.NewGameRepo(pgPool)

	// ── Category provider ─────────────────────────────────────────────────────
	classifier, err := ai.NewProviderFromConfig(
		cfg.Provider,
		cfg.GeminiAPIKey, cfg.GeminiModel,
		cfg.OllamaURL, cfg.OllamaModel,
		log,
	)
	if err != nil {
		log.Fatal("init category provider", zap.Error(err))
	}

	// ── WebSocket hub ─────────────────────────────────────────────────────────
	hub := adapws.NewHub(log)
	go hub.Run()

	// ── Auth infrastructure ───────────────────────────────────────────────────
	jwtSvc     := auth.NewJWTService(cfg.JWTSecret)
	tokenStore := auth.NewRedisTokenStore(rdb)

	// ── Use cases ─────────────────────────────────────────────────────────────
	engine := usecase.NewGameEngine(txRepo, accountRepo, gameRepo, classifier, hub, log)

	importer := usecase.NewTransactionImport(txRepo, pool, engine, log)

	receiptUploader, err := usecase.NewReceiptUpload("uploads", pool, log)
	if err != nil {
		log.Fatal("init receipt uploader", zap.Error(err))
	}

	getProfile := usecase.NewGetProfile(gameRepo)
	getQuests  := usecase.NewGetQuests(gameRepo, txRepo, accountRepo)

	registerUC := usecase.NewRegisterUser(userRepo, gameRepo)
	loginUC    := usecase.NewLoginUser(userRepo, jwtSvc, tokenStore)
	refreshUC  := usecase.NewRefreshToken(userRepo, jwtSvc, tokenStore)
	logoutUC   := usecase.NewLogout(tokenStore)

	// ── HTTP handlers ─────────────────────────────────────────────────────────
	handlers := adapthttp.Handlers{
		Auth:        adapthttp.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, log),
		Transaction: adapthttp.NewTransactionHandler(importer, txRepo, accountRepo, log),
		Receipt:     adapthttp.NewReceiptHandler(receiptUploader, log),
		Profile:     adapthttp.NewProfileHandler(getProfile, log),
		Quests:      adapthttp.NewQuestHandler(getQuests, log),
		WebSocket:   adapthttp.NewWebSocketHandler(hub, log),
	}

	// ── Embedded static files ─────────────────────────────────────────────────
	distFS, err := fs.Sub(hillscheck.StaticFiles, "web/dist")
	if err != nil {
		log.Fatal("sub static fs", zap.Error(err))
	}

	router := adapthttp.NewRouter(handlers, jwtSvc, rdb, log, distFS)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Start ─────────────────────────────────────────────────────────────────
	log.Info("starting server", zap.String("addr", srv.Addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info("shutting down…")

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("http shutdown", zap.Error(err))
	}
	if err := pool.Shutdown(shutCtx); err != nil {
		log.Error("worker pool shutdown", zap.Error(err))
	}

	log.Info("bye")
}

func runMigrations(databaseURL, direction string, log *zap.Logger) error {
	migrationsFS, err := fs.Sub(hillscheck.StaticFiles, "migrations")
	if err != nil {
		// Migrations FS from embed is optional; fall back to file system.
		log.Warn("no embedded migrations, skipping auto-migrate")
		return nil
	}

	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}

	switch direction {
	case "up", "":
		err = m.Up()
	case "down":
		err = m.Down()
	default:
		return fmt.Errorf("unknown migrate direction: %q", direction)
	}

	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Info("migrations done", zap.String("direction", direction))
	return nil
}
