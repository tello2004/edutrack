package edutrack

import "gorm.io/gorm"

// Subject represents a subject or course unit in the system.
// Subjects are linked to one or more careers.
type Subject struct {
	gorm.Model

	// Name of the subject (e.g., "Matemáticas I", "Programación").
	Name string

	// Unique code for the subject (e.g., "MAT101").
	Code string `gorm:"uniqueIndex:idx_subject_tenant"`

	// Description of the subject content.
	Description string

	// Number of credits or hours for this subject.
	Credits int

	// Foreign keys.

	// TenantID links the subject to an institution.
	TenantID string `gorm:"uniqueIndex:idx_subject_tenant"`
	Tenant   Tenant

	// Relationships.

	// Careers this subject belongs to (many-to-many).
	Careers []Career `gorm:"many2many:career_subjects;"`

	// Teacher assigned to this subject.
	TeacherID *uint
	Teacher   *Teacher
}
