package mock

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"ncoe/internal/domain"
)

type Repositories struct {
	User    *UserRepository
	Session *SessionRepository
	Case    *CaseRepository
}

func NewRepositories() *Repositories {
	return &Repositories{
		User:    NewUserRepository(),
		Session: NewSessionRepository(),
		Case:    NewCaseRepository(),
	}
}

// UserRepository is an in-memory user store
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewUserRepository() *UserRepository {
	r := &UserRepository{users: make(map[string]*domain.User)}
	// Add demo users
	r.users["demo@ncoe.nv.gov"] = &domain.User{
		ID:        "user_1",
		Email:     "demo@ncoe.nv.gov",
		FirstName: "Demo",
		LastName:  "Admin",
		Role:      domain.RoleAdmin,
		Title:     "System Administrator",
		IsActive:  true,
	}
	return r
}

func (r *UserRepository) GetByEmail(email string) *domain.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.users[email]
}

func (r *UserRepository) GetByID(id string) *domain.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.ID == id {
			return u
		}
	}
	// Demo mode: return mock user
	return &domain.User{
		ID:        id,
		Email:     "demo@ncoe.nv.gov",
		FirstName: "Demo",
		LastName:  "User",
		Role:      domain.RoleAdmin,
		IsActive:  true,
	}
}

// SessionRepository is an in-memory session store
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: make(map[string]*domain.Session)}
}

func (r *SessionRepository) Create(s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.Token] = s
	return nil
}

func (r *SessionRepository) GetByToken(token string) *domain.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sessions[token]
}

func (r *SessionRepository) Delete(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
	return nil
}

// CaseRepository is an in-memory case store
type CaseRepository struct {
	mu       sync.RWMutex
	cases    map[string]*domain.Case
	counters map[domain.CaseType]int
}

func NewCaseRepository() *CaseRepository {
	r := &CaseRepository{
		cases:    make(map[string]*domain.Case),
		counters: make(map[domain.CaseType]int),
	}
	r.seedDemoData()
	return r
}

