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

// setupTeacherTestDB creates an in-memory SQLite database for testing.
func setupTeacherTestDB(t *testing.T) *gorm.DB {
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

// createTeacherTestTenant creates a test tenant with a valid license.
func createTeacherTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createTeacherTestAccount creates a test account for a tenant.
func createTeacherTestAccount(t *testing.T, db *gorm.DB, tenantID, email, name string, role edutrack.Role) *edutrack.Account {
	hashedPassword, err := edutrack.HashPassword("password123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	account := &edutrack.Account{
		Name:     name,
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

// createTestTeacher creates a test teacher for a tenant.
func createTestTeacher(t *testing.T, db *gorm.DB, tenantID string, accountID uint) *edutrack.Teacher {
	teacher := &edutrack.Teacher{
		AccountID: accountID,
		TenantID:  tenantID,
	}

	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("Failed to create test teacher: %v", err)
	}

	return teacher
}

// makeTeacherAuthenticatedRequest creates an HTTP request with the account in context.
func makeTeacherAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListTeachers_Success(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	teacherAccount1 := createTeacherTestAccount(t, db, tenant.ID, "teacher1@test.com", "Teacher One", edutrack.RoleTeacher)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant.ID, "teacher2@test.com", "Teacher Two", edutrack.RoleTeacher)

	createTestTeacher(t, db, tenant.ID, teacherAccount1.ID)
	createTestTeacher(t, db, tenant.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, "/teachers", nil, account)
	w := httptest.NewRecorder()

	server.handleListTeachers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListTeachers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var teachers []edutrack.Teacher
	if err := json.NewDecoder(w.Body).Decode(&teachers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(teachers) != 2 {
		t.Errorf("handleListTeachers() returned %d teachers, want 2", len(teachers))
	}
}

func TestHandleListTeachers_FilterByAccountID(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	teacherAccount1 := createTeacherTestAccount(t, db, tenant.ID, "teacher1@test.com", "Teacher One", edutrack.RoleTeacher)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant.ID, "teacher2@test.com", "Teacher Two", edutrack.RoleTeacher)

	createTestTeacher(t, db, tenant.ID, teacherAccount1.ID)
	createTestTeacher(t, db, tenant.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/teachers?account_id=%d", teacherAccount1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListTeachers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListTeachers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var teachers []edutrack.Teacher
	if err := json.NewDecoder(w.Body).Decode(&teachers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(teachers) != 1 {
		t.Errorf("handleListTeachers() with account_id filter returned %d teachers, want 1", len(teachers))
	}
}

func TestHandleListTeachers_Unauthorized(t *testing.T) {
	db := setupTeacherTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
	w := httptest.NewRecorder()

	server.handleListTeachers(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListTeachers() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListTeachers_TenantIsolation(t *testing.T) {
	db := setupTeacherTestDB(t)

	tenant1 := createTeacherTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTeacherTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)

	teacherAccount1 := createTeacherTestAccount(t, db, tenant1.ID, "teacher1@test.com", "Teacher 1", edutrack.RoleTeacher)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant2.ID, "teacher2@test.com", "Teacher 2", edutrack.RoleTeacher)

	createTestTeacher(t, db, tenant1.ID, teacherAccount1.ID)
	createTestTeacher(t, db, tenant2.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, "/teachers", nil, account1)
	w := httptest.NewRecorder()

	server.handleListTeachers(w, req)

	var teachers []edutrack.Teacher
	json.NewDecoder(w.Body).Decode(&teachers)

	if len(teachers) != 1 {
		t.Errorf("handleListTeachers() should only return teachers from same tenant, got %d", len(teachers))
	}
}

func TestHandleGetTeacher_Success(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createTeacherTestAccount(t, db, tenant.ID, "teacher@test.com", "Test Teacher", edutrack.RoleTeacher)
	teacher := createTestTeacher(t, db, tenant.ID, teacherAccount.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/teachers/%d", teacher.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher.ID))
	w := httptest.NewRecorder()

	server.handleGetTeacher(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetTeacher() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Teacher
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.ID != teacher.ID {
		t.Errorf("handleGetTeacher() id = %d, want %d", found.ID, teacher.ID)
	}
}

func TestHandleGetTeacher_NotFound(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, "/teachers/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetTeacher(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetTeacher() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetTeacher_InvalidID(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, "/teachers/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetTeacher() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetTeacher_ForbiddenCrossTenant(t *testing.T) {
	db := setupTeacherTestDB(t)

	tenant1 := createTeacherTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTeacherTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant2.ID, "teacher2@test.com", "Teacher 2", edutrack.RoleTeacher)
	teacher2 := createTestTeacher(t, db, tenant2.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/teachers/%d", teacher2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher2.ID))
	w := httptest.NewRecorder()

	server.handleGetTeacher(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetTeacher() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateTeacher_Success(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createTeacherTestAccount(t, db, tenant.ID, "newteacher@test.com", "New Teacher", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateTeacherRequest{
		AccountID: teacherAccount.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPost, "/teachers", body, account)
	w := httptest.NewRecorder()

	server.handleCreateTeacher(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateTeacher() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Teacher
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.AccountID != teacherAccount.ID {
		t.Errorf("handleCreateTeacher() account_id = %d, want %d", created.AccountID, teacherAccount.ID)
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateTeacher() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}
}

func TestHandleCreateTeacher_MissingAccountID(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateTeacherRequest{
		AccountID: 0,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPost, "/teachers", body, account)
	w := httptest.NewRecorder()

	server.handleCreateTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateTeacher() missing account_id status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateTeacher_AccountNotFound(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateTeacherRequest{
		AccountID: 99999,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPost, "/teachers", body, account)
	w := httptest.NewRecorder()

	server.handleCreateTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateTeacher() account not found status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateTeacher_InvalidJSON(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodPost, "/teachers", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateTeacher() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateTeacher_CrossTenantAccount(t *testing.T) {
	db := setupTeacherTestDB(t)

	tenant1 := createTeacherTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTeacherTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	account2 := createTeacherTestAccount(t, db, tenant2.ID, "teacher2@test.com", "Teacher 2", edutrack.RoleTeacher)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateTeacherRequest{
		AccountID: account2.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPost, "/teachers", body, account1)
	w := httptest.NewRecorder()

	server.handleCreateTeacher(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleCreateTeacher() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleUpdateTeacher_Success(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount1 := createTeacherTestAccount(t, db, tenant.ID, "teacher1@test.com", "Teacher One", edutrack.RoleTeacher)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant.ID, "teacher2@test.com", "Teacher Two", edutrack.RoleTeacher)
	teacher := createTestTeacher(t, db, tenant.ID, teacherAccount1.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	newAccountID := teacherAccount2.ID
	reqBody := UpdateTeacherRequest{
		AccountID: &newAccountID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/teachers/%d", teacher.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTeacher(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateTeacher() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Teacher
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.AccountID != teacherAccount2.ID {
		t.Errorf("handleUpdateTeacher() account_id = %d, want %d", updated.AccountID, teacherAccount2.ID)
	}
}

func TestHandleUpdateTeacher_NotFound(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newAccountID := uint(1)
	reqBody := UpdateTeacherRequest{AccountID: &newAccountID}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPut, "/teachers/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateTeacher(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateTeacher() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateTeacher_ForbiddenCrossTenant(t *testing.T) {
	db := setupTeacherTestDB(t)

	tenant1 := createTeacherTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTeacherTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant2.ID, "teacher2@test.com", "Teacher 2", edutrack.RoleTeacher)
	teacher2 := createTestTeacher(t, db, tenant2.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	newAccountID := uint(1)
	reqBody := UpdateTeacherRequest{AccountID: &newAccountID}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/teachers/%d", teacher2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTeacher(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateTeacher() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleUpdateTeacher_AccountNotFound(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createTeacherTestAccount(t, db, tenant.ID, "teacher@test.com", "Teacher", edutrack.RoleTeacher)
	teacher := createTestTeacher(t, db, tenant.ID, teacherAccount.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	nonExistentAccountID := uint(99999)
	reqBody := UpdateTeacherRequest{AccountID: &nonExistentAccountID}
	body, _ := json.Marshal(reqBody)

	req := makeTeacherAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/teachers/%d", teacher.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleUpdateTeacher() account not found status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteTeacher_Success(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createTeacherTestAccount(t, db, tenant.ID, "teacher@test.com", "Test Teacher", edutrack.RoleTeacher)
	teacher := createTestTeacher(t, db, tenant.ID, teacherAccount.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/teachers/%d", teacher.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher.ID))
	w := httptest.NewRecorder()

	server.handleDeleteTeacher(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteTeacher() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify teacher was deleted
	var found edutrack.Teacher
	err := db.First(&found, teacher.ID).Error
	if err == nil {
		t.Error("handleDeleteTeacher() teacher was not deleted")
	}
}

func TestHandleDeleteTeacher_NotFound(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodDelete, "/teachers/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteTeacher(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteTeacher() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteTeacher_InvalidID(t *testing.T) {
	db := setupTeacherTestDB(t)
	tenant := createTeacherTestTenant(t, db)
	account := createTeacherTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodDelete, "/teachers/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteTeacher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteTeacher() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteTeacher_ForbiddenCrossTenant(t *testing.T) {
	db := setupTeacherTestDB(t)

	tenant1 := createTeacherTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTeacherTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	teacherAccount2 := createTeacherTestAccount(t, db, tenant2.ID, "teacher2@test.com", "Teacher 2", edutrack.RoleTeacher)
	teacher2 := createTestTeacher(t, db, tenant2.ID, teacherAccount2.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTeacherAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/teachers/%d", teacher2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", teacher2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteTeacher(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteTeacher() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListTeachers(b *testing.B) {
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

	// Create 50 teachers for benchmark
	for i := 0; i < 50; i++ {
		teacherAccount := &edutrack.Account{
			Name:     fmt.Sprintf("Teacher %d", i),
			Email:    fmt.Sprintf("teacher%d@test.com", i),
			Password: hashedPassword,
			Role:     edutrack.RoleTeacher,
			Active:   true,
			TenantID: tenant.ID,
		}
		db.Create(teacherAccount)

		db.Create(&edutrack.Teacher{
			AccountID: teacherAccount.ID,
			TenantID:  tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListTeachers(w, req)
	}
}
