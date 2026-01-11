package handler

import (
	"net/http"
	"strings"
	"time"

	"ncoe/internal/config"
	"ncoe/internal/domain"
	"ncoe/internal/service"
	"ncoe/internal/templates"
)

type StaffHandler struct {
	caseService      *service.CaseService
	dashboardService *service.DashboardService
	tmpl             *templates.Renderer
	branding         config.Branding
}

func NewStaffHandler(cs *service.CaseService, ds *service.DashboardService, tmpl *templates.Renderer, b config.Branding) *StaffHandler {
	return &StaffHandler{
		caseService:      cs,
		dashboardService: ds,
		tmpl:             tmpl,
		branding:         b,
	}
}

// Dashboard shows the staff dashboard with KPIs
func (h *StaffHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats := h.dashboardService.GetStats()
	recentCases := h.caseService.GetRecent(10)
	deadlines := h.caseService.GetUpcomingDeadlines(5)

	data := map[string]interface{}{
		"Title":     "Dashboard",
		"Branding":  h.branding,
		"Dashboard": stats, // Template expects .Dashboard
		"Stats":     stats,
		"Recent":    recentCases,
		"Deadlines": deadlines,
		"User":      getUserFromContext(r),
		"ActiveNav": "dashboard",
	}

	h.render(w, "staff/dashboard", data)
}

// CaseList shows all cases with filtering
func (h *StaffHandler) CaseList(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")
	searchQuery := r.URL.Query().Get("q")

	cases := h.caseService.List(typeFilter, statusFilter, searchQuery)

	// Build filter object for template
	filter := map[string]string{
		"Type":   typeFilter,
		"Status": statusFilter,
		"Query":  searchQuery,
	}

	// Calculate counts (simplified - in production these would come from service)
	totalCount := len(cases)
	submittedCount := 0
	underReviewCount := 0
	overdueCount := 0
	for _, c := range cases {
		switch c.Status {
		case domain.StatusSubmitted:
			submittedCount++
		case domain.StatusUnderReview:
			underReviewCount++
		}
		if c.IsOverdue() {
			overdueCount++
		}
	}

	data := map[string]interface{}{
		"Title":            "Cases",
		"Branding":         h.branding,
		"Cases":            cases,
		"Filter":           filter,
		"TotalCount":       totalCount,
		"SubmittedCount":   submittedCount,
		"UnderReviewCount": underReviewCount,
		"OverdueCount":     overdueCount,
		"CurrentPage":      1,
		"TotalPages":       1,
		"PageNumbers":      []int{1},
		"User":             getUserFromContext(r),
		"ActiveNav":        "cases",
	}

	h.render(w, "staff/cases", data)
}

// CaseDetail shows a single case (or routes to panel/fragments)
func (h *StaffHandler) CaseDetail(w http.ResponseWriter, r *http.Request) {
	// Extract case ID from URL path: /staff/cases/{id} or /staff/cases/{id}/_panel
	path := strings.TrimPrefix(r.URL.Path, "/staff/cases/")
	parts := strings.Split(path, "/")
	caseID := parts[0]

	// Check if this is a fragment request (HTMX partials use /_prefix)
	if len(parts) > 1 && parts[1] == "_panel" {
		h.CasePanel(w, r, caseID)
		return
	}
	if len(parts) > 1 && parts[1] == "_status" {
		h.CaseStatusUpdate(w, r, caseID)
		return
	}

	c := h.caseService.GetByID(caseID)
	if c == nil {
		http.NotFound(w, r)
		return
	}

	documents := h.caseService.GetDocuments(caseID)
	notes := h.caseService.GetNotes(caseID)
	activity := h.caseService.GetActivity(caseID)

	data := map[string]interface{}{
		"Title":     c.CaseNumber,
		"Branding":  h.branding,
		"Case":      c,
		"Documents": documents,
		"Notes":     notes,
		"Activity":  activity,
		"User":      getUserFromContext(r),
	}

	h.render(w, "staff/case_detail", data)
}

// CasePanel returns the case detail panel (for HTMX offcanvas)
func (h *StaffHandler) CasePanel(w http.ResponseWriter, r *http.Request, caseID string) {
	c := h.caseService.GetByID(caseID)
	if c == nil {
		http.NotFound(w, r)
		return
	}

	documents := h.caseService.GetDocuments(caseID)
	activity := h.caseService.GetActivity(caseID)

	data := map[string]interface{}{
		"Branding":  h.branding,
		"Case":      c,
		"Documents": documents,
		"Activity":  activity,
		"User":      getUserFromContext(r),
	}

	h.render(w, "staff/case_panel", data)
}

