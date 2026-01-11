package main

import (
	"log"
	"net/http"
	"os"

	"ncoe/internal/config"
	"ncoe/internal/handler"
	"ncoe/internal/middleware"
	"ncoe/internal/repository/mock"
	"ncoe/internal/service"
	"ncoe/internal/templates"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize repositories (mock for demo, postgres for production)
	var repos *mock.Repositories
	if os.Getenv("DATABASE_URL") == "" {
		log.Println("DATABASE_URL not set, using mock repositories (demo mode)")
		repos = mock.NewRepositories()
	} else {
		log.Fatal("PostgreSQL repositories not yet implemented")
	}

	// Initialize services
	authService := service.NewAuthService(repos.User, repos.Session)
	caseService := service.NewCaseService(repos.Case)
	dashboardService := service.NewDashboardService(repos.Case)

	// Load templates
	tmpl := templates.NewRenderer(cfg.TemplateDir)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, tmpl, cfg.Branding)
	staffHandler := handler.NewStaffHandler(caseService, dashboardService, tmpl, cfg.Branding)
	publicHandler := handler.NewPublicHandler(caseService, tmpl, cfg.Branding)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Auth routes
	mux.HandleFunc("/", publicHandler.Home)
	mux.HandleFunc("/staff/login", authHandler.StaffLogin)
	mux.HandleFunc("/staff/logout", authHandler.Logout)

	// Public submission forms (no login required)
	mux.HandleFunc("/submit/advisory-opinion", publicHandler.SubmitAdvisoryOpinion)
	mux.HandleFunc("/submit/ethics-complaint", publicHandler.SubmitEthicsComplaint)
	mux.HandleFunc("/submit/acknowledgment", publicHandler.SubmitAcknowledgment)
	mux.HandleFunc("/submit/records-request", publicHandler.SubmitRecordsRequest)
	mux.HandleFunc("/submit/confirmation", publicHandler.Confirmation)

	// Public search
	mux.HandleFunc("/search", publicHandler.Search)
	mux.HandleFunc("/opinions/", publicHandler.ViewOpinion)

	// Staff routes (protected)
	staffMux := http.NewServeMux()
	staffMux.HandleFunc("/staff/dashboard", staffHandler.Dashboard)
	staffMux.HandleFunc("/staff/cases", staffHandler.CaseList)
	staffMux.HandleFunc("/staff/cases/", staffHandler.CaseDetail)          // Handles /{id} and /{id}/_panel, /{id}/_status
	staffMux.HandleFunc("/staff/acknowledgments", staffHandler.Acknowledgments)
	staffMux.HandleFunc("/staff/acknowledgments/", staffHandler.AcknowledgmentsDetail) // Handles /{id}/_panel
	staffMux.HandleFunc("/staff/deadlines", staffHandler.Deadlines)
	staffMux.HandleFunc("/staff/reports", staffHandler.Reports)
	staffMux.HandleFunc("/staff/users", staffHandler.Users)
	staffMux.HandleFunc("/staff/settings", staffHandler.Settings)

	// Wrap staff routes with auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	mux.Handle("/staff/", authMiddleware.RequireAuth(staffMux))

	// Apply global middleware (order: outermost first)
	// Recovery -> RequestID -> Logging -> mux
	// RequestID runs before Logging so request_id is available for log output
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)

	// Start server
	addr := cfg.ServerAddress
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("Starting NCOE Case Management System on %s", addr)
	log.Printf("Demo mode: any credentials accepted for staff login")
	log.Fatal(http.ListenAndServe(addr, h))
}

