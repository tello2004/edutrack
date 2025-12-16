package edutrack

import "gorm.io/gorm"

// Grade represents a student's grade for a specific topic within a subject.
type Grade struct {
	gorm.Model

	// The numeric grade value (e.g., 0-100 or 0-10).
	Value float64

	// Optional description or notes about the grade.
	Notes string

	// Foreign keys.

	// The student who received this grade.
	StudentID uint
	Student   Student

	// The topic this grade is for.
	TopicID uint
	Topic   Topic

	// TenantID for multi-tenant support.
	TenantID string
	Tenant   Tenant
}
