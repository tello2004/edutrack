package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// setupAccountTestDB creates an in-memory SQLite database for testing.
func setupAccountTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := edutrack.Migrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// createAccountTestTenant creates a test tenant with a valid license.
func createAccountTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createAccountTestAccount creates a test account for a tenant.
func createAccountTestAccount(t *testing.T, db *gorm.DB, tenantID, email, password string, role edutrack.Role) *edutrack.Account {
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

// makeAuthenticatedRequest creates an HTTP request with the account in context.
func makeAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set the account in the context
	ctx := edutrack.NewContextWithAccount(req.Context(), account)
	return req.WithContext(ctx)
}

func TestHandleListAccounts_Success(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	// Create additional accounts
	createAccountTestAccount(t, db, tenant.ID, "user1@test.com", "password123", edutrack.RoleTeacher)
	createAccountTestAccount(t, db, tenant.ID, "user2@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts", nil, account)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(accounts) != 3 {
		t.Errorf("handleListAccounts() returned %d accounts, want 3", len(accounts))
	}
}

func TestHandleListAccounts_FilterByName(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	// Create accounts with specific names
	acc1 := &edutrack.Account{Name: "Juan García", Email: "juan@test.com", Password: "hash", Active: true, TenantID: tenant.ID}
	acc2 := &edutrack.Account{Name: "María López", Email: "maria@test.com", Password: "hash", Active: true, TenantID: tenant.ID}
	db.Create(acc1)
	db.Create(acc2)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts?name=Juan", nil, account)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("handleListAccounts() with filter returned %d accounts, want 1", len(accounts))
	}
}

func TestHandleListAccounts_FilterByEmail(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	createAccountTestAccount(t, db, tenant.ID, "specific@example.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts?email=specific", nil, account)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("handleListAccounts() with email filter returned %d accounts, want 1", len(accounts))
	}
}

func TestHandleListAccounts_FilterByActive(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	// Create an account and then set it to inactive (similar to career test pattern)
	inactiveAccount := createAccountTestAccount(t, db, tenant.ID, "inactive@test.com", "password123", edutrack.RoleTeacher)
	inactiveAccount.Active = false
	db.Save(inactiveAccount)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts?active=false", nil, account)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("handleListAccounts() with active filter returned %d accounts, want 1", len(accounts))
	}
}

func TestHandleListAccounts_Unauthorized(t *testing.T) {
	db := setupAccountTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListAccounts() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListAccounts_TenantIsolation(t *testing.T) {
	db := setupAccountTestDB(t)

	// Create two tenants
	tenant1 := createAccountTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	// Create accounts in different tenants
	account1 := createAccountTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	createAccountTestAccount(t, db, tenant2.ID, "user2@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts", nil, account1)
	w := httptest.NewRecorder()

	server.handleListAccounts(w, req)

	var accounts []edutrack.Account
	json.NewDecoder(w.Body).Decode(&accounts)

	// Should only see accounts from tenant1
	if len(accounts) != 1 {
		t.Errorf("handleListAccounts() should only return accounts from same tenant, got %d", len(accounts))
	}
}

func TestHandleGetAccount_Success(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/accounts/%d", targetAccount.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleGetAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetAccount() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Email != "target@test.com" {
		t.Errorf("handleGetAccount() email = %q, want %q", found.Email, "target@test.com")
	}
}

