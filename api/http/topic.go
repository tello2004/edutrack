package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListTopics handles GET /topics.
func (s *Server) handleListTopics(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.Where("tenant_id = ?", account.TenantID)

	// Optional filters.
	if subjectID := r.URL.Query().Get("subject_id"); subjectID != "" {
		query = query.Where("subject_id = ?", subjectID)
	}

	var topics []edutrack.Topic
	if err := query.Preload("Subject").Find(&topics).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, topics)
}

// handleGetTopic handles GET /topics/{id}.
func (s *Server) handleGetTopic(w http.ResponseWriter, r *http.Request) {
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

	var topic edutrack.Topic
	if err := s.DB.Preload("Subject").First(&topic, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if topic.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, topic)
}

// CreateTopicRequest represents the request body for creating a topic.
type CreateTopicRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SubjectID   uint   `json:"subject_id"`
}

// handleCreateTopic handles POST /topics.
func (s *Server) handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateTopicRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.SubjectID == 0 || req.Name == "" {
		sendErrorMessage(w, http.StatusBadRequest, "El nombre y el tema son requeridos.")
		return
	}

	// Verify subject exists and belongs to the same tenant.
	var subject edutrack.Subject
	if err := s.DB.First(&subject, req.SubjectID).Error; err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "El tema especificado no existe.")
		return
	}
	if subject.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	topic := &edutrack.Topic{
		Name:        req.Name,
		Description: req.Description,
		SubjectID:   req.SubjectID,
		TenantID:    account.TenantID,
	}

	if err := s.DB.Create(topic).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Subject").First(topic, topic.ID)

	sendJSON(w, http.StatusCreated, topic)
}

// UpdateTopicRequest represents the request body for updating a topic.
type UpdateTopicRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// handleUpdateTopic handles PUT /topics/{id}.
func (s *Server) handleUpdateTopic(w http.ResponseWriter, r *http.Request) {
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

	var topic edutrack.Topic
	if err := s.DB.First(&topic, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if topic.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateTopicRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name != nil {
		topic.Name = *req.Name
	}
	if req.Description != nil {
		topic.Description = *req.Description
	}

	if err := s.DB.Save(&topic).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Subject").First(&topic, topic.ID)

	sendJSON(w, http.StatusOK, topic)
}

// handleDeleteTopic handles DELETE /topics/{id}.
func (s *Server) handleDeleteTopic(w http.ResponseWriter, r *http.Request) {
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

	var topic edutrack.Topic
	if err := s.DB.First(&topic, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if topic.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&topic).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
