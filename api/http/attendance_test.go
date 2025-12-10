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

// setupAttendanceTestDB creates an in-memory SQLite database for testing.
func setupAttendanceTestDB(t *testing.T) *gorm.DB {
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

// createAttendanceTestTenant creates a test tenant with a valid license.
func createAttendanceTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save test tenant: %v", err)
	}

	return tenant
}

// createAttendanceTestAccount creates a test account for a tenant.
func createAttendanceTestAccount(t *testing.T, db *gorm.DB, tenantID, email, name string, role edutrack.Role) *edutrack.Account {
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

// createAttendanceTestCareer creates a test career for a tenant.
func createAttendanceTestCareer(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Career {
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

// createAttendanceTestStudent creates a test student for a tenant.
func createAttendanceTestStudent(t *testing.T, db *gorm.DB, tenantID string, accountID, careerID uint) *edutrack.Student {
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

// createAttendanceTestSubject creates a test subject for a tenant.
func createAttendanceTestSubject(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Subject {
	subject := &edutrack.Subject{
		Name:        "Matemáticas I",
		Code:        fmt.Sprintf("MAT-%d", time.Now().UnixNano()),
		Description: "Test subject",
		Credits:     5,
		TenantID:    tenantID,
	}

	if err := db.Create(subject).Error; err != nil {
		t.Fatalf("Failed to create test subject: %v", err)
	}

	return subject
}

// createTestAttendance creates a test attendance record.
func createTestAttendance(t *testing.T, db *gorm.DB, tenantID string, studentID, subjectID uint, date time.Time, status edutrack.AttendanceStatus) *edutrack.Attendance {
	attendance := &edutrack.Attendance{
		Date:      date,
		Status:    status,
		Notes:     "Test attendance",
		StudentID: studentID,
		SubjectID: subjectID,
		TenantID:  tenantID,
	}

	if err := db.Create(attendance).Error; err != nil {
		t.Fatalf("Failed to create test attendance: %v", err)
	}

	return attendance
}

// makeAttendanceAuthenticatedRequest creates an HTTP request with the account in context.
func makeAttendanceAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListAttendances_Success(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)
	createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today.AddDate(0, 0, -1), edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, "/attendances", nil, account)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAttendances() status = %d, want %d", w.Code, http.StatusOK)
	}

	var attendances []edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&attendances); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(attendances) != 2 {
		t.Errorf("handleListAttendances() returned %d attendances, want 2", len(attendances))
	}
}

func TestHandleListAttendances_FilterByStudentID(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)

	studentAccount1 := createAttendanceTestAccount(t, db, tenant.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createAttendanceTestAccount(t, db, tenant.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student1 := createAttendanceTestStudent(t, db, tenant.ID, studentAccount1.ID, career.ID)
	student2 := createAttendanceTestStudent(t, db, tenant.ID, studentAccount2.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	createTestAttendance(t, db, tenant.ID, student1.ID, subject.ID, today, edutrack.AttendancePresent)
	createTestAttendance(t, db, tenant.ID, student2.ID, subject.ID, today, edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/attendances?student_id=%d", student1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAttendances() status = %d, want %d", w.Code, http.StatusOK)
	}

	var attendances []edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&attendances); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(attendances) != 1 {
		t.Errorf("handleListAttendances() with student_id filter returned %d attendances, want 1", len(attendances))
	}
}

func TestHandleListAttendances_FilterBySubjectID(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject1 := createAttendanceTestSubject(t, db, tenant.ID)
	subject2 := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	createTestAttendance(t, db, tenant.ID, student.ID, subject1.ID, today, edutrack.AttendancePresent)
	createTestAttendance(t, db, tenant.ID, student.ID, subject2.ID, today, edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/attendances?subject_id=%d", subject1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAttendances() status = %d, want %d", w.Code, http.StatusOK)
	}

	var attendances []edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&attendances); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(attendances) != 1 {
		t.Errorf("handleListAttendances() with subject_id filter returned %d attendances, want 1", len(attendances))
	}
}

func TestHandleListAttendances_FilterByDate(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)
	createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, yesterday, edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/attendances?date=%s", today.Format("2006-01-02")), nil, account)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListAttendances() status = %d, want %d", w.Code, http.StatusOK)
	}

	var attendances []edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&attendances); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(attendances) != 1 {
		t.Errorf("handleListAttendances() with date filter returned %d attendances, want 1", len(attendances))
	}
}

