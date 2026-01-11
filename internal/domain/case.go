package domain

import "time"

// CaseType represents the type of ethics case
type CaseType string

const (
	CaseTypeAdvisoryOpinion     CaseType = "AO"  // Advisory Opinion Request
	CaseTypeEthicsComplaint     CaseType = "EC"  // Ethics Complaint
	CaseTypeEthicsAcknowledgment CaseType = "EA"  // Ethics Acknowledgment
	CaseTypePublicRecordsRequest CaseType = "PRR" // Public Records Request
)

// CaseStatus represents the current status of a case
type CaseStatus string

const (
	StatusSubmitted    CaseStatus = "submitted"
	StatusUnderReview  CaseStatus = "under_review"
	StatusInvestigation CaseStatus = "investigation"
	StatusPendingHearing CaseStatus = "pending_hearing"
	StatusDraftPrepared CaseStatus = "draft_prepared"
	StatusPublished    CaseStatus = "published"
	StatusClosed       CaseStatus = "closed"
	StatusWithdrawn    CaseStatus = "withdrawn"
)

// Case represents an ethics case in the system
type Case struct {
	ID              string
	CaseNumber      string     // e.g., "AO-2024-015"
	Type            CaseType
	Status          CaseStatus

	// Submitter Information
	SubmitterName   string
	SubmitterTitle  string
	SubmitterAgency string
	SubmitterEmail  string
	SubmitterPhone  string

	// For Ethics Complaints - Subject Official
	SubjectName     string
	SubjectTitle    string
	SubjectAgency   string

	// Case Content
	Summary         string
	Description     string
	StatuteCitations string // NRS 281A references

	// Dates
	SubmittedAt     time.Time
	DueDate         time.Time
	ClosedAt        *time.Time
	PublishedAt     *time.Time

	// Assignment
	AssignedTo      string // Staff user ID
	AssignedToName  string

	// Metadata
	IsPublic        bool   // Whether published to public search
	IsConfidential  bool
	Priority        string // "normal", "high", "critical"
	Tags            []string

	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Document represents a file attached to a case
type Document struct {
	ID          string
	CaseID      string
	Filename    string
	ContentType string
	Size        int64
	Category    string // "submission", "evidence", "draft", "final", "correspondence"
	IsPublic    bool
	UploadedBy  string
	UploadedAt  time.Time
}

// CaseNote represents an internal note on a case
type CaseNote struct {
	ID        string
	CaseID    string
	AuthorID  string
	AuthorName string
	Content   string
	CreatedAt time.Time
}

// CaseActivity represents a timeline entry for case history
type CaseActivity struct {
	ID          string
	CaseID      string
	Action      string // "created", "status_changed", "assigned", "document_added", "note_added"
	Description string
	UserID      string
	UserName    string
	OldValue    string
	NewValue    string
	CreatedAt   time.Time
}

// Deadline represents a deadline for a case
type Deadline struct {
	ID           string
	CaseID       string
	CaseNumber   string
	CaseType     CaseType // Case type (AO, EC, EA, PRR) for display
	Summary      string   // Case summary for display
	Type         string   // Deadline type: "response_due", "hearing", "extension"
	DueDate      time.Time
	Status       string // "upcoming", "due_soon", "overdue", "completed"
	ReminderSent bool
	CompletedAt  *time.Time
}

// DaysUntilDue returns the number of days until the deadline
func (d *Deadline) DaysUntilDue() int {
	return int(time.Until(d.DueDate).Hours() / 24)
}

// IsOverdue returns true if the deadline has passed
func (d *Deadline) IsOverdue() bool {
	return time.Now().After(d.DueDate) && d.CompletedAt == nil
}

// IsDueSoon returns true if the deadline is within 7 days
func (d *Deadline) IsDueSoon() bool {
	days := d.DaysUntilDue()
	return days >= 0 && days <= 7 && d.CompletedAt == nil
}

// IsOverdue returns true if the case due date has passed
func (c *Case) IsOverdue() bool {
	if c.DueDate.IsZero() {
		return false
	}
	return time.Now().After(c.DueDate) && c.Status != StatusClosed && c.Status != StatusPublished
}

// CaseStats holds dashboard statistics
type CaseStats struct {
	TotalOpen         int
	TotalPending      int
	TotalOverdue      int
	TotalClosed       int
	ByType            map[string]int // Use string keys for template compatibility
	ByStatus          map[string]int // Use string keys for template compatibility
	RecentCases       []Case
	RecentActivity    []CaseActivity
	UpcomingDeadlines []Deadline
}
