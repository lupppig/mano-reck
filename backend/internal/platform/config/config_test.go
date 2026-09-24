package config_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/lupppig/mano-reck/backend/internal/platform/config"
)

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	t.Parallel()

	loaded, err := config.Load(emptyEnvironment)
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	if loaded.Environment != "local" {
		t.Fatalf("expected local environment, got %q", loaded.Environment)
	}
	if loaded.LogLevel != slog.LevelInfo {
		t.Fatalf("expected info log level, got %s", loaded.LogLevel)
	}
	if loaded.HTTPAddress != ":8080" {
		t.Fatalf("expected :8080 HTTP address, got %q", loaded.HTTPAddress)
	}
	if loaded.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected 10 second shutdown timeout, got %s", loaded.ShutdownTimeout)
	}
	if loaded.ReadinessTimeout != 2*time.Second {
		t.Fatalf("expected 2 second readiness timeout, got %s", loaded.ReadinessTimeout)
	}
	if loaded.Database.MaxConnections != 10 || loaded.Database.MinConnections != 1 {
		t.Fatalf("unexpected database defaults: %#v", loaded.Database)
	}
	if loaded.Database.URL.Reveal() == "" {
		t.Fatal("expected default database URL")
	}
	if loaded.NATS.URL.Reveal() != "nats://127.0.0.1:4222" {
		t.Fatalf("unexpected default NATS URL")
	}
	if loaded.Outbox.BatchSize != 100 || loaded.Outbox.MaxAttempts != 8 {
		t.Fatalf("unexpected default outbox configuration: %#v", loaded.Outbox)
	}
}

func TestLoadAcceptsExplicitConfiguration(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"MANORECK_ENVIRONMENT":                   "sandbox",
		"MANORECK_LOG_LEVEL":                     "debug",
		"MANORECK_HTTP_ADDRESS":                  "127.0.0.1:9090",
		"MANORECK_HTTP_SHUTDOWN_TIMEOUT_SECONDS": "15",
		"MANORECK_READINESS_TIMEOUT_SECONDS":     "3",
		"MANORECK_DATABASE_URL":                  "postgres://user:secret@database:5432/testdb?sslmode=disable",
		"MANORECK_DATABASE_MAX_CONNECTIONS":      "20",
		"MANORECK_DATABASE_MIN_CONNECTIONS":      "2",
		"MANORECK_NATS_URL":                      "tls://user:secret@nats.example:4222",
		"MANORECK_OUTBOX_POLL_MILLISECONDS":      "100",
		"MANORECK_OUTBOX_LEASE_SECONDS":          "20",
		"MANORECK_OUTBOX_BATCH_SIZE":             "25",
		"MANORECK_OUTBOX_MAX_ATTEMPTS":           "6",
		"MANORECK_OUTBOX_BASE_BACKOFF_SECONDS":   "2",
		"MANORECK_OUTBOX_MAX_BACKOFF_SECONDS":    "120",
	}
	loaded, err := config.Load(mapEnvironment(values))
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	if loaded.Environment != "sandbox" || loaded.LogLevel != slog.LevelDebug {
		t.Fatalf("unexpected identity configuration: %#v", loaded)
	}
	if loaded.HTTPAddress != "127.0.0.1:9090" {
		t.Fatalf("unexpected HTTP address %q", loaded.HTTPAddress)
	}
	if loaded.ShutdownTimeout != 15*time.Second || loaded.ReadinessTimeout != 3*time.Second {
		t.Fatalf("unexpected timeouts: %#v", loaded)
	}
	if loaded.Database.MaxConnections != 20 || loaded.Database.MinConnections != 2 {
		t.Fatalf("unexpected database configuration: %#v", loaded.Database)
	}
	if strings.Contains(loaded.Database.URL.String(), "secret") || loaded.Database.URL.String() != "<redacted>" {
		t.Fatalf("database URL was not redacted: %s", loaded.Database.URL)
	}
	if rendered := fmt.Sprintf("%#v", loaded); strings.Contains(rendered, "secret") {
		t.Fatalf("formatted configuration exposed the database password: %s", rendered)
	}
	if loaded.NATS.URL.String() != "<redacted>" {
		t.Fatalf("NATS URL was not redacted: %s", loaded.NATS.URL)
	}
	if loaded.Outbox.PollInterval != 100*time.Millisecond || loaded.Outbox.MaxBackoff != 120*time.Second {
		t.Fatalf("unexpected outbox configuration: %#v", loaded.Outbox)
	}
}

func TestLoadRejectsInvalidValuesWithVariableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variable string
		value    string
	}{
		{name: "environment", variable: "MANORECK_ENVIRONMENT", value: "production"},
		{name: "log level", variable: "MANORECK_LOG_LEVEL", value: "verbose"},
		{name: "address", variable: "MANORECK_HTTP_ADDRESS", value: "localhost"},
		{name: "shutdown timeout", variable: "MANORECK_HTTP_SHUTDOWN_TIMEOUT_SECONDS", value: "0"},
		{name: "readiness timeout", variable: "MANORECK_READINESS_TIMEOUT_SECONDS", value: "many"},
		{name: "database URL", variable: "MANORECK_DATABASE_URL", value: "mysql://database/test"},
		{name: "database maximum", variable: "MANORECK_DATABASE_MAX_CONNECTIONS", value: "101"},
		{name: "database minimum", variable: "MANORECK_DATABASE_MIN_CONNECTIONS", value: "11"},
		{name: "NATS URL", variable: "MANORECK_NATS_URL", value: "http://nats.example"},
		{name: "outbox poll", variable: "MANORECK_OUTBOX_POLL_MILLISECONDS", value: "0"},
		{name: "outbox attempts", variable: "MANORECK_OUTBOX_MAX_ATTEMPTS", value: "101"},
		{name: "outbox backoff", variable: "MANORECK_OUTBOX_MAX_BACKOFF_SECONDS", value: "0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := config.Load(mapEnvironment(map[string]string{test.variable: test.value}))
			if err == nil {
				t.Fatal("expected configuration error")
			}
			if !strings.Contains(err.Error(), test.variable) {
				t.Fatalf("expected error to name %s, got %q", test.variable, err)
			}
		})
	}
}

func emptyEnvironment(string) (string, bool) {
	return "", false
}

func mapEnvironment(values map[string]string) config.LookupEnv {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
