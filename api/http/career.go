package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListCareers handles GET /careers.
func (s *Server) handleListCareers(w http.ResponseWriter, r *http.Request) {
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
	if active := r.URL.Query().Get("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	var careers []edutrack.Career
	if err := query.Find(&careers).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, careers)
}

// handleGetCareer handles GET /careers/{id}.
func (s *Server) handleGetCareer(w http.ResponseWriter, r *http.Request) {
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

	var career edutrack.Career
	if err := s.DB.First(&career, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if career.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, career)
}

// CreateCareerRequest represents the request body for creating a career.
type CreateCareerRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
}

// handleCreateCareer handles POST /careers.
func (s *Server) handleCreateCareer(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateCareerRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name == "" || req.Code == "" {
		sendErrorMessage(w, http.StatusBadRequest, "El nombre y código son requeridos.")
		return
	}

	career := &edutrack.Career{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Duration:    req.Duration,
		Active:      true,
		TenantID:    account.TenantID,
	}

	if err := s.DB.Create(career).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusCreated, career)
}

// UpdateCareerRequest represents the request body for updating a career.
type UpdateCareerRequest struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Duration    *int    `json:"duration"`
	Active      *bool   `json:"active"`
}

// handleUpdateCareer handles PUT /careers/{id}.
func (s *Server) handleUpdateCareer(w http.ResponseWriter, r *http.Request) {
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

	var career edutrack.Career
	if err := s.DB.First(&career, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if career.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateCareerRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name != nil {
		career.Name = *req.Name
	}
	if req.Code != nil {
		career.Code = *req.Code
	}
	if req.Description != nil {
		career.Description = *req.Description
	}
	if req.Duration != nil {
		career.Duration = *req.Duration
	}
	if req.Active != nil {
		career.Active = *req.Active
	}

	if err := s.DB.Save(&career).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, career)
}

// handleDeleteCareer handles DELETE /careers/{id}.
func (s *Server) handleDeleteCareer(w http.ResponseWriter, r *http.Request) {
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

	var career edutrack.Career
	if err := s.DB.First(&career, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if career.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&career).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
