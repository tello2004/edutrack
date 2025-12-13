package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
	"lahuerta.tecmm.edu.mx/edutrack/database/sqlite"
)

func main() {
	// Parse command line flags.
	dbPath := flag.String("db", "edutrack.db", "Path to the SQLite database file")
	flag.Parse()

	logger := log.New(os.Stdout, "[seed] ", log.LstdFlags)
	errLogger := log.New(os.Stderr, "[seed] ERROR: ", log.LstdFlags)

	// Open the database.
	logger.Printf("Opening database: %s", *dbPath)
	db, err := sqlite.Open(*dbPath)
	if err != nil {
		errLogger.Fatalf("Failed to open database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize edutrack app and run migrations.
	app := edutrack.New(db)

	logger.Println("Running migrations...")
	if err := app.Migrate(); err != nil {
		errLogger.Fatalf("Failed to run migrations: %v", err)
	}

	// Create a tenant with a license.
	logger.Println("Creating tenant...")
	tenant, err := edutrack.NewTenant("Instituto Tecnológico de La Huerta", edutrack.LicenseTypePro, 365*24*time.Hour)
	if err != nil {
		errLogger.Fatalf("Failed to create tenant: %v", err)
	}

	if err := db.Create(tenant).Error; err != nil {
		errLogger.Fatalf("Failed to save tenant: %v", err)
	}
	logger.Printf("Created tenant: %s (ID: %s)", tenant.Name, tenant.ID)
	logger.Printf("License key: %s", tenant.GetLicenseKey())

	// Create careers.
	logger.Println("Creating careers...")
	careers := []edutrack.Career{
		{
			Name:        "Ingeniería en Sistemas Computacionales",
			Code:        "ISC-2024",
			Description: "Formación de profesionales en el desarrollo de software y sistemas computacionales.",
			Duration:    9,
			Active:      true,
			TenantID:    tenant.ID,
		},
		{
			Name:        "Licenciatura en Administración",
			Code:        "LAD-2024",
			Description: "Formación de profesionales en la gestión y administración de empresas.",
			Duration:    8,
			Active:      true,
			TenantID:    tenant.ID,
		},
		{
			Name:        "Ingeniería Industrial",
			Code:        "IIN-2024",
			Description: "Formación de profesionales en optimización de procesos industriales.",
			Duration:    9,
			Active:      true,
			TenantID:    tenant.ID,
		},
	}

	for i := range careers {
		if err := db.Create(&careers[i]).Error; err != nil {
			errLogger.Fatalf("Failed to create career %s: %v", careers[i].Code, err)
		}
		logger.Printf("Created career: %s (%s)", careers[i].Name, careers[i].Code)
	}

	// Create subjects.
	logger.Println("Creating subjects...")
	subjects := []edutrack.Subject{
		{Name: "Cálculo Diferencial", Code: "MAT101", Description: "Fundamentos de cálculo diferencial", Credits: 5, TenantID: tenant.ID},
		{Name: "Cálculo Integral", Code: "MAT102", Description: "Fundamentos de cálculo integral", Credits: 5, TenantID: tenant.ID},
		{Name: "Programación I", Code: "PRG101", Description: "Introducción a la programación", Credits: 5, TenantID: tenant.ID},
		{Name: "Programación II", Code: "PRG102", Description: "Programación orientada a objetos", Credits: 5, TenantID: tenant.ID},
		{Name: "Base de Datos", Code: "BDD101", Description: "Fundamentos de bases de datos", Credits: 4, TenantID: tenant.ID},
		{Name: "Redes de Computadoras", Code: "RED101", Description: "Fundamentos de redes", Credits: 4, TenantID: tenant.ID},
		{Name: "Contabilidad I", Code: "CON101", Description: "Fundamentos de contabilidad", Credits: 4, TenantID: tenant.ID},
		{Name: "Administración I", Code: "ADM101", Description: "Fundamentos de administración", Credits: 4, TenantID: tenant.ID},
		{Name: "Física I", Code: "FIS101", Description: "Mecánica clásica", Credits: 5, TenantID: tenant.ID},
		{Name: "Química", Code: "QUI101", Description: "Química general", Credits: 4, TenantID: tenant.ID},
	}

	for i := range subjects {
		if err := db.Create(&subjects[i]).Error; err != nil {
			errLogger.Fatalf("Failed to create subject %s: %v", subjects[i].Code, err)
		}
		logger.Printf("Created subject: %s (%s)", subjects[i].Name, subjects[i].Code)
	}

	// Associate subjects with careers.
	logger.Println("Associating subjects with careers...")
	// ISC: MAT101, MAT102, PRG101, PRG102, BDD101, RED101, FIS101
	db.Model(&careers[0]).Association("Subjects").Append(&subjects[0], &subjects[1], &subjects[2], &subjects[3], &subjects[4], &subjects[5], &subjects[8])
	// LAD: MAT101, CON101, ADM101
	db.Model(&careers[1]).Association("Subjects").Append(&subjects[0], &subjects[6], &subjects[7])
	// IIN: MAT101, MAT102, FIS101, QUI101, ADM101
	db.Model(&careers[2]).Association("Subjects").Append(&subjects[0], &subjects[1], &subjects[8], &subjects[9], &subjects[7])

	// Create accounts and teachers.
	logger.Println("Creating accounts...")

	// Secretary account.
	secretaryPassword, _ := edutrack.HashPassword("secretary123")
	secretary := edutrack.Account{
		Name:     "María García López",
		Email:    "maria.garcia@lahuerta.tecmm.edu.mx",
		Password: secretaryPassword,
		Role:     edutrack.RoleSecretary,
		Active:   true,
		TenantID: tenant.ID,
	}
	if err := db.Create(&secretary).Error; err != nil {
		errLogger.Fatalf("Failed to create secretary account: %v", err)
	}
	logger.Printf("Created secretary: %s (%s)", secretary.Name, secretary.Email)

	// Teacher accounts.
	teacherData := []struct {
		name  string
		email string
	}{
		{"Dr. Juan Pérez Hernández", "juan.perez@lahuerta.tecmm.edu.mx"},
		{"Mtra. Ana Rodríguez Martínez", "ana.rodriguez@lahuerta.tecmm.edu.mx"},
		{"Ing. Carlos López Sánchez", "carlos.lopez@lahuerta.tecmm.edu.mx"},
		{"Lic. Patricia Ramírez García", "patricia.ramirez@lahuerta.tecmm.edu.mx"},
		{"Dr. Roberto Flores Díaz", "roberto.flores@lahuerta.tecmm.edu.mx"},
	}

	var teachers []edutrack.Teacher
	for _, td := range teacherData {
		password, _ := edutrack.HashPassword("teacher123")
		account := edutrack.Account{
			Name:     td.name,
			Email:    td.email,
			Password: password,
			Role:     edutrack.RoleTeacher,
			Active:   true,
			TenantID: tenant.ID,
		}
		if err := db.Create(&account).Error; err != nil {
			errLogger.Fatalf("Failed to create teacher account %s: %v", td.email, err)
		}

		teacher := edutrack.Teacher{
			TenantID:  tenant.ID,
			AccountID: account.ID,
		}
		if err := db.Create(&teacher).Error; err != nil {
			errLogger.Fatalf("Failed to create teacher %s: %v", td.name, err)
		}
		teachers = append(teachers, teacher)
		logger.Printf("Created teacher: %s (%s)", td.name, td.email)
	}

	// Assign subjects to teachers.
	logger.Println("Assigning subjects to teachers...")
	// Teacher 0 (Juan): MAT101, MAT102
	db.Model(&teachers[0]).Association("Subjects").Append(&subjects[0], &subjects[1])
	db.Model(&subjects[0]).Update("TeacherID", teachers[0].ID)
	db.Model(&subjects[1]).Update("TeacherID", teachers[0].ID)

	// Teacher 1 (Ana): PRG101, PRG102
	db.Model(&teachers[1]).Association("Subjects").Append(&subjects[2], &subjects[3])
	db.Model(&subjects[2]).Update("TeacherID", teachers[1].ID)
	db.Model(&subjects[3]).Update("TeacherID", teachers[1].ID)

	// Teacher 2 (Carlos): BDD101, RED101
	db.Model(&teachers[2]).Association("Subjects").Append(&subjects[4], &subjects[5])
	db.Model(&subjects[4]).Update("TeacherID", teachers[2].ID)
	db.Model(&subjects[5]).Update("TeacherID", teachers[2].ID)

	// Teacher 3 (Patricia): CON101, ADM101
	db.Model(&teachers[3]).Association("Subjects").Append(&subjects[6], &subjects[7])
	db.Model(&subjects[6]).Update("TeacherID", teachers[3].ID)
	db.Model(&subjects[7]).Update("TeacherID", teachers[3].ID)

	// Teacher 4 (Roberto): FIS101, QUI101
	db.Model(&teachers[4]).Association("Subjects").Append(&subjects[8], &subjects[9])
	db.Model(&subjects[8]).Update("TeacherID", teachers[4].ID)
	db.Model(&subjects[9]).Update("TeacherID", teachers[4].ID)

	// Create students.
	logger.Println("Creating students...")
	studentData := []struct {
		studentID string
		name      string
		email     string
		careerIdx int
	}{
		{"2024ISC001", "Luis Fernando Torres Vega", "luis.torres@estudiante.tecmm.edu.mx", 0},
		{"2024ISC002", "Andrea Sofía Mendoza Ruiz", "andrea.mendoza@estudiante.tecmm.edu.mx", 0},
		{"2024ISC003", "Diego Alejandro Navarro Cruz", "diego.navarro@estudiante.tecmm.edu.mx", 0},
		{"2024ISC004", "Valeria Campos Herrera", "valeria.campos@estudiante.tecmm.edu.mx", 0},
		{"2024ISC005", "Miguel Ángel Ortiz Reyes", "miguel.ortiz@estudiante.tecmm.edu.mx", 0},
		{"2024LAD001", "Gabriela Ríos Morales", "gabriela.rios@estudiante.tecmm.edu.mx", 1},
		{"2024LAD002", "Fernando Castillo León", "fernando.castillo@estudiante.tecmm.edu.mx", 1},
		{"2024LAD003", "Mariana Delgado Vargas", "mariana.delgado@estudiante.tecmm.edu.mx", 1},
		{"2024IIN001", "Javier Eduardo Soto Jiménez", "javier.soto@estudiante.tecmm.edu.mx", 2},
		{"2024IIN002", "Paulina Aguilar Domínguez", "paulina.aguilar@estudiante.tecmm.edu.mx", 2},
	}

	for _, sd := range studentData {
		password, _ := edutrack.HashPassword("student123")
		account := edutrack.Account{
			Name:     sd.name,
			Email:    sd.email,
			Password: password,
			Role:     edutrack.RoleTeacher, // Students use teacher role for now (could add RoleStudent later).
			Active:   true,
			TenantID: tenant.ID,
		}
		if err := db.Create(&account).Error; err != nil {
			errLogger.Fatalf("Failed to create student account %s: %v", sd.email, err)
		}

		student := edutrack.Student{
			StudentID: sd.studentID,
			TenantID:  tenant.ID,
			AccountID: account.ID,
			CareerID:  careers[sd.careerIdx].ID,
		}
		if err := db.Create(&student).Error; err != nil {
			errLogger.Fatalf("Failed to create student %s: %v", sd.studentID, err)
		}
		logger.Printf("Created student: %s (%s) - %s", sd.name, sd.studentID, careers[sd.careerIdx].Code)
	}

	// Print summary.
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("         SEED COMPLETED SUCCESSFULLY")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Tenant: %s\n", tenant.Name)
	fmt.Printf("Tenant ID: %s\n", tenant.ID)
	fmt.Printf("License Key: %s\n", tenant.GetLicenseKey())
	fmt.Println()
	fmt.Println("Test Accounts:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Secretary: %s\n", secretary.Email)
	fmt.Println("Password: secretary123")
	fmt.Println()
	fmt.Println("Teachers: (Password: teacher123)")
	for _, td := range teacherData {
		fmt.Printf("  - %s\n", td.email)
	}
	fmt.Println()
	fmt.Println("Students: (Password: student123)")
	for _, sd := range studentData {
		fmt.Printf("  - %s (%s)\n", sd.email, sd.studentID)
	}
	fmt.Println("========================================")
}