func TestHandleGetAccount_NotFound(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetAccount() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetAccount_InvalidID(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodGet, "/accounts/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetAccount() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetAccount_ForbiddenCrossTenant(t *testing.T) {
	db := setupAccountTestDB(t)

	tenant1 := createAccountTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAccountTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	account2 := createAccountTestAccount(t, db, tenant2.ID, "user2@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	// User from tenant1 tries to access user from tenant2
	req := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/accounts/%d", account2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", account2.ID))
	w := httptest.NewRecorder()

	server.handleGetAccount(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetAccount() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateAccount_Success(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAccountRequest{
		Name:     "New User",
		Email:    "newuser@test.com",
		Password: "securepassword123",
	}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPost, "/accounts", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAccount(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateAccount() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Email != "newuser@test.com" {
		t.Errorf("handleCreateAccount() email = %q, want %q", created.Email, "newuser@test.com")
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateAccount() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}

	// Verify password was hashed
	if created.Password == "securepassword123" {
		t.Error("handleCreateAccount() password was not hashed")
	}
}

func TestHandleCreateAccount_MissingFields(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateAccountRequest
	}{
		{"missing name", CreateAccountRequest{Email: "test@test.com", Password: "password"}},
		{"missing email", CreateAccountRequest{Name: "Test", Password: "password"}},
		{"missing password", CreateAccountRequest{Name: "Test", Email: "test@test.com"}},
		{"all empty", CreateAccountRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeAuthenticatedRequest(t, http.MethodPost, "/accounts", body, account)
			w := httptest.NewRecorder()

			server.handleCreateAccount(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateAccount() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateAccount_InvalidJSON(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodPost, "/accounts", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAccount() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateAccount_Success(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	reqBody := UpdateAccountRequest{
		Name: &newName,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/accounts/%d", targetAccount.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAccount() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Account
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("handleUpdateAccount() name = %q, want %q", updated.Name, "Updated Name")
	}
}

func TestHandleUpdateAccount_UpdatePassword(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "oldpassword", edutrack.RoleTeacher)

	oldPasswordHash := targetAccount.Password

	server := NewServer(":8080", db, []byte("test-secret"))

	newPassword := "newpassword456"
	reqBody := UpdateAccountRequest{
		Password: &newPassword,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/accounts/%d", targetAccount.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAccount() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify password was changed and hashed
	var updated edutrack.Account
	db.First(&updated, targetAccount.ID)

	if updated.Password == oldPasswordHash {
		t.Error("handleUpdateAccount() password was not changed")
	}

	if updated.Password == "newpassword456" {
		t.Error("handleUpdateAccount() password was not hashed")
	}
}

func TestHandleUpdateAccount_UpdateActive(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	active := false
	reqBody := UpdateAccountRequest{
		Active: &active,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/accounts/%d", targetAccount.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAccount() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Account
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Active != false {
		t.Errorf("handleUpdateAccount() active = %v, want false", updated.Active)
	}
}

func TestHandleUpdateAccount_NotFound(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	reqBody := UpdateAccountRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPut, "/accounts/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateAccount() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateAccount_ForbiddenCrossTenant(t *testing.T) {
	db := setupAccountTestDB(t)

	tenant1 := createAccountTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAccountTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	account2 := createAccountTestAccount(t, db, tenant2.ID, "user2@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Hacked Name"
	reqBody := UpdateAccountRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/accounts/%d", account2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", account2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAccount(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateAccount() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteAccount_Success(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/accounts/%d", targetAccount.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAccount(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteAccount() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify account was deleted
	var found edutrack.Account
	err := db.First(&found, targetAccount.ID).Error
	if err == nil {
		t.Error("handleDeleteAccount() account was not deleted")
	}
}

func TestHandleDeleteAccount_CannotDeleteSelf(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/accounts/%d", account.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", account.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteAccount() self-delete status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteAccount_NotFound(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	account := createAccountTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodDelete, "/accounts/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteAccount() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteAccount_ForbiddenCrossTenant(t *testing.T) {
	db := setupAccountTestDB(t)

	tenant1 := createAccountTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAccountTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	account2 := createAccountTestAccount(t, db, tenant2.ID, "user2@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/accounts/%d", account2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", account2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAccount(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteAccount() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteAccount_ForbiddenForTeacher(t *testing.T) {
	db := setupAccountTestDB(t)
	tenant := createAccountTestTenant(t, db)
	deleterAccount := createAccountTestAccount(t, db, tenant.ID, "teacher@test.com", "password123", edutrack.RoleTeacher)
	targetAccount := createAccountTestAccount(t, db, tenant.ID, "target@test.com", "password123", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/accounts/%d", targetAccount.ID), nil, deleterAccount)
	req.SetPathValue("id", fmt.Sprintf("%d", targetAccount.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAccount(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteAccount() by teacher status = %d, want %d", w.Code, http.StatusForbidden)
	}

	// Verify account was NOT deleted
	var found edutrack.Account
	err := db.First(&found, targetAccount.ID).Error
	if err != nil {
		t.Error("handleDeleteAccount() account was deleted by teacher")
	}
}

func BenchmarkHandleListAccounts(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	edutrack.Migrate(db)

	tenant, _ := edutrack.NewTenant("Benchmark Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant)

	hashedPassword, _ := edutrack.HashPassword("password123")
	account := &edutrack.Account{
		Name:     "Benchmark User",
		Email:    "bench@test.com",
		Password: hashedPassword,
		Role:     edutrack.RoleSecretary,
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(account)

	// Create 100 accounts for benchmark
	for i := 0; i < 100; i++ {
		db.Create(&edutrack.Account{
			Name:     fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@test.com", i),
			Password: hashedPassword,
			Role:     edutrack.RoleTeacher,
			Active:   true,
			TenantID: tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListAccounts(w, req)
	}
}
