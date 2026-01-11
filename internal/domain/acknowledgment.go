package domain

import "time"

// EthicsAcknowledgment represents a filed ethics acknowledgment
type EthicsAcknowledgment struct {
	ID              string
	CaseNumber      string // EA-YYYY-NNN

	// Official Information
	OfficialName    string
	OfficialTitle   string
	Agency          string
	AgencyType      string // "state", "county", "city", "district"

	// Term Information
	TermStartDate   time.Time
	TermEndDate     *time.Time

	// Acknowledgment Details
	AcknowledgedAt  time.Time
	SignatureOnFile bool

	// Contact
	Email           string
	Phone           string
	Address         string

	// Status
	IsActive        bool

	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PublishedOpinion represents a published advisory opinion or order
type PublishedOpinion struct {
	ID           string
	CaseNumber   string
	Type         CaseType // AO or EC
	Title        string
	Summary      string
	Topics       []string // "conflicts of interest", "gifts", "voting", etc.
	Statutes     []string // NRS 281A.xxx citations
	DocumentURL  string
	PublishedAt  time.Time
	Year         int
}

// SearchResult represents a search result for public search
type SearchResult struct {
	CaseNumber  string
	Type        string
	Title       string
	Summary     string
	Topics      []string
	PublishedAt time.Time
	Relevance   float64
}
