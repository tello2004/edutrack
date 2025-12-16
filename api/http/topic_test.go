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

func setupTopicTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	edutrack.Migrate(db)
	return db
}

func createTopicTestTenant(t *testing.T, db *gorm.DB) *edutrack.Tenant {
	tenant, err := edutrack.NewTenant("Test Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to save tenant: %v", err)
	}
	return tenant
}

func createTopicTestAccount(t *testing.T, db *gorm.DB, tenantID string, email, name string, role edutrack.Role) *edutrack.Account {
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
		t.Fatalf("Failed to create account: %v", err)
	}
	return account
}

func createTopicTestCareer(t *testing.T, db *gorm.DB, tenantID string) *edutrack.Career {
	career := &edutrack.Career{
		Name:     "Test Career",
		Code:     "TEST-2024",
		Active:   true,
		TenantID: tenantID,
	}
	if err := db.Create(career).Error; err != nil {
		t.Fatalf("Failed to create career: %v", err)
	}
	return career
}

func createTopicTestSubject(t *testing.T, db *gorm.DB, tenantID string, careerID uint) *edutrack.Subject {
	subject := &edutrack.Subject{
		Name:        "Test Subject",
		Code:        fmt.Sprintf("SUB-%d", time.Now().UnixNano()),
		Description: "Test subject",
		Credits:     5,
		CareerID:    careerID,
		TenantID:    tenantID,
	}
	if err := db.Create(subject).Error; err != nil {
		t.Fatalf("Failed to create subject: %v", err)
	}
	return subject
}

func createTestTopic(t *testing.T, db *gorm.DB, tenantID string, subjectID uint, name string) *edutrack.Topic {
	topic := &edutrack.Topic{
		Name:      name,
		SubjectID: subjectID,
		TenantID:  tenantID,
	}
	if err := db.Create(topic).Error; err != nil {
		t.Fatalf("Failed to create test topic: %v", err)
	}
	return topic
}

