// Package requestcontext stores transport identifiers on standard contexts so
// application and infrastructure code can correlate downstream work without a
// dependency on Gin.
package requestcontext

import "context"

type key uint8

const (
	requestIDKey key = iota
	correlationIDKey
)

// WithIdentifiers returns a child context containing request and correlation
// identifiers.
func WithIdentifiers(parent context.Context, requestID string, correlationID string) context.Context {
	withRequestID := context.WithValue(parent, requestIDKey, requestID)
	return context.WithValue(withRequestID, correlationIDKey, correlationID)
}

// RequestID returns the request identifier, or an empty string when none is
// associated with the context.
func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

// CorrelationID returns the correlation identifier, or an empty string when
// none is associated with the context.
func CorrelationID(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDKey).(string)
	return value
}
