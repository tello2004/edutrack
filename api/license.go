package edutrack

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

	// Unique license key (used for institutional login).
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

// GenerateLicenseKey generates a new unique license key.
// Format: XXXX-XXXX-XXXX-XXXX (16 hex characters with dashes).
func GenerateLicenseKey() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate license key: %w", err)
	}

	hex := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s-%s-%s-%s",
		hex[0:4],
		hex[4:8],
		hex[8:12],
		hex[12:16],
	), nil
}

// NewLicense creates a new license with default values.
func NewLicense(licenseType LicenseType, duration time.Duration) (*License, error) {
	key, err := GenerateLicenseKey()
	if err != nil {
		return nil, err
	}

	license := &License{
		Key:      key,
		Type:     licenseType,
		ExpiryAt: time.Now().Add(duration),
		Active:   true,
	}

	// Set limits based on license type.
	switch licenseType {
	case LicenseTypeTrial:
		license.MaxUsers = 3
		license.MaxStudents = 25
		license.MaxCourses = 5
	case LicenseTypeBasic:
		license.MaxUsers = 10
		license.MaxStudents = 100
		license.MaxCourses = 20
	case LicenseTypePro:
		license.MaxUsers = 50
		license.MaxStudents = 500
		license.MaxCourses = 100
	case LicenseTypeEnterprise:
		license.MaxUsers = -1 // Unlimited
		license.MaxStudents = -1
		license.MaxCourses = -1
	}

	return license, nil
}

// Regenerate generates a new key for an existing license and optionally extends it.
func (l *License) Regenerate(extendDuration time.Duration) error {
	key, err := GenerateLicenseKey()
	if err != nil {
		return err
	}

	l.Key = key
	l.Active = true

	if extendDuration > 0 {
		// Extend from now or from expiry date, whichever is later.
		baseTime := time.Now()
		if l.ExpiryAt.After(baseTime) {
			baseTime = l.ExpiryAt
		}
		l.ExpiryAt = baseTime.Add(extendDuration)
	}

	return nil
}

// DaysUntilExpiry returns the number of days until the license expires.
// Returns 0 if already expired.
func (l *License) DaysUntilExpiry() int {
	if l.IsExpired() {
		return 0
	}
	return int(time.Until(l.ExpiryAt).Hours() / 24)
}
