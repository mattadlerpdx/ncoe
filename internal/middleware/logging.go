package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging middleware logs all requests with structured output.
// Format: REQ=request_id METHOD path STATUS duration [user_email]
// Requires RequestID middleware to run first in the chain.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(wrapped, r)

		// Get request ID from context (set by RequestID middleware)
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			requestID = "-"
		}

		// Get user email from context if available
		userEmail := "-"
		if u := r.Context().Value("user"); u != nil {
			if user, ok := u.(interface{ GetEmail() string }); ok {
				userEmail = user.GetEmail()
			}
		}

		// Log in structured format
		log.Printf("REQ=%s %s %s %d %.1fms %s",
			requestID,
			r.Method,
			r.URL.Path,
			wrapped.status,
			float64(time.Since(start).Microseconds())/1000.0,
			userEmail,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Recovery middleware recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
