package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

type AssignSubjectsRequest struct {
	SubjectIDs []uint `json:"subject_ids"`
}

func (s *Server) handleAssignSubjectsToGroup(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	groupID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var group edutrack.Group
	if err := s.DB.Preload("Subjects").First(&group, groupID).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if group.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req AssignSubjectsRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var subjects []edutrack.Subject
	if err := s.DB.
		Where("id IN ? AND tenant_id = ?", req.SubjectIDs, account.TenantID).
		Find(&subjects).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	if err := s.DB.Model(&group).Association("Subjects").Replace(subjects); err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, subjects)
}

func (s *Server) handleListGroupSubjects(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	groupID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var group edutrack.Group
	if err := s.DB.Preload("Subjects").First(&group, groupID).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if group.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, group.Subjects)
}
