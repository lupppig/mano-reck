// Package appruntime owns API process startup and shutdown ordering.
package appruntime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Dependencies is the lifecycle boundary implemented by the required runtime
// dependency set.
type Dependencies interface {
	Start(context.Context) error
	Close(context.Context) error
}

// Run starts dependencies before accepting HTTP work and always attempts
// graceful server and dependency shutdown.
func Run(
	ctx context.Context,
	server *http.Server,
	dependencies Dependencies,
	shutdownTimeout time.Duration,
	logger *slog.Logger,
) error {
	if err := dependencies.Start(ctx); err != nil {
		return closeAfterStartupFailure(dependencies, shutdownTimeout, err)
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return closeAfterStartupFailure(
			dependencies,
			shutdownTimeout,
			fmt.Errorf("listen on %s: %w", server.Addr, err),
		)
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server started", "address", listener.Addr().String())
		serverErrors <- server.Serve(listener)
	}()

	var runError error
	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			runError = fmt.Errorf("serve HTTP: %w", err)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	serverError := server.Shutdown(shutdownContext)
	if serverError != nil {
		serverError = fmt.Errorf("shutdown HTTP server: %w", serverError)
	}
	dependencyError := dependencies.Close(shutdownContext)

	return errors.Join(runError, serverError, dependencyError)
}

func closeAfterStartupFailure(
	dependencies Dependencies,
	shutdownTimeout time.Duration,
	startupError error,
) error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return errors.Join(startupError, dependencies.Close(shutdownContext))
}
