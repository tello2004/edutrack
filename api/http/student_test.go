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

// setupStudentTestDB creates an in-memory SQLite database for testing.
func setupStudentTestDB(t *testing.T) *gorm.DB {
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

// createStudentTestTenant creates a test tenant with a valid license.
func createStudentTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createStudentTestAccount creates a test account for a tenant.
func createStudentTestAccount(t *testing.T, db *gorm.DB, tenantID, email, name string, role edutrack.Role) *edutrack.Account {
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

// createStudentTestCareer creates a test career for a tenant.
func createStudentTestCareer(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Career {
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

// createTestStudent creates a test student for a tenant.
func createTestStudent(t *testing.T, db *gorm.DB, tenantID, studentID string, accountID, careerID uint, semester int) *edutrack.Student {
	student := &edutrack.Student{
		StudentID: studentID,
		AccountID: accountID,
		CareerID:  careerID,
		TenantID:  tenantID,
		Semester:  semester,
	}

	if err := db.Create(student).Error; err != nil {
		t.Fatalf("Failed to create test student: %v", err)
	}

	return student
}

// makeStudentAuthenticatedRequest creates an HTTP request with the account in context.
func makeStudentAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListStudents_Success(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)

	studentAccount1 := createStudentTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createStudentTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)

	createTestStudent(t, db, tenant.ID, "2024001", studentAccount1.ID, career.ID, 1)
	createTestStudent(t, db, tenant.ID, "2024002", studentAccount2.ID, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students", nil, account)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListStudents() status = %d, want %d", w.Code, http.StatusOK)
	}

	var students []edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(students) != 2 {
		t.Errorf("handleListStudents() returned %d students, want 2", len(students))
	}
}

func TestHandleListStudents_FilterByCareerID(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career1 := createStudentTestCareer(t, db, tenant.ID)
	career2 := &edutrack.Career{
		Name:     "Licenciatura en Administración",
		Code:     fmt.Sprintf("LAD-%d", time.Now().UnixNano()),
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(career2)

	studentAccount1 := createStudentTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createStudentTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)

	createTestStudent(t, db, tenant.ID, "2024001", studentAccount1.ID, career1.ID, 1)
	createTestStudent(t, db, tenant.ID, "2024002", studentAccount2.ID, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/students?career_id=%d", career1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListStudents() status = %d, want %d", w.Code, http.StatusOK)
	}

	var students []edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(students) != 1 {
		t.Errorf("handleListStudents() with career_id filter returned %d students, want 1", len(students))
	}
}

func TestHandleListStudents_FilterBySemester(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)

	studentAccount1 := createStudentTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createStudentTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)

	createTestStudent(t, db, tenant.ID, "2024001", studentAccount1.ID, career.ID, 1)
	createTestStudent(t, db, tenant.ID, "2024002", studentAccount2.ID, career.ID, 2)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students?semester=2", nil, account)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListStudents() status = %d, want %d", w.Code, http.StatusOK)
	}

	var students []edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(students) != 1 {
		t.Errorf("handleListStudents() with semester filter returned %d students, want 1", len(students))
	}
}

func TestHandleListStudents_FilterByStudentID(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)

	studentAccount1 := createStudentTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createStudentTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)

	createTestStudent(t, db, tenant.ID, "2024001", studentAccount1.ID, career.ID, 1)
	createTestStudent(t, db, tenant.ID, "2024002", studentAccount2.ID, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students?student_id=2024001", nil, account)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListStudents() status = %d, want %d", w.Code, http.StatusOK)
	}

	var students []edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&students); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(students) != 1 {
		t.Errorf("handleListStudents() with student_id filter returned %d students, want 1", len(students))
	}
}

