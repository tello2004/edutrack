package edutrack

import (
	"testing"
	"time"
)

func TestGenerateTenantID(t *testing.T) {
	id1, err := GenerateTenantID()
	if err != nil {
		t.Fatalf("GenerateTenantID() error = %v", err)
	}

	// Check length: 8 hex characters
	if len(id1) != 8 {
		t.Errorf("GenerateTenantID() length = %d, want 8", len(id1))
	}

	// Check that it's valid hex
	for _, c := range id1 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("GenerateTenantID() contains non-hex character: %c", c)
		}
	}

	// Generate another ID and ensure they're different
	id2, err := GenerateTenantID()
	if err != nil {
		t.Fatalf("GenerateTenantID() second call error = %v", err)
	}

	if id1 == id2 {
		t.Errorf("GenerateTenantID() generated duplicate IDs: %s", id1)
	}
}

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name            string
		tenantName      string
		licenseType     LicenseType
		licenseDuration time.Duration
	}{
		{
			name:            "trial tenant",
			tenantName:      "Test Institution",
			licenseType:     LicenseTypeTrial,
			licenseDuration: 30 * 24 * time.Hour,
		},
		{
			name:            "basic tenant",
			tenantName:      "Basic School",
			licenseType:     LicenseTypeBasic,
			licenseDuration: 365 * 24 * time.Hour,
		},
		{
			name:            "pro tenant",
			tenantName:      "Professional Academy",
			licenseType:     LicenseTypePro,
			licenseDuration: 365 * 24 * time.Hour,
		},
		{
			name:            "enterprise tenant",
			tenantName:      "Enterprise University",
			licenseType:     LicenseTypeEnterprise,
			licenseDuration: 3 * 365 * 24 * time.Hour,
		},
		{
			name:            "tenant with special characters",
			tenantName:      "Instituto Técnico 'La Huerta' #1",
			licenseType:     LicenseTypeBasic,
			licenseDuration: 365 * 24 * time.Hour,
		},
		{
			name:            "tenant with unicode name",
			tenantName:      "学校日本語",
			licenseType:     LicenseTypeTrial,
			licenseDuration: 30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := NewTenant(tt.tenantName, tt.licenseType, tt.licenseDuration)
			if err != nil {
				t.Fatalf("NewTenant() error = %v", err)
			}

			if tenant.ID == "" {
				t.Error("NewTenant() ID is empty")
			}

			if len(tenant.ID) != 8 {
				t.Errorf("NewTenant() ID length = %d, want 8", len(tenant.ID))
			}

			if tenant.Name != tt.tenantName {
				t.Errorf("NewTenant() Name = %v, want %v", tenant.Name, tt.tenantName)
			}

			if tenant.License.Type != tt.licenseType {
				t.Errorf("NewTenant() License.Type = %v, want %v", tenant.License.Type, tt.licenseType)
			}

			if tenant.License.Key == "" {
				t.Error("NewTenant() License.Key is empty")
			}

			if !tenant.License.Active {
				t.Error("NewTenant() License.Active = false, want true")
			}

			// Check expiry is approximately correct
			expectedExpiry := time.Now().Add(tt.licenseDuration)
			if tenant.License.ExpiryAt.Before(expectedExpiry.Add(-time.Minute)) ||
				tenant.License.ExpiryAt.After(expectedExpiry.Add(time.Minute)) {
				t.Errorf("NewTenant() License.ExpiryAt = %v, want approximately %v",
					tenant.License.ExpiryAt, expectedExpiry)
			}
		})
	}
}

func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		name         string
		licenseValid bool
		want         bool
	}{
		{
			name:         "active tenant with valid license",
			licenseValid: true,
			want:         true,
		},
		{
			name:         "inactive tenant with invalid license",
			licenseValid: false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				ID:   "test1234",
				Name: "Test Institution",
				License: License{
					Active: tt.licenseValid,
					ExpiryAt: func() time.Time {
						if tt.licenseValid {
							return time.Now().Add(24 * time.Hour)
						}
						return time.Now().Add(-24 * time.Hour)
					}(),
				},
			}

			if got := tenant.IsActive(); got != tt.want {
				t.Errorf("Tenant.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenant_IsActive_ExpiredLicense(t *testing.T) {
	tenant := &Tenant{
		ID:   "test1234",
		Name: "Test Institution",
		License: License{
			Active:   true, // License is active but expired
			ExpiryAt: time.Now().Add(-24 * time.Hour),
		},
	}

	if tenant.IsActive() {
		t.Error("Tenant.IsActive() = true for expired license, want false")
	}
}

func TestTenant_IsActive_InactiveLicense(t *testing.T) {
	tenant := &Tenant{
		ID:   "test1234",
		Name: "Test Institution",
		License: License{
			Active:   false, // License is inactive but not expired
			ExpiryAt: time.Now().Add(24 * time.Hour),
		},
	}

	if tenant.IsActive() {
		t.Error("Tenant.IsActive() = true for inactive license, want false")
	}
}

func TestTenant_GetLicenseKey(t *testing.T) {
	expectedKey := "1234-5678-abcd-efgh"
	tenant := &Tenant{
		ID:   "test1234",
		Name: "Test Institution",
		License: License{
			Key: expectedKey,
		},
	}

	if got := tenant.GetLicenseKey(); got != expectedKey {
		t.Errorf("Tenant.GetLicenseKey() = %v, want %v", got, expectedKey)
	}
}

func TestTenant_GetLicenseKey_Empty(t *testing.T) {
	tenant := &Tenant{
		ID:      "test1234",
		Name:    "Test Institution",
		License: License{},
	}

	if got := tenant.GetLicenseKey(); got != "" {
		t.Errorf("Tenant.GetLicenseKey() = %v, want empty string", got)
	}
}

func TestMultipleTenantCreation(t *testing.T) {
	// Test creating multiple tenants to ensure unique IDs and license keys
	tenants := make([]*Tenant, 10)
	ids := make(map[string]bool)
	keys := make(map[string]bool)

	for i := 0; i < 10; i++ {
		tenant, err := NewTenant("Test Institution", LicenseTypeTrial, 30*24*time.Hour)
		if err != nil {
			t.Fatalf("NewTenant() iteration %d error = %v", i, err)
		}
		tenants[i] = tenant

		if ids[tenant.ID] {
			t.Errorf("Duplicate tenant ID found: %s", tenant.ID)
		}
		ids[tenant.ID] = true

		if keys[tenant.License.Key] {
			t.Errorf("Duplicate license key found: %s", tenant.License.Key)
		}
		keys[tenant.License.Key] = true
	}
}

func BenchmarkGenerateTenantID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateTenantID()
	}
}

func BenchmarkNewTenant(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewTenant("Benchmark Institution", LicenseTypeBasic, 365*24*time.Hour)
	}
}
