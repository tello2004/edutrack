package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := edutrack.Migrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// createTestTenant creates a test tenant with a valid license.
func createTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createTestAccount creates a test account for a tenant.
func createTestAccount(t *testing.T, db *gorm.DB, tenantID, email, password string, role edutrack.Role) *edutrack.Account {
	hashedPassword, err := edutrack.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	account := &edutrack.Account{
		Name:     "Test User",
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		Active:   true,
		TenantID: tenantID,
	}

	if err := db.Create(account).Error; err != nil {
		t.Fatalf("Failed to save test account: %v", err)
	}

	return account
}

func TestHandleLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Error("handleLogin() returned empty token")
	}

	if resp.Role != edutrack.RoleSecretary {
		t.Errorf("handleLogin() role = %q, want %q", resp.Role, edutrack.RoleSecretary)
	}

	if resp.User.Email != "test@example.com" {
		t.Errorf("handleLogin() user.email = %q, want %q", resp.User.Email, "test@example.com")
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", "test@example.com", "wrongpassword"},
		{"wrong email", "wrong@example.com", "password123"},
		{"both wrong", "wrong@example.com", "wrongpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleLogin(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestHandleLogin_InactiveAccount(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	account := createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	// Deactivate the account
	account.Active = false
	db.Save(account)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogin_ExpiredLicense(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	// Expire the license
	tenant.License.ExpiryAt = time.Now().Add(-24 * time.Hour)
	db.Save(&tenant.License)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogin_MissingFields(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		reqBody LoginRequest
	}{
		{"missing email", LoginRequest{Password: "password123"}},
		{"missing password", LoginRequest{Email: "test@example.com"}},
		{"missing both", LoginRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleLogin(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleLogin() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleLicenseLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	createTestAccount(t, db, tenant.ID, "secretary@example.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: tenant.License.Key,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp LicenseLoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.TenantID != tenant.ID {
		t.Errorf("handleLicenseLogin() tenant_id = %q, want %q", resp.TenantID, tenant.ID)
	}

	if resp.TenantName != tenant.Name {
		t.Errorf("handleLicenseLogin() tenant_name = %q, want %q", resp.TenantName, tenant.Name)
	}
}

func TestHandleLicenseLogin_NoSecretary(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	// No secretary account created

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: tenant.License.Key,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp LicenseLoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Message should indicate that a secretary account is needed
	if resp.Message == "" {
		t.Error("handleLicenseLogin() should return a message about creating secretary account")
	}
}

func TestHandleLicenseLogin_InvalidKey(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: "invalid-key-that-does-not-exist",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLicenseLogin_ExpiredLicense(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)

	// Expire the license
	tenant.License.ExpiryAt = time.Now().Add(-24 * time.Hour)
	db.Save(&tenant.License)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: tenant.License.Key,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLicenseLogin_InactiveLicense(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)

	// Deactivate the license
	tenant.License.Active = false
	db.Save(&tenant.License)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: tenant.License.Key,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLicenseLogin_MissingKey(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := LicenseLoginRequest{
		LicenseKey: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/license", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleLicenseLogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleLicenseLogin() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestWithAuth_ValidToken(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	account := createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	// Generate a valid token
	token, err := server.generateToken(account)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create a handler that checks if account is in context
	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		acc := edutrack.AccountFromContext(r.Context())
		if acc == nil {
			t.Error("Account not found in context")
			return
		}
		if acc.ID != account.ID {
			t.Errorf("Account ID = %d, want %d", acc.ID, account.ID)
		}
		w.WriteHeader(http.StatusOK)
	}

	wrappedHandler := server.withAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("withAuth() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestWithAuth_MissingToken(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}

	wrappedHandler := server.withAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("withAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWithAuth_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}

	wrappedHandler := server.withAuth(testHandler)

	tests := []struct {
		name   string
		header string
	}{
		{"invalid format", "invalid-token"},
		{"wrong prefix", "Basic token123"},
		{"empty bearer", "Bearer "},
		{"malformed jwt", "Bearer not.a.valid.jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			wrappedHandler(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("withAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestWithAuth_WrongSecret(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	account := createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	// Generate token with one secret
	server1 := NewServer(":8080", db, []byte("secret-one"))
	token, err := server1.generateToken(account)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify with different secret
	server2 := NewServer(":8080", db, []byte("secret-two"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}

	wrappedHandler := server2.withAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("withAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWithRole_Secretary(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	secretary := createTestAccount(t, db, tenant.ID, "secretary@example.com", "password123", edutrack.RoleSecretary)
	teacher := createTestAccount(t, db, tenant.ID, "teacher@example.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrappedHandler := server.withSecretary(testHandler)

	t.Run("secretary can access", func(t *testing.T) {
		token, _ := server.generateToken(secretary)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("withSecretary() status = %d, want %d for secretary", w.Code, http.StatusOK)
		}
	})

	t.Run("teacher cannot access", func(t *testing.T) {
		token, _ := server.generateToken(teacher)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("withSecretary() status = %d, want %d for teacher", w.Code, http.StatusForbidden)
		}
	})
}

func TestWithRole_Teacher(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	secretary := createTestAccount(t, db, tenant.ID, "secretary@example.com", "password123", edutrack.RoleSecretary)
	teacher := createTestAccount(t, db, tenant.ID, "teacher@example.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrappedHandler := server.withTeacher(testHandler)

	t.Run("teacher can access", func(t *testing.T) {
		token, _ := server.generateToken(teacher)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("withTeacher() status = %d, want %d for teacher", w.Code, http.StatusOK)
		}
	})

	t.Run("secretary cannot access", func(t *testing.T) {
		token, _ := server.generateToken(secretary)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("withTeacher() status = %d, want %d for secretary", w.Code, http.StatusForbidden)
		}
	})
}

func TestWithRole_Student(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	secretary := createTestAccount(t, db, tenant.ID, "secretary@example.com", "password123", edutrack.RoleSecretary)
	teacher := createTestAccount(t, db, tenant.ID, "teacher@example.com", "password123", edutrack.RoleTeacher)
	student := createTestAccount(t, db, tenant.ID, "student@example.com", "password123", edutrack.RoleStudent)

	server := NewServer(":8080", db, []byte("test-secret"))

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrappedHandler := server.withStudent(testHandler)

	t.Run("student can access", func(t *testing.T) {
		token, _ := server.generateToken(student)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("withStudent() status = %d, want %d for student", w.Code, http.StatusOK)
		}
	})

	t.Run("secretary cannot access", func(t *testing.T) {
		token, _ := server.generateToken(secretary)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("withStudent() status = %d, want %d for secretary", w.Code, http.StatusForbidden)
		}
	})

	t.Run("teacher cannot access", func(t *testing.T) {
		token, _ := server.generateToken(teacher)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		wrappedHandler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("withStudent() status = %d, want %d for teacher", w.Code, http.StatusForbidden)
		}
	})
}

func TestGenerateToken(t *testing.T) {
	db := setupTestDB(t)
	tenant := createTestTenant(t, db)
	account := createTestAccount(t, db, tenant.ID, "test@example.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	token, err := server.generateToken(account)
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	if token == "" {
		t.Error("generateToken() returned empty token")
	}

	// Token should have 3 parts separated by dots
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("generateToken() token has %d dots, want 2", parts)
	}
}

func BenchmarkHandleLogin(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = edutrack.Migrate(db)

	tenant, _ := edutrack.NewTenant("Bench Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant)

	hashedPassword, _ := edutrack.HashPassword("password123")
	account := &edutrack.Account{
		Name:     "Bench User",
		Email:    "bench@example.com",
		Password: hashedPassword,
		Role:     edutrack.RoleSecretary,
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(account)

	server := NewServer(":8080", db, []byte("bench-secret"))

	reqBody := LoginRequest{
		Email:    "bench@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.handleLogin(w, req)
	}
}

func BenchmarkWithAuth(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = edutrack.Migrate(db)

	tenant, _ := edutrack.NewTenant("Bench Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant)

	hashedPassword, _ := edutrack.HashPassword("password123")
	account := &edutrack.Account{
		Name:     "Bench User",
		Email:    "bench@example.com",
		Password: hashedPassword,
		Role:     edutrack.RoleSecretary,
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(account)

	server := NewServer(":8080", db, []byte("bench-secret"))
	token, _ := server.generateToken(account)

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	wrappedHandler := server.withAuth(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		wrappedHandler(w, req)
	}
}
