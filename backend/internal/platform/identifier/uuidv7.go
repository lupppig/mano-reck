// Package identifier creates and validates the platform's opaque identifiers.
package identifier

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"
)

// NewUUIDv7 returns a canonical lowercase UUIDv7 using the current UTC time
// and cryptographically secure randomness.
func NewUUIDv7() (string, error) {
	return newUUIDv7(time.Now(), rand.Reader)
}

func newUUIDv7(now time.Time, randomness io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(randomness, value[:]); err != nil {
		return "", err
	}

	timestamp := uint64(now.UnixMilli())
	value[0] = byte(timestamp >> 40)
	value[1] = byte(timestamp >> 32)
	value[2] = byte(timestamp >> 24)
	value[3] = byte(timestamp >> 16)
	value[4] = byte(timestamp >> 8)
	value[5] = byte(timestamp)
	value[6] = (value[6] & 0x0f) | 0x70
	value[8] = (value[8] & 0x3f) | 0x80

	var encoded [36]byte
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])

	return string(encoded[:]), nil
}

// IsUUIDv7 reports whether value is a canonical lowercase UUIDv7.
func IsUUIDv7(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	if value[14] != '7' || (value[19] != '8' && value[19] != '9' && value[19] != 'a' && value[19] != 'b') {
		return false
	}

	for index, character := range []byte(value) {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}

	return true
}
