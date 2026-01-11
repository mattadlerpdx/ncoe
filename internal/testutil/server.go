// Package testutil provides test utilities for integration testing the NCOE web app.
package testutil

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"ncoe/internal/config"
	"ncoe/internal/handler"
	"ncoe/internal/middleware"
	"ncoe/internal/repository/mock"
	"ncoe/internal/service"
	"ncoe/internal/templates"
)

// TestServer provides an httptest.Server configured with the full app stack.
// It exposes Repos for data validation in tests.
type TestServer struct {
	*httptest.Server
	Repos  *mock.Repositories
	Client *http.Client
	t      *testing.T
}

// projectRoot returns the absolute path to the ncoe project root.
// It uses runtime.Caller to find the path relative to this source file.
func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("testutil: unable to determine project root")
	}
	// This file is at internal/testutil/server.go, so go up 2 levels
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// NewTestServer creates a fully configured test server with mock repositories.
// Templates are loaded using an absolute path, so tests work from any directory.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	root := projectRoot()

	// Initialize mock repositories
	repos := mock.NewRepositories()

	// Initialize services
	authService := service.NewAuthService(repos.User, repos.Session)
	caseService := service.NewCaseService(repos.Case)
	dashboardService := service.NewDashboardService(repos.Case)

	// Load templates from absolute path (quiet mode for tests)
	templateDir := filepath.Join(root, "templates")
	tmpl := templates.NewQuietRenderer(templateDir)

	// Default test branding
	branding := config.Branding{
		AgencyName:   "Test Ethics Commission",
		ShortName:    "TEC",
		Tagline:      "Test Tagline",
		PrimaryColor: "#003366",
		ContactEmail: "test@test.gov",
		ContactPhone: "(555) 555-5555",
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, tmpl, branding)
	staffHandler := handler.NewStaffHandler(caseService, dashboardService, tmpl, branding)
	publicHandler := handler.NewPublicHandler(caseService, tmpl, branding)

	// Setup routes (mirrors cmd/server/main.go)
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/", publicHandler.Home)
	mux.HandleFunc("/staff/login", authHandler.StaffLogin)
	mux.HandleFunc("/staff/logout", authHandler.Logout)
	mux.HandleFunc("/submit/advisory-opinion", publicHandler.SubmitAdvisoryOpinion)
	mux.HandleFunc("/submit/ethics-complaint", publicHandler.SubmitEthicsComplaint)
	mux.HandleFunc("/submit/acknowledgment", publicHandler.SubmitAcknowledgment)
	mux.HandleFunc("/submit/records-request", publicHandler.SubmitRecordsRequest)
	mux.HandleFunc("/submit/confirmation", publicHandler.Confirmation)
	mux.HandleFunc("/search", publicHandler.Search)
	mux.HandleFunc("/opinions/", publicHandler.ViewOpinion)

	// Staff routes (protected)
	staffMux := http.NewServeMux()
	staffMux.HandleFunc("/staff/dashboard", staffHandler.Dashboard)
	staffMux.HandleFunc("/staff/cases", staffHandler.CaseList)
	staffMux.HandleFunc("/staff/cases/", staffHandler.CaseDetail)
	staffMux.HandleFunc("/staff/acknowledgments", staffHandler.Acknowledgments)
	staffMux.HandleFunc("/staff/acknowledgments/", staffHandler.AcknowledgmentsDetail)
	staffMux.HandleFunc("/staff/deadlines", staffHandler.Deadlines)
	staffMux.HandleFunc("/staff/reports", staffHandler.Reports)
	staffMux.HandleFunc("/staff/users", staffHandler.Users)
	staffMux.HandleFunc("/staff/settings", staffHandler.Settings)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	mux.Handle("/staff/", authMiddleware.RequireAuth(staffMux))

	server := httptest.NewServer(mux)

	// Create cookie jar for session management
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	return &TestServer{
		Server: server,
		Repos:  repos,
		Client: client,
		t:      t,
	}
}

// Response wraps an HTTP response with the body as a string for convenience.
type Response struct {
	*http.Response
	Body string
}

// GET performs a GET request and returns the response with body.
func (ts *TestServer) GET(path string) *Response {
	ts.t.Helper()
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	if err != nil {
		ts.t.Fatalf("GET %s: failed to create request: %v", path, err)
	}
	return ts.do(req)
}

// POST performs a POST request with form data and returns the response with body.
func (ts *TestServer) POST(path string, data url.Values) *Response {
	ts.t.Helper()
	req, err := http.NewRequest("POST", ts.URL+path, strings.NewReader(data.Encode()))
	if err != nil {
		ts.t.Fatalf("POST %s: failed to create request: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return ts.do(req)
}

// HTMX performs a GET request with HX-Request header for HTMX fragment requests.
func (ts *TestServer) HTMX(path string) *Response {
	ts.t.Helper()
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	if err != nil {
		ts.t.Fatalf("HTMX %s: failed to create request: %v", path, err)
	}
	req.Header.Set("HX-Request", "true")
	return ts.do(req)
}

func (ts *TestServer) do(req *http.Request) *Response {
	ts.t.Helper()
	resp, err := ts.Client.Do(req)
	if err != nil {
		ts.t.Fatalf("%s %s: request failed: %v", req.Method, req.URL.Path, err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		ts.t.Fatalf("%s %s: failed to read body: %v", req.Method, req.URL.Path, err)
	}

	return &Response{
		Response: resp,
		Body:     string(body),
	}
}

// Login authenticates with the given credentials.
func (ts *TestServer) Login(email, password string) {
	ts.t.Helper()
	resp := ts.POST("/staff/login", url.Values{
		"email":    {email},
		"password": {password},
	})
	if resp.StatusCode != http.StatusSeeOther {
		ts.t.Fatalf("Login failed: expected 303, got %d", resp.StatusCode)
	}
}

// SessionToken returns the current session token from cookies, or empty string.
func (ts *TestServer) SessionToken() string {
	u, _ := url.Parse(ts.URL)
	for _, c := range ts.Client.Jar.Cookies(u) {
		if c.Name == "session" {
			return c.Value
		}
	}
	return ""
}

// ClearCookies removes all cookies from the cookie jar.
func (ts *TestServer) ClearCookies() {
	jar, _ := cookiejar.New(nil)
	ts.Client.Jar = jar
}