func TestHandleListStudents_Unauthorized(t *testing.T) {
	db := setupStudentTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListStudents() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListStudents_TenantIsolation(t *testing.T) {
	db := setupStudentTestDB(t)

	tenant1 := createStudentTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createStudentTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career1 := createStudentTestCareer(t, db, tenant1.ID)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)

	studentAccount1 := createStudentTestAccount(t, db, tenant1.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createStudentTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)

	createTestStudent(t, db, tenant1.ID, "2024001", studentAccount1.ID, career1.ID, 1)
	createTestStudent(t, db, tenant2.ID, "2024002", studentAccount2.ID, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students", nil, account1)
	w := httptest.NewRecorder()

	server.handleListStudents(w, req)

	var students []edutrack.Student
	json.NewDecoder(w.Body).Decode(&students)

	if len(students) != 1 {
		t.Errorf("handleListStudents() should only return students from same tenant, got %d", len(students))
	}
}

func TestHandleGetStudent_Success(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)
	studentAccount := createStudentTestAccount(t, db, tenant.ID, "student@test.com", "Test Student", edutrack.RoleTeacher)
	student := createTestStudent(t, db, tenant.ID, "2024001", studentAccount.ID, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/students/%d", student.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", student.ID))
	w := httptest.NewRecorder()

	server.handleGetStudent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetStudent() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.StudentID != "2024001" {
		t.Errorf("handleGetStudent() student_id = %q, want %q", found.StudentID, "2024001")
	}
}

