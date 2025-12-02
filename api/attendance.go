package edutrack

import (
	"time"

	"gorm.io/gorm"
)

// AttendanceStatus represents the attendance state of a student.
type AttendanceStatus string

const (
	// AttendancePresent indicates the student was present.
	AttendancePresent AttendanceStatus = "present"

	// AttendanceAbsent indicates the student was absent.
	AttendanceAbsent AttendanceStatus = "absent"

	// AttendanceLate indicates the student arrived late.
	AttendanceLate AttendanceStatus = "late"

	// AttendanceExcused indicates the student had an excused absence.
	AttendanceExcused AttendanceStatus = "excused"
)

// Attendance represents a daily attendance record for a student in a subject.
type Attendance struct {
	gorm.Model

	// Date of the attendance record.
	Date time.Time `gorm:"index"`

	// Status of the attendance (present, absent, late, excused).
	Status AttendanceStatus `gorm:"default:'absent'"`

	// Optional notes about the attendance.
	Notes string

	// Foreign keys.

	// StudentID links to the student.
	StudentID uint `gorm:"index"`
	Student   Student

	// SubjectID links to the subject/class.
	SubjectID uint `gorm:"index"`
	Subject   Subject

	// TenantID links the record to an institution.
	TenantID string `gorm:"index"`
	Tenant   Tenant
}
