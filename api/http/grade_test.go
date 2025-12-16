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

// setupGradeTestDB creates an in-memory SQLite database for testing.
func setupGradeTestDB(t *testing.T) *gorm.DB {
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

// createGradeTestTenant creates a test tenant with a valid license.
func createGradeTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createGradeTestAccount creates a test account for a tenant.
func createGradeTestAccount(t *testing.T, db *gorm.DB, tenantID, email, name string, role edutrack.Role) *edutrack.Account {
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

// createGradeTestCareer creates a test career for a tenant.
func createGradeTestCareer(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Career {
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

// createGradeTestStudent creates a test student for a tenant.
func createGradeTestStudent(t *testing.T, db *gorm.DB, tenantID string, accountID, careerID uint) *edutrack.Student {
	student := &edutrack.Student{
		StudentID: fmt.Sprintf("STU-%d", time.Now().UnixNano()),
		AccountID: accountID,
		CareerID:  careerID,
		TenantID:  tenantID,
	}

	if err := db.Create(student).Error; err != nil {
		t.Fatalf("Failed to create test student: %v", err)
	}

	return student
}

// createGradeTestTeacher creates a test teacher for a tenant.
func createGradeTestTeacher(t *testing.T, db *gorm.DB, tenantID string, accountID uint) *edutrack.Teacher {
	teacher := &edutrack.Teacher{
		AccountID: accountID,
		TenantID:  tenantID,
	}

	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("Failed to create test teacher: %v", err)
	}

	return teacher
}

// createGradeTestSubject creates a test subject for a tenant.
func createGradeTestSubject(t *testing.T, db *gorm.DB, tenantID string, careerID uint) *edutrack.Subject {
	subject := &edutrack.Subject{
		Name:        "Matemáticas I",
		Code:        fmt.Sprintf("MAT-%d", time.Now().UnixNano()),
		Description: "Test subject",
		Credits:     5,
		TenantID:    tenantID,
		CareerID:    careerID,
	}

	if err := db.Create(subject).Error; err != nil {
		t.Fatalf("Failed to create test subject: %v", err)
	}

	return subject
}

// createGradeTestTopic creates a test topic for a subject.
func createGradeTestTopic(t *testing.T, db *gorm.DB, tenantID string, subjectID uint) *edutrack.Topic {
	topic := &edutrack.Topic{
		Name:      "Test Topic",
		SubjectID: subjectID,
		TenantID:  tenantID,
	}

	if err := db.Create(topic).Error; err != nil {
		t.Fatalf("Failed to create test topic: %v", err)
	}

	return topic
}

// createTestGrade creates a test grade record.
func createTestGrade(t *testing.T, db *gorm.DB, tenantID string, studentID, topicID uint, value float64) *edutrack.Grade {
	grade := &edutrack.Grade{
		Value:     value,
		Notes:     "Test grade",
		StudentID: studentID,
		TopicID:   topicID,
		TenantID:  tenantID,
	}

	if err := db.Create(grade).Error; err != nil {
		t.Fatalf("Failed to create test grade: %v", err)
	}

	return grade
}

// makeGradeAuthenticatedRequest creates an HTTP request with the account in context.
func makeGradeAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListGrades_Success(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 85.5)
	createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 92.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, "/grades", nil, account)
	w := httptest.NewRecorder()

	server.handleListGrades(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListGrades() status = %d, want %d", w.Code, http.StatusOK)
	}

	var grades []edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&grades); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(grades) != 2 {
		t.Errorf("handleListGrades() returned %d grades, want 2", len(grades))
	}
}

func TestHandleListGrades_FilterByStudentID(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)

	studentAccount1 := createGradeTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createGradeTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student1 := createGradeTestStudent(t, db, tenant.ID, studentAccount1.ID, career.ID)
	student2 := createGradeTestStudent(t, db, tenant.ID, studentAccount2.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	createTestGrade(t, db, tenant.ID, student1.ID, topic.ID, 85.5)
	createTestGrade(t, db, tenant.ID, student2.ID, topic.ID, 92.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/grades?student_id=%d", student1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListGrades(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListGrades() status = %d, want %d", w.Code, http.StatusOK)
	}

	var grades []edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&grades); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(grades) != 1 {
		t.Errorf("handleListGrades() with student_id filter returned %d grades, want 1", len(grades))
	}
}

func TestHandleListGrades_FilterByTopicID(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic1 := createGradeTestTopic(t, db, tenant.ID, subject.ID)
	topic2 := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	createTestGrade(t, db, tenant.ID, student.ID, topic1.ID, 85.5)
	createTestGrade(t, db, tenant.ID, student.ID, topic2.ID, 92.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/grades?topic_id=%d", topic1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListGrades(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListGrades() status = %d, want %d", w.Code, http.StatusOK)
	}

	var grades []edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&grades); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(grades) != 1 {
		t.Errorf("handleListGrades() with topic_id filter returned %d grades, want 1", len(grades))
	}
}

