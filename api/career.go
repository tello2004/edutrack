package edutrack

import "gorm.io/gorm"

// Career represents an academic program or degree (e.g., "Ingeniería en Sistemas").
// Students enroll in a career and take subjects associated with it.
type Career struct {
	gorm.Model

	// Name of the career (e.g., "Licenciatura en Administración").
	Name string

	// Unique code for the career (e.g., "LAD-2024").
	Code string `gorm:"uniqueIndex:idx_career_tenant"`

	// Description of the career.
	Description string

	// Duration in semesters or periods.
	Duration int

	// Whether the career is currently active.
	Active bool `gorm:"default:true"`

	// Foreign keys.

	TenantID string `gorm:"uniqueIndex:idx_career_tenant"`
	Tenant   Tenant

	// Associations.

	// Subjects that belong to this career.
	Subjects []Subject `gorm:"many2many:career_subjects;"`

	// Students enrolled in this career.
	Students []Student
}
