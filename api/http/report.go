package http

import (
	"net/http"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

type StudentAverage struct {
	StudentID uint    `json:"student_id"`
	Name      string  `json:"name"`
	Average   float64 `json:"average"`
}

func (s *Server) handleStudentAverages(w http.ResponseWriter, r *http.Request) {
	account := edutrack.AccountFromContext(r.Context())
	if account == nil {
		sendError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("group_id")

	query := s.DB.
		Table("grades").
		Select(`
			students.id as student_id,
			accounts.name as name,
			AVG(grades.value) as average
		`).
		Joins("JOIN students ON students.id = grades.student_id").
		Joins("JOIN accounts ON accounts.id = students.account_id").
		Where("grades.tenant_id = ?", account.TenantID).
		Group("students.id, accounts.name")

	if groupID != "" {
		query = query.Where("grades.group_id = ?", groupID)
	}

	var result []StudentAverage
	query.Scan(&result)

	sendJSON(w, http.StatusOK, result)
}
