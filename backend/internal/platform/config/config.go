// Package config loads and validates process configuration at startup.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
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
	defaultDatabaseURL             = "postgres://manoreck:manoreck_local@127.0.0.1:5432/manoreck?sslmode=disable"
	defaultDatabaseMaxConnections  = 10
	defaultDatabaseMinConnections  = 1
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
	Database         DatabaseConfig
}

// DatabaseConfig contains the PostgreSQL connection and pool policy.
type DatabaseConfig struct {
	URL            Secret
	MaxConnections int32
	MinConnections int32
}

// Secret prevents credentials from being exposed by ordinary string and Go
// formatting. Reveal is reserved for the adapter that consumes the value.
type Secret struct {
	value string
}

// Reveal returns the sensitive configuration value to its owning adapter.
func (secret Secret) Reveal() string {
	return secret.value
}

func (Secret) String() string {
	return "<redacted>"
}

func (Secret) GoString() string {
	return "<redacted>"
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

	databaseURL := valueOrDefault(lookup, "MANORECK_DATABASE_URL", defaultDatabaseURL)
	if err := validateDatabaseURL(databaseURL); err != nil {
		return Config{}, fmt.Errorf("MANORECK_DATABASE_URL is invalid: %w", err)
	}
	databaseMaxConnections, err := integer(
		lookup,
		"MANORECK_DATABASE_MAX_CONNECTIONS",
		defaultDatabaseMaxConnections,
		1,
		100,
	)
	if err != nil {
		return Config{}, err
	}
	databaseMinConnections, err := integer(
		lookup,
		"MANORECK_DATABASE_MIN_CONNECTIONS",
		defaultDatabaseMinConnections,
		0,
		databaseMaxConnections,
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
		Database: DatabaseConfig{
			URL:            Secret{value: databaseURL},
			MaxConnections: int32(databaseMaxConnections),
			MinConnections: int32(databaseMinConnections),
		},
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
	parsed, err := integer(lookup, name, fallback, 1, 60)
	if err != nil {
		return 0, err
	}
	return time.Duration(parsed) * time.Second, nil
}

func integer(lookup LookupEnv, name string, fallback int, minimum int, maximum int) (int, error) {
	value := strconv.Itoa(fallback)
	if configured, ok := lookup(name); ok {
		value = strings.TrimSpace(configured)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf(
			"%s must be an integer from %d to %d, got %q",
			name,
			minimum,
			maximum,
			value,
		)
	}
	return parsed, nil
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

func validateDatabaseURL(databaseURL string) error {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("must be a PostgreSQL URL")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("scheme must be postgres or postgresql")
	}
	if parsed.Host == "" || parsed.Path == "" || parsed.Path == "/" {
		return fmt.Errorf("host and database name are required")
	}
	return nil
}
