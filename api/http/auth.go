package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// Claims represents the JWT claims.
type Claims struct {
	AccountID uint          `json:"account_id"`
	TenantID  string        `json:"tenant_id"`
	Role      edutrack.Role `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response body.
type LoginResponse struct {
	Token string            `json:"token"`
	Role  edutrack.Role     `json:"role"`
	User  LoginResponseUser `json:"user"`
}

// LoginResponseUser represents user info in login response.
type LoginResponseUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// LicenseLoginRequest represents the license login request body.
type LicenseLoginRequest struct {
	LicenseKey string `json:"license_key"`
}

// LicenseLoginResponse represents the license login response body.
type LicenseLoginResponse struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Message    string `json:"message"`
}

// handleLogin handles POST /auth/login.
// This is for secretaries and teachers to log in with email/password.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		sendErrorMessage(w, http.StatusBadRequest, "Email y contraseña son requeridos.")
		return
	}

	// Find the account by email.
	var account edutrack.Account
	if err := s.DB.Preload("Tenant").Preload("Tenant.License").Where("email = ?", req.Email).First(&account).Error; err != nil {
		sendError(w, http.StatusUnauthorized, ErrInvalidCredentials)
		return
	}

	// Check if the account is active.
	if !account.Active {
		sendErrorMessage(w, http.StatusUnauthorized, "La cuenta está desactivada.")
		return
	}

	// Check if the tenant's license is valid.
	if !account.Tenant.License.IsValid() {
		sendErrorMessage(w, http.StatusUnauthorized, "La licencia de la institución ha expirado.")
		return
	}

	// Verify password.
	if !edutrack.PasswordMatches(req.Password, account.Password) {
		sendError(w, http.StatusUnauthorized, ErrInvalidCredentials)
		return
	}

	// Generate JWT token.
	token, err := s.generateToken(&account)
	if err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		Role:  account.Role,
		User: LoginResponseUser{
			ID:    account.ID,
			Name:  account.Name,
			Email: account.Email,
		},
	})
}

// handleLicenseLogin handles POST /auth/license.
// This is for institutional login using the license key.
// It validates the license and returns tenant info.
// After this, a secretary account must be created or used for further operations.
func (s *Server) handleLicenseLogin(w http.ResponseWriter, r *http.Request) {
	var req LicenseLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.LicenseKey == "" {
		sendErrorMessage(w, http.StatusBadRequest, "La llave de licencia es requerida.")
		return
	}

	// Find the license by key.
	var license edutrack.License
	if err := s.DB.Where("key = ?", req.LicenseKey).First(&license).Error; err != nil {
		sendErrorMessage(w, http.StatusUnauthorized, "Llave de licencia inválida.")
		return
	}

	// Check if the license is valid.
	if !license.IsValid() {
		if license.IsExpired() {
			sendErrorMessage(w, http.StatusUnauthorized, "La licencia ha expirado.")
		} else {
			sendErrorMessage(w, http.StatusUnauthorized, "La licencia está desactivada.")
		}
		return
	}

	// Find the tenant associated with this license.
	var tenant edutrack.Tenant
	if err := s.DB.Where("license_id = ?", license.ID).First(&tenant).Error; err != nil {
		sendErrorMessage(w, http.StatusUnauthorized, "No se encontró la institución asociada a esta licencia.")
		return
	}

	// Check if there's at least one secretary account for this tenant.
	var secretaryCount int64
	s.DB.Model(&edutrack.Account{}).Where("tenant_id = ? AND role = ?", tenant.ID, edutrack.RoleSecretary).Count(&secretaryCount)

	message := "Licencia válida. Inicie sesión con su cuenta."
	if secretaryCount == 0 {
		message = "Licencia válida. Es necesario crear una cuenta de secretario."
	}

	sendJSON(w, http.StatusOK, LicenseLoginResponse{
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		Message:    message,
	})
}

// generateToken generates a JWT token for the given account.
func (s *Server) generateToken(account *edutrack.Account) (string, error) {
	claims := &Claims{
		AccountID: account.ID,
		TenantID:  account.TenantID,
		Role:      account.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}

// withAuth is a middleware that validates the JWT token and adds the account to the context.
func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		// Check for Bearer prefix.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate the token.
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return s.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		// Load the account from the database.
		var account edutrack.Account
		if err := s.DB.Preload("Tenant").Preload("Tenant.License").First(&account, claims.AccountID).Error; err != nil {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		// Check if the account is still active.
		if !account.Active {
			sendErrorMessage(w, http.StatusUnauthorized, "La cuenta está desactivada.")
			return
		}

		// Check if the tenant's license is still valid.
		if !account.Tenant.License.IsValid() {
			sendErrorMessage(w, http.StatusUnauthorized, "La licencia de la institución ha expirado.")
			return
		}

		// Add the account to the context.
		ctx := edutrack.NewContextWithAccount(r.Context(), &account)
		next(w, r.WithContext(ctx))
	}
}

// withRole is a middleware that checks if the user has the required role.
func (s *Server) withRole(role edutrack.Role, next http.HandlerFunc) http.HandlerFunc {
	return s.withAuth(func(w http.ResponseWriter, r *http.Request) {
		account := edutrack.AccountFromContext(r.Context())
		if account == nil {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		if account.Role != role {
			sendError(w, http.StatusForbidden, ErrForbidden)
			return
		}

		next(w, r)
	})
}

// withSecretary is a middleware that ensures only secretaries can access the endpoint.
func (s *Server) withSecretary(next http.HandlerFunc) http.HandlerFunc {
	return s.withRole(edutrack.RoleSecretary, next)
}

// withTeacher is a middleware that ensures only teachers can access the endpoint.
func (s *Server) withTeacher(next http.HandlerFunc) http.HandlerFunc {
	return s.withRole(edutrack.RoleTeacher, next)
}

// withStudent is a middleware that ensures only students can access the endpoint.
func (s *Server) withStudent(next http.HandlerFunc) http.HandlerFunc {
	return s.withRole(edutrack.RoleStudent, next)
}
