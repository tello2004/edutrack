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

// setupCareerTestDB creates an in-memory SQLite database for testing.
func setupCareerTestDB(t *testing.T) *gorm.DB {
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

// createCareerTestTenant creates a test tenant with a valid license.
func createCareerTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createCareerTestAccount creates a test account for a tenant.
func createCareerTestAccount(t *testing.T, db *gorm.DB, tenantID, email, password string, role edutrack.Role) *edutrack.Account {
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

// createTestCareer creates a test career for a tenant.
func createTestCareer(t *testing.T, db *gorm.DB, tenantID, name, code string) *edutrack.Career {
	career := &edutrack.Career{
		Name:        name,
		Code:        code,
		Description: "Test career description",
		Duration:    8,
		Active:      true,
		TenantID:    tenantID,
	}

	if err := db.Create(career).Error; err != nil {
		t.Fatalf("Failed to create test career: %v", err)
	}

	return career
}

// makeCareerAuthenticatedRequest creates an HTTP request with the account in context.
func makeCareerAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx := edutrack.NewContextWithAccount(req.Context(), account)
	return req.WithContext(ctx)
}

func TestHandleListCareers_Success(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	createTestCareer(t, db, tenant.ID, "Ingeniería en Sistemas", "ISC-2024")
	createTestCareer(t, db, tenant.ID, "Licenciatura en Administración", "LAD-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers", nil, account)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListCareers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var careers []edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&careers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(careers) != 2 {
		t.Errorf("handleListCareers() returned %d careers, want 2", len(careers))
	}
}

func TestHandleListCareers_FilterByName(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	createTestCareer(t, db, tenant.ID, "Ingeniería en Sistemas", "ISC-2024")
	createTestCareer(t, db, tenant.ID, "Licenciatura en Administración", "LAD-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers?name=Ingeniería", nil, account)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListCareers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var careers []edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&careers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(careers) != 1 {
		t.Errorf("handleListCareers() with name filter returned %d careers, want 1", len(careers))
	}
}

func TestHandleListCareers_FilterByCode(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	createTestCareer(t, db, tenant.ID, "Ingeniería en Sistemas", "ISC-2024")
	createTestCareer(t, db, tenant.ID, "Licenciatura en Administración", "LAD-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers?code=LAD", nil, account)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListCareers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var careers []edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&careers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(careers) != 1 {
		t.Errorf("handleListCareers() with code filter returned %d careers, want 1", len(careers))
	}
}

func TestHandleListCareers_FilterByActive(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	createTestCareer(t, db, tenant.ID, "Active Career", "ACT-2024")
	inactiveCareer := createTestCareer(t, db, tenant.ID, "Inactive Career", "INA-2024")
	inactiveCareer.Active = false
	db.Save(inactiveCareer)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers?active=false", nil, account)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListCareers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var careers []edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&careers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(careers) != 1 {
		t.Errorf("handleListCareers() with active filter returned %d careers, want 1", len(careers))
	}
}

func TestHandleListCareers_Unauthorized(t *testing.T) {
	db := setupCareerTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/careers", nil)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListCareers() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListCareers_TenantIsolation(t *testing.T) {
	db := setupCareerTestDB(t)

	tenant1 := createCareerTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createCareerTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	createTestCareer(t, db, tenant1.ID, "Career Tenant 1", "CT1-2024")
	createTestCareer(t, db, tenant2.ID, "Career Tenant 2", "CT2-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers", nil, account1)
	w := httptest.NewRecorder()

	server.handleListCareers(w, req)

	var careers []edutrack.Career
	json.NewDecoder(w.Body).Decode(&careers)

	if len(careers) != 1 {
		t.Errorf("handleListCareers() should only return careers from same tenant, got %d", len(careers))
	}
}

