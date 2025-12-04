package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListAccounts handles GET /accounts.
func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
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
	if email := r.URL.Query().Get("email"); email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if active := r.URL.Query().Get("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	var accounts []edutrack.Account
	if err := query.Find(&accounts).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, accounts)
}

// handleGetAccount handles GET /accounts/{id}.
func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
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

	var found edutrack.Account
	if err := s.DB.First(&found, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	// Ensure the account belongs to the same tenant.
	if found.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, found)
}

// CreateAccountRequest represents the request body for creating an account.
type CreateAccountRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleCreateAccount handles POST /accounts.
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		sendErrorMessage(w, http.StatusBadRequest, "Nombre, email y contraseña son requeridos.")
		return
	}

	// Hash the password.
	hashedPassword, err := edutrack.HashPassword(req.Password)
	if err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	newAccount := &edutrack.Account{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Active:   true,
		TenantID: account.TenantID,
	}

	if err := s.DB.Create(newAccount).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusCreated, newAccount)
}

// UpdateAccountRequest represents the request body for updating an account.
type UpdateAccountRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Active   *bool   `json:"active"`
}

// handleUpdateAccount handles PUT /accounts/{id}.
func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
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

	var existing edutrack.Account
	if err := s.DB.First(&existing, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if existing.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Password != nil {
		hashedPassword, err := edutrack.HashPassword(*req.Password)
		if err != nil {
			sendError(w, http.StatusInternalServerError, ErrInternalServer)
			return
		}
		existing.Password = hashedPassword
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}

	if err := s.DB.Save(&existing).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, existing)
}

// handleDeleteAccount handles DELETE /accounts/{id}.
func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
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

	var existing edutrack.Account
	if err := s.DB.First(&existing, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if existing.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	// Prevent self-deletion.
	if existing.ID == account.ID {
		sendErrorMessage(w, http.StatusBadRequest, "No puedes eliminar tu propia cuenta.")
		return
	}

	if err := s.DB.Delete(&existing).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
