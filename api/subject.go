package edutrack

import "gorm.io/gorm"

// Subject represents a subject or course unit in the system.
// A subject belongs to a specific career and semester.
type Subject struct {
	gorm.Model

	// Name of the subject (e.g., "Matemáticas I", "Programación").
	Name string

	// Unique code for the subject within a career and tenant.
	Code string `gorm:"uniqueIndex:idx_subject_code_career_tenant"`

	// Description of the subject content.
	Description string

	// Number of credits or hours for this subject.
	Credits int

	// The semester this subject is taught in.
	Semester int `gorm:"not null;default:1"`

	// Foreign keys.

	// TenantID links the subject to an institution.
	TenantID string `gorm:"uniqueIndex:idx_subject_code_career_tenant"`
	Tenant   Tenant

	// Career this subject belongs to.
	CareerID uint `gorm:"uniqueIndex:idx_subject_code_career_tenant"`
	Career   Career

	// Teacher assigned to this subject.
	TeacherID *uint
	Teacher   *Teacher

	// Students attending this subject.
	Students []Student `gorm:"many2many:student_subjects;"`

	// Topics that belong to this subject.
	Topics []Topic
}
