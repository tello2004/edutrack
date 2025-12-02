package edutrack

import "gorm.io/gorm"

// Teacher represents an instructor in the system.
// Personal information (name, email, etc.) is stored in the linked Account.
type Teacher struct {
	gorm.Model

	// The institution this teacher belongs to.
	TenantID string
	Tenant   Tenant

	// The account used for authentication and personal info.
	AccountID uint
	Account   Account

	// Subjects this teacher can teach.
	Subjects []Subject `gorm:"many2many:teacher_subjects;"`
}
