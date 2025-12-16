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

// CreateGradeRequest represents the request body for creating a grade.
type CreateGradeRequest struct {
	Value     float64 `json:"value"`
	Notes     string  `json:"notes"`
	StudentID uint    `json:"student_id"`
	TopicID   uint    `json:"topic_id"`
}

// handleCreateGrade handles POST /grades.
func (s *Server) handleCreateGrade(w http.ResponseWriter, r *http.Request) {
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

	grade := &edutrack.Grade{
		Value:     req.Value,
		Notes:     req.Notes,
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

// UpdateGradeRequest represents the request body for updating a grade.
type UpdateGradeRequest struct {
	Value *float64 `json:"value"`
	Notes *string  `json:"notes"`
}

// handleUpdateGrade handles PUT /grades/{id}.
func (s *Server) handleUpdateGrade(w http.ResponseWriter, r *http.Request) {
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

	sendJSON(w, http.StatusOK, grade)
}

// handleDeleteGrade handles DELETE /grades/{id}.
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

	if err := s.DB.Delete(&grade).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
