package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	platformhttp "github.com/lupppig/mano-reck/backend/internal/platform/httpserver"
)

const shutdownTimeout = 10 * time.Second

func main() {
	address := os.Getenv("MANORECK_HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := platformhttp.New(address, logger)

	shutdownSignals, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "address", address)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-shutdownSignals.Done():
		logger.Info("shutdown requested")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("http server shutdown failed", "error", err)
		os.Exit(1)
	}
}
