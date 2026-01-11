package handler

import (
	"net/http"

	"ncoe/internal/config"
	"ncoe/internal/service"
	"ncoe/internal/templates"
)

type AuthHandler struct {
	authService *service.AuthService
	tmpl        *templates.Renderer
	branding    config.Branding
}

func NewAuthHandler(as *service.AuthService, tmpl *templates.Renderer, b config.Branding) *AuthHandler {
	return &AuthHandler{
		authService: as,
		tmpl:        tmpl,
		branding:    b,
	}
}

// StaffLogin handles staff login
func (h *AuthHandler) StaffLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleStaffLogin(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "Staff Login",
		"Branding": h.branding,
	}

	h.render(w, "auth/staff_login", data)
}

func (h *AuthHandler) handleStaffLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	session, err := h.authService.LoginStaff(email, password)
	if err != nil {
		data := map[string]interface{}{
			"Title":    "Staff Login",
			"Branding": h.branding,
			"Error":    "Invalid credentials",
		}
		h.render(w, "auth/staff_login", data)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/staff/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/staff/login", http.StatusSeeOther)
}

func (h *AuthHandler) render(w http.ResponseWriter, name string, data interface{}) {
	err := h.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