// CaseStatusUpdate handles case status changes (HTMX fragment: /_status)
func (h *StaffHandler) CaseStatusUpdate(w http.ResponseWriter, r *http.Request, caseID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse new status from form
	r.ParseForm()
	newStatus := domain.CaseStatus(r.FormValue("status"))

	// Update the case status in the repository
	err := h.caseService.UpdateStatus(caseID, newStatus)
	if err != nil {
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	// Return empty response with HX-Trigger to refresh the panel
	w.Header().Set("HX-Trigger", "caseUpdated")
	w.WriteHeader(http.StatusOK)
}

// Deadlines shows all upcoming deadlines
func (h *StaffHandler) Deadlines(w http.ResponseWriter, r *http.Request) {
	deadlines := h.caseService.GetAllDeadlines()

	data := map[string]interface{}{
		"Title":     "Deadlines",
		"Branding":  h.branding,
		"Deadlines": deadlines,
		"User":      getUserFromContext(r),
		"ActiveNav": "deadlines",
	}

	h.render(w, "staff/deadlines", data)
}

// Reports shows reporting interface
func (h *StaffHandler) Reports(w http.ResponseWriter, r *http.Request) {
	stats := h.dashboardService.GetStats()

	data := map[string]interface{}{
		"Title":     "Reports",
		"Branding":  h.branding,
		"Stats":     stats,
		"User":      getUserFromContext(r),
		"ActiveNav": "reports",
	}

	h.render(w, "staff/reports", data)
}

// Users shows user management (admin only)
func (h *StaffHandler) Users(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":     "User Management",
		"Branding":  h.branding,
		"User":      getUserFromContext(r),
		"ActiveNav": "users",
	}

	h.render(w, "staff/users", data)
}

// Settings shows system settings
func (h *StaffHandler) Settings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":     "Settings",
		"Branding":  h.branding,
		"User":      getUserFromContext(r),
		"ActiveNav": "settings",
	}

	h.render(w, "staff/settings", data)
}

// AcknowledgmentsDetail handles /staff/acknowledgments/{id} and fragments
func (h *StaffHandler) AcknowledgmentsDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/staff/acknowledgments/")
	parts := strings.Split(path, "/")
	ackID := parts[0]

	// Check if this is a fragment request (HTMX partials use /_prefix)
	if len(parts) > 1 && parts[1] == "_panel" {
		h.AcknowledgmentPanel(w, r, ackID)
		return
	}

	// Full page view (not implemented yet)
	http.NotFound(w, r)
}

// AcknowledgmentPanel returns acknowledgment detail panel (HTMX fragment: /_panel)
func (h *StaffHandler) AcknowledgmentPanel(w http.ResponseWriter, r *http.Request, ackID string) {
	// Get mock acknowledgment by ID
	acks := getMockAcknowledgments("", "", "")
	var ack *domain.EthicsAcknowledgment
	for _, a := range acks {
		if a.ID == ackID {
			ack = a
			break
		}
	}
	if ack == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Branding":       h.branding,
		"Acknowledgment": ack,
		"User":           getUserFromContext(r),
	}

	h.render(w, "staff/acknowledgment_panel", data)
}

// Acknowledgments shows filed ethics acknowledgments
func (h *StaffHandler) Acknowledgments(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	agencyType := r.URL.Query().Get("agency_type")
	query := r.URL.Query().Get("q")
	year := r.URL.Query().Get("year")

	// Get mock acknowledgments (in production, this would come from a service)
	acknowledgments := getMockAcknowledgments(agencyType, query, year)

	// Build filter object for template
	filter := map[string]string{
		"AgencyType": agencyType,
		"Query":      query,
		"Year":       year,
	}

	data := map[string]interface{}{
		"Title":           "Acknowledgments",
		"Branding":        h.branding,
		"Acknowledgments": acknowledgments,
		"Filter":          filter,
		"TotalCount":      len(acknowledgments),
		"ActiveCount":     countActive(acknowledgments),
		"ThisMonthCount":  3, // Mock data
		"ExpiringCount":   2, // Mock data
		"CurrentPage":     1,
		"TotalPages":      1,
		"PageNumbers":     []int{1},
		"User":            getUserFromContext(r),
		"ActiveNav":       "acknowledgments",
	}

	h.render(w, "staff/acknowledgments", data)
}

