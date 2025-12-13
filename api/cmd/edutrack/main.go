package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	edutrack "lahuerta.tecmm.edu.mx/edutrack"
	"lahuerta.tecmm.edu.mx/edutrack/database/sqlite"
)

const usage = `edutrack - EduTrack Administration CLI

Usage:
  edutrack <command> [options]

Commands:
  tenant      Manage tenants (institutions)
  license     Manage licenses
  account     Manage accounts
  stats       Show tenant statistics

Use "edutrack <command> -h" for more information about a command.
`

const tenantUsage = `Usage: edutrack tenant <subcommand> [options]

Subcommands:
  add       Add a new tenant
  list      List all tenants
  get       Get tenant details

Examples:
  edutrack tenant add "La Huerta"
  edutrack tenant add -type=pro -days=365 "Instituto Tecnológico"
  edutrack tenant list
  edutrack tenant get -id=abc12345
`

const licenseUsage = `Usage: edutrack license <subcommand> [options]

Subcommands:
  generate     Generate a new license for a tenant
  regenerate   Regenerate (renew) a license key
  info         Show license information

Examples:
  edutrack license generate -for=abc12345
  edutrack license regenerate -for=abc12345 -extend=365
  edutrack license info -for=abc12345
`

const accountUsage = `Usage: edutrack account <subcommand> [options]

Subcommands:
  add       Add a new account to a tenant
  list      List accounts for a tenant

Examples:
  edutrack account add -tenant=abc12345 -role=secretary -email=admin@example.com -name="Admin" -password=secret
  edutrack account list -tenant=abc12345
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	// Initialize database connection.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "edutrack.db"
	}

	db, err := sqlite.Open(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	app := edutrack.New(db)

	// Run migrations to ensure tables exist.
	if err := app.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "tenant":
		handleTenant(app, os.Args[2:])
	case "license":
		handleLicense(app, os.Args[2:])
	case "account":
		handleAccount(app, os.Args[2:])
	case "stats":
		handleStats(app, os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func handleTenant(app *edutrack.App, args []string) {
	if len(args) < 1 {
		fmt.Print(tenantUsage)
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		tenantAdd(app, args[1:])
	case "list":
		tenantList(app)
	case "get":
		tenantGet(app, args[1:])
	case "-h", "--help", "help":
		fmt.Print(tenantUsage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown tenant subcommand: %s\n", args[0])
		fmt.Print(tenantUsage)
		os.Exit(1)
	}
}

func tenantAdd(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("tenant add", flag.ExitOnError)
	licenseType := fs.String("type", "trial", "License type: trial, basic, pro, enterprise")
	days := fs.Int("days", 30, "License duration in days")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: tenant name is required")
		fmt.Println("Usage: edutrack tenant add [options] <name>")
		os.Exit(1)
	}

	name := fs.Arg(0)
	lt := edutrack.LicenseType(*licenseType)

	tenant, err := app.CreateTenant(name, lt, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Tenant created successfully!")
	fmt.Println()
	printTenant(tenant)
}

func tenantList(app *edutrack.App) {
	tenants, err := app.ListTenants()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(tenants) == 0 {
		fmt.Println("No tenants found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tLICENSE TYPE\tLICENSE KEY\tEXPIRES\tSTATUS")
	fmt.Fprintln(w, "--\t----\t------------\t-----------\t-------\t------")

	for _, t := range tenants {
		status := "Active"
		if !t.License.IsValid() {
			if t.License.IsExpired() {
				status = "Expired"
			} else {
				status = "Inactive"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			t.Name,
			t.License.Type,
			t.License.Key,
			t.License.ExpiryAt.Format("2006-01-02"),
			status,
		)
	}
	w.Flush()
}

func tenantGet(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("tenant get", flag.ExitOnError)
	id := fs.String("id", "", "Tenant ID")
	fs.Parse(args)

	if *id == "" {
		fmt.Fprintln(os.Stderr, "Error: tenant ID is required (-id)")
		os.Exit(1)
	}

	tenant, err := app.FindTenantByID(*id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: tenant not found: %v\n", err)
		os.Exit(1)
	}

	printTenant(tenant)
}

func printTenant(t *edutrack.Tenant) {
	fmt.Printf("Tenant ID:       %s\n", t.ID)
	fmt.Printf("Name:            %s\n", t.Name)
	fmt.Printf("License Key:     %s\n", t.License.Key)
	fmt.Printf("License Type:    %s\n", t.License.Type)
	fmt.Printf("Expires:         %s\n", t.License.ExpiryAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Days Remaining:  %d\n", t.License.DaysUntilExpiry())
	fmt.Printf("Max Users:       %d\n", t.License.MaxUsers)
	fmt.Printf("Max Students:    %d\n", t.License.MaxStudents)
	fmt.Printf("Max Courses:     %d\n", t.License.MaxCourses)

	status := "Active"
	if !t.License.IsValid() {
		if t.License.IsExpired() {
			status = "Expired"
		} else {
			status = "Inactive"
		}
	}
	fmt.Printf("Status:          %s\n", status)
}

func handleLicense(app *edutrack.App, args []string) {
	if len(args) < 1 {
		fmt.Print(licenseUsage)
		os.Exit(1)
	}

	switch args[0] {
	case "generate":
		licenseGenerate(app, args[1:])
	case "regenerate":
		licenseRegenerate(app, args[1:])
	case "info":
		licenseInfo(app, args[1:])
	case "-h", "--help", "help":
		fmt.Print(licenseUsage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown license subcommand: %s\n", args[0])
		fmt.Print(licenseUsage)
		os.Exit(1)
	}
}

func licenseGenerate(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("license generate", flag.ExitOnError)
	tenantID := fs.String("for", "", "Tenant ID to generate license for")
	fs.Parse(args)

	if *tenantID == "" {
		fmt.Fprintln(os.Stderr, "Error: tenant ID is required (-for)")
		os.Exit(1)
	}

	// This regenerates with 0 days extension (just new key).
	license, err := app.RegenerateLicense(*tenantID, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("License key generated successfully!")
	fmt.Printf("New Key: %s\n", license.Key)
}

func licenseRegenerate(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("license regenerate", flag.ExitOnError)
	tenantID := fs.String("for", "", "Tenant ID to regenerate license for")
	extend := fs.Int("extend", 365, "Days to extend the license")
	fs.Parse(args)

	if *tenantID == "" {
		fmt.Fprintln(os.Stderr, "Error: tenant ID is required (-for)")
		os.Exit(1)
	}

	license, err := app.RegenerateLicense(*tenantID, *extend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("License regenerated successfully!")
	fmt.Printf("New Key:     %s\n", license.Key)
	fmt.Printf("Expires:     %s\n", license.ExpiryAt.Format("2006-01-02"))
	fmt.Printf("Days Left:   %d\n", license.DaysUntilExpiry())
}

func licenseInfo(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("license info", flag.ExitOnError)
	tenantID := fs.String("for", "", "Tenant ID to get license info for")
	fs.Parse(args)

	if *tenantID == "" {
		fmt.Fprintln(os.Stderr, "Error: tenant ID is required (-for)")
		os.Exit(1)
	}

	tenant, err := app.FindTenantByID(*tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: tenant not found: %v\n", err)
		os.Exit(1)
	}

	l := &tenant.License
	fmt.Printf("License Key:     %s\n", l.Key)
	fmt.Printf("Type:            %s\n", l.Type)
	fmt.Printf("Expires:         %s\n", l.ExpiryAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Days Remaining:  %d\n", l.DaysUntilExpiry())
	fmt.Printf("Active:          %t\n", l.Active)
	fmt.Printf("Valid:           %t\n", l.IsValid())
	fmt.Printf("Max Users:       %d\n", l.MaxUsers)
	fmt.Printf("Max Students:    %d\n", l.MaxStudents)
	fmt.Printf("Max Courses:     %d\n", l.MaxCourses)
}

func handleAccount(app *edutrack.App, args []string) {
	if len(args) < 1 {
		fmt.Print(accountUsage)
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		accountAdd(app, args[1:])
	case "list":
		accountList(app, args[1:])
	case "-h", "--help", "help":
		fmt.Print(accountUsage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown account subcommand: %s\n", args[0])
		fmt.Print(accountUsage)
		os.Exit(1)
	}
}

func accountAdd(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("account add", flag.ExitOnError)
	tenantID := fs.String("tenant", "", "Tenant ID")
	role := fs.String("role", "teacher", "Role: secretary or teacher")
	email := fs.String("email", "", "Email address")
	name := fs.String("name", "", "Full name")
	password := fs.String("password", "", "Password")
	fs.Parse(args)

	if *tenantID == "" || *email == "" || *name == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: all fields are required (-tenant, -email, -name, -password)")
		os.Exit(1)
	}

	r := edutrack.Role(*role)
	if r != edutrack.RoleSecretary && r != edutrack.RoleTeacher {
		fmt.Fprintln(os.Stderr, "Error: role must be 'secretary' or 'teacher'")
		os.Exit(1)
	}

	account, err := app.CreateAccount(*tenantID, *name, *email, *password, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Account created successfully!")
	fmt.Printf("ID:       %d\n", account.ID)
	fmt.Printf("Name:     %s\n", account.Name)
	fmt.Printf("Email:    %s\n", account.Email)
	fmt.Printf("Role:     %s\n", account.Role)
	fmt.Printf("Tenant:   %s\n", account.TenantID)
}

func accountList(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("account list", flag.ExitOnError)
	tenantID := fs.String("tenant", "", "Tenant ID")
	fs.Parse(args)

	if *tenantID == "" {
		fmt.Fprintln(os.Stderr, "Error: tenant ID is required (-tenant)")
		os.Exit(1)
	}

	var accounts []edutrack.Account
	if err := app.DB.Where("tenant_id = ?", *tenantID).Find(&accounts).Error; err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts found for this tenant.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tROLE\tACTIVE")
	fmt.Fprintln(w, "--\t----\t-----\t----\t------")

	for _, a := range accounts {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%t\n",
			a.ID,
			a.Name,
			a.Email,
			a.Role,
			a.Active,
		)
	}
	w.Flush()
}

func handleStats(app *edutrack.App, args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	tenantID := fs.String("tenant", "", "Tenant ID (optional, shows all if not specified)")
	fs.Parse(args)

	if *tenantID != "" {
		stats, err := app.GetTenantStats(*tenantID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printStats(stats)
		return
	}

	// Show stats for all tenants.
	tenants, err := app.ListTenants()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(tenants) == 0 {
		fmt.Println("No tenants found.")
		return
	}

	for i, t := range tenants {
		stats, err := app.GetTenantStats(t.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting stats for %s: %v\n", t.ID, err)
			continue
		}

		if i > 0 {
			fmt.Println()
			fmt.Println("---")
			fmt.Println()
		}

		printStats(stats)
	}
}

func printStats(s *edutrack.Stats) {
	fmt.Printf("Tenant:          %s (%s)\n", s.TenantName, s.TenantID)
	fmt.Printf("License Type:    %s\n", s.LicenseType)
	fmt.Printf("Days Remaining:  %d\n", s.DaysRemaining)
	fmt.Printf("Accounts:        %d\n", s.AccountCount)
	fmt.Printf("Students:        %d\n", s.StudentCount)
	fmt.Printf("Teachers:        %d\n", s.TeacherCount)
	fmt.Printf("Careers:         %d\n", s.CareerCount)
	fmt.Printf("Subjects:        %d\n", s.SubjectCount)
}