func (r *CaseRepository) seedDemoData() {
	now := time.Now()

	// Seed demo cases - these match the dashboard display exactly
	// IDs are simple numbers for easy URL navigation (/staff/cases/1, /staff/cases/2, etc.)
	demoCases := []*domain.Case{
		// Recent submissions shown on dashboard
		{
			ID:              "1",
			CaseNumber:      "AO-2024-042",
			Type:            domain.CaseTypeAdvisoryOpinion,
			Status:          domain.StatusSubmitted,
			SubmitterName:   "John Smith",
			SubmitterTitle:  "City Manager",
			SubmitterAgency: "City of Henderson",
			SubmitterEmail:  "jsmith@henderson.gov",
			Summary:         "Question regarding contractor relationships",
			Description:     "May I participate in discussions regarding a contract with a company where my brother-in-law is employed?",
			SubmittedAt:     now.AddDate(0, 0, -1),
			DueDate:         now.AddDate(0, 0, 3), // Upcoming deadline
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "normal",
		},
		{
			ID:              "2",
			CaseNumber:      "EC-2024-018",
			Type:            domain.CaseTypeEthicsComplaint,
			Status:          domain.StatusUnderReview,
			SubmitterName:   "Jane Doe",
			SubmitterEmail:  "concerned@example.com",
			SubjectName:     "Robert Johnson",
			SubjectTitle:    "County Commissioner",
			SubjectAgency:   "Clark County",
			Summary:         "Alleged gift violation",
			Description:     "Commissioner Johnson allegedly accepted tickets to a Las Vegas show from a vendor seeking county contracts.",
			SubmittedAt:     now.AddDate(0, 0, -2),
			DueDate:         now.AddDate(0, 0, 5), // Investigation deadline
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "high",
		},
		{
			ID:              "3",
			CaseNumber:      "PRR-2024-089",
			Type:            domain.CaseTypePublicRecordsRequest,
			Status:          domain.StatusSubmitted,
			SubmitterName:   "City of Henderson",
			SubmitterEmail:  "records@cityofhenderson.com",
			Summary:         "Request for ethics training records",
			Description:     "Requesting copies of all ethics training materials and attendance records from 2023-2024.",
			SubmittedAt:     now.AddDate(0, 0, -3),
			DueDate:         now.AddDate(0, 0, 2),
			Priority:        "normal",
		},
		{
			ID:              "4",
			CaseNumber:      "EA-2024-156",
			Type:            domain.CaseTypeEthicsAcknowledgment,
			Status:          domain.StatusUnderReview,
			SubmitterName:   "Robert Johnson",
			SubmitterTitle:  "Board Member",
			SubmitterAgency: "Nevada Gaming Control Board",
			SubmitterEmail:  "rjohnson@gcb.nv.gov",
			Summary:         "Annual Ethics Acknowledgment",
			SubmittedAt:     now.AddDate(0, 0, -4),
			Priority:        "normal",
		},
		{
			ID:              "5",
			CaseNumber:      "AO-2024-041",
			Type:            domain.CaseTypeAdvisoryOpinion,
			Status:          domain.StatusInvestigation,
			SubmitterName:   "Maria Garcia",
			SubmitterTitle:  "State Employee",
			SubmitterAgency: "Department of Motor Vehicles",
			SubmitterEmail:  "mgarcia@dmv.nv.gov",
			Summary:         "Outside employment with DMV vendor",
			Description:     "I have been offered a weekend consulting position with an IT firm that has contracts with DMV. Is this permissible?",
			SubmittedAt:     now.AddDate(0, 0, -5),
			DueDate:         now.AddDate(0, 0, 40),
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "normal",
		},
		// Additional cases for deadlines display
		{
			ID:              "6",
			CaseNumber:      "PRR-2024-088",
			Type:            domain.CaseTypePublicRecordsRequest,
			Status:          domain.StatusUnderReview,
			SubmitterName:   "Nevada Press Association",
			SubmitterEmail:  "records@nvpress.org",
			Summary:         "Request for complaint statistics",
			Description:     "Requesting all complaint statistics from 2020-2024.",
			SubmittedAt:     now.AddDate(0, 0, -7),
			DueDate:         now.AddDate(0, 0, -2), // Overdue!
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "high",
		},
		{
			ID:              "7",
			CaseNumber:      "AO-2024-039",
			Type:            domain.CaseTypeAdvisoryOpinion,
			Status:          domain.StatusUnderReview,
			SubmitterName:   "David Chen",
			SubmitterTitle:  "Director",
			SubmitterAgency: "Department of Transportation",
			SubmitterEmail:  "dchen@dot.nv.gov",
			Summary:         "Family member employment at vendor",
			Description:     "My daughter has been offered employment at a firm that frequently bids on NDOT contracts. What are my obligations?",
			SubmittedAt:     now.AddDate(0, 0, -14),
			DueDate:         now.AddDate(0, 0, 7), // Hearing scheduled
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "normal",
		},
		// More cases for realistic case list
		{
			ID:              "8",
			CaseNumber:      "EC-2024-017",
			Type:            domain.CaseTypeEthicsComplaint,
			Status:          domain.StatusDraftPrepared,
			SubmitterName:   "Anonymous",
			SubmitterEmail:  "anonymous@protonmail.com",
			SubjectName:     "Lisa Wong",
			SubjectTitle:    "City Councilwoman",
			SubjectAgency:   "City of Reno",
			Summary:         "Misuse of public resources",
			Description:     "Councilwoman Wong allegedly used city staff to plan her daughter's wedding.",
			SubmittedAt:     now.AddDate(0, 0, -21),
			DueDate:         now.AddDate(0, 0, 14),
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "high",
		},
		{
			ID:              "9",
			CaseNumber:      "AO-2024-040",
			Type:            domain.CaseTypeAdvisoryOpinion,
			Status:          domain.StatusClosed,
			SubmitterName:   "Thomas Anderson",
			SubmitterTitle:  "Sheriff",
			SubmitterAgency: "Washoe County Sheriff's Office",
			SubmitterEmail:  "tanderson@washoesheriff.gov",
			Summary:         "Charitable organization board membership",
			Description:     "May I serve on the board of a charitable organization that occasionally applies for county grants?",
			SubmittedAt:     now.AddDate(0, -1, 0),
			DueDate:         now.AddDate(0, 0, -15),
			AssignedTo:      "user_1",
			AssignedToName:  "Ross Armstrong",
			Priority:        "normal",
		},
		{
			ID:              "10",
			CaseNumber:      "EA-2024-155",
			Type:            domain.CaseTypeEthicsAcknowledgment,
			Status:          domain.StatusClosed,
			SubmitterName:   "Sarah Miller",
			SubmitterTitle:  "Commissioner",
			SubmitterAgency: "Public Utilities Commission",
			SubmitterEmail:  "smiller@puc.nv.gov",
			Summary:         "Annual Ethics Acknowledgment",
			SubmittedAt:     now.AddDate(0, 0, -10),
			Priority:        "normal",
		},
	}

	for _, c := range demoCases {
		r.cases[c.ID] = c
	}

	// Set counters to continue numbering correctly
	r.counters[domain.CaseTypeAdvisoryOpinion] = 42
	r.counters[domain.CaseTypeEthicsComplaint] = 18
	r.counters[domain.CaseTypeEthicsAcknowledgment] = 156
	r.counters[domain.CaseTypePublicRecordsRequest] = 89
}

