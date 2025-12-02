package edutrack

import (
	"time"

	"gorm.io/gorm"
)

// LicenseType represents the type of license.
type LicenseType string

const (
	// LicenseTypeTrial is a trial license with limited features.
	LicenseTypeTrial LicenseType = "trial"

	// LicenseTypeBasic is a basic license for small institutions.
	LicenseTypeBasic LicenseType = "basic"

	// LicenseTypePro is a professional license with more features.
	LicenseTypePro LicenseType = "pro"

	// LicenseTypeEnterprise is an enterprise license with all features.
	LicenseTypeEnterprise LicenseType = "enterprise"
)

// License represents a software license for an institution.
type License struct {
	gorm.Model

	// Unique license key.
	Key string `gorm:"uniqueIndex"`

	// Type of license (trial, basic, pro, enterprise).
	Type LicenseType `gorm:"default:'trial'"`

	// Expiration date of the license.
	ExpiryAt time.Time

	// Maximum number of users allowed under this license.
	MaxUsers int `gorm:"default:5"`

	// Maximum number of students allowed under this license.
	MaxStudents int `gorm:"default:50"`

	// Maximum number of courses allowed under this license.
	MaxCourses int `gorm:"default:10"`

	// Whether the license is currently active.
	Active bool `gorm:"default:true"`

	// Additional notes or comments about the license.
	Notes string
}

// IsExpired returns true if the license has expired.
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiryAt)
}

// IsValid returns true if the license is active and not expired.
func (l *License) IsValid() bool {
	return l.Active && !l.IsExpired()
}
