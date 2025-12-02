package edutrack

import "gorm.io/gorm"

// Student represents a student enrolled in the institution.
type Student struct {
	gorm.Model

	// Unique student ID or registration number within the institution.
	StudentID string `gorm:"uniqueIndex:idx_student_tenant"`

	// Foreign keys.

	// TenantID links the student to an institution.
	TenantID string `gorm:"uniqueIndex:idx_student_tenant"`
	Tenant   Tenant

	// AccountID links the student to their login account.
	AccountID uint
	Account   Account

	// CareerID links the student to the career they are enrolled in.
	CareerID uint
	Career   Career
}
