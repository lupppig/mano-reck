// Package redisclient owns the process-wide Redis connection used only for
// bounded ephemeral state.
package redisclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound reports an absent ephemeral key without exposing the driver.
var ErrNotFound = errors.New("ephemeral value not found")

// Config defines the connection and isolated key namespace.
type Config struct {
	URL       string
	KeyPrefix string
}

// Client provides a deliberately small expiring key-value surface.
type Client struct {
	options   *redis.Options
	keyPrefix string

	mu     sync.RWMutex
	client *redis.Client
}

// New validates Redis connection settings without opening a connection.
func New(configuration Config) (*Client, error) {
	options, err := redis.ParseURL(configuration.URL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis connection configuration")
	}
	prefix := strings.TrimSpace(configuration.KeyPrefix)
	if prefix == "" || prefix != configuration.KeyPrefix || strings.Contains(prefix, ":") {
		return nil, fmt.Errorf("Redis key prefix must be non-empty and must not contain whitespace or colons")
	}
	return &Client{options: options, keyPrefix: prefix}, nil
}

// Name identifies this dependency in readiness output.
func (*Client) Name() string {
	return "redis"
}

// Start constructs the connection pool. Connectivity remains a readiness
// concern so an ordinary outage does not turn into a process restart loop.
func (client *Client) Start(context.Context) error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.client != nil {
		return fmt.Errorf("Redis client has already started")
	}
	client.client = redis.NewClient(client.options)
	return nil
}

// Check verifies Redis connectivity without mutating application state.
func (client *Client) Check(ctx context.Context) error {
	started, err := client.startedClient()
	if err != nil {
		return err
	}
	if err := started.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping Redis: %w", err)
	}
	return nil
}

// Close releases pooled Redis connections.
func (client *Client) Close(context.Context) error {
	client.mu.Lock()
	started := client.client
	client.client = nil
	client.mu.Unlock()
	if started == nil {
		return nil
	}
	if err := started.Close(); err != nil {
		return fmt.Errorf("close Redis client: %w", err)
	}
	return nil
}

// Put stores an isolated ephemeral value. A positive expiry is mandatory so
// Redis cannot accidentally become authoritative state.
func (client *Client) Put(ctx context.Context, key string, value []byte, expiry time.Duration) error {
	if expiry <= 0 {
		return fmt.Errorf("ephemeral value expiry must be positive")
	}
	qualified, err := client.qualifiedKey(key)
	if err != nil {
		return err
	}
	started, err := client.startedClient()
	if err != nil {
		return err
	}
	if err := started.Set(ctx, qualified, value, expiry).Err(); err != nil {
		return fmt.Errorf("put ephemeral value: %w", err)
	}
	return nil
}

// Get retrieves one isolated ephemeral value.
func (client *Client) Get(ctx context.Context, key string) ([]byte, error) {
	qualified, err := client.qualifiedKey(key)
	if err != nil {
		return nil, err
	}
	started, err := client.startedClient()
	if err != nil {
		return nil, err
	}
	value, err := started.Get(ctx, qualified).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ephemeral value: %w", err)
	}
	return value, nil
}

// Delete removes one isolated ephemeral value.
func (client *Client) Delete(ctx context.Context, key string) error {
	qualified, err := client.qualifiedKey(key)
	if err != nil {
		return err
	}
	started, err := client.startedClient()
	if err != nil {
		return err
	}
	if err := started.Del(ctx, qualified).Err(); err != nil {
		return fmt.Errorf("delete ephemeral value: %w", err)
	}
	return nil
}

func (client *Client) startedClient() (*redis.Client, error) {
	client.mu.RLock()
	defer client.mu.RUnlock()
	if client.client == nil {
		return nil, fmt.Errorf("Redis client is not started")
	}
	return client.client, nil
}

func (client *Client) qualifiedKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" || trimmed != key || len(key) > 128 {
		return "", fmt.Errorf("ephemeral key must be 1-128 characters without surrounding whitespace")
	}
	return client.keyPrefix + ":" + key, nil
}
