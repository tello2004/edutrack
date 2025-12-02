package edutrack

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents an institute in the system.
type Tenant struct {
	ID string `gorm:"primarykey"`

	// The name of the institute.
	Name string

	// The institute's Logo
	LogoURL string

	// License linked to the institute.
	License License

	// Foreign keys.

	LicenseID uint

	// Built ins, extracted from gorm.Model.

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
