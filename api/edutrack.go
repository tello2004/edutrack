package edutrack

import (
	"fmt"

	"gorm.io/gorm"
)

// Version is the current version of the application.
const Version = "0.1.0"

// App represents the main application instance.
type App struct {
	DB *gorm.DB
}

// New creates a new application instance with the given database connection.
func New(db *gorm.DB) *App {
	return &App{
		DB: db,
	}
}

// Migrate runs all database migrations.
func (a *App) Migrate() error {
	return Migrate(a.DB)
}

// Migrate runs all database migrations on the given database connection.
func Migrate(db *gorm.DB) error {
	models := []any{
		&License{},
		&Tenant{},
		&Account{},
		&Career{},
		&Subject{},
		&Topic{},
		&Teacher{},
		&Student{},
		&Attendance{},
		&Grade{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return nil
}

// CreateTenant creates a new tenant with the specified license type.
func (a *App) CreateTenant(name string, licenseType LicenseType, licenseDuration int) (*Tenant, error) {
	// Convert days to duration.
	duration := DaysToYears(licenseDuration)

	tenant, err := NewTenant(name, licenseType, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	if err := a.DB.Create(tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to save tenant: %w", err)
	}

	return tenant, nil
}

// FindTenantByID finds a tenant by its ID.
func (a *App) FindTenantByID(id string) (*Tenant, error) {
	var tenant Tenant
	if err := a.DB.Preload("License").First(&tenant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// FindTenantByLicenseKey finds a tenant by its license key.
func (a *App) FindTenantByLicenseKey(key string) (*Tenant, error) {
	var license License
	if err := a.DB.Where("key = ?", key).First(&license).Error; err != nil {
		return nil, err
	}

	var tenant Tenant
	if err := a.DB.Preload("License").Where("license_id = ?", license.ID).First(&tenant).Error; err != nil {
		return nil, err
	}

	return &tenant, nil
}

// RegenerateLicense regenerates the license key for a tenant.
func (a *App) RegenerateLicense(tenantID string, extendDays int) (*License, error) {
	tenant, err := a.FindTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	duration := DaysToYears(extendDays)
	if err := tenant.License.Regenerate(duration); err != nil {
		return nil, fmt.Errorf("failed to regenerate license: %w", err)
	}

	if err := a.DB.Save(&tenant.License).Error; err != nil {
		return nil, fmt.Errorf("failed to save license: %w", err)
	}

	return &tenant.License, nil
}

// CreateAccount creates a new account for a tenant.
func (a *App) CreateAccount(tenantID, name, email, password string, role Role) (*Account, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		Active:   true,
		TenantID: tenantID,
	}

	if err := a.DB.Create(account).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// ListTenants returns all tenants in the system.
func (a *App) ListTenants() ([]Tenant, error) {
	var tenants []Tenant
	if err := a.DB.Preload("License").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

// Stats returns basic statistics about a tenant.
type Stats struct {
	TenantID      string
	TenantName    string
	LicenseType   LicenseType
	DaysRemaining int
	AccountCount  int64
	StudentCount  int64
	TeacherCount  int64
	CareerCount   int64
	SubjectCount  int64
}

// GetTenantStats returns statistics for a specific tenant.
func (a *App) GetTenantStats(tenantID string) (*Stats, error) {
	tenant, err := a.FindTenantByID(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		TenantID:      tenant.ID,
		TenantName:    tenant.Name,
		LicenseType:   tenant.License.Type,
		DaysRemaining: tenant.License.DaysUntilExpiry(),
	}

	a.DB.Model(&Account{}).Where("tenant_id = ?", tenantID).Count(&stats.AccountCount)
	a.DB.Model(&Student{}).Where("tenant_id = ?", tenantID).Count(&stats.StudentCount)
	a.DB.Model(&Teacher{}).Where("tenant_id = ?", tenantID).Count(&stats.TeacherCount)
	a.DB.Model(&Career{}).Where("tenant_id = ?", tenantID).Count(&stats.CareerCount)
	a.DB.Model(&Subject{}).Where("tenant_id = ?", tenantID).Count(&stats.SubjectCount)

	return stats, nil
}
