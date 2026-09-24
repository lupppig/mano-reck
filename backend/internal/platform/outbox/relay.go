package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lupppig/mano-reck/backend/internal/platform/identifier"
)

// RelayRepository is the persistence surface consumed by the relay worker.
type RelayRepository interface {
	Claim(context.Context, string, int, time.Duration) ([]Record, error)
	MarkPublished(context.Context, string, string) error
	MarkFailed(context.Context, string, string, FailureClass, int, time.Time, string) error
}

// Publisher sends one stable event to the asynchronous transport.
type Publisher interface {
	Publish(context.Context, string, string, []byte) error
}

// RelayConfig bounds polling, leasing, batching, and retry behavior.
type RelayConfig struct {
	PollInterval  time.Duration
	LeaseDuration time.Duration
	BatchSize     int
	MaxAttempts   int
	BaseBackoff   time.Duration
	MaxBackoff    time.Duration
}

// Relay publishes committed outbox rows without holding a database
// transaction across the NATS call.
type Relay struct {
	repository RelayRepository
	publisher  Publisher
	config     RelayConfig
	workerID   string

	mu        sync.RWMutex
	cancel    context.CancelFunc
	done      chan struct{}
	lastError error
}

// NewRelay validates worker policy and allocates a stable lease owner ID.
func NewRelay(repository RelayRepository, publisher Publisher, configuration RelayConfig) (*Relay, error) {
	if repository == nil || publisher == nil {
		return nil, fmt.Errorf("outbox relay repository and publisher are required")
	}
	if configuration.PollInterval <= 0 || configuration.LeaseDuration <= 0 {
		return nil, fmt.Errorf("outbox relay intervals must be positive")
	}
	if configuration.BatchSize < 1 || configuration.BatchSize > 1000 {
		return nil, fmt.Errorf("outbox relay batch size must be between 1 and 1000")
	}
	if configuration.MaxAttempts < 1 {
		return nil, fmt.Errorf("outbox relay maximum attempts must be positive")
	}
	if configuration.BaseBackoff <= 0 || configuration.MaxBackoff < configuration.BaseBackoff {
		return nil, fmt.Errorf("outbox relay backoff policy is invalid")
	}
	workerID, err := identifier.NewUUIDv7()
	if err != nil {
		return nil, fmt.Errorf("create outbox worker ID: %w", err)
	}
	return &Relay{
		repository: repository,
		publisher:  publisher,
		config:     configuration,
		workerID:   workerID,
	}, nil
}

// Name identifies the worker in process readiness.
func (*Relay) Name() string {
	return "outbox-relay"
}

// Start begins polling after the database and NATS dependencies are ready.
func (relay *Relay) Start(parent context.Context) error {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	if relay.cancel != nil {
		return fmt.Errorf("outbox relay has already started")
	}
	ctx, cancel := context.WithCancel(parent)
	relay.cancel = cancel
	relay.done = make(chan struct{})
	go relay.run(ctx, relay.done)
	return nil
}

// Check reports worker-loop failures while leaving transport and database
// connectivity to their owning dependency checks.
func (relay *Relay) Check(context.Context) error {
	relay.mu.RLock()
	defer relay.mu.RUnlock()
	if relay.cancel == nil {
		return fmt.Errorf("outbox relay is not started")
	}
	return relay.lastError
}

// Close stops polling and waits for the active batch to finish or for the
// shutdown deadline.
func (relay *Relay) Close(ctx context.Context) error {
	relay.mu.Lock()
	cancel := relay.cancel
	done := relay.done
	relay.cancel = nil
	relay.done = nil
	relay.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stop outbox relay: %w", ctx.Err())
	}
}

func (relay *Relay) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(relay.config.PollInterval)
	defer ticker.Stop()

	for {
		err := relay.processBatch(ctx)
		relay.mu.Lock()
		relay.lastError = err
		relay.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (relay *Relay) processBatch(ctx context.Context) error {
	records, err := relay.repository.Claim(
		ctx,
		relay.workerID,
		relay.config.BatchSize,
		relay.config.LeaseDuration,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}

	for _, record := range records {
		payload, err := json.Marshal(record.Envelope)
		if err != nil {
			if failureErr := relay.repository.MarkFailed(
				ctx,
				record.Envelope.ID,
				relay.workerID,
				FailurePermanent,
				relay.config.MaxAttempts,
				time.Now().UTC(),
				"event_encoding_failed",
			); failureErr != nil {
				return errors.Join(err, failureErr)
			}
			continue
		}

		if err := relay.publisher.Publish(ctx, record.Subject, record.Envelope.ID, payload); err != nil {
			nextAttempt := time.Now().UTC().Add(relay.backoff(record.Attempt))
			if failureErr := relay.repository.MarkFailed(
				ctx,
				record.Envelope.ID,
				relay.workerID,
				FailureTransient,
				relay.config.MaxAttempts,
				nextAttempt,
				"nats_publish_failed",
			); failureErr != nil {
				return errors.Join(err, failureErr)
			}
			continue
		}

		if err := relay.repository.MarkPublished(ctx, record.Envelope.ID, relay.workerID); err != nil {
			return err
		}
	}
	return nil
}

func (relay *Relay) backoff(attempt int) time.Duration {
	backoff := relay.config.BaseBackoff
	for current := 1; current < attempt && backoff < relay.config.MaxBackoff; current++ {
		if backoff > relay.config.MaxBackoff/2 {
			return relay.config.MaxBackoff
		}
		backoff *= 2
	}
	if backoff > relay.config.MaxBackoff {
		return relay.config.MaxBackoff
	}
	return backoff
}
