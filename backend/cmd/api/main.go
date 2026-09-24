package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lupppig/mano-reck/backend/internal/platform/appruntime"
	"github.com/lupppig/mano-reck/backend/internal/platform/config"
	"github.com/lupppig/mano-reck/backend/internal/platform/database"
	"github.com/lupppig/mano-reck/backend/internal/platform/dependency"
	platformhttp "github.com/lupppig/mano-reck/backend/internal/platform/httpserver"
	"github.com/lupppig/mano-reck/backend/internal/platform/logging"
)

func main() {
	configuration, err := config.Load(os.LookupEnv)
	if err != nil {
		bootstrapLogger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		bootstrapLogger.Error("invalid runtime configuration", "error", err)
		os.Exit(1)
	}

	logger := logging.New(os.Stdout, configuration.LogLevel, configuration.Environment)
	databaseConnection, err := database.New(database.Config{
		ConnectionString: configuration.Database.URL.Reveal(),
		MaxConnections:   configuration.Database.MaxConnections,
		MinConnections:   configuration.Database.MinConnections,
	})
	if err != nil {
		logger.Error("configure PostgreSQL", "error", err)
		os.Exit(1)
	}
	dependencies, err := dependency.NewSet(configuration.ReadinessTimeout, databaseConnection)
	if err != nil {
		logger.Error("configure runtime dependencies", "error", err)
		os.Exit(1)
	}
	server := platformhttp.New(configuration.HTTPAddress, logger, dependencies)

	shutdownSignals, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := appruntime.Run(
		shutdownSignals,
		server,
		dependencies,
		configuration.ShutdownTimeout,
		logger,
	); err != nil {
		logger.Error("API process stopped with an error", "error", err)
		os.Exit(1)
	}
}
