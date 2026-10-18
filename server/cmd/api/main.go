package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/db"
	"github.com/mddanish004/jamzy/server/internal/httpapi"
	"github.com/mddanish004/jamzy/server/internal/httpx"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/messages"
	"github.com/mddanish004/jamzy/server/internal/music"
	"github.com/mddanish004/jamzy/server/internal/playback"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
	"github.com/mddanish004/jamzy/server/internal/realtime"
	"github.com/mddanish004/jamzy/server/internal/rooms"
	"github.com/mddanish004/jamzy/server/internal/users"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

const (
	shutdownTimeout = 10 * time.Second
	maxConnTries    = 5
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	log := logging.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.DatabaseURL == "" {
		return errors.New("DATABASE_URL is not configured")
	}
	if cfg.ClerkSecretKey == "" {
		log.Warn("CLERK_SECRET_KEY is not configured — every authenticated request will be rejected")
	}
	if cfg.YouTubeAPIKey == "" {
		log.Warn("YOUTUBE_API_KEY is not configured — music endpoints will return 503")
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, log); err != nil {
		return err
	}

	limiter := ratelimit.NewLimiter()
	defer limiter.Close()

	verifier := auth.NewClerkVerifier(cfg.ClerkSecretKey, cfg.ClerkAuthorizedParties)

	usersService := users.NewService(pool)
	roomsService := rooms.NewService(pool, usersService)

	hub := realtime.NewHub(
		roomsService,
		playback.NewService(pool),
		messages.NewService(pool),
		usersService,
		verifier,
		limiter,
		log,
	)

	socketOptions := socket.DefaultServerOptions()
	socketOptions.SetCors(&types.Cors{Origin: originList(cfg.CORSOrigins), Credentials: true})
	socketOptions.SetMaxHttpBufferSize(100_000)

	server := hub.Server(socketOptions)

	handler := httpapi.New(httpapi.Deps{
		Middleware: &httpx.Middleware{
			Log:         log,
			Verifier:    verifier,
			Limiter:     limiter,
			CORSOrigins: cfg.CORSOrigins,
		},
		Users:  &users.Handlers{Service: usersService, Log: log},
		Rooms:  &rooms.Handlers{Service: roomsService, Hooks: hub, Log: log},
		Music:  &music.Handlers{Service: music.NewService(cfg.YouTubeAPIKey, log)},
		Socket: server.ServeHandler(nil),
		Log:    log,
	})

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	failures := make(chan error, 1)

	go func() {
		log.Info(fmt.Sprintf("Server running on 0.0.0.0:%d", cfg.Port))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failures <- err
		}
	}()

	select {
	case err := <-failures:
		return err
	case <-ctx.Done():
		log.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	server.Close(nil)

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func originList(origins []string) []any {
	values := make([]any, 0, len(origins))
	for _, origin := range origins {
		values = append(values, origin)
	}
	return values
}
