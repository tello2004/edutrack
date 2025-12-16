package edutrack

import "gorm.io/gorm"

// Topic represents a specific topic within a subject, created by a teacher.
// Grades are given based on these topics.
type Topic struct {
	gorm.Model

	// Name of the topic (e.g., "Unit 1: Derivatives", "Final Project").
	Name string

	// Description of the topic.
	Description string

	// Foreign keys.

	// SubjectID links the topic to a subject.
	SubjectID uint `gorm:"index"`
	Subject   Subject

	// TenantID links the topic to an institution.
	TenantID string `gorm:"index"`
	Tenant   Tenant

	// Grades associated with this topic.
	Grades []Grade
}
