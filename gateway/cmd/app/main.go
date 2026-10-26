package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/godnota/gateway/internal/adapter"
	"github.com/godnota/gateway/internal/config"
	"github.com/godnota/gateway/internal/httpapi"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("read gateway config: %v", err)
	}

	backend, closeBackend, err := adapter.New(
		cfg.Clients.TravelAddr,
		cfg.Clients.SessionAddr,
		cfg.Clients.UserAddr,
		cfg.Clients.OrderAddr,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer closeBackend()

	const headerTimeout = 5 * time.Second
	const readTimeout = 10 * time.Second
	const writeTimeout = 15 * time.Second

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           httpapi.Handler(backend),
		ReadHeaderTimeout: headerTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), headerTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	slog.Info("gateway listening", "port", cfg.HTTP.Port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Print(err)
	}
}
