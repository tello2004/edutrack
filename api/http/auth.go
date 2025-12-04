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
	AccountID uint   `json:"account_id"`
	TenantID  string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response body.
type LoginResponse struct {
	Token string `json:"token"`
}

// handleLogin handles POST /auth/login.
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
	if err := s.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		sendError(w, http.StatusUnauthorized, ErrInvalidCredentials)
		return
	}

	// Check if the account is active.
	if !account.Active {
		sendErrorMessage(w, http.StatusUnauthorized, "La cuenta está desactivada.")
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

	sendJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// generateToken generates a JWT token for the given account.
func (s *Server) generateToken(account *edutrack.Account) (string, error) {
	claims := &Claims{
		AccountID: account.ID,
		TenantID:  account.TenantID,
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
		if err := s.DB.First(&account, claims.AccountID).Error; err != nil {
			sendError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		// Check if the account is still active.
		if !account.Active {
			sendErrorMessage(w, http.StatusUnauthorized, "La cuenta está desactivada.")
			return
		}

		// Add the account to the context.
		ctx := edutrack.NewContextWithAccount(r.Context(), &account)
		next(w, r.WithContext(ctx))
	}
}
