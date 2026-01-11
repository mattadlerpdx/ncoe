package middleware

import (
	"context"
	"net/http"

	"ncoe/internal/service"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(as *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: as}
}

// RequireAuth wraps a handler to require authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login page
		if r.URL.Path == "/staff/login" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
			return
		}

		user, err := m.authService.ValidateSession(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