func TestHandleListGrades_Unauthorized(t *testing.T) {
	db := setupGradeTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/grades", nil)
	w := httptest.NewRecorder()

	server.handleListGrades(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListGrades() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListGrades_TenantIsolation(t *testing.T) {
	db := setupGradeTestDB(t)

	tenant1 := createGradeTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createGradeTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career1 := createGradeTestCareer(t, db, tenant1.ID)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)

	studentAccount1 := createGradeTestAccount(t, db, tenant1.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createGradeTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student1 := createGradeTestStudent(t, db, tenant1.ID, studentAccount1.ID, career1.ID)
	student2 := createGradeTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)

	subject1 := createGradeTestSubject(t, db, tenant1.ID, career1.ID)
	subject2 := createGradeTestSubject(t, db, tenant2.ID, career2.ID)
	topic1 := createGradeTestTopic(t, db, tenant1.ID, subject1.ID)
	topic2 := createGradeTestTopic(t, db, tenant2.ID, subject2.ID)

	createTestGrade(t, db, tenant1.ID, student1.ID, topic1.ID, 85.5)
	createTestGrade(t, db, tenant2.ID, student2.ID, topic2.ID, 92.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, "/grades", nil, account1)
	w := httptest.NewRecorder()

	server.handleListGrades(w, req)

	var grades []edutrack.Grade
	json.NewDecoder(w.Body).Decode(&grades)

	if len(grades) != 1 {
		t.Errorf("handleListGrades() should only return grades from same tenant, got %d", len(grades))
	}
}

func TestHandleGetGrade_Success(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	grade := createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/grades/%d", grade.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", grade.ID))
	w := httptest.NewRecorder()

	server.handleGetGrade(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetGrade() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Value != 85.5 {
		t.Errorf("handleGetGrade() value = %f, want %f", found.Value, 85.5)
	}
}

