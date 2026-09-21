package httpserver_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	platformhttp "github.com/lupppig/mano-reck/backend/internal/platform/httpserver"
	"github.com/lupppig/mano-reck/backend/internal/platform/identifier"
)

func TestHealthEndpointReportsProcessHealth(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := platformhttp.New(":0", logger)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	application.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}
	if requestID := response.Header().Get("X-Request-ID"); !identifier.IsUUIDv7(requestID) {
		t.Fatalf("expected a UUIDv7 request ID, got %q", requestID)
	}
	if correlationID := response.Header().Get("X-Correlation-ID"); !identifier.IsUUIDv7(correlationID) {
		t.Fatalf("expected a UUIDv7 correlation ID, got %q", correlationID)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("expected ok health status, got %q", body.Status)
	}
}

func TestHealthEndpointPreservesValidClientIdentifiers(t *testing.T) {
	t.Parallel()

	const requestID = "01991a4b-fa00-7000-8000-000000000001"
	const correlationID = "01991a4b-fa00-7000-8000-000000000002"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	application := platformhttp.New(":0", logger)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", requestID)
	request.Header.Set("X-Correlation-ID", correlationID)
	response := httptest.NewRecorder()

	application.Handler.ServeHTTP(response, request)

	if received := response.Header().Get("X-Request-ID"); received != requestID {
		t.Fatalf("expected request ID %q, got %q", requestID, received)
	}
	if received := response.Header().Get("X-Correlation-ID"); received != correlationID {
		t.Fatalf("expected correlation ID %q, got %q", correlationID, received)
	}
}