func (r *CaseRepository) Create(c *domain.Case) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cases[c.ID] = c
	return nil
}

func (r *CaseRepository) Update(c *domain.Case) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.cases[c.ID]; !exists {
		return fmt.Errorf("case not found: %s", c.ID)
	}
	r.cases[c.ID] = c
	return nil
}

func (r *CaseRepository) GetByID(id string) *domain.Case {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cases[id]
}

func (r *CaseRepository) GetByCaseNumber(num string) *domain.Case {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.cases {
		if c.CaseNumber == num {
			return c
		}
	}
	return nil
}

func (r *CaseRepository) List(typeFilter, statusFilter, query string) []*domain.Case {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Case
	for _, c := range r.cases {
		if typeFilter != "" && string(c.Type) != typeFilter {
			continue
		}
		if statusFilter != "" && string(c.Status) != statusFilter {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(c.CaseNumber+c.Summary+c.SubmitterName), strings.ToLower(query)) {
			continue
		}
		result = append(result, c)
	}
	return result
}

func (r *CaseRepository) GetRecent(limit int) []*domain.Case {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Case
	for _, c := range r.cases {
		result = append(result, c)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func (r *CaseRepository) GetDocuments(caseID string) []*domain.Document {
	return []*domain.Document{} // Demo: no documents
}

func (r *CaseRepository) GetNotes(caseID string) []*domain.CaseNote {
	return []*domain.CaseNote{} // Demo: no notes
}

func (r *CaseRepository) GetActivity(caseID string) []*domain.CaseActivity {
	return []*domain.CaseActivity{
		{
			ID:          "act_1",
			CaseID:      caseID,
			Action:      "created",
			Description: "Case created from public submission",
			CreatedAt:   time.Now().AddDate(0, 0, -5),
		},
	}
}

func (r *CaseRepository) GetDeadlines(limit int) []*domain.Deadline {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var deadlines []*domain.Deadline
	for _, c := range r.cases {
		if !c.DueDate.IsZero() && c.Status != domain.StatusClosed {
			status := "upcoming"
			if time.Now().After(c.DueDate) {
				status = "overdue"
			} else if time.Until(c.DueDate).Hours() < 7*24 {
				status = "due_soon"
			}
			deadlines = append(deadlines, &domain.Deadline{
				ID:         "dl_" + c.ID,
				CaseID:     c.ID,
				CaseNumber: c.CaseNumber,
				CaseType:   c.Type,
				Summary:    c.Summary,
				Type:       "response_due",
				DueDate:    c.DueDate,
				Status:     status,
			})
		}
		if len(deadlines) >= limit {
			break
		}
	}
	return deadlines
}

func (r *CaseRepository) GetAllDeadlines() []*domain.Deadline {
	return r.GetDeadlines(100)
}

func (r *CaseRepository) SearchPublished(query, docType, year, topic string) []domain.PublishedOpinion {
	// Demo: return sample published opinions
	return []domain.PublishedOpinion{
		{
			CaseNumber:  "AO-2024-010",
			Type:        domain.CaseTypeAdvisoryOpinion,
			Title:       "Advisory Opinion: Contractor Relationships",
			Summary:     "A public officer may not use their position to secure unwarranted privileges for a family member's business.",
			Topics:      []string{"Conflicts of Interest", "Family Members"},
			Statutes:    []string{"NRS 281A.400"},
			PublishedAt: time.Now().AddDate(0, -1, 0),
			Year:        2024,
		},
		{
			CaseNumber:  "EC-2024-005",
			Type:        domain.CaseTypeEthicsComplaint,
			Title:       "Final Order: Gift Violations",
			Summary:     "The Commission finds a willful violation of the Ethics in Government Law occurred when the subject accepted gifts exceeding $50.",
			Topics:      []string{"Gifts", "NRS 281A.400"},
			Statutes:    []string{"NRS 281A.400", "NRS 281A.480"},
			PublishedAt: time.Now().AddDate(0, -2, 0),
			Year:        2024,
		},
	}
}

func (r *CaseRepository) GetPublishedOpinion(caseNumber string) *domain.PublishedOpinion {
	// Demo: return sample
	return &domain.PublishedOpinion{
		CaseNumber:  caseNumber,
		Type:        domain.CaseTypeAdvisoryOpinion,
		Title:       "Advisory Opinion: " + caseNumber,
		Summary:     "Sample published opinion text.",
		Topics:      []string{"Conflicts of Interest"},
		PublishedAt: time.Now().AddDate(0, -1, 0),
		Year:        2024,
	}
}

func (r *CaseRepository) NextCaseNumber(caseType domain.CaseType) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counters[caseType]++
	year := time.Now().Year()
	return fmt.Sprintf("%s-%d-%03d", caseType, year, r.counters[caseType])
}
