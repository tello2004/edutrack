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

// setupSubjectTestDB creates an in-memory SQLite database for testing.
func setupSubjectTestDB(t *testing.T) *gorm.DB {
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

// createSubjectTestTenant creates a test tenant with a valid license.
func createSubjectTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createSubjectTestAccount creates a test account for a tenant.
func createSubjectTestAccount(t *testing.T, db *gorm.DB, tenantID, email, name string, role edutrack.Role) *edutrack.Account {
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

// createSubjectTestCareer creates a test career for a tenant.
func createSubjectTestCareer(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Career {
	career := &edutrack.Career{
		Name:        "Ingeniería en Sistemas",
		Code:        fmt.Sprintf("ISC-%d", time.Now().UnixNano()),
		Description: "Test career",
		Duration:    8,
		Active:      true,
		TenantID:    tenantID,
	}

	if err := db.Create(career).Error; err != nil {
		t.Fatalf("Failed to create test career: %v", err)
	}

	return career
}

// createSubjectTestTeacher creates a test teacher for a tenant.
func createSubjectTestTeacher(t *testing.T, db *gorm.DB, tenantID string, accountID uint) *edutrack.Teacher {
	teacher := &edutrack.Teacher{
		AccountID: accountID,
		TenantID:  tenantID,
	}

	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("Failed to create test teacher: %v", err)
	}

	return teacher
}

// createTestSubject creates a test subject for a tenant.
func createTestSubject(t *testing.T, db *gorm.DB, tenantID, name, code string, teacherID *uint, careerID uint, semester int) *edutrack.Subject {
	subject := &edutrack.Subject{
		Name:        name,
		Code:        code,
		Description: "Test subject description",
		Credits:     5,
		TeacherID:   teacherID,
		TenantID:    tenantID,
		CareerID:    careerID,
		Semester:    semester,
	}

	if err := db.Create(subject).Error; err != nil {
		t.Fatalf("Failed to create test subject: %v", err)
	}

	return subject
}

// makeSubjectAuthenticatedRequest creates an HTTP request with the account in context.
func makeSubjectAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListSubjects_Success(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects", nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 2 {
		t.Errorf("handleListSubjects() returned %d subjects, want 2", len(subjects))
	}
}

func TestHandleListSubjects_FilterByName(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects?name=Matemáticas", nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() with name filter returned %d subjects, want 1", len(subjects))
	}
}

func TestHandleListSubjects_FilterByCode(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects?code=PRO", nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() with code filter returned %d subjects, want 1", len(subjects))
	}
}

func TestHandleListSubjects_FilterByTeacherID(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createSubjectTestAccount(t, db, tenant.ID, "teacher@test.com", "Teacher", edutrack.RoleTeacher)
	teacher := createSubjectTestTeacher(t, db, tenant.ID, teacherAccount.ID)
	career := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", &teacher.ID, career.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/subjects?teacher_id=%d", teacher.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() with teacher_id filter returned %d subjects, want 1", len(subjects))
	}
}

func TestHandleListSubjects_FilterByCareerID(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career1 := createSubjectTestCareer(t, db, tenant.ID)
	career2 := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career1.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/subjects?career_id=%d", career1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() with career_id filter returned %d subjects, want 1", len(subjects))
	}
}

func TestHandleListSubjects_FilterBySemester(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career.ID, 1)
	createTestSubject(t, db, tenant.ID, "Programación I", "PRO101", nil, career.ID, 2)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects?semester=1", nil, account)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var subjects []edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() with semester filter returned %d subjects, want 1", len(subjects))
	}
}

