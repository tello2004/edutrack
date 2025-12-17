package edutrack

import "gorm.io/gorm"

// Grade represents a student's grade for a specific topic within a subject.
type Grade struct {
	gorm.Model
	Value     float64
	Period    int

	StudentID uint
	Student   Student

	// The subject this grade is for.
	SubjectID uint
	Subject   Subject

	// The teacher who assigned this grade.
	TeacherID uint
	Teacher   Teacher

	TenantID string
}

/*
type Grade struct {
	gorm.Model
	Value     float64
	Period    int

	StudentID uint
	Student   Student

	SubjectID uint
	Subject   Subject

	GroupID uint
	Group   Group

	TenantID string
}
*/
