package edutrack

import "gorm.io/gorm"

// Student represents a student enrolled in the institution.
type Student struct {
	gorm.Model

	// Unique student ID or registration number within the institution.
	StudentID string `gorm:"uniqueIndex:idx_student_tenant"`

	// The current semester of the student.
	Semester int `gorm:"not null;default:1"`

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

	// Subjects this student is attending.
	Subjects []Subject `gorm:"many2many:student_subjects;"`

	// Calculated fields (not stored in the database).

	// OverallAverage is the general average grade of the student across all subjects.
	OverallAverage float64 `gorm:"-"`

	// SubjectAverages holds the average grade for each subject.
	SubjectAverages []SubjectAverage `gorm:"-"`
}

// SubjectAverage is a helper struct to hold the average grade for a subject.
type SubjectAverage struct {
	SubjectID   uint    `json:"subjectId"`
	SubjectName string  `json:"subjectName"`
	Average     float64 `json:"average"`
}

// CalculateAverages computes the overall average and per-subject averages for the student.
func (s *Student) CalculateAverages(db *gorm.DB) {
	var grades []Grade
	// Fetch all grades for the student, preloading the topic and its subject.
	db.Preload("Topic.Subject").Where("student_id = ?", s.ID).Find(&grades)

	if len(grades) == 0 {
		s.OverallAverage = 0
		s.SubjectAverages = []SubjectAverage{}
		return
	}

	var totalSum float64
	subjectGrades := make(map[uint]struct {
		sum   float64
		count int
		name  string
	})

	for _, grade := range grades {
		totalSum += grade.Value

		// Ensure Topic and Subject are loaded to prevent panics
		if grade.Topic.Subject.ID != 0 {
			subjectData := subjectGrades[grade.Topic.SubjectID]
			subjectData.sum += grade.Value
			subjectData.count++
			subjectData.name = grade.Topic.Subject.Name
			subjectGrades[grade.Topic.SubjectID] = subjectData
		}
	}

	// Calculate overall average
	s.OverallAverage = totalSum / float64(len(grades))

	// Calculate per-subject averages
	s.SubjectAverages = make([]SubjectAverage, 0, len(subjectGrades))
	for subjectID, data := range subjectGrades {
		if data.count > 0 {
			s.SubjectAverages = append(s.SubjectAverages, SubjectAverage{
				SubjectID:   subjectID,
				SubjectName: data.name,
				Average:     data.sum / float64(data.count),
			})
		}
	}
}
