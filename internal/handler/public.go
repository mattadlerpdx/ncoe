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

type PublicHandler struct {
	caseService *service.CaseService
	tmpl        *templates.Renderer
	branding    config.Branding
}

func NewPublicHandler(cs *service.CaseService, tmpl *templates.Renderer, b config.Branding) *PublicHandler {
	return &PublicHandler{
		caseService: cs,
		tmpl:        tmpl,
		branding:    b,
	}
}

// Home shows the public landing page
func (h *PublicHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "Home",
		"Branding": h.branding,
	}

	h.render(w, "public/home", data)
}

// SubmitAdvisoryOpinion handles advisory opinion request submissions
func (h *PublicHandler) SubmitAdvisoryOpinion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleAdvisoryOpinionSubmission(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "Request Advisory Opinion",
		"Branding": h.branding,
	}

	h.render(w, "public/submit_advisory", data)
}

func (h *PublicHandler) handleAdvisoryOpinionSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max

	c := &domain.Case{
		Type:            domain.CaseTypeAdvisoryOpinion,
		Status:          domain.StatusSubmitted,
		SubmitterName:   r.FormValue("name"),
		SubmitterTitle:  r.FormValue("title"),
		SubmitterAgency: r.FormValue("agency"),
		SubmitterEmail:  r.FormValue("email"),
		SubmitterPhone:  r.FormValue("phone"),
		Summary:         r.FormValue("question_summary"),
		Description:     r.FormValue("question_detail"),
		SubmittedAt:     time.Now(),
	}

	caseNumber, err := h.caseService.Create(c)
	if err != nil {
		http.Error(w, "Failed to submit request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/submit/confirmation?case="+caseNumber+"&type=advisory", http.StatusSeeOther)
}

// SubmitEthicsComplaint handles ethics complaint submissions
func (h *PublicHandler) SubmitEthicsComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleComplaintSubmission(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "File Ethics Complaint",
		"Branding": h.branding,
	}

	h.render(w, "public/submit_complaint", data)
}

func (h *PublicHandler) handleComplaintSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max

	c := &domain.Case{
		Type:             domain.CaseTypeEthicsComplaint,
		Status:           domain.StatusSubmitted,
		SubmitterName:    r.FormValue("complainant_name"),
		SubmitterEmail:   r.FormValue("complainant_email"),
		SubmitterPhone:   r.FormValue("complainant_phone"),
		SubjectName:      r.FormValue("subject_name"),
		SubjectTitle:     r.FormValue("subject_title"),
		SubjectAgency:    r.FormValue("subject_agency"),
		Summary:          r.FormValue("allegation_summary"),
		Description:      r.FormValue("allegation_detail"),
		StatuteCitations: r.FormValue("statutes"),
		SubmittedAt:      time.Now(),
	}

	caseNumber, err := h.caseService.Create(c)
	if err != nil {
		http.Error(w, "Failed to submit complaint", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/submit/confirmation?case="+caseNumber+"&type=complaint", http.StatusSeeOther)
}

// SubmitAcknowledgment handles ethics acknowledgment submissions
func (h *PublicHandler) SubmitAcknowledgment(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleAcknowledgmentSubmission(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "File Ethics Acknowledgment",
		"Branding": h.branding,
	}

	h.render(w, "public/submit_acknowledgment", data)
}

func (h *PublicHandler) handleAcknowledgmentSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max

	c := &domain.Case{
		Type:            domain.CaseTypeEthicsAcknowledgment,
		Status:          domain.StatusSubmitted,
		SubmitterName:   r.FormValue("official_name"),
		SubmitterTitle:  r.FormValue("official_title"),
		SubmitterAgency: r.FormValue("agency"),
		SubmitterEmail:  r.FormValue("email"),
		SubmitterPhone:  r.FormValue("phone"),
		Summary:         "Ethics Acknowledgment Filing",
		SubmittedAt:     time.Now(),
	}

	caseNumber, err := h.caseService.Create(c)
	if err != nil {
		http.Error(w, "Failed to submit acknowledgment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/submit/confirmation?case="+caseNumber+"&type=acknowledgment", http.StatusSeeOther)
}

// SubmitRecordsRequest handles public records request submissions
func (h *PublicHandler) SubmitRecordsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleRecordsRequestSubmission(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    "Public Records Request",
		"Branding": h.branding,
	}

	h.render(w, "public/submit_records", data)
}

func (h *PublicHandler) handleRecordsRequestSubmission(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max

	c := &domain.Case{
		Type:            domain.CaseTypePublicRecordsRequest,
		Status:          domain.StatusSubmitted,
		SubmitterName:   r.FormValue("requester_name"),
		SubmitterEmail:  r.FormValue("requester_email"),
		SubmitterPhone:  r.FormValue("requester_phone"),
		Summary:         r.FormValue("request_summary"),
		Description:     r.FormValue("request_detail"),
		SubmittedAt:     time.Now(),
	}

	caseNumber, err := h.caseService.Create(c)
	if err != nil {
		http.Error(w, "Failed to submit request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/submit/confirmation?case="+caseNumber+"&type=records", http.StatusSeeOther)
}

// Confirmation shows submission confirmation
func (h *PublicHandler) Confirmation(w http.ResponseWriter, r *http.Request) {
	caseNumber := r.URL.Query().Get("case")
	submissionType := r.URL.Query().Get("type")

	data := map[string]interface{}{
		"Title":      "Submission Received",
		"Branding":   h.branding,
		"CaseNumber": caseNumber,
		"Type":       submissionType,
	}

	h.render(w, "public/confirmation", data)
}

// Search handles public search for published opinions
func (h *PublicHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	docType := r.URL.Query().Get("type")
	year := r.URL.Query().Get("year")
	topic := r.URL.Query().Get("topic")

	var results []domain.PublishedOpinion
	if query != "" || docType != "" || year != "" || topic != "" {
		results = h.caseService.SearchPublished(query, docType, year, topic)
	}

	data := map[string]interface{}{
		"Title":    "Search Published Opinions & Orders",
		"Branding": h.branding,
		"Query":    query,
		"DocType":  docType,
		"Year":     year,
		"Topic":    topic,
		"Results":  results,
		"Topics":   []string{"Conflicts of Interest", "Gifts", "Voting", "Employment", "Financial Disclosure"},
		"Years":    []string{"2024", "2023", "2022", "2021", "2020"},
	}

	h.render(w, "public/search", data)
}

// ViewOpinion shows a single published opinion
func (h *PublicHandler) ViewOpinion(w http.ResponseWriter, r *http.Request) {
	caseNumber := strings.TrimPrefix(r.URL.Path, "/opinions/")

	opinion := h.caseService.GetPublishedOpinion(caseNumber)
	if opinion == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":    opinion.Title,
		"Branding": h.branding,
		"Opinion":  opinion,
	}

	h.render(w, "public/opinion", data)
}

func (h *PublicHandler) render(w http.ResponseWriter, name string, data interface{}) {
	err := h.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
