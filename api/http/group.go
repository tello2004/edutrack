package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

func (s *Server) handleListGroups(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.Where("tenant_id = ?", account.TenantID)

	if careerID := r.URL.Query().Get("career_id"); careerID != "" {
		query = query.Where("career_id = ?", careerID)
	}

	var groups []edutrack.Group
	if err := query.Find(&groups).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, groups)
}

func (s *Server) handleListGroupStudents(w http.ResponseWriter, r *http.Request) {
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

	var group edutrack.Group
	if err := s.DB.Preload("Students.Account").First(&group, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	sendJSON(w, http.StatusOK, group.Students)
}
