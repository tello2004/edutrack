package http

import (
	"net/http"
	"strconv"
	"time"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

// handleListAttendances handles GET /attendances.
func (s *Server) handleListAttendances(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var attendances []edutrack.Attendance
	query := s.DB.Where("tenant_id = ?", account.TenantID).Preload("Student").Preload("Subject")

	// Optional filters.
	if studentID := r.URL.Query().Get("student_id"); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}
	if subjectID := r.URL.Query().Get("subject_id"); subjectID != "" {
		query = query.Where("subject_id = ?", subjectID)
	}
	if date := r.URL.Query().Get("date"); date != "" {
		query = query.Where("DATE(date) = ?", date)
	}

	if err := query.Find(&attendances).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	sendJSON(w, http.StatusOK, attendances)
}

// handleGetAttendance handles GET /attendances/{id}.
func (s *Server) handleGetAttendance(w http.ResponseWriter, r *http.Request) {
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

	var attendance edutrack.Attendance
	if err := s.DB.Preload("Student").Preload("Subject").First(&attendance, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if attendance.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	sendJSON(w, http.StatusOK, attendance)
}

// CreateAttendanceRequest represents the request body for creating an attendance record.
type CreateAttendanceRequest struct {
	Date      string                    `json:"date"` // Format: "2006-01-02"
	Status    edutrack.AttendanceStatus `json:"status"`
	Notes     string                    `json:"notes"`
	StudentID uint                      `json:"student_id"`
	SubjectID uint                      `json:"subject_id"`
}

// handleCreateAttendance handles POST /attendances.
func (s *Server) handleCreateAttendance(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	var req CreateAttendanceRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	// Parse the date.
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "Formato de fecha inválido. Use YYYY-MM-DD.")
		return
	}

	// Validate status.
	if !isValidAttendanceStatus(req.Status) {
		sendErrorMessage(w, http.StatusBadRequest, "Estado de asistencia inválido.")
		return
	}

	if req.StudentID == 0 || req.SubjectID == 0 {
		sendErrorMessage(w, http.StatusBadRequest, "El ID de estudiante y materia son requeridos.")
		return
	}

	// Verify student belongs to the same tenant.
	var student edutrack.Student
	if err := s.DB.First(&student, req.StudentID).Error; err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "El estudiante especificado no existe.")
		return
	}
	if student.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	// Verify subject belongs to the same tenant.
	var subject edutrack.Subject
	if err := s.DB.First(&subject, req.SubjectID).Error; err != nil {
		sendErrorMessage(w, http.StatusBadRequest, "La materia especificada no existe.")
		return
	}
	if subject.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	attendance := &edutrack.Attendance{
		Date:      date,
		Status:    req.Status,
		Notes:     req.Notes,
		StudentID: req.StudentID,
		SubjectID: req.SubjectID,
		TenantID:  account.TenantID,
	}

	if err := s.DB.Create(attendance).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Student").Preload("Subject").First(attendance, attendance.ID)

	sendJSON(w, http.StatusCreated, attendance)
}

// UpdateAttendanceRequest represents the request body for updating an attendance record.
type UpdateAttendanceRequest struct {
	Date   *string                    `json:"date"` // Format: "2006-01-02"
	Status *edutrack.AttendanceStatus `json:"status"`
	Notes  *string                    `json:"notes"`
}

// handleUpdateAttendance handles PUT /attendances/{id}.
func (s *Server) handleUpdateAttendance(w http.ResponseWriter, r *http.Request) {
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

	var attendance edutrack.Attendance
	if err := s.DB.First(&attendance, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if attendance.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	var req UpdateAttendanceRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			sendErrorMessage(w, http.StatusBadRequest, "Formato de fecha inválido. Use YYYY-MM-DD.")
			return
		}
		attendance.Date = date
	}

	if req.Status != nil {
		if !isValidAttendanceStatus(*req.Status) {
			sendErrorMessage(w, http.StatusBadRequest, "Estado de asistencia inválido.")
			return
		}
		attendance.Status = *req.Status
	}

	if req.Notes != nil {
		attendance.Notes = *req.Notes
	}

	if err := s.DB.Save(&attendance).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	// Reload with associations.
	s.DB.Preload("Student").Preload("Subject").First(&attendance, attendance.ID)

	sendJSON(w, http.StatusOK, attendance)
}

// handleDeleteAttendance handles DELETE /attendances/{id}.
func (s *Server) handleDeleteAttendance(w http.ResponseWriter, r *http.Request) {
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

	var attendance edutrack.Attendance
	if err := s.DB.First(&attendance, id).Error; err != nil {
		sendError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	if attendance.TenantID != account.TenantID {
		sendError(w, http.StatusForbidden, ErrForbidden)
		return
	}

	if err := s.DB.Delete(&attendance).Error; err != nil {
		sendError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isValidAttendanceStatus checks if the given status is a valid attendance status.
func isValidAttendanceStatus(status edutrack.AttendanceStatus) bool {
	switch status {
	case edutrack.AttendancePresent,
		edutrack.AttendanceAbsent,
		edutrack.AttendanceLate,
		edutrack.AttendanceExcused:
		return true
	default:
		return false
	}
}