func TestHandleGetCareer_Success(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	career := createTestCareer(t, db, tenant.ID, "Ingeniería en Sistemas", "ISC-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/careers/%d", career.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", career.ID))
	w := httptest.NewRecorder()

	server.handleGetCareer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetCareer() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Name != "Ingeniería en Sistemas" {
		t.Errorf("handleGetCareer() name = %q, want %q", found.Name, "Ingeniería en Sistemas")
	}
}

func TestHandleGetCareer_NotFound(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetCareer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetCareer() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetCareer_InvalidID(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, "/careers/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetCareer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetCareer() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetCareer_ForbiddenCrossTenant(t *testing.T) {
	db := setupCareerTestDB(t)

	tenant1 := createCareerTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createCareerTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	career2 := createTestCareer(t, db, tenant2.ID, "Career Tenant 2", "CT2-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/careers/%d", career2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", career2.ID))
	w := httptest.NewRecorder()

	server.handleGetCareer(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetCareer() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateCareer_Success(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateCareerRequest{
		Name:        "Nueva Carrera",
		Code:        "NC-2024",
		Description: "Descripción de la carrera",
		Duration:    8,
	}
	body, _ := json.Marshal(reqBody)

	req := makeCareerAuthenticatedRequest(t, http.MethodPost, "/careers", body, account)
	w := httptest.NewRecorder()

	server.handleCreateCareer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateCareer() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Name != "Nueva Carrera" {
		t.Errorf("handleCreateCareer() name = %q, want %q", created.Name, "Nueva Carrera")
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateCareer() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}

	if created.Active != true {
		t.Error("handleCreateCareer() active should default to true")
	}
}

func TestHandleCreateCareer_MissingFields(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateCareerRequest
	}{
		{"missing name", CreateCareerRequest{Code: "TEST-2024"}},
		{"missing code", CreateCareerRequest{Name: "Test Career"}},
		{"all empty", CreateCareerRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeCareerAuthenticatedRequest(t, http.MethodPost, "/careers", body, account)
			w := httptest.NewRecorder()

			server.handleCreateCareer(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateCareer() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateCareer_InvalidJSON(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodPost, "/careers", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateCareer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateCareer() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateCareer_Success(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	career := createTestCareer(t, db, tenant.ID, "Old Name", "OLD-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Career Name"
	reqBody := UpdateCareerRequest{
		Name: &newName,
	}
	body, _ := json.Marshal(reqBody)

	req := makeCareerAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/careers/%d", career.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", career.ID))
	w := httptest.NewRecorder()

	server.handleUpdateCareer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateCareer() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Career
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "Updated Career Name" {
		t.Errorf("handleUpdateCareer() name = %q, want %q", updated.Name, "Updated Career Name")
	}
}

func TestHandleUpdateCareer_UpdateAllFields(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	career := createTestCareer(t, db, tenant.ID, "Old Name", "OLD-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	newCode := "UPD-2024"
	newDescription := "Updated description"
	newDuration := 10
	newActive := false

	reqBody := UpdateCareerRequest{
		Name:        &newName,
		Code:        &newCode,
		Description: &newDescription,
		Duration:    &newDuration,
		Active:      &newActive,
	}
	body, _ := json.Marshal(reqBody)

	req := makeCareerAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/careers/%d", career.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", career.ID))
	w := httptest.NewRecorder()

	server.handleUpdateCareer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateCareer() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Career
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Name != newName {
		t.Errorf("handleUpdateCareer() name = %q, want %q", updated.Name, newName)
	}
	if updated.Code != newCode {
		t.Errorf("handleUpdateCareer() code = %q, want %q", updated.Code, newCode)
	}
	if updated.Description != newDescription {
		t.Errorf("handleUpdateCareer() description = %q, want %q", updated.Description, newDescription)
	}
	if updated.Duration != newDuration {
		t.Errorf("handleUpdateCareer() duration = %d, want %d", updated.Duration, newDuration)
	}
	if updated.Active != newActive {
		t.Errorf("handleUpdateCareer() active = %v, want %v", updated.Active, newActive)
	}
}

func TestHandleUpdateCareer_NotFound(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	reqBody := UpdateCareerRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeCareerAuthenticatedRequest(t, http.MethodPut, "/careers/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateCareer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateCareer() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateCareer_ForbiddenCrossTenant(t *testing.T) {
	db := setupCareerTestDB(t)

	tenant1 := createCareerTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createCareerTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	career2 := createTestCareer(t, db, tenant2.ID, "Career Tenant 2", "CT2-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Hacked Name"
	reqBody := UpdateCareerRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeCareerAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/careers/%d", career2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", career2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateCareer(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateCareer() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteCareer_Success(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)
	career := createTestCareer(t, db, tenant.ID, "Career to Delete", "DEL-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/careers/%d", career.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", career.ID))
	w := httptest.NewRecorder()

	server.handleDeleteCareer(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteCareer() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify career was deleted
	var found edutrack.Career
	err := db.First(&found, career.ID).Error
	if err == nil {
		t.Error("handleDeleteCareer() career was not deleted")
	}
}

func TestHandleDeleteCareer_NotFound(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodDelete, "/careers/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteCareer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteCareer() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteCareer_InvalidID(t *testing.T) {
	db := setupCareerTestDB(t)
	tenant := createCareerTestTenant(t, db)
	account := createCareerTestAccount(t, db, tenant.ID, "admin@test.com", "password123", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodDelete, "/careers/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteCareer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteCareer() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteCareer_ForbiddenCrossTenant(t *testing.T) {
	db := setupCareerTestDB(t)

	tenant1 := createCareerTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createCareerTestAccount(t, db, tenant1.ID, "user1@test.com", "password123", edutrack.RoleSecretary)
	career2 := createTestCareer(t, db, tenant2.ID, "Career Tenant 2", "CT2-2024")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeCareerAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/careers/%d", career2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", career2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteCareer(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteCareer() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListCareers(b *testing.B) {
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

	// Create 50 careers for benchmark
	for i := 0; i < 50; i++ {
		db.Create(&edutrack.Career{
			Name:        fmt.Sprintf("Career %d", i),
			Code:        fmt.Sprintf("CAR-%d", i),
			Description: "Benchmark career",
			Duration:    8,
			Active:      true,
			TenantID:    tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/careers", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListCareers(w, req)
	}
}