func makeTopicAuthenticatedRequest(t *testing.T, method, path string, body []byte, account *edutrack.Account) *http.Request {
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

func TestHandleListTopics_Success(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	createTestTopic(t, db, tenant.ID, subject.ID, "Topic 1")
	createTestTopic(t, db, tenant.ID, subject.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, "/topics", nil, account)
	w := httptest.NewRecorder()

	server.handleListTopics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListTopics() status = %d, want %d", w.Code, http.StatusOK)
	}

	var topics []edutrack.Topic
	if err := json.NewDecoder(w.Body).Decode(&topics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(topics) != 2 {
		t.Errorf("handleListTopics() returned %d topics, want 2", len(topics))
	}
}

func TestHandleListTopics_FilterBySubjectID(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject1 := createTopicTestSubject(t, db, tenant.ID, career.ID)
	subject2 := createTopicTestSubject(t, db, tenant.ID, career.ID)

	createTestTopic(t, db, tenant.ID, subject1.ID, "Topic 1")
	createTestTopic(t, db, tenant.ID, subject2.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/topics?subject_id=%d", subject1.ID), nil, account)
	w := httptest.NewRecorder()

	server.handleListTopics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleListTopics() status = %d, want %d", w.Code, http.StatusOK)
	}

	var topics []edutrack.Topic
	if err := json.NewDecoder(w.Body).Decode(&topics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(topics) != 1 {
		t.Errorf("handleListTopics() with subject_id filter returned %d topics, want 1", len(topics))
	}
}

func TestHandleListTopics_Unauthorized(t *testing.T) {
	db := setupTopicTestDB(t)
	server := NewServer(":8080", db, []byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/topics", nil)
	w := httptest.NewRecorder()

	server.handleListTopics(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("handleListTopics() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleListTopics_TenantIsolation(t *testing.T) {
	db := setupTopicTestDB(t)

	tenant1 := createTopicTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTopicTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career1 := createTopicTestCareer(t, db, tenant1.ID)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)

	subject1 := createTopicTestSubject(t, db, tenant1.ID, career1.ID)
	subject2 := createTopicTestSubject(t, db, tenant2.ID, career2.ID)

	createTestTopic(t, db, tenant1.ID, subject1.ID, "Topic 1")
	createTestTopic(t, db, tenant2.ID, subject2.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, "/topics", nil, account1)
	w := httptest.NewRecorder()

	server.handleListTopics(w, req)

	var topics []edutrack.Topic
	json.NewDecoder(w.Body).Decode(&topics)

	if len(topics) != 1 {
		t.Errorf("handleListTopics() should only return topics from same tenant, got %d", len(topics))
	}
}

func TestHandleGetTopic_Success(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	topic := createTestTopic(t, db, tenant.ID, subject.ID, "Test Topic")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/topics/%d", topic.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", topic.ID))
	w := httptest.NewRecorder()

	server.handleGetTopic(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetTopic() status = %d, want %d", w.Code, http.StatusOK)
	}

	var found edutrack.Topic
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if found.Name != "Test Topic" {
		t.Errorf("handleGetTopic() name = %q, want %q", found.Name, "Test Topic")
	}
}

func TestHandleGetTopic_NotFound(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, "/topics/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleGetTopic(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleGetTopic() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetTopic_InvalidID(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, "/topics/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleGetTopic(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleGetTopic() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetTopic_ForbiddenCrossTenant(t *testing.T) {
	db := setupTopicTestDB(t)

	tenant1 := createTopicTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTopicTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	subject2 := createTopicTestSubject(t, db, tenant2.ID, career2.ID)

	topic2 := createTestTopic(t, db, tenant2.ID, subject2.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/topics/%d", topic2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", topic2.ID))
	w := httptest.NewRecorder()

	server.handleGetTopic(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleGetTopic() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleCreateTopic_Success(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	server := NewServer(":8080", db, []byte("test-secret"))

	reqBody := CreateTopicRequest{
		Name:        "New Topic",
		Description: "A new topic",
		SubjectID:   subject.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTopicAuthenticatedRequest(t, http.MethodPost, "/topics", body, account)
	w := httptest.NewRecorder()

	server.handleCreateTopic(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("handleCreateTopic() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var created edutrack.Topic
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Name != "New Topic" {
		t.Errorf("handleCreateTopic() name = %q, want %q", created.Name, "New Topic")
	}

	if created.TenantID != tenant.ID {
		t.Errorf("handleCreateTopic() tenant_id = %q, want %q", created.TenantID, tenant.ID)
	}
}

func TestHandleCreateTopic_MissingFields(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	tests := []struct {
		name    string
		request CreateTopicRequest
	}{
		{"missing name", CreateTopicRequest{SubjectID: 1}},
		{"missing subject_id", CreateTopicRequest{Name: "Topic"}},
		{"missing both", CreateTopicRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := makeTopicAuthenticatedRequest(t, http.MethodPost, "/topics", body, account)
			w := httptest.NewRecorder()

			server.handleCreateTopic(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("handleCreateTopic() %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleCreateTopic_InvalidJSON(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodPost, "/topics", []byte("invalid json"), account)
	w := httptest.NewRecorder()

	server.handleCreateTopic(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleCreateTopic() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdateTopic_Success(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	topic := createTestTopic(t, db, tenant.ID, subject.ID, "Old Topic")

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated Topic"
	reqBody := UpdateTopicRequest{
		Name: &newName,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTopicAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/topics/%d", topic.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", topic.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTopic(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateTopic() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Topic
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "Updated Topic" {
		t.Errorf("handleUpdateTopic() name = %q, want %q", updated.Name, "Updated Topic")
	}
}

func TestHandleUpdateTopic_UpdateDescription(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	topic := createTestTopic(t, db, tenant.ID, subject.ID, "Topic")

	server := NewServer(":8080", db, []byte("test-secret"))

	newDesc := "Updated description"
	reqBody := UpdateTopicRequest{
		Description: &newDesc,
	}
	body, _ := json.Marshal(reqBody)

	req := makeTopicAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/topics/%d", topic.ID), body, account)
	req.SetPathValue("id", fmt.Sprintf("%d", topic.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTopic(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleUpdateTopic() status = %d, want %d", w.Code, http.StatusOK)
	}

	var updated edutrack.Topic
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Description != newDesc {
		t.Errorf("handleUpdateTopic() description = %q, want %q", updated.Description, newDesc)
	}
}

func TestHandleUpdateTopic_NotFound(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated"
	reqBody := UpdateTopicRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeTopicAuthenticatedRequest(t, http.MethodPut, "/topics/99999", body, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleUpdateTopic(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleUpdateTopic() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdateTopic_ForbiddenCrossTenant(t *testing.T) {
	db := setupTopicTestDB(t)

	tenant1 := createTopicTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTopicTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	subject2 := createTopicTestSubject(t, db, tenant2.ID, career2.ID)

	topic2 := createTestTopic(t, db, tenant2.ID, subject2.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	newName := "Updated"
	reqBody := UpdateTopicRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := makeTopicAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/topics/%d", topic2.ID), body, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", topic2.ID))
	w := httptest.NewRecorder()

	server.handleUpdateTopic(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleUpdateTopic() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleDeleteTopic_Success(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)
	career := createTopicTestCareer(t, db, tenant.ID)
	subject := createTopicTestSubject(t, db, tenant.ID, career.ID)

	topic := createTestTopic(t, db, tenant.ID, subject.ID, "Topic to Delete")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/topics/%d", topic.ID), nil, account)
	req.SetPathValue("id", fmt.Sprintf("%d", topic.ID))
	w := httptest.NewRecorder()

	server.handleDeleteTopic(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("handleDeleteTopic() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify topic was deleted
	var found edutrack.Topic
	err := db.First(&found, topic.ID).Error
	if err == nil {
		t.Error("handleDeleteTopic() topic was not deleted")
	}
}

func TestHandleDeleteTopic_NotFound(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodDelete, "/topics/99999", nil, account)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	server.handleDeleteTopic(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("handleDeleteTopic() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleDeleteTopic_InvalidID(t *testing.T) {
	db := setupTopicTestDB(t)
	tenant := createTopicTestTenant(t, db)
	account := createTopicTestAccount(t, db, tenant.ID, "admin@test.com", "Admin", edutrack.RoleSecretary)

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodDelete, "/topics/invalid", nil, account)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	server.handleDeleteTopic(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleDeleteTopic() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleDeleteTopic_ForbiddenCrossTenant(t *testing.T) {
	db := setupTopicTestDB(t)

	tenant1 := createTopicTestTenant(t, db)
	tenant2, _ := edutrack.NewTenant("Other Institution", edutrack.LicenseTypeTrial, 30*24*time.Hour)
	db.Create(tenant2)

	account1 := createTopicTestAccount(t, db, tenant1.ID, "admin1@test.com", "Admin 1", edutrack.RoleSecretary)
	career2 := &edutrack.Career{Name: "Career 2", Code: "C2", TenantID: tenant2.ID, Active: true}
	db.Create(career2)
	subject2 := createTopicTestSubject(t, db, tenant2.ID, career2.ID)

	topic2 := createTestTopic(t, db, tenant2.ID, subject2.ID, "Topic 2")

	server := NewServer(":8080", db, []byte("test-secret"))

	req := makeTopicAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/topics/%d", topic2.ID), nil, account1)
	req.SetPathValue("id", fmt.Sprintf("%d", topic2.ID))
	w := httptest.NewRecorder()

	server.handleDeleteTopic(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("handleDeleteTopic() cross-tenant status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func BenchmarkHandleListTopics(b *testing.B) {
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

	subject := &edutrack.Subject{
		Name:     "Benchmark Subject",
		Code:     "BENCHSUB-101",
		TenantID: tenant.ID,
	}
	db.Create(subject)

	// Create 100 topic records for benchmark
	for i := 0; i < 100; i++ {
		db.Create(&edutrack.Topic{
			Name:      fmt.Sprintf("Benchmark Topic %d", i),
			SubjectID: subject.ID,
			TenantID:  tenant.ID,
		})
	}

	server := NewServer(":8080", db, []byte("test-secret"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/topics", nil)
		ctx := edutrack.NewContextWithAccount(req.Context(), account)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.handleListTopics(w, req)
	}
}