func TestHandleListAttendances_Unauthorized(t *testing.T) {
	db := setupAttendanceTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/attendances", nil)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListAttendances() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListAttendances_TenantIsolation(t *testing.T) {
	db := setupAttendanceTestDB(t)

	tenant1 := createAttendanceTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAttendanceTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career1 := createAttendanceTestCareer(t, db, tenant1.ID)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)

	studentAccount1 := createAttendanceTestAccount(t, db, tenant1.ID, "student1@test.com", "Student 1", edutrack.RoleTeacher)
	studentAccount2 := createAttendanceTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student1 := createAttendanceTestStudent(t, db, tenant1.ID, studentAccount1.ID, career1.ID)
	student2 := createAttendanceTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)

	subject1 := createAttendanceTestSubject(t, db, tenant1.ID)
	subject2 := &edutrack.Subject{Name: "Subject 2", Code: "S2", TenantID: tenant2.ID}
	db.Create(subject2)

	today := time.Now().Truncate(24 * time.Hour)
	createTestAttendance(t, db, tenant1.ID, student1.ID, subject1.ID, today, edutrack.AttendancePresent)
	createTestAttendance(t, db, tenant2.ID, student2.ID, subject2.ID, today, edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, "/attendances", nil, account1)
	w := httptest.NewRecorder()

	server.handleListAttendances(w, req)

	var attendances []edutrack.Attendance
	json.NewDecoder(w.Body).Decode(&attendances)

	if len(attendances) != 1 {
		t.Errorf("handleListAttendances() should only return attendances from same tenant, got %d", len(attendances))
	}
}

func TestHandleGetAttendance_Success(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/attendances/%d", attendance.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleGetAttendance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetAttendance() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Status != edutrack.AttendancePresent {
		t.Errorf("handleGetAttendance() status = %q, want %q", found.Status, edutrack.AttendancePresent)
	}
}

