// Package config loads and validates process configuration at startup.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEnvironment             = "local"
	defaultLogLevel                = "info"
	defaultHTTPAddress             = ":8080"
	defaultShutdownTimeoutSeconds  = 10
	defaultReadinessTimeoutSeconds = 2
)

// LookupEnv matches os.LookupEnv and makes configuration loading deterministic
// in tests.
type LookupEnv func(string) (string, bool)

// Config contains the validated values consumed by the API process.
type Config struct {
	Environment      string
	LogLevel         slog.Level
	HTTPAddress      string
	ShutdownTimeout  time.Duration
	ReadinessTimeout time.Duration
}

// Load reads every runtime value once and rejects malformed configuration.
func Load(lookup LookupEnv) (Config, error) {
	environment := valueOrDefault(lookup, "MANORECK_ENVIRONMENT", defaultEnvironment)
	switch environment {
	case "local", "test", "sandbox":
	default:
		return Config{}, fmt.Errorf(
			"MANORECK_ENVIRONMENT must be one of local, test, or sandbox, got %q",
			environment,
		)
	}

	logLevelValue := valueOrDefault(lookup, "MANORECK_LOG_LEVEL", defaultLogLevel)
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(logLevelValue)); err != nil {
		return Config{}, fmt.Errorf("MANORECK_LOG_LEVEL %q is invalid: %w", logLevelValue, err)
	}

	httpAddress := valueOrDefault(lookup, "MANORECK_HTTP_ADDRESS", defaultHTTPAddress)
	if err := validateAddress(httpAddress); err != nil {
		return Config{}, fmt.Errorf("MANORECK_HTTP_ADDRESS %q is invalid: %w", httpAddress, err)
	}

	shutdownTimeout, err := seconds(
		lookup,
		"MANORECK_HTTP_SHUTDOWN_TIMEOUT_SECONDS",
		defaultShutdownTimeoutSeconds,
	)
	if err != nil {
		return Config{}, err
	}

	readinessTimeout, err := seconds(
		lookup,
		"MANORECK_READINESS_TIMEOUT_SECONDS",
		defaultReadinessTimeoutSeconds,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:      environment,
		LogLevel:         logLevel,
		HTTPAddress:      httpAddress,
		ShutdownTimeout:  shutdownTimeout,
		ReadinessTimeout: readinessTimeout,
	}, nil
}

func valueOrDefault(lookup LookupEnv, name string, fallback string) string {
	value, ok := lookup(name)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(value)
}

func seconds(lookup LookupEnv, name string, fallback int) (time.Duration, error) {
	value := strconv.Itoa(fallback)
	if configured, ok := lookup(name); ok {
		value = strings.TrimSpace(configured)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 60 {
		return 0, fmt.Errorf("%s must be an integer from 1 to 60, got %q", name, value)
	}
	return time.Duration(parsed) * time.Second, nil
}

func validateAddress(address string) error {
	if strings.TrimSpace(address) != address || address == "" {
		return fmt.Errorf("must be a non-empty host:port value without surrounding whitespace")
	}

	_, portValue, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("must be a host:port value: %w", err)
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("port must be an integer from 1 to 65535")
	}
	return nil
}
