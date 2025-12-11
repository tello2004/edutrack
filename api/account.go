package edutrack

import "gorm.io/gorm"

// Role represents the type of user in the system.
type Role string

const (
	// RoleSecretary is a secretary/admin who can manage students and administrative tasks.
	RoleSecretary Role = "secretary"

	// RoleTeacher is a teacher who can manage academic tasks (grades, attendance).
	RoleTeacher Role = "teacher"
)

// Account represents a user account for authentication.
type Account struct {
	gorm.Model

	// The user's full name.
	Name string

	// The user's email address (used for login).
	Email string `gorm:"uniqueIndex:idx_account_email_tenant"`

	// The hashed password.
	Password string

	// The role of the user in the system (secretary or teacher).
	Role Role `gorm:"default:'teacher'"`

	// Whether the account is active.
	Active bool `gorm:"default:true"`

	// Foreign keys.

	// TenantID links the account to an institution.
	TenantID string `gorm:"uniqueIndex:idx_account_email_tenant"`
	Tenant   Tenant
}

// IsSecretary returns true if the account has secretary role.
func (a *Account) IsSecretary() bool {
	return a.Role == RoleSecretary
}

// IsTeacher returns true if the account has teacher role.
func (a *Account) IsTeacher() bool {
	return a.Role == RoleTeacher
}
