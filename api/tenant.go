package edutrack

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Tenant represents an institute in the system.
type Tenant struct {
	// ID is a unique identifier for the tenant (8 character hex string).
	ID string `gorm:"primarykey;size:8"`

	// The name of the institute.
	Name string

	// The institute's logo URL.
	LogoURL string

	// License linked to the institute.
	License   License
	LicenseID uint

	// Built-ins, extracted from gorm.Model.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// GenerateTenantID generates a new unique tenant ID.
// Format: 8 character hex string.
func GenerateTenantID() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate tenant ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// NewTenant creates a new tenant with a generated ID and associated license.
func NewTenant(name string, licenseType LicenseType, licenseDuration time.Duration) (*Tenant, error) {
	id, err := GenerateTenantID()
	if err != nil {
		return nil, err
	}

	license, err := NewLicense(licenseType, licenseDuration)
	if err != nil {
		return nil, err
	}

	return &Tenant{
		ID:      id,
		Name:    name,
		License: *license,
	}, nil
}

// IsActive returns true if the tenant's license is valid.
func (t *Tenant) IsActive() bool {
	return t.License.IsValid()
}

// GetLicenseKey returns the license key for institutional login.
func (t *Tenant) GetLicenseKey() string {
	return t.License.Key
}
