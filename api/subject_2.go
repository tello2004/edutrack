type Subject struct {
	gorm.Model
	Name        string
	Code        string
	Description string
	Credits     int

	TeacherID *uint
	Teacher   *Teacher

	CareerID uint
	Career   Career

	Semester int
	TenantID string

	Students []Student `gorm:"many2many:subject_students;"`
	Grades   []Grade
}
