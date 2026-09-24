package logging_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/lupppig/mano-reck/backend/internal/platform/logging"
)

func TestNewAppliesLevelAndRuntimeAttributes(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := logging.New(&output, slog.LevelInfo, "test")
	logger.Debug("hidden")
	logger.Info("visible")

	entry := output.String()
	if strings.Contains(entry, "hidden") {
		t.Fatal("expected debug entry to be filtered")
	}
	for _, expected := range []string{
		`"msg":"visible"`,
		`"service":"manoreck-api"`,
		`"environment":"test"`,
	} {
		if !strings.Contains(entry, expected) {
			t.Fatalf("expected log entry to contain %s, got %s", expected, entry)
		}
	}
}
