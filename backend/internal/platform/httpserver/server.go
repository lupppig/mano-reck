package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/lupppig/mano-reck/backend/internal/platform/identifier"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

// New returns the minimal HTTP server used to prove the backend build and
// runtime. Domain routes are added by their owning modules in later phases.
func New(address string, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)

	return &http.Server{
		Addr:              address,
		Handler:           requestIdentifiers(logRequests(logger, mux)),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func requestIdentifiers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get("X-Request-ID")
		if !identifier.IsUUIDv7(requestID) {
			var err error
			requestID, err = identifier.NewUUIDv7()
			if err != nil {
				http.Error(response, "could not create request identifier", http.StatusInternalServerError)
				return
			}
		}

		correlationID := request.Header.Get("X-Correlation-ID")
		if !identifier.IsUUIDv7(correlationID) {
			correlationID = requestID
		}

		response.Header().Set("X-Request-ID", requestID)
		response.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(response, request)
	})
}

func health(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(response).Encode(map[string]string{"status": "ok"})
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(response, request)
		logger.InfoContext(
			request.Context(),
			"http request completed",
			"method", request.Method,
			"path", request.URL.Path,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}
