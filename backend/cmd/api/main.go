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
	"github.com/lupppig/mano-reck/backend/internal/platform/eventconsumer"
	platformhttp "github.com/lupppig/mano-reck/backend/internal/platform/httpserver"
	"github.com/lupppig/mano-reck/backend/internal/platform/logging"
	"github.com/lupppig/mano-reck/backend/internal/platform/natsclient"
	"github.com/lupppig/mano-reck/backend/internal/platform/objectstorage"
	"github.com/lupppig/mano-reck/backend/internal/platform/outbox"
	"github.com/lupppig/mano-reck/backend/internal/platform/redisclient"
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
	natsConnection, err := natsclient.New(configuration.NATS.URL.Reveal())
	if err != nil {
		logger.Error("configure NATS", "error", err)
		os.Exit(1)
	}
	redisConnection, err := redisclient.New(redisclient.Config{
		URL:       configuration.Redis.URL.Reveal(),
		KeyPrefix: configuration.Redis.KeyPrefix,
	})
	if err != nil {
		logger.Error("configure Redis", "error", err)
		os.Exit(1)
	}
	objectStore, err := objectstorage.New(objectstorage.Config{
		Endpoint:  configuration.ObjectStorage.Endpoint,
		Bucket:    configuration.ObjectStorage.Bucket,
		AccessKey: configuration.ObjectStorage.AccessKey.Reveal(),
		SecretKey: configuration.ObjectStorage.SecretKey.Reveal(),
	})
	if err != nil {
		logger.Error("configure object storage", "error", err)
		os.Exit(1)
	}
	outboxRepository := outbox.NewRepository(databaseConnection)
	outboxRelay, err := outbox.NewRelay(outboxRepository, natsConnection, outbox.RelayConfig{
		PollInterval:  configuration.Outbox.PollInterval,
		LeaseDuration: configuration.Outbox.LeaseDuration,
		BatchSize:     configuration.Outbox.BatchSize,
		MaxAttempts:   configuration.Outbox.MaxAttempts,
		BaseBackoff:   configuration.Outbox.BaseBackoff,
		MaxBackoff:    configuration.Outbox.MaxBackoff,
	})
	if err != nil {
		logger.Error("configure outbox relay", "error", err)
		os.Exit(1)
	}
	proofConsumer, err := eventconsumer.New(databaseConnection, natsConnection)
	if err != nil {
		logger.Error("configure infrastructure event consumer", "error", err)
		os.Exit(1)
	}
	dependencies, err := dependency.NewSet(
		configuration.ReadinessTimeout,
		databaseConnection,
		natsConnection,
		redisConnection,
		objectStore,
		proofConsumer,
		outboxRelay,
	)
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
