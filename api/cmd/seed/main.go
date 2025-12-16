package main

import (
	"fmt"
	"time"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
)

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
		app.errLogger.Fatal("No database configured. Build with -tags sqlite or -tags postgres.")
	}

	// Initialize edutrack app and run migrations.
	edutrackApp := edutrack.New(app.db)

	app.logger.Println("Running migrations...")
	if err := edutrackApp.Migrate(); err != nil {
		app.errLogger.Fatalf("Failed to run migrations: %v", err)
	}

	// Create a tenant with a license.
	app.logger.Println("Creating tenant...")
	tenant, err := edutrack.NewTenant("Instituto Tecnológico de La Huerta", edutrack.LicenseTypePro, 365*24*time.Hour)
	if err != nil {
		app.errLogger.Fatalf("Failed to create tenant: %v", err)
	}

	if err := app.db.Create(tenant).Error; err != nil {
		app.errLogger.Fatalf("Failed to save tenant: %v", err)
	}
	app.logger.Printf("Created tenant: %s (ID: %s)", tenant.Name, tenant.ID)
	app.logger.Printf("License key: %s", tenant.GetLicenseKey())

	// Create careers.
	app.logger.Println("Creating careers...")
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
		if err := app.db.Create(&careers[i]).Error; err != nil {
			app.errLogger.Fatalf("Failed to create career %s: %v", careers[i].Code, err)
		}
		app.logger.Printf("Created career: %s (%s)", careers[i].Name, careers[i].Code)
	}

	// Create subjects.
	app.logger.Println("Creating subjects...")
	subjects := []edutrack.Subject{
		{Name: "Cálculo Diferencial", Code: "MAT101", Description: "Fundamentos de cálculo diferencial", Credits: 5, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 1},
		{Name: "Cálculo Integral", Code: "MAT102", Description: "Fundamentos de cálculo integral", Credits: 5, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 2},
		{Name: "Programación I", Code: "PRG101", Description: "Introducción a la programación", Credits: 5, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 1},
		{Name: "Programación II", Code: "PRG102", Description: "Programación orientada a objetos", Credits: 5, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 2},
		{Name: "Base de Datos", Code: "BDD101", Description: "Fundamentos de bases de datos", Credits: 4, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 3},
		{Name: "Redes de Computadoras", Code: "RED101", Description: "Fundamentos de redes", Credits: 4, TenantID: tenant.ID, CareerID: careers[0].ID, Semester: 4},
		{Name: "Contabilidad I", Code: "CON101", Description: "Fundamentos de contabilidad", Credits: 4, TenantID: tenant.ID, CareerID: careers[1].ID, Semester: 1},
		{Name: "Administración I", Code: "ADM101", Description: "Fundamentos de administración", Credits: 4, TenantID: tenant.ID, CareerID: careers[1].ID, Semester: 1},
		{Name: "Física I", Code: "FIS101", Description: "Mecánica clásica", Credits: 5, TenantID: tenant.ID, CareerID: careers[2].ID, Semester: 1},
		{Name: "Química", Code: "QUI101", Description: "Química general", Credits: 4, TenantID: tenant.ID, CareerID: careers[2].ID, Semester: 1},
	}

	for i := range subjects {
		if err := app.db.Create(&subjects[i]).Error; err != nil {
			app.errLogger.Fatalf("Failed to create subject %s: %v", subjects[i].Code, err)
		}
		app.logger.Printf("Created subject: %s (%s)", subjects[i].Name, subjects[i].Code)
	}

	// Create accounts and teachers.
	app.logger.Println("Creating accounts...")

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
	if err := app.db.Create(&secretary).Error; err != nil {
		app.errLogger.Fatalf("Failed to create secretary account: %v", err)
	}
	app.logger.Printf("Created secretary: %s (%s)", secretary.Name, secretary.Email)

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
		if err := app.db.Create(&account).Error; err != nil {
			app.errLogger.Fatalf("Failed to create teacher account %s: %v", td.email, err)
		}

		teacher := edutrack.Teacher{
			TenantID:  tenant.ID,
			AccountID: account.ID,
		}
		if err := app.db.Create(&teacher).Error; err != nil {
			app.errLogger.Fatalf("Failed to create teacher %s: %v", td.name, err)
		}
		teachers = append(teachers, teacher)
		app.logger.Printf("Created teacher: %s (%s)", td.name, td.email)
	}

	// Assign subjects to teachers.
	app.logger.Println("Assigning subjects to teachers...")
	app.db.Model(&subjects[0]).Update("TeacherID", teachers[0].ID)
	app.db.Model(&subjects[1]).Update("TeacherID", teachers[0].ID)
	app.db.Model(&subjects[2]).Update("TeacherID", teachers[1].ID)
	app.db.Model(&subjects[3]).Update("TeacherID", teachers[1].ID)
	app.db.Model(&subjects[4]).Update("TeacherID", teachers[2].ID)
	app.db.Model(&subjects[5]).Update("TeacherID", teachers[2].ID)
	app.db.Model(&subjects[6]).Update("TeacherID", teachers[3].ID)
	app.db.Model(&subjects[7]).Update("TeacherID", teachers[3].ID)
	app.db.Model(&subjects[8]).Update("TeacherID", teachers[4].ID)
	app.db.Model(&subjects[9]).Update("TeacherID", teachers[4].ID)

	// Create students.
	app.logger.Println("Creating students...")
	studentData := []struct {
		studentID string
		name      string
		email     string
		careerIdx int
		semester  int
	}{
		{"2024ISC001", "Luis Fernando Torres Vega", "luis.torres@estudiante.tecmm.edu.mx", 0, 1},
		{"2024ISC002", "Andrea Sofía Mendoza Ruiz", "andrea.mendoza@estudiante.tecmm.edu.mx", 0, 1},
		{"2024ISC003", "Diego Alejandro Navarro Cruz", "diego.navarro@estudiante.tecmm.edu.mx", 0, 2},
		{"2024ISC004", "Valeria Campos Herrera", "valeria.campos@estudiante.tecmm.edu.mx", 0, 2},
		{"2024ISC005", "Miguel Ángel Ortiz Reyes", "miguel.ortiz@estudiante.tecmm.edu.mx", 0, 3},
		{"2024LAD001", "Gabriela Ríos Morales", "gabriela.rios@estudiante.tecmm.edu.mx", 1, 1},
		{"2024LAD002", "Fernando Castillo León", "fernando.castillo@estudiante.tecmm.edu.mx", 1, 2},
		{"2024LAD003", "Mariana Delgado Vargas", "mariana.delgado@estudiante.tecmm.edu.mx", 1, 3},
		{"2024IIN001", "Javier Eduardo Soto Jiménez", "javier.soto@estudiante.tecmm.edu.mx", 2, 1},
		{"2024IIN002", "Paulina Aguilar Domínguez", "paulina.aguilar@estudiante.tecmm.edu.mx", 2, 2},
	}

	var students []edutrack.Student
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
		if err := app.db.Create(&account).Error; err != nil {
			app.errLogger.Fatalf("Failed to create student account %s: %v", sd.email, err)
		}

		student := edutrack.Student{
			StudentID: sd.studentID,
			TenantID:  tenant.ID,
			AccountID: account.ID,
			CareerID:  careers[sd.careerIdx].ID,
			Semester:  sd.semester,
		}
		if err := app.db.Create(&student).Error; err != nil {
			app.errLogger.Fatalf("Failed to create student %s: %v", sd.studentID, err)
		}
		app.logger.Printf("Created student: %s (%s) - %s", sd.name, sd.studentID, careers[sd.careerIdx].Code)
		students = append(students, student)
	}

	// Enroll students in subjects.
	app.logger.Println("Enrolling students in subjects...")
	for _, s := range students {
		var subjectsToEnroll []edutrack.Subject
		app.db.Where("career_id = ? AND semester = ?", s.CareerID, s.Semester).Find(&subjectsToEnroll)
		if len(subjectsToEnroll) > 0 {
			app.db.Model(&s).Association("Subjects").Append(subjectsToEnroll)
			app.logger.Printf("Enrolled student %s in %d subjects for semester %d", s.StudentID, len(subjectsToEnroll), s.Semester)
		}
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
