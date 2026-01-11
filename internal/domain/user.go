package domain

import "time"

// Role represents a user's role in the system
type Role string

const (
	RoleAdmin          Role = "admin"           // Full system access
	RoleCommissionCounsel Role = "commission_counsel" // All case access, publishing
	RoleStaffAttorney  Role = "staff_attorney"  // Assigned case access
	RoleInvestigator   Role = "investigator"    // Complaint investigation
	RoleAdminStaff     Role = "admin_staff"     // Case intake, PRR handling
	RoleReadOnly       Role = "readonly"        // View only
	RoleAuditor        Role = "auditor"         // Audit logs only
)

// User represents a staff user in the system
type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	Title        string
	Phone        string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// CanManageCases returns true if the user can create/edit cases
func (u *User) CanManageCases() bool {
	switch u.Role {
	case RoleAdmin, RoleCommissionCounsel, RoleStaffAttorney, RoleInvestigator, RoleAdminStaff:
		return true
	}
	return false
}

// CanPublish returns true if the user can publish opinions/orders
func (u *User) CanPublish() bool {
	return u.Role == RoleAdmin || u.Role == RoleCommissionCounsel
}

// CanViewAuditLogs returns true if the user can view audit logs
func (u *User) CanViewAuditLogs() bool {
	return u.Role == RoleAdmin || u.Role == RoleAuditor
}

// CanManageUsers returns true if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

// Session represents a user session
type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
