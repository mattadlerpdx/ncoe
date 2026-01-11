package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// ctxKeyRequestID is the context key for request IDs
type ctxKeyRequestID struct{}

// RequestID middleware adds a unique request ID to each request.
// If X-Request-Id header is present, it uses that value.
// Otherwise, it generates a new unique ID.
// The request ID is:
// - Stored in context (use GetRequestID to retrieve)
// - Added to response header X-Request-Id
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID in header
		requestID := r.Header.Get("X-Request-Id")

		// Generate new ID if not provided
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-Id", requestID)

		// Store in context
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context.
// Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return id
	}
	return ""
}

// generateRequestID creates a unique request ID.
// Format: timestamp_hex + random_hex (e.g., "1703500000_a1b2c3d4")
func generateRequestID() string {
	// Time component (seconds since epoch, base36 for compactness)
	ts := time.Now().Unix()

	// Random component (4 bytes = 8 hex chars)
	randBytes := make([]byte, 4)
	rand.Read(randBytes)

	return hex.EncodeToString([]byte{
		byte(ts >> 24), byte(ts >> 16), byte(ts >> 8), byte(ts),
	}) + hex.EncodeToString(randBytes)
}
