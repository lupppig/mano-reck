package identifier

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestNewUUIDv7UsesTimestampVersionAndVariant(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	generated, err := newUUIDv7(now, bytes.NewReader(make([]byte, 16)))
	if err != nil {
		t.Fatalf("generate UUIDv7: %v", err)
	}

	if generated != "01a0c3d6-5a00-7000-8000-000000000000" {
		t.Fatalf("unexpected UUIDv7: %s", generated)
	}
	if !IsUUIDv7(generated) {
		t.Fatalf("generated value is not a canonical UUIDv7: %s", generated)
	}
}

func TestNewUUIDv7ReturnsRandomnessFailure(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("random source unavailable")
	_, err := newUUIDv7(time.UnixMilli(0), failingReader{err: expectedError})
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected randomness error, got %v", err)
	}
}

func TestIsUUIDv7RejectsNonCanonicalValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty":           "",
		"uuid version 4":  "01991a4b-fa00-4000-8000-000000000000",
		"invalid variant": "01991a4b-fa00-7000-7000-000000000000",
		"uppercase":       "01991A4B-FA00-7000-8000-000000000000",
		"missing hyphen":  "01991a4bfa00-7000-8000-000000000000",
	}

	for name, value := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if IsUUIDv7(value) {
				t.Fatalf("expected %q to be rejected", value)
			}
		})
	}
}

type failingReader struct {
	err error
}

func (reader failingReader) Read(_ []byte) (int, error) {
	return 0, reader.err
}
