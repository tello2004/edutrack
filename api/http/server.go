package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// Server represents the HTTP server for the API.
type Server struct {
	server *http.Server
	router *http.ServeMux

	// Database connection.
	DB *gorm.DB

	// JWT secret key for authentication.
	JWTSecret []byte

	// CORS configuration.
	CORSConfig *CORSConfig
}

// NewServer creates a new HTTP server.
func NewServer(addr string, db *gorm.DB, jwtSecret []byte) *Server {
	s := &Server{
		router:     http.NewServeMux(),
		DB:         db,
		JWTSecret:  jwtSecret,
		CORSConfig: DefaultCORSConfig(),
	}

	s.registerRoutes()

	// Wrap the router with CORS middleware.
	handler := withCORS(s.CORSConfig, s.router)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// NewServerWithCORS creates a new HTTP server with custom CORS configuration.
func NewServerWithCORS(addr string, db *gorm.DB, jwtSecret []byte, corsConfig *CORSConfig) *Server {
	s := &Server{
		router:     http.NewServeMux(),
		DB:         db,
		JWTSecret:  jwtSecret,
		CORSConfig: corsConfig,
	}

	if s.CORSConfig == nil {
		s.CORSConfig = DefaultCORSConfig()
	}

	s.registerRoutes()

	// Wrap the router with CORS middleware.
	handler := withCORS(s.CORSConfig, s.router)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Close shuts down the HTTP server immediately.
func (s *Server) Close() error {
	return s.server.Close()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerRoutes registers all API routes.
func (s *Server) registerRoutes() {
	// Auth (public)
	s.router.HandleFunc("POST /auth/login", s.handleLogin)
	s.router.HandleFunc("POST /auth/license", s.handleLicenseLogin)

	// Protected routes (require authentication)
	protected := s.withAuth

	// Accounts
	s.router.HandleFunc("GET /accounts", protected(s.handleListAccounts))
	s.router.HandleFunc("GET /accounts/{id}", protected(s.handleGetAccount))
	s.router.HandleFunc("POST /accounts", protected(s.handleCreateAccount))
	s.router.HandleFunc("PUT /accounts/{id}", protected(s.handleUpdateAccount))
	s.router.HandleFunc("DELETE /accounts/{id}", protected(s.handleDeleteAccount))

	// Students
	s.router.HandleFunc("GET /students", protected(s.handleListStudents))
	s.router.HandleFunc("GET /students/{id}", protected(s.handleGetStudent))
	s.router.HandleFunc("POST /students", protected(s.handleCreateStudent))
	s.router.HandleFunc("PUT /students/{id}", protected(s.handleUpdateStudent))
	s.router.HandleFunc("DELETE /students/{id}", protected(s.handleDeleteStudent))

	// Teachers
	s.router.HandleFunc("GET /teachers", protected(s.handleListTeachers))
	s.router.HandleFunc("GET /teachers/{id}", protected(s.handleGetTeacher))
	s.router.HandleFunc("POST /teachers", protected(s.handleCreateTeacher))
	s.router.HandleFunc("PUT /teachers/{id}", protected(s.handleUpdateTeacher))
	s.router.HandleFunc("DELETE /teachers/{id}", protected(s.handleDeleteTeacher))

	// Careers
	s.router.HandleFunc("GET /careers", protected(s.handleListCareers))
	s.router.HandleFunc("GET /careers/{id}", protected(s.handleGetCareer))
	s.router.HandleFunc("POST /careers", protected(s.handleCreateCareer))
	s.router.HandleFunc("PUT /careers/{id}", protected(s.handleUpdateCareer))
	s.router.HandleFunc("DELETE /careers/{id}", protected(s.handleDeleteCareer))

	// Subjects
	s.router.HandleFunc("GET /subjects", protected(s.handleListSubjects))
	s.router.HandleFunc("GET /subjects/{id}", protected(s.handleGetSubject))
	s.router.HandleFunc("POST /subjects", protected(s.handleCreateSubject))
	s.router.HandleFunc("PUT /subjects/{id}", protected(s.handleUpdateSubject))
	s.router.HandleFunc("DELETE /subjects/{id}", protected(s.handleDeleteSubject))
	s.router.HandleFunc("GET /subjects/{id}/students", protected(s.handleListSubjectStudents))
	s.router.HandleFunc("POST /subjects/{id}/students", protected(s.handleAddStudentToSubject))
	s.router.HandleFunc("DELETE /subjects/{id}/students/{student_id}", protected(s.handleRemoveStudentFromSubject))

	// Topics
	s.router.HandleFunc("GET /topics", protected(s.handleListTopics))
	s.router.HandleFunc("GET /topics/{id}", protected(s.handleGetTopic))
	s.router.HandleFunc("POST /topics", protected(s.handleCreateTopic))
	s.router.HandleFunc("PUT /topics/{id}", protected(s.handleUpdateTopic))
	s.router.HandleFunc("DELETE /topics/{id}", protected(s.handleDeleteTopic))

	// Attendances
	s.router.HandleFunc("GET /attendances", protected(s.handleListAttendances))
	s.router.HandleFunc("GET /attendances/{id}", protected(s.handleGetAttendance))
	s.router.HandleFunc("POST /attendances", protected(s.handleCreateAttendance))
	s.router.HandleFunc("PUT /attendances/{id}", protected(s.handleUpdateAttendance))
	s.router.HandleFunc("DELETE /attendances/{id}", protected(s.handleDeleteAttendance))

	// test
	mux.Handle("POST /groups/{id}/subjects", s.auth(s.handleAssignSubjectsToGroup))
	mux.Handle("GET /groups/{id}/subjects", s.auth(s.handleListGroupSubjects))
	mux.Handle("POST /groups/{groupId}/subjects/{subjectId}/grades", s.auth(s.handleCreateGrades))
	mux.Handle("GET /reports/student-averages", s.auth(s.handleStudentAverages))

	// Grades
	s.router.HandleFunc("POST /grades", protected(s.handleCreateOrUpdateGrade))
	s.router.HandleFunc("GET /grades/average", protected(s.handleStudentAverage))

	s.router.HandleFunc("GET /grades", protected(s.handleListGrades))
	s.router.HandleFunc("GET /grades/{id}", protected(s.handleGetGrade))
	//s.router.HandleFunc("POST /grades", protected(s.handleCreateGrade))
	s.router.HandleFunc("PUT /grades/{id}", protected(s.handleUpdateGrade))
	s.router.HandleFunc("DELETE /grades/{id}", protected(s.handleDeleteGrade))
}

// decodeJSON decodes a JSON request body into the given destination.
func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