// Mock acknowledgments data
func getMockAcknowledgments(agencyType, query, year string) []*domain.EthicsAcknowledgment {
	now := time.Now()
	termEnd1 := now.AddDate(2, 0, 0)
	termEnd2 := now.AddDate(1, 6, 0)

	acknowledgments := []*domain.EthicsAcknowledgment{
		{
			ID:              "ack_1",
			CaseNumber:      "EA-2024-089",
			OfficialName:    "Maria Garcia",
			OfficialTitle:   "Board Member",
			Agency:          "Nevada State Board of Education",
			AgencyType:      "state",
			TermStartDate:   now.AddDate(-1, 0, 0),
			TermEndDate:     &termEnd1,
			AcknowledgedAt:  now.AddDate(0, 0, -1),
			SignatureOnFile: true,
			Email:           "mgarcia@doe.nv.gov",
			IsActive:        true,
		},
		{
			ID:              "ack_2",
			CaseNumber:      "EA-2024-088",
			OfficialName:    "James Wilson",
			OfficialTitle:   "County Commissioner",
			Agency:          "Clark County",
			AgencyType:      "county",
			TermStartDate:   now.AddDate(-2, 0, 0),
			TermEndDate:     &termEnd2,
			AcknowledgedAt:  now.AddDate(0, 0, -5),
			SignatureOnFile: true,
			Email:           "jwilson@clarkcounty.gov",
			IsActive:        true,
		},
		{
			ID:              "ack_3",
			CaseNumber:      "EA-2024-087",
			OfficialName:    "Patricia Chen",
			OfficialTitle:   "City Councilwoman",
			Agency:          "City of Las Vegas",
			AgencyType:      "city",
			TermStartDate:   now.AddDate(-1, 6, 0),
			AcknowledgedAt:  now.AddDate(0, 0, -10),
			SignatureOnFile: true,
			Email:           "pchen@lasvegasnevada.gov",
			IsActive:        true,
		},
		{
			ID:              "ack_4",
			CaseNumber:      "EA-2024-086",
			OfficialName:    "Robert Thompson",
			OfficialTitle:   "Board Trustee",
			Agency:          "Las Vegas Valley Water District",
			AgencyType:      "district",
			TermStartDate:   now.AddDate(-3, 0, 0),
			AcknowledgedAt:  now.AddDate(0, 0, -15),
			SignatureOnFile: true,
			Email:           "rthompson@lvvwd.com",
			IsActive:        true,
		},
		{
			ID:              "ack_5",
			CaseNumber:      "EA-2024-085",
			OfficialName:    "Sarah Martinez",
			OfficialTitle:   "Director",
			Agency:          "Nevada Department of Motor Vehicles",
			AgencyType:      "state",
			TermStartDate:   now.AddDate(-1, 0, 0),
			AcknowledgedAt:  now.AddDate(0, 0, -20),
			SignatureOnFile: true,
			Email:           "smartinez@dmv.nv.gov",
			IsActive:        true,
		},
	}

	// Apply filters
	var filtered []*domain.EthicsAcknowledgment
	for _, a := range acknowledgments {
		if agencyType != "" && a.AgencyType != agencyType {
			continue
		}
		if query != "" {
			queryLower := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(a.OfficialName), queryLower) &&
				!strings.Contains(strings.ToLower(a.Agency), queryLower) &&
				!strings.Contains(strings.ToLower(a.CaseNumber), queryLower) {
				continue
			}
		}
		filtered = append(filtered, a)
	}

	return filtered
}

func countActive(acks []*domain.EthicsAcknowledgment) int {
	count := 0
	for _, a := range acks {
		if a.IsActive {
			count++
		}
	}
	return count
}

func (h *StaffHandler) render(w http.ResponseWriter, name string, data interface{}) {
	err := h.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func getUserFromContext(r *http.Request) *domain.User {
	// Get user from context (set by auth middleware)
	if u := r.Context().Value("user"); u != nil {
		return u.(*domain.User)
	}
	return nil
}