func TestHandleGetAttendance_NotFound(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, "/attendances/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetAttendance(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetAttendance() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetAttendance_InvalidID(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, "/attendances/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetAttendance() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetAttendance_ForbiddenCrossTenant(t *testing.T) {
	db := setupAttendanceTestDB(t)

	tenant1 := createAttendanceTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAttendanceTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createAttendanceTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createAttendanceTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := &edutrack.Subject{Name: "Subject 2", Code: "S2", TenantID: tenant2.ID}
	db.Create(subject2)

	today := time.Now().Truncate(24 * time.Hour)
	attendance2 := createTestAttendance(t, db, tenant2.ID, student2.ID, subject2.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/attendances/%d", attendance2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance2.ID))
	w := httptest.NewRecorder()

	server.handleGetAttendance(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetAttendance() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateAttendance_Success(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAttendanceRequest{
		Date:      "2024-01-15",
		Status:    edutrack.AttendancePresent,
		Notes:     "On time",
		StudentID: student.ID,
		SubjectID: subject.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateAttendance() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Status != edutrack.AttendancePresent {
		t.Errorf("handleCreateAttendance() status = %q, want %q", created.Status, edutrack.AttendancePresent)
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateAttendance() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}
}

func TestHandleCreateAttendance_AllStatuses(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	statuses := []edutrack.AttendanceStatus{
		edutrack.AttendancePresent,
		edutrack.AttendanceAbsent,
		edutrack.AttendanceLate,
		edutrack.AttendanceExcused,
	}

	for i, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			reqBody := CreateAttendanceRequest{
				Date:      fmt.Sprintf("2024-01-%02d", i+1),
				Status:    status,
				StudentID: student.ID,
				SubjectID: subject.ID,
			}
			body, _ := json.Marshal(reqBody)

			req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
			w := httptest.NewRecorder()

			server.handleCreateAttendance(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("handleCreateAttendance() status = %d, want %d", w.Code, http.StatusCreated)
			}

			var created edutrack.Attendance
			json.NewDecoder(w.Body).Decode(&created)

			if created.Status != status {
				t.Errorf("handleCreateAttendance() status = %q, want %q", created.Status, status)
			}
		})
	}
}

func TestHandleCreateAttendance_InvalidDate(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAttendanceRequest{
		Date:      "invalid-date",
		Status:    edutrack.AttendancePresent,
		StudentID: student.ID,
		SubjectID: subject.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAttendance() invalid date status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateAttendance_InvalidStatus(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAttendanceRequest{
		Date:      "2024-01-15",
		Status:    "invalid_status",
		StudentID: student.ID,
		SubjectID: subject.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAttendance() invalid status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateAttendance_MissingFields(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateAttendanceRequest
	}{
		{"missing student_id", CreateAttendanceRequest{Date: "2024-01-15", Status: edutrack.AttendancePresent, SubjectID: 1}},
		{"missing subject_id", CreateAttendanceRequest{Date: "2024-01-15", Status: edutrack.AttendancePresent, StudentID: 1}},
		{"missing both ids", CreateAttendanceRequest{Date: "2024-01-15", Status: edutrack.AttendancePresent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
			w := httptest.NewRecorder()

			server.handleCreateAttendance(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateAttendance() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateAttendance_StudentNotFound(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAttendanceRequest{
		Date:      "2024-01-15",
		Status:    edutrack.AttendancePresent,
		StudentID: 99999,
		SubjectID: subject.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAttendance() student not found status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateAttendance_SubjectNotFound(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateAttendanceRequest{
		Date:      "2024-01-15",
		Status:    edutrack.AttendancePresent,
		StudentID: student.ID,
		SubjectID: 99999,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", body, account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAttendance() subject not found status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCreateAttendance_InvalidJSON(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPost, "/attendances", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateAttendance() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateAttendance_Success(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendanceAbsent)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStatus := edutrack.AttendanceExcused
	reqBody := UpdateAttendanceRequest{
		Status: &newStatus,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAttendance() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Attendance
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Status != edutrack.AttendanceExcused {
		t.Errorf("handleUpdateAttendance() status = %q, want %q", updated.Status, edutrack.AttendanceExcused)
	}
}

func TestHandleUpdateAttendance_UpdateDate(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	newDate := "2024-02-20"
	reqBody := UpdateAttendanceRequest{
		Date: &newDate,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAttendance() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Attendance
	json.NewDecoder(w.Body).Decode(&updated)

	expectedDate, _ := time.Parse("2006-01-02", newDate)
	if !updated.Date.Equal(expectedDate) {
		t.Errorf("handleUpdateAttendance() date = %v, want %v", updated.Date, expectedDate)
	}
}

func TestHandleUpdateAttendance_UpdateNotes(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	newNotes := "Updated notes"
	reqBody := UpdateAttendanceRequest{
		Notes: &newNotes,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateAttendance() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Attendance
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Notes != newNotes {
		t.Errorf("handleUpdateAttendance() notes = %q, want %q", updated.Notes, newNotes)
	}
}

func TestHandleUpdateAttendance_InvalidDate(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	invalidDate := "invalid-date"
	reqBody := UpdateAttendanceRequest{
		Date: &invalidDate,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleUpdateAttendance() invalid date status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateAttendance_InvalidStatus(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	invalidStatus := edutrack.AttendanceStatus("invalid_status")
	reqBody := UpdateAttendanceRequest{
		Status: &invalidStatus,
	}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleUpdateAttendance() invalid status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateAttendance_NotFound(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStatus := edutrack.AttendancePresent
	reqBody := UpdateAttendanceRequest{Status: &newStatus}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, "/attendances/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateAttendance() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateAttendance_ForbiddenCrossTenant(t *testing.T) {
	db := setupAttendanceTestDB(t)

	tenant1 := createAttendanceTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAttendanceTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createAttendanceTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createAttendanceTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := &edutrack.Subject{Name: "Subject 2", Code: "S2", TenantID: tenant2.ID}
	db.Create(subject2)

	today := time.Now().Truncate(24 * time.Hour)
	attendance2 := createTestAttendance(t, db, tenant2.ID, student2.ID, subject2.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	newStatus := edutrack.AttendanceAbsent
	reqBody := UpdateAttendanceRequest{Status: &newStatus}
	body, _ := json.Marshal(reqBody)

	req := makeAttendanceAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/attendances/%d", attendance2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateAttendance(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateAttendance() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteAttendance_Success(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createAttendanceTestCareer(t, db, tenant.ID)
	studentAccount := createAttendanceTestAccount(t, db, tenant.ID, "student@test.com", "Student", edutrack.RoleTeacher)
	student := createAttendanceTestStudent(t, db, tenant.ID, studentAccount.ID, career.ID)
	subject := createAttendanceTestSubject(t, db, tenant.ID)

	today := time.Now().Truncate(24 * time.Hour)
	attendance := createTestAttendance(t, db, tenant.ID, student.ID, subject.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/attendances/%d", attendance.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAttendance(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteAttendance() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify attendance was deleted
	var found edutrack.Attendance
	err := db.First(&found, attendance.ID).Error
	if err == nil {
		t.Error("handleDeleteAttendance() attendance was not deleted")
	}
}

func TestHandleDeleteAttendance_NotFound(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodDelete, "/attendances/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteAttendance(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteAttendance() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteAttendance_InvalidID(t *testing.T) {
	db := setupAttendanceTestDB(t)
	tenant := createAttendanceTestTenant(t, db)
	account := createAttendanceTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodDelete, "/attendances/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteAttendance(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteAttendance() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteAttendance_ForbiddenCrossTenant(t *testing.T) {
	db := setupAttendanceTestDB(t)

	tenant1 := createAttendanceTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createAttendanceTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	studentAccount2 := createAttendanceTestAccount(t, db, tenant2.ID, "student2@test.com", "Student 2", edutrack.RoleTeacher)
	student2 := createAttendanceTestStudent(t, db, tenant2.ID, studentAccount2.ID, career2.ID)
	subject2 := &edutrack.Subject{Name: "Subject 2", Code: "S2", TenantID: tenant2.ID}
	db.Create(subject2)

	today := time.Now().Truncate(24 * time.Hour)
	attendance2 := createTestAttendance(t, db, tenant2.ID, student2.ID, subject2.ID, today, edutrack.AttendancePresent)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeAttendanceAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/attendances/%d", attendance2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", attendance2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteAttendance(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteAttendance() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListAttendances(b *testing.B) {
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

	subject := &edutrack.Subject{
		Name:     "Benchmark Subject",
		Code:     "BENCHSUB-101",
		TenantID: tenant.ID,
	}
	db.Create(subject)

	// Create 100 attendance records for benchmark
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 100; i++ {
		db.Create(&edutrack.Attendance{
			Date:      today.AddDate(0, 0, -i),
			Status:    edutrack.AttendancePresent,
			StudentID: student.ID,
			SubjectID: subject.ID,
			TenantID:  tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/attendances", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListAttendances(w, req)
	}
}
