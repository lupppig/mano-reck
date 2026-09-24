// Package dependency coordinates required runtime dependencies and reports
// their readiness without exposing provider-specific details to HTTP handlers.
package dependency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Component is a required process dependency with an explicit lifecycle and
// readiness check.
type Component interface {
	Name() string
	Start(context.Context) error
	Check(context.Context) error
	Close(context.Context) error
}

// CheckResult is the non-sensitive readiness state exposed to operators.
type CheckResult struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
}

// ReadinessReport is a point-in-time view of required dependencies.
type ReadinessReport struct {
	Ready  bool
	Checks []CheckResult
}

// Set owns ordered startup, reverse-order shutdown, and bounded readiness
// checks for process dependencies.
type Set struct {
	components []Component
	timeout    time.Duration

	mu      sync.RWMutex
	active  bool
	started int
}

// NewSet constructs a dependency set and rejects ambiguous component names.
func NewSet(timeout time.Duration, components ...Component) (*Set, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("dependency readiness timeout must be positive")
	}

	seen := make(map[string]struct{}, len(components))
	for _, component := range components {
		if component == nil {
			return nil, fmt.Errorf("dependency component must not be nil")
		}
		name := component.Name()
		if name == "" {
			return nil, fmt.Errorf("dependency component name must not be empty")
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("dependency component name %q is duplicated", name)
		}
		seen[name] = struct{}{}
	}

	return &Set{components: append([]Component(nil), components...), timeout: timeout}, nil
}

// Start initializes dependencies in declaration order.
func (set *Set) Start(ctx context.Context) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if set.active {
		return fmt.Errorf("dependencies have already started")
	}
	set.active = true

	for index, component := range set.components {
		if err := component.Start(ctx); err != nil {
			return fmt.Errorf("start dependency %s: %w", component.Name(), err)
		}
		set.started = index + 1
	}
	return nil
}

// Close releases started dependencies in reverse order. It attempts every
// close even when an earlier component reports an error.
func (set *Set) Close(ctx context.Context) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	var closeErrors []error
	for index := set.started - 1; index >= 0; index-- {
		component := set.components[index]
		if err := component.Close(ctx); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close dependency %s: %w", component.Name(), err))
		}
	}
	set.started = 0
	set.active = false
	return errors.Join(closeErrors...)
}

// Readiness checks every started dependency within a shared deadline. Error
// details remain internal to avoid exposing connection data through probes.
func (set *Set) Readiness(ctx context.Context) ReadinessReport {
	set.mu.RLock()
	active := set.active
	started := set.started
	components := append([]Component(nil), set.components...)
	set.mu.RUnlock()

	report := ReadinessReport{
		Ready:  active && started == len(components),
		Checks: make([]CheckResult, len(components)),
	}
	if !active || started != len(components) {
		for index, component := range components {
			report.Checks[index] = CheckResult{Name: component.Name(), Ready: index < started}
		}
		return report
	}

	checkContext, cancel := context.WithTimeout(ctx, set.timeout)
	defer cancel()

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(components))
	for index, component := range components {
		go func() {
			defer waitGroup.Done()
			ready := component.Check(checkContext) == nil
			report.Checks[index] = CheckResult{Name: component.Name(), Ready: ready}
		}()
	}
	waitGroup.Wait()

	for _, check := range report.Checks {
		if !check.Ready {
			report.Ready = false
			break
		}
	}
	return report
}