func TestHandleGetStudent_NotFound(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetStudent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetStudent() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetStudent_InvalidID(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, "/students/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetStudent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetStudent() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetStudent_ForbiddenCrossTenant(t *testing.T) {
	db := setupStudentTestDB(t)

	tenant1 := createStudentTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createStudentTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createStudentTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createTestStudent(t, db, tenant2.ID, "2024002", studentAccount2.ID, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/students/%d", student2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", student2.ID))
	w := httptest.NewRecorder()

	server.handleGetStudent(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetStudent() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateStudent_Success(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateStudentRequest{
		StudentID: "2024001",
		Name:      "New Student",
		Email:     "newstudent@test.com",
		Password:  "password123",
		CareerID:  career.ID,
		Semester:  1,
	}
	body, _ := json.Marshal(reqBody)

	req := makeStudentAuthenticatedRequest(t, http.MethodPost, "/students", body, account)
	w := httptest.NewRecorder()

	server.handleCreateStudent(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateStudent() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.StudentID != "2024001" {
		t.Errorf("handleCreateStudent() student_id = %q, want %q", created.StudentID, "2024001")
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateStudent() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}

	if created.Semester != 1 {
		t.Errorf("handleCreateStudent() semester = %d, want %d", created.Semester, 1)
	}
}

func TestHandleCreateStudent_MissingFields(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateStudentRequest
	}{
		{"missing student_id", CreateStudentRequest{Name: "Name", Email: "email@test.com", Password: "pass", CareerID: 1, Semester: 1}},
		{"missing name", CreateStudentRequest{StudentID: "2024001", Email: "email@test.com", Password: "pass", CareerID: 1, Semester: 1}},
		{"missing email", CreateStudentRequest{StudentID: "2024001", Name: "Name", Password: "pass", CareerID: 1, Semester: 1}},
		{"missing password", CreateStudentRequest{StudentID: "2024001", Name: "Name", Email: "email@test.com", CareerID: 1, Semester: 1}},
		{"all empty", CreateStudentRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeStudentAuthenticatedRequest(t, http.MethodPost, "/students", body, account)
			w := httptest.NewRecorder()

			server.handleCreateStudent(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateStudent() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateStudent_InvalidJSON(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodPost, "/students", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateStudent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateStudent() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateStudent_Success(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)
	studentAccount := createStudentTestAccount(t, db, tenant.ID, "student@test.com", "Test Student", edutrack.RoleTeacher)
	student := createTestStudent(t, db, tenant.ID, "2024001", studentAccount.ID, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStudentID := "2024999"
	newSemester := 2
	reqBody := UpdateStudentRequest{
		StudentID: &newStudentID,
		Semester:  &newSemester,
	}
	body, _ := json.Marshal(reqBody)

	req := makeStudentAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/students/%d", student.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", student.ID))
	w := httptest.NewRecorder()

	server.handleUpdateStudent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateStudent() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Student
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.StudentID != "2024999" {
		t.Errorf("handleUpdateStudent() student_id = %q, want %q", updated.StudentID, "2024999")
	}

	if updated.Semester != 2 {
		t.Errorf("handleUpdateStudent() semester = %d, want %d", updated.Semester, 2)
	}
}

func TestHandleUpdateStudent_UpdateCareer(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career1 := createStudentTestCareer(t, db, tenant.ID)
	career2 := &edutrack.Career{
		Name:     "New Career",
		Code:     fmt.Sprintf("NEW-%d", time.Now().UnixNano()),
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(career2)

	studentAccount := createStudentTestAccount(t, db, tenant.ID, "student@test.com", "Test Student", edutrack.RoleTeacher)
	student := createTestStudent(t, db, tenant.ID, "2024001", studentAccount.ID, career1.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newCareerID := career2.ID
	reqBody := UpdateStudentRequest{
		CareerID: &newCareerID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeStudentAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/students/%d", student.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", student.ID))
	w := httptest.NewRecorder()

	server.handleUpdateStudent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateStudent() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Student
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.CareerID != career2.ID {
		t.Errorf("handleUpdateStudent() career_id = %d, want %d", updated.CareerID, career2.ID)
	}
}

func TestHandleUpdateStudent_NotFound(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStudentID := "2024999"
	reqBody := UpdateStudentRequest{StudentID: &newStudentID}
	body, _ := json.Marshal(reqBody)

	req := makeStudentAuthenticatedRequest(t, http.MethodPut, "/students/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateStudent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateStudent() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateStudent_ForbiddenCrossTenant(t *testing.T) {
	db := setupStudentTestDB(t)

	tenant1 := createStudentTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createStudentTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createStudentTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createTestStudent(t, db, tenant2.ID, "2024002", studentAccount2.ID, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStudentID := "HACKED"
	reqBody := UpdateStudentRequest{StudentID: &newStudentID}
	body, _ := json.Marshal(reqBody)

	req := makeStudentAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/students/%d", student2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", student2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateStudent(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateStudent() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteStudent_Success(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createStudentTestCareer(t, db, tenant.ID)
	studentAccount := createStudentTestAccount(t, db, tenant.ID, "student@test.com", "Test Student", edutrack.RoleTeacher)
	student := createTestStudent(t, db, tenant.ID, "2024001", studentAccount.ID, career.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/students/%d", student.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", student.ID))
	w := httptest.NewRecorder()

	server.handleDeleteStudent(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteStudent() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify student was deleted
	var found edutrack.Student
	err := db.First(&found, student.ID).Error
	if err == nil {
		t.Error("handleDeleteStudent() student was not deleted")
	}
}

func TestHandleDeleteStudent_NotFound(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodDelete, "/students/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteStudent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteStudent() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteStudent_InvalidID(t *testing.T) {
	db := setupStudentTestDB(t)
	tenant := createStudentTestTenant(t, db)
	account := createStudentTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodDelete, "/students/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteStudent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteStudent() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteStudent_ForbiddenCrossTenant(t *testing.T) {
	db := setupStudentTestDB(t)

	tenant1 := createStudentTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createStudentTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createStudentTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createTestStudent(t, db, tenant2.ID, "2024002", studentAccount2.ID, career2.ID, 1)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeStudentAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/students/%d", student2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", student2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteStudent(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteStudent() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListStudents(b *testing.B) {
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
		Name:     "Benchmark Career",
		Code:     "BENCH-2024",
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(career)

	// Create 100 students for benchmark
	for i := 0; i < 100; i++ {
		studentAccount := &edutrack.Account{
			Name:     fmt.Sprintf("Student %d", i),
			Email:    fmt.Sprintf("student%d@test.com", i),
			Password: hashedPassword,
			Role:     edutrack.RoleTeacher,
			Active:   true,
			TenantID: tenant.ID,
		}
		db.Create(studentAccount)

		db.Create(&edutrack.Student{
			StudentID: fmt.Sprintf("2024%03d", i),
			AccountID: studentAccount.ID,
			CareerID:  career.ID,
			TenantID:  tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/students", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListStudents(w, req)
	}
}
