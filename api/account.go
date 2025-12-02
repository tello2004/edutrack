package edutrack

import "gorm.io/gorm"

// Account represents a user account for authentication.
type Account struct {
	gorm.Model

	// The user's full name.
	Name string

	// The user's email address (used for login).
	Email string `gorm:"uniqueIndex"`

	// The hashed password.
	Password string

	// Whether the account is active.
	Active bool `gorm:"default:true"`

	// Foreign keys.

	// TenantID links the account to an institution.
	TenantID string
	Tenant   Tenant
}