func TestHandleGetGrade_NotFound(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, "/grades/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetGrade(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetGrade() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetGrade_InvalidID(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, "/grades/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetGrade(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetGrade() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetGrade_ForbiddenCrossTenant(t *testing.T) {
	db := setupGradeTestDB(t)

	tenant1 := createGradeTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createGradeTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createGradeTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createGradeTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := createGradeTestSubject(t, db, tenant2.ID, career2.ID)
	topic2 := createGradeTestTopic(t, db, tenant2.ID, subject2.ID)

	grade2 := createTestGrade(t, db, tenant2.ID, student2.ID, topic2.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/grades/%d", grade2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", grade2.ID))
	w := httptest.NewRecorder()

	server.handleGetGrade(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetGrade() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateGrade_Success(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateGradeRequest{
		Value:     95.5,
		Notes:     "Excellent work",
		StudentID: student.ID,
		TopicID:   topic.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPost, "/grades", body, account)
	w := httptest.NewRecorder()

	server.handleCreateGrade(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateGrade() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Value != 95.5 {
		t.Errorf("handleCreateGrade() value = %f, want %f", created.Value, 95.5)
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateGrade() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}
}

func TestHandleCreateGrade_VariousValues(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	values := []float64{0.0, 50.5, 70.0, 85.5, 100.0}

	for _, value := range values {
		t.Run(fmt.Sprintf("value_%.1f", value), func(t *testing.T) {
			reqBody := CreateGradeRequest{
				Value:     value,
				StudentID: student.ID,
				TopicID:   topic.ID,
			}
			body, _ := json.Marshal(reqBody)

			req := makeGradeAuthenticatedRequest(t, http.MethodPost, "/grades", body, account)
			w := httptest.NewRecorder()

			server.handleCreateGrade(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("handleCreateGrade() status = %d, want %d", w.Code, http.StatusCreated)
			}

			var created edutrack.Grade
			json.NewDecoder(w.Body).Decode(&created)

			if created.Value != value {
				t.Errorf("handleCreateGrade() value = %f, want %f", created.Value, value)
			}
		})
	}
}

func TestHandleCreateGrade_MissingFields(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateGradeRequest
	}{
		{"missing topic_id", CreateGradeRequest{Value: 85.5, TopicID: 1}},
		{"missing subject_id", CreateGradeRequest{Value: 85.5, StudentID: 1}},
		{"missing both ids", CreateGradeRequest{Value: 85.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeGradeAuthenticatedRequest(t, http.MethodPost, "/grades", body, account)
			w := httptest.NewRecorder()

			server.handleCreateGrade(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateGrade() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateGrade_InvalidJSON(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodPost, "/grades", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateGrade(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateGrade() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateGrade_Success(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	grade := createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 70.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	newValue := 85.5
	reqBody := UpdateGradeRequest{
		Value: &newValue,
	}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/grades/%d", grade.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", grade.ID))
	w := httptest.NewRecorder()

	server.handleUpdateGrade(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateGrade() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Grade
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Value != 85.5 {
		t.Errorf("handleUpdateGrade() value = %f, want %f", updated.Value, 85.5)
	}
}

func TestHandleUpdateGrade_UpdateNotes(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	grade := createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	newNotes := "Updated notes - excellent improvement"
	reqBody := UpdateGradeRequest{
		Notes: &newNotes,
	}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/grades/%d", grade.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", grade.ID))
	w := httptest.NewRecorder()

	server.handleUpdateGrade(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateGrade() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Grade
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Notes != newNotes {
		t.Errorf("handleUpdateGrade() notes = %q, want %q", updated.Notes, newNotes)
	}
}

func TestHandleUpdateGrade_UpdateAllFields(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	grade := createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 70.0)

	server := NewServer(":8080", db, []byte("test-secret"))

	newValue := 95.0
	newNotes := "Perfect score"
	reqBody := UpdateGradeRequest{
		Value: &newValue,
		Notes: &newNotes,
	}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/grades/%d", grade.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", grade.ID))
	w := httptest.NewRecorder()

	server.handleUpdateGrade(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateGrade() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Grade
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Value != newValue {
		t.Errorf("handleUpdateGrade() value = %f, want %f", updated.Value, newValue)
	}
	if updated.Notes != newNotes {
		t.Errorf("handleUpdateGrade() notes = %q, want %q", updated.Notes, newNotes)
	}
}

func TestHandleUpdateGrade_NotFound(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newValue := 85.5
	reqBody := UpdateGradeRequest{Value: &newValue}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPut, "/grades/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateGrade(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateGrade() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateGrade_ForbiddenCrossTenant(t *testing.T) {
	db := setupGradeTestDB(t)

	tenant1 := createGradeTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createGradeTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createGradeTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createGradeTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := createGradeTestSubject(t, db, tenant2.ID, career2.ID)
	topic2 := createGradeTestTopic(t, db, tenant2.ID, subject2.ID)

	grade2 := createTestGrade(t, db, tenant2.ID, student2.ID, topic2.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	newValue := 100.0
	reqBody := UpdateGradeRequest{Value: &newValue}
	body, _ := json.Marshal(reqBody)

	req := makeGradeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/grades/%d", grade2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", grade2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateGrade(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateGrade() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteGrade_Success(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createGradeTestCareer(t, db, tenant.ID)
	studentAccount := createGradeTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createGradeTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createGradeTestSubject(t, db, tenant.ID, career.ID)
	topic := createGradeTestTopic(t, db, tenant.ID, subject.ID)

	grade := createTestGrade(t, db, tenant.ID, student.ID, topic.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/grades/%d", grade.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", grade.ID))
	w := httptest.NewRecorder()

	server.handleDeleteGrade(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteGrade() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify grade was deleted
	var found edutrack.Grade
	err := db.First(&found, grade.ID).Error
	if err == nil {
		t.Error("handleDeleteGrade() grade was not deleted")
	}
}

func TestHandleDeleteGrade_NotFound(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodDelete, "/grades/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteGrade(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteGrade() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteGrade_InvalidID(t *testing.T) {
	db := setupGradeTestDB(t)
	tenant := createGradeTestTenant(t, db)
	account := createGradeTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodDelete, "/grades/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteGrade(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteGrade() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteGrade_ForbiddenCrossTenant(t *testing.T) {
	db := setupGradeTestDB(t)

	tenant1 := createGradeTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createGradeTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createGradeTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createGradeTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := &edutrack.Subject{Name: "Subject 2", Code: "S2", TenantID: tenant2.ID}
	db.Create(subject2)
	topic2 := createGradeTestTopic(t, db, tenant2.ID, subject2.ID)

	grade2 := createTestGrade(t, db, tenant2.ID, student2.ID, topic2.ID, 85.5)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeGradeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/grades/%d", grade2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", grade2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteGrade(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteGrade() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListGrades(b *testing.B) {
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

	studentAccount := &edutrack.Account{
		Name:     "Benchmark Student",
		Email:    "student@test.com",
		Password: hashedPassword,
		Role:     edutrack.RoleTeacher,
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(studentAccount)

	student := &edutrack.Student{
		StudentID: "BENCH-001",
		AccountID: studentAccount.ID,
		CareerID:  career.ID,
		TenantID:  tenant.ID,
	}
	db.Create(student)

	teacherAccount := &edutrack.Account{
		Name:     "Benchmark Teacher",
		Email:    "teacher@test.com",
		Password: hashedPassword,
		Role:     edutrack.RoleTeacher,
		Active:   true,
		TenantID: tenant.ID,
	}
	db.Create(teacherAccount)

	teacher := &edutrack.Teacher{
		AccountID: teacherAccount.ID,
		TenantID:  tenant.ID,
	}
	db.Create(teacher)

	subject := &edutrack.Subject{
		Name:     "Benchmark Subject",
		Code:     "BENCHSUB-101",
		TenantID: tenant.ID,
	}
	db.Create(subject)

	topic := &edutrack.Topic{
		Name:      "Benchmark Topic",
		SubjectID: subject.ID,
		TenantID:  tenant.ID,
	}
	db.Create(topic)

	// Create 100 grade records for benchmark
	for i := 0; i < 100; i++ {
		db.Create(&edutrack.Grade{
			Value:     float64(50 + i%51),
			Notes:     "Benchmark grade",
			StudentID: student.ID,
			TopicID:   topic.ID,
			TenantID:  tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/grades", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListGrades(w, req)
	}
}
