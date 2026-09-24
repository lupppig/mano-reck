package config_test

import (
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
}

func TestLoadAcceptsExplicitConfiguration(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"MANORECK_ENVIRONMENT":                   "sandbox",
		"MANORECK_LOG_LEVEL":                     "debug",
		"MANORECK_HTTP_ADDRESS":                  "127.0.0.1:9090",
		"MANORECK_HTTP_SHUTDOWN_TIMEOUT_SECONDS": "15",
		"MANORECK_READINESS_TIMEOUT_SECONDS":     "3",
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
