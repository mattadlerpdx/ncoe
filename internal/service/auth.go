package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"ncoe/internal/domain"
)

type UserRepository interface {
	GetByEmail(email string) *domain.User
	GetByID(id string) *domain.User
}

type SessionRepository interface {
	Create(s *domain.Session) error
	GetByToken(token string) *domain.Session
	Delete(token string) error
}

type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
}

func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// LoginStaff authenticates a staff user
func (s *AuthService) LoginStaff(email, password string) (*domain.Session, error) {
	// In demo mode, accept any credentials
	user := s.userRepo.GetByEmail(email)
	if user == nil {
		// Demo mode: create a mock user
		user = &domain.User{
			ID:        "demo_user",
			Email:     email,
			FirstName: "Demo",
			LastName:  "User",
			Role:      domain.RoleAdmin,
			IsActive:  true,
		}
	}

	// Create session
	token := generateToken()
	session := &domain.Session{
		ID:        generateToken(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession checks if a session is valid
func (s *AuthService) ValidateSession(token string) (*domain.User, error) {
	session := s.sessionRepo.GetByToken(token)
	if session == nil {
		return nil, errors.New("invalid session")
	}

	if session.IsExpired() {
		s.sessionRepo.Delete(token)
		return nil, errors.New("session expired")
	}

	user := s.userRepo.GetByID(session.UserID)
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// Logout invalidates a session
func (s *AuthService) Logout(token string) error {
	return s.sessionRepo.Delete(token)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
