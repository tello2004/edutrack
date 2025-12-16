package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListSubjects handles GET /subjects.
func (s *Server) handleListSubjects(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.Where("tenant_id = ?", account.TenantID)

	// Optional filters.
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code := r.URL.Query().Get("code"); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if teacherID := r.URL.Query().Get("teacher_id"); teacherID != "" {
		query = query.Where("teacher_id = ?", teacherID)
	}
	if careerID := r.URL.Query().Get("career_id"); careerID != "" {
		query = query.Where("career_id = ?", careerID)
	}
	if semester := r.URL.Query().Get("semester"); semester != "" {
		query = query.Where("semester = ?", semester)
	}

	var subjects []edutrack.Subject
	if err := query.Preload("Teacher").Preload("Career").Find(&subjects).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, subjects)
}

// handleGetSubject handles GET /subjects/{id}.
func (s *Server) handleGetSubject(w http.ResponseWriter, r *http.Request) {
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

	var subject edutrack.Subject
	if err := s.DB.Preload("Teacher").Preload("Career").First(&subject, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if subject.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, subject)
}

// CreateSubjectRequest represents the request body for creating a subject.
type CreateSubjectRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Credits     int    `json:"credits"`
	TeacherID   *uint  `json:"teacher_id"`
	CareerID    uint   `json:"career_id"`
	Semester    int    `json:"semester"`
}

// handleCreateSubject handles POST /subjects.
func (s *Server) handleCreateSubject(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateSubjectRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name == "" || req.Code == "" || req.CareerID == 0 || req.Semester <= 0 {
		sendErrorMessage(w, http.StatusBadRequest, "Nombre, código, ID de carrera y un semestre positivo son requeridos.")
		return
	}

	subject := &edutrack.Subject{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Credits:     req.Credits,
		TeacherID:   req.TeacherID,
		CareerID:    req.CareerID,
		Semester:    req.Semester,
		TenantID:    account.TenantID,
	}

	if err := s.DB.Create(subject).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Teacher").Preload("Career").First(subject, subject.ID)

	sendJSON(w, http.StatusCreated, subject)
}

// UpdateSubjectRequest represents the request body for updating a subject.
type UpdateSubjectRequest struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Credits     *int    `json:"credits"`
	TeacherID   *uint   `json:"teacher_id"`
	CareerID    *uint   `json:"career_id"`
	Semester    *int    `json:"semester"`
}

// handleUpdateSubject handles PUT /subjects/{id}.
func (s *Server) handleUpdateSubject(w http.ResponseWriter, r *http.Request) {
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

	var subject edutrack.Subject
	if err := s.DB.First(&subject, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if subject.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateSubjectRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name != nil {
		subject.Name = *req.Name
	}
	if req.Code != nil {
		subject.Code = *req.Code
	}
	if req.Description != nil {
		subject.Description = *req.Description
	}
	if req.Credits != nil {
		subject.Credits = *req.Credits
	}
	if req.TeacherID != nil {
		subject.TeacherID = req.TeacherID
	}
	if req.CareerID != nil {
		subject.CareerID = *req.CareerID
	}
	if req.Semester != nil {
		if *req.Semester <= 0 {
			sendErrorMessage(w, http.StatusBadRequest, "El semestre debe ser un número positivo.")
			return
		}
		subject.Semester = *req.Semester
	}

	if err := s.DB.Save(&subject).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Teacher").Preload("Career").First(&subject, subject.ID)

	sendJSON(w, http.StatusOK, subject)
}

// handleDeleteSubject handles DELETE /subjects/{id}.
func (s *Server) handleDeleteSubject(w http.ResponseWriter, r *http.Request) {
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

	var subject edutrack.Subject
	if err := s.DB.First(&subject, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if subject.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&subject).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
