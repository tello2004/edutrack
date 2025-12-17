package http

import (
	"net/http"
	"strconv"

	"gorm.io/gorm"
	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListGrades handles GET /grades.
func (s *Server) handleListGrades(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var query *gorm.DB
	if account.IsStudent() {
		// Students can only see their own grades.
		var student edutrack.Student
		if err := s.DB.Where("account_id = ? AND tenant_id = ?", account.ID, account.TenantID).First(&student).Error; err != nil {
			sendError(w, http.StatusNotFound, ErrNotFound)
			return
		}
		query = s.DB.Where("student_id = ? AND tenant_id = ?", student.ID, account.TenantID)
	} else {
		query = s.DB.Where("tenant_id = ?", account.TenantID)

		// Optional filters for teachers/secretaries.
		if studentID := r.URL.Query().Get("student_id"); studentID != "" {
			query = query.Where("student_id = ?", studentID)
		}
		if topicID := r.URL.Query().Get("topic_id"); topicID != "" {
			query = query.Where("topic_id = ?", topicID)
		}
	}

	var grades []edutrack.Grade
	if err := query.Preload("Student").Preload("Topic.Subject").Find(&grades).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, grades)
}

// handleGetGrade handles GET /grades/{id}.
func (s *Server) handleGetGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var grade edutrack.Grade
	if err := s.DB.Preload("Student").Preload("Topic.Subject").First(&grade, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if grade.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if account.IsStudent() {
		// Students can only access their own grades.
		var student edutrack.Student
		if err := s.DB.Where("account_id = ? AND tenant_id = ?", account.ID, account.TenantID).First(&student).Error; err != nil {
			sendError(w, http.StatusNotFound, ErrNotFound)
			return
		}
		if grade.StudentID != student.ID {
			sendError(w, http.StatusForbidden, ErrForbidden)
			return
		}
	}

	sendJSON(w, http.StatusOK, grade)
}

/*
	CREAR O ACTUALIZAR CALIFICACIÓN
	(1 alumno, 1 materia, 1 periodo)
*/
type CreateOrUpdateGradeRequest struct {
	StudentID uint    `json:"student_id"`
	TopicID   uint    `json:"topic_id"`
}

func (s *Server) handleCreateOrUpdateGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Only teachers and secretaries can create grades.
	if account.IsStudent() {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req CreateGradeRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.StudentID == 0 || req.TopicID == 0 {
		sendErrorMessage(w, http.StatusBadRequest, "El estudiante y el tema son requeridos.")
		return
	}

	// Verify topic exists and belongs to the same tenant.
	var topic edutrack.Topic
	if err := s.DB.First(&topic, req.TopicID).Error; err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "El tema especificado no existe.")
		return
	}
	if topic.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	grade = edutrack.Grade{
		StudentID: req.StudentID,
		TopicID:   req.TopicID,
		TenantID:  account.TenantID,
	}

	if err := s.DB.Create(grade).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Student").Preload("Topic.Subject").First(grade, grade.ID)

	sendJSON(w, http.StatusCreated, grade)
}

/*
	PROMEDIO GENERAL POR ALUMNO
	Filtros:
	- group_id
*/
type StudentAverageResult struct {
	StudentID uint    `json:"student_id"`
	Name      string  `json:"name"`
	Average   float64 `json:"average"`
}

func (s *Server) handleStudentAverage(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Only teachers and secretaries can update grades.
	if account.IsStudent() {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var grade edutrack.Grade
	if err := s.DB.First(&grade, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if grade.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateGradeRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Value != nil {
		grade.Value = *req.Value
	}
	if req.Notes != nil {
		grade.Notes = *req.Notes
	}

	if err := s.DB.Save(&grade).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Student").Preload("Topic.Subject").First(&grade, grade.ID)

	sendJSON(w, http.StatusOK, result)
}

/*
	ELIMINAR CALIFICACIÓN
*/
func (s *Server) handleDeleteGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Only teachers and secretaries can delete grades.
	if account.IsStudent() {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var grade edutrack.Grade
	if err := s.DB.First(&grade, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if grade.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	s.DB.Delete(&grade)
	w.WriteHeader(http.StatusNoContent)
}

/*
package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListGrades handles GET /grades.
func (s *Server) handleListGrades(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.
		Where("grades.tenant_id = ?", account.TenantID).
		Preload("Student.Account").
		Preload("Subject.Teacher.Account").
		Preload("Group")

	if studentID := r.URL.Query().Get("student_id"); studentID != "" {
		query = query.Where("grades.student_id = ?", studentID)
	}

	if subjectID := r.URL.Query().Get("subject_id"); subjectID != "" {
		query = query.Where("grades.subject_id = ?", subjectID)
	}

	if groupID := r.URL.Query().Get("group_id"); groupID != "" {
		query = query.Where("grades.group_id = ?", groupID)
	}

	var grades []edutrack.Grade
	if err := query.Find(&grades).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, grades)
}

// handleGetGrade handles GET /grades/{id}.
func (s *Server) handleGetGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var grade edutrack.Grade
	if err := s.DB.
		Preload("Student.Account").
		Preload("Subject.Teacher.Account").
		Preload("Group").
		First(&grade, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if grade.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, grade)
}

type CreateOrUpdateGradeRequest struct {
	StudentID uint    `json:"student_id"`
	SubjectID uint    `json:"subject_id"`
	GroupID   uint    `json:"group_id"`
	Period    int     `json:"period"`
	Value     float64 `json:"value"`
}

func (s *Server) handleCreateOrUpdateGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	if !account.IsSecretary() && !account.IsTeacher() {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req CreateOrUpdateGradeRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.StudentID == 0 || req.SubjectID == 0 || req.GroupID == 0 || req.Period <= 0 {
		sendErrorMessage(w, http.StatusBadRequest, "Datos incompletos")
		return
	}

	var grade edutrack.Grade
	err := s.DB.
		Where(
			"student_id = ? AND subject_id = ? AND period = ? AND tenant_id = ?",
			req.StudentID, req.SubjectID, req.Period, account.TenantID,
		).
		First(&grade).Error

	if err == nil {
		grade.Value = req.Value
		s.DB.Save(&grade)
		sendJSON(w, http.StatusOK, grade)
		return
	}

	grade = edutrack.Grade{
		StudentID: req.StudentID,
		SubjectID: req.SubjectID,
		GroupID:   req.GroupID,
		Period:    req.Period,
		Value:     req.Value,
		TenantID:  account.TenantID,
	}

	s.DB.Create(&grade)
	sendJSON(w, http.StatusCreated, grade)
}

type StudentAverageResult struct {
	StudentID uint    `json:"student_id"`
	Name      string  `json:"name"`
	Average   float64 `json:"average"`
}

func (s *Server) handleStudentAverage(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("group_id")

	query := s.DB.
		Table("grades").
		Select(`
			students.id AS student_id,
			accounts.name AS name,
			AVG(grades.value) AS average
		`).
		Joins("JOIN students ON students.id = grades.student_id").
		Joins("JOIN accounts ON accounts.id = students.account_id").
		Where("grades.tenant_id = ?", account.TenantID).
		Group("students.id, accounts.name")

	if groupID != "" {
		query = query.Where("grades.group_id = ?", groupID)
	}

	var result []StudentAverageResult
	query.Scan(&result)

	sendJSON(w, http.StatusOK, result)
}

func (s *Server) handleDeleteGrade(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var grade edutrack.Grade
	if err := s.DB.First(&grade, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if grade.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	s.DB.Delete(&grade)
	w.WriteHeader(http.StatusNoContent)
}
*/