func TestHandleListSubjects_Unauthorized(t *testing.T) {
	db := setupSubjectTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/subjects", nil)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListSubjects() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListSubjects_TenantIsolation(t *testing.T) {
	db := setupSubjectTestDB(t)

	tenant1 := createSubjectTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createSubjectTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career1 := createSubjectTestCareer(t, db, tenant1.ID)
	career2 := createSubjectTestCareer(t, db, tenant2.ID)

	createTestSubject(t, db, tenant1.ID, "Subject Tenant 1", "ST1-101", nil, career1.ID, 1)
	createTestSubject(t, db, tenant2.ID, "Subject Tenant 2", "ST2-101", nil, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects", nil, account1)
	w := httptest.NewRecorder()

	server.handleListSubjects(w, req)

	var subjects []edutrack.Subject
	json.NewDecoder(w.Body).Decode(&subjects)

	if len(subjects) != 1 {
		t.Errorf("handleListSubjects() should only return subjects from same tenant, got %d", len(subjects))
	}
}

func TestHandleGetSubject_Success(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)
	subject := createTestSubject(t, db, tenant.ID, "Matemáticas I", "MAT101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/subjects/%d", subject.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", subject.ID))
	w := httptest.NewRecorder()

	server.handleGetSubject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetSubject() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Name != "Matemáticas I" {
		t.Errorf("handleGetSubject() name = %q, want %q", found.Name, "Matemáticas I")
	}
}

func TestHandleGetSubject_NotFound(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetSubject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetSubject() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetSubject_InvalidID(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, "/subjects/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetSubject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetSubject() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetSubject_ForbiddenCrossTenant(t *testing.T) {
	db := setupSubjectTestDB(t)

	tenant1 := createSubjectTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createSubjectTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := createSubjectTestCareer(t, db, tenant2.ID)
	subject2 := createTestSubject(t, db, tenant2.ID, "Subject Tenant 2", "ST2-101", nil, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/subjects/%d", subject2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", subject2.ID))
	w := httptest.NewRecorder()

	server.handleGetSubject(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetSubject() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateSubject_Success(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateSubjectRequest{
		Name:        "Nueva Materia",
		Code:        "NEW-101",
		Description: "Descripción de la materia",
		Credits:     6,
		CareerID:    career.ID,
		Semester:    1,
	}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPost, "/subjects", body, account)
	w := httptest.NewRecorder()

	server.handleCreateSubject(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateSubject() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Name != "Nueva Materia" {
		t.Errorf("handleCreateSubject() name = %q, want %q", created.Name, "Nueva Materia")
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateSubject() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}

	if created.CareerID != career.ID {
		t.Errorf("handleCreateSubject() career_id = %d, want %d", created.CareerID, career.ID)
	}

	if created.Semester != 1 {
		t.Errorf("handleCreateSubject() semester = %d, want %d", created.Semester, 1)
	}
}

func TestHandleCreateSubject_WithTeacher(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createSubjectTestAccount(t, db, tenant.ID, "teacher@test.com", "Teacher", edutrack.RoleTeacher)
	teacher := createSubjectTestTeacher(t, db, tenant.ID, teacherAccount.ID)
	career := createSubjectTestCareer(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateSubjectRequest{
		Name:      "Nueva Materia",
		Code:      "NEW-101",
		Credits:   6,
		TeacherID: &teacher.ID,
		CareerID:  career.ID,
		Semester:  1,
	}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPost, "/subjects", body, account)
	w := httptest.NewRecorder()

	server.handleCreateSubject(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateSubject() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Subject
	json.NewDecoder(w.Body).Decode(&created)

	if created.TeacherID == nil || *created.TeacherID != teacher.ID {
		t.Errorf("handleCreateSubject() teacher_id should be %d", teacher.ID)
	}
}

func TestHandleCreateSubject_MissingFields(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateSubjectRequest
	}{
		{"missing name", CreateSubjectRequest{Code: "TEST-101", CareerID: career.ID, Semester: 1}},
		{"missing code", CreateSubjectRequest{Name: "Test Subject", CareerID: career.ID, Semester: 1}},
		{"missing career_id", CreateSubjectRequest{Name: "Test Subject", Code: "TEST-101", Semester: 1}},
		{"missing semester", CreateSubjectRequest{Name: "Test Subject", Code: "TEST-101", CareerID: career.ID}},
		{"zero semester", CreateSubjectRequest{Name: "Test Subject", Code: "TEST-101", CareerID: career.ID, Semester: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeSubjectAuthenticatedRequest(t, http.MethodPost, "/subjects", body, account)
			w := httptest.NewRecorder()

			server.handleCreateSubject(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateSubject() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateSubject_InvalidJSON(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodPost, "/subjects", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateSubject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateSubject() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateSubject_Success(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)
	subject := createTestSubject(t, db, tenant.ID, "Old Name", "OLD-101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Subject Name"
	reqBody := UpdateSubjectRequest{
		Name: &newName,
	}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/subjects/%d", subject.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", subject.ID))
	w := httptest.NewRecorder()

	server.handleUpdateSubject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateSubject() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Subject
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "Updated Subject Name" {
		t.Errorf("handleUpdateSubject() name = %q, want %q", updated.Name, "Updated Subject Name")
	}
}

func TestHandleUpdateSubject_UpdateAllFields(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	teacherAccount := createSubjectTestAccount(t, db, tenant.ID, "teacher@test.com", "Teacher", edutrack.RoleTeacher)
	teacher := createSubjectTestTeacher(t, db, tenant.ID, teacherAccount.ID)
	career1 := createSubjectTestCareer(t, db, tenant.ID)
	career2 := createSubjectTestCareer(t, db, tenant.ID)
	subject := createTestSubject(t, db, tenant.ID, "Old Name", "OLD-101", nil, career1.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	newCode := "UPD-101"
	newDescription := "Updated description"
	newCredits := 8
	newSemester := 2

	reqBody := UpdateSubjectRequest{
		Name:        &newName,
		Code:        &newCode,
		Description: &newDescription,
		Credits:     &newCredits,
		TeacherID:   &teacher.ID,
		CareerID:    &career2.ID,
		Semester:    &newSemester,
	}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/subjects/%d", subject.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", subject.ID))
	w := httptest.NewRecorder()

	server.handleUpdateSubject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateSubject() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Subject
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Name != newName {
		t.Errorf("handleUpdateSubject() name = %q, want %q", updated.Name, newName)
	}
	if updated.Code != newCode {
		t.Errorf("handleUpdateSubject() code = %q, want %q", updated.Code, newCode)
	}
	if updated.Description != newDescription {
		t.Errorf("handleUpdateSubject() description = %q, want %q", updated.Description, newDescription)
	}
	if updated.Credits != newCredits {
		t.Errorf("handleUpdateSubject() credits = %d, want %d", updated.Credits, newCredits)
	}
	if updated.CareerID != career2.ID {
		t.Errorf("handleUpdateSubject() career_id = %d, want %d", updated.CareerID, career2.ID)
	}
	if updated.Semester != newSemester {
		t.Errorf("handleUpdateSubject() semester = %d, want %d", updated.Semester, newSemester)
	}
}

func TestHandleUpdateSubject_NotFound(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Name"
	reqBody := UpdateSubjectRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPut, "/subjects/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateSubject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateSubject() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateSubject_ForbiddenCrossTenant(t *testing.T) {
	db := setupSubjectTestDB(t)

	tenant1 := createSubjectTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createSubjectTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := createSubjectTestCareer(t, db, tenant2.ID)
	subject2 := createTestSubject(t, db, tenant2.ID, "Subject Tenant 2", "ST2-101", nil, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Hacked Name"
	reqBody := UpdateSubjectRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeSubjectAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/subjects/%d", subject2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", subject2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateSubject(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateSubject() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteSubject_Success(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createSubjectTestCareer(t, db, tenant.ID)
	subject := createTestSubject(t, db, tenant.ID, "Subject to Delete", "DEL-101", nil, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/subjects/%d", subject.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", subject.ID))
	w := httptest.NewRecorder()

	server.handleDeleteSubject(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteSubject() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify subject was deleted
	var found edutrack.Subject
	err := db.First(&found, subject.ID).Error
	if err == nil {
		t.Error("handleDeleteSubject() subject was not deleted")
	}
}

func TestHandleDeleteSubject_NotFound(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodDelete, "/subjects/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteSubject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteSubject() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteSubject_InvalidID(t *testing.T) {
	db := setupSubjectTestDB(t)
	tenant := createSubjectTestTenant(t, db)
	account := createSubjectTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodDelete, "/subjects/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteSubject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteSubject() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteSubject_ForbiddenCrossTenant(t *testing.T) {
	db := setupSubjectTestDB(t)

	tenant1 := createSubjectTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createSubjectTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := createSubjectTestCareer(t, db, tenant2.ID)
	subject2 := createTestSubject(t, db, tenant2.ID, "Subject Tenant 2", "ST2-101", nil, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeSubjectAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/subjects/%d", subject2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", subject2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteSubject(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteSubject() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListSubjects(b *testing.B) {
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

	career := &edutrack.Career{
		Name:        "Benchmark Career",
		Code:        fmt.Sprintf("BENCH-%d", time.Now().UnixNano()),
		Description: "Benchmark career",
		Duration:    8,
		Active:      true,
		TenantID:    tenant.ID,
	}
	if err := db.Create(career).Error; err != nil {
		b.Fatalf("Failed to create test career: %v", err)
	}

	// Create 50 subjects for benchmark
	for i := 0; i < 50; i++ {
		db.Create(&edutrack.Subject{
			Name:        fmt.Sprintf("Subject %d", i),
			Code:        fmt.Sprintf("SUB-%d", i),
			Description: "Benchmark subject",
			Credits:     5,
			TenantID:    tenant.ID,
			CareerID:    career.ID,
			Semester:    1,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/subjects", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListSubjects(w, req)
	}
}
