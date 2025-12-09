package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
	"lahuerta.tecmm.edu.mx/edutrack/http"

	"gorm.io/gorm"
)

var app *application

type application struct {
	db        *gorm.DB
	logger    *log.Logger
	errLogger *log.Logger
	server    *http.Server
	edutrack  *edutrack.App
}

func init() {
	app = &application{
		logger:    log.New(os.Stdout, "[edutrackd] ", log.LstdFlags),
		errLogger: log.New(os.Stderr, "[edutrackd] ERROR: ", log.LstdFlags),
	}
}

func main() {
	// Ensure database is closed on exit.
	defer func() {
		if app.db != nil {
			db, _ := app.db.DB()
			_ = db.Close()
		}
	}()

	// Wait for database initialization from build tags.
	if app.db == nil {
		app.errLogger.Fatal("No database configured. Build with -tags sqlite or another database driver.")
	}

	// Initialize the edutrack application.
	app.edutrack = edutrack.New(app.db)

	// Run migrations.
	app.logger.Println("Running database migrations...")
	if err := app.edutrack.Migrate(); err != nil {
		app.errLogger.Fatalf("Failed to run migrations: %v", err)
	}
	app.logger.Println("Migrations completed successfully.")

	// Get configuration from environment.
	addr := os.Getenv("EDUTRACK_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	jwtSecret := os.Getenv("EDUTRACK_JWT_SECRET")
	if jwtSecret == "" {
		app.errLogger.Println("WARNING: EDUTRACK_JWT_SECRET not set, using insecure default.")
		jwtSecret = "edutrack-dev-secret-change-in-production"
	}

	// Create and configure the HTTP server.
	app.server = http.NewServer(addr, app.db, []byte(jwtSecret))

	// Channel to listen for shutdown signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine.
	go func() {
		app.logger.Printf("Starting server on %s", addr)
		app.logger.Printf("EduTrack version %s", edutrack.Version)
		if err := app.server.Start(); err != nil {
			app.errLogger.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal.
	<-quit
	app.logger.Println("Shutting down server...")

	// Create a deadline for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown.
	if err := app.server.Shutdown(ctx); err != nil {
		app.errLogger.Printf("Server forced to shutdown: %v", err)
	}

	app.logger.Println("Server stopped.")
}
