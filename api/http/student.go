package http

import (
	"net/http"
	"strconv"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListStudents handles GET /students.
func (s *Server) handleListStudents(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	query := s.DB.Where("students.tenant_id = ?", account.TenantID)

	// Optional filters.
	if careerID := r.URL.Query().Get("career_id"); careerID != "" {
		query = query.Where("career_id = ?", careerID)
	}
	if semester := r.URL.Query().Get("semester"); semester != "" {
		query = query.Where("semester = ?", semester)
	}
	if studentID := r.URL.Query().Get("student_id"); studentID != "" {
		query = query.Where("student_id LIKE ?", "%"+studentID+"%")
	}
	if name := r.URL.Query().Get("name"); name != "" {
		query = query.Joins("Account").Where("Account.name LIKE ?", "%"+name+"%")
	}

	var students []edutrack.Student
	if err := query.Preload("Account").Preload("Career").Find(&students).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Calculate averages for each student.
	for i := range students {
		students[i].CalculateAverages(s.DB)
	}

	sendJSON(w, http.StatusOK, students)
}

// handleGetStudent handles GET /students/{id}.
func (s *Server) handleGetStudent(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var student edutrack.Student
	if err := s.DB.Preload("Account").Preload("Career").First(&student, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if student.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	// Calculate the student's averages.
	student.CalculateAverages(s.DB)

	sendJSON(w, http.StatusOK, student)
}

// CreateStudentRequest represents the request body for creating a student.
type CreateStudentRequest struct {
	StudentID string `json:"student_id"`
	AccountID uint   `json:"account_id"`
	CareerID  uint   `json:"career_id"`
	Semester  int    `json:"semester"`
}

// handleCreateStudent handles POST /students.
func (s *Server) handleCreateStudent(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateStudentRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.StudentID == "" || req.AccountID == 0 {
		sendErrorMessage(w, http.StatusBadRequest, "El ID de estudiante y la cuenta son requeridos.")
		return
	}

	if req.Semester <= 0 {
		sendErrorMessage(w, http.StatusBadRequest, "El semestre debe ser un número positivo.")
		return
	}

	student := &edutrack.Student{
		StudentID: req.StudentID,
		AccountID: req.AccountID,
		CareerID:  req.CareerID,
		Semester:  req.Semester,
		TenantID:  account.TenantID,
	}

	if err := s.DB.Create(student).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusCreated, student)
}

// UpdateStudentRequest represents the request body for updating a student.
type UpdateStudentRequest struct {
	StudentID *string `json:"student_id"`
	CareerID  *uint   `json:"career_id"`
	Semester  *int    `json:"semester"`
}

// handleUpdateStudent handles PUT /students/{id}.
func (s *Server) handleUpdateStudent(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var student edutrack.Student
	if err := s.DB.First(&student, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if student.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateStudentRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.StudentID != nil {
		student.StudentID = *req.StudentID
	}
	if req.CareerID != nil {
		student.CareerID = *req.CareerID
	}
	if req.Semester != nil {
		if *req.Semester <= 0 {
			sendErrorMessage(w, http.StatusBadRequest, "El semestre debe ser un número positivo.")
			return
		}
		student.Semester = *req.Semester
	}

	if err := s.DB.Save(&student).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, student)
}

// handleDeleteStudent handles DELETE /students/{id}.
func (s *Server) handleDeleteStudent(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	var student edutrack.Student
	if err := s.DB.First(&student, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if student.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&student).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
