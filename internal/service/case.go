package service

import (
	"fmt"
	"log"
	"time"

	"ncoe/internal/domain"
)

type CaseRepository interface {
	Create(c *domain.Case) error
	Update(c *domain.Case) error
	GetByID(id string) *domain.Case
	GetByCaseNumber(num string) *domain.Case
	List(typeFilter, statusFilter, query string) []*domain.Case
	GetRecent(limit int) []*domain.Case
	GetDocuments(caseID string) []*domain.Document
	GetNotes(caseID string) []*domain.CaseNote
	GetActivity(caseID string) []*domain.CaseActivity
	GetDeadlines(limit int) []*domain.Deadline
	GetAllDeadlines() []*domain.Deadline
	SearchPublished(query, docType, year, topic string) []domain.PublishedOpinion
	GetPublishedOpinion(caseNumber string) *domain.PublishedOpinion
	NextCaseNumber(caseType domain.CaseType) string
}

type CaseService struct {
	repo CaseRepository
}

func NewCaseService(repo CaseRepository) *CaseService {
	return &CaseService{repo: repo}
}

// Create creates a new case and returns the case number
func (s *CaseService) Create(c *domain.Case) (string, error) {
	// Generate case number
	c.CaseNumber = s.repo.NextCaseNumber(c.Type)
	c.ID = fmt.Sprintf("case_%d", time.Now().UnixNano())
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	// Calculate deadline based on case type
	switch c.Type {
	case domain.CaseTypeAdvisoryOpinion:
		c.DueDate = calculateBusinessDays(c.SubmittedAt, 45)
	case domain.CaseTypePublicRecordsRequest:
		c.DueDate = calculateBusinessDays(c.SubmittedAt, 5)
	}

	if err := s.repo.Create(c); err != nil {
		return "", err
	}

	log.Printf("[CASE CREATED] ID=%s Number=%s Type=%s Submitter=%s", c.ID, c.CaseNumber, c.Type, c.SubmitterName)
	return c.CaseNumber, nil
}

// GetByID retrieves a case by ID
func (s *CaseService) GetByID(id string) *domain.Case {
	return s.repo.GetByID(id)
}

// List returns cases with optional filters
func (s *CaseService) List(typeFilter, statusFilter, query string) []*domain.Case {
	return s.repo.List(typeFilter, statusFilter, query)
}

// GetRecent returns the most recent cases
func (s *CaseService) GetRecent(limit int) []*domain.Case {
	return s.repo.GetRecent(limit)
}

// GetDocuments returns documents for a case
func (s *CaseService) GetDocuments(caseID string) []*domain.Document {
	return s.repo.GetDocuments(caseID)
}

// GetNotes returns notes for a case
func (s *CaseService) GetNotes(caseID string) []*domain.CaseNote {
	return s.repo.GetNotes(caseID)
}

// GetActivity returns activity log for a case
func (s *CaseService) GetActivity(caseID string) []*domain.CaseActivity {
	return s.repo.GetActivity(caseID)
}

// GetUpcomingDeadlines returns upcoming deadlines
func (s *CaseService) GetUpcomingDeadlines(limit int) []*domain.Deadline {
	return s.repo.GetDeadlines(limit)
}

// GetAllDeadlines returns all deadlines
func (s *CaseService) GetAllDeadlines() []*domain.Deadline {
	return s.repo.GetAllDeadlines()
}

// SearchPublished searches published opinions
func (s *CaseService) SearchPublished(query, docType, year, topic string) []domain.PublishedOpinion {
	return s.repo.SearchPublished(query, docType, year, topic)
}

// GetPublishedOpinion retrieves a published opinion
func (s *CaseService) GetPublishedOpinion(caseNumber string) *domain.PublishedOpinion {
	return s.repo.GetPublishedOpinion(caseNumber)
}

// UpdateStatus updates the status of a case
func (s *CaseService) UpdateStatus(caseID string, status domain.CaseStatus) error {
	c := s.repo.GetByID(caseID)
	if c == nil {
		return fmt.Errorf("case not found: %s", caseID)
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	return s.repo.Update(c)
}

// calculateBusinessDays adds business days to a date
func calculateBusinessDays(start time.Time, days int) time.Time {
	result := start
	added := 0
	for added < days {
		result = result.AddDate(0, 0, 1)
		// Skip weekends
		if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
			added++
		}
	}
	return result
}
