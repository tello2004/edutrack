package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListTeachers handles GET /teachers.
func (s *Server) handleListTeachers(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.Where("tenant_id = ?", account.TenantID).Preload("Account")

	// Optional filters.
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Joins("JOIN accounts ON accounts.id = teachers.account_id").
			Where("accounts.name LIKE ?", "%"+name+"%")
	}
	if accountID := r.URL.Query().Get("account_id"); accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	var teachers []edutrack.Teacher
	if err := query.Find(&teachers).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, teachers)
}

// handleGetTeacher handles GET /teachers/{id}.
func (s *Server) handleGetTeacher(w http.ResponseWriter, r *http.Request) {
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

	var teacher edutrack.Teacher
	if err := s.DB.Preload("Account").First(&teacher, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if teacher.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, teacher)
}

// CreateTeacherRequest represents the request body for creating a teacher.
type CreateTeacherRequest struct {
	AccountID uint `json:"account_id"`
}

// handleCreateTeacher handles POST /teachers.
func (s *Server) handleCreateTeacher(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateTeacherRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.AccountID == 0 {
		sendErrorMessage(w, http.StatusBadRequest, "El ID de cuenta es requerido.")
		return
	}

	// Verify the account exists and belongs to the same tenant.
	var linkedAccount edutrack.Account
	if err := s.DB.First(&linkedAccount, req.AccountID).Error; err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "La cuenta especificada no existe.")
		return
	}

	if linkedAccount.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	teacher := &edutrack.Teacher{
		AccountID: req.AccountID,
		TenantID:  account.TenantID,
	}

	if err := s.DB.Create(teacher).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Load the account relationship for the response.
	s.DB.Preload("Account").First(teacher, teacher.ID)

	sendJSON(w, http.StatusCreated, teacher)
}

// UpdateTeacherRequest represents the request body for updating a teacher.
type UpdateTeacherRequest struct {
	AccountID *uint `json:"account_id"`
}

// handleUpdateTeacher handles PUT /teachers/{id}.
func (s *Server) handleUpdateTeacher(w http.ResponseWriter, r *http.Request) {
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

	var teacher edutrack.Teacher
	if err := s.DB.First(&teacher, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if teacher.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateTeacherRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.AccountID != nil {
		// Verify the new account exists and belongs to the same tenant.
		var linkedAccount edutrack.Account
		if err := s.DB.First(&linkedAccount, *req.AccountID).Error; err != nil {
			sendErrorMessage(w, http.StatusBadRequest, "La cuenta especificada no existe.")
			return
		}

		if linkedAccount.TenantID != account.TenantID {
			sendError(w, http.StatusForbidden, ErrForbidden)
			return
		}

		teacher.AccountID = *req.AccountID
	}

	if err := s.DB.Save(&teacher).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Load the account relationship for the response.
	s.DB.Preload("Account").First(&teacher, teacher.ID)

	sendJSON(w, http.StatusOK, teacher)
}

// handleDeleteTeacher handles DELETE /teachers/{id}.
func (s *Server) handleDeleteTeacher(w http.ResponseWriter, r *http.Request) {
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

	var teacher edutrack.Teacher
	if err := s.DB.First(&teacher, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if teacher.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&teacher).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
