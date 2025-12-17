package edutrack

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name     string
	CareerID uint
	Career   Career
	Semester int
	TenantID string

	Subjects []Subject `gorm:"many2many:group_subjects;"`
	Students []Student
}
