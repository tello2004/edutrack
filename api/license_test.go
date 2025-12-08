package edutrack

import (
	"testing"
	"time"
)

func TestGenerateLicenseKey(t *testing.T) {
	key1, err := GenerateLicenseKey()
	if err != nil {
		t.Fatalf("GenerateLicenseKey() error = %v", err)
	}

	// Check format: XXXX-XXXX-XXXX-XXXX
	if len(key1) != 19 {
		t.Errorf("GenerateLicenseKey() key length = %d, want 19", len(key1))
	}

	// Check dashes are in correct positions
	if key1[4] != '-' || key1[9] != '-' || key1[14] != '-' {
		t.Errorf("GenerateLicenseKey() key format invalid: %s", key1)
	}

	// Generate another key and ensure they're different
	key2, err := GenerateLicenseKey()
	if err != nil {
		t.Fatalf("GenerateLicenseKey() second call error = %v", err)
	}

	if key1 == key2 {
		t.Errorf("GenerateLicenseKey() generated duplicate keys: %s", key1)
	}
}

func TestNewLicense(t *testing.T) {
	tests := []struct {
		name        string
		licenseType LicenseType
		duration    time.Duration
		wantUsers   int
		wantStuds   int
		wantCourses int
	}{
		{
			name:        "trial license",
			licenseType: LicenseTypeTrial,
			duration:    30 * 24 * time.Hour,
			wantUsers:   3,
			wantStuds:   25,
			wantCourses: 5,
		},
		{
			name:        "basic license",
			licenseType: LicenseTypeBasic,
			duration:    365 * 24 * time.Hour,
			wantUsers:   10,
			wantStuds:   100,
			wantCourses: 20,
		},
		{
			name:        "pro license",
			licenseType: LicenseTypePro,
			duration:    365 * 24 * time.Hour,
			wantUsers:   50,
			wantStuds:   500,
			wantCourses: 100,
		},
		{
			name:        "enterprise license",
			licenseType: LicenseTypeEnterprise,
			duration:    365 * 24 * time.Hour,
			wantUsers:   -1,
			wantStuds:   -1,
			wantCourses: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license, err := NewLicense(tt.licenseType, tt.duration)
			if err != nil {
				t.Fatalf("NewLicense() error = %v", err)
			}

			if license.Type != tt.licenseType {
				t.Errorf("NewLicense() Type = %v, want %v", license.Type, tt.licenseType)
			}

			if license.MaxUsers != tt.wantUsers {
				t.Errorf("NewLicense() MaxUsers = %v, want %v", license.MaxUsers, tt.wantUsers)
			}

			if license.MaxStudents != tt.wantStuds {
				t.Errorf("NewLicense() MaxStudents = %v, want %v", license.MaxStudents, tt.wantStuds)
			}

			if license.MaxCourses != tt.wantCourses {
				t.Errorf("NewLicense() MaxCourses = %v, want %v", license.MaxCourses, tt.wantCourses)
			}

			if !license.Active {
				t.Error("NewLicense() Active = false, want true")
			}

			if license.Key == "" {
				t.Error("NewLicense() Key is empty")
			}
		})
	}
}

func TestLicense_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiryAt time.Time
		want     bool
	}{
		{
			name:     "not expired - future date",
			expiryAt: time.Now().Add(24 * time.Hour),
			want:     false,
		},
		{
			name:     "expired - past date",
			expiryAt: time.Now().Add(-24 * time.Hour),
			want:     true,
		},
		{
			name:     "expired - exactly now (edge case)",
			expiryAt: time.Now().Add(-time.Second),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &License{ExpiryAt: tt.expiryAt}
			if got := l.IsExpired(); got != tt.want {
				t.Errorf("License.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicense_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		expiryAt time.Time
		want     bool
	}{
		{
			name:     "valid - active and not expired",
			active:   true,
			expiryAt: time.Now().Add(24 * time.Hour),
			want:     true,
		},
		{
			name:     "invalid - inactive",
			active:   false,
			expiryAt: time.Now().Add(24 * time.Hour),
			want:     false,
		},
		{
			name:     "invalid - expired",
			active:   true,
			expiryAt: time.Now().Add(-24 * time.Hour),
			want:     false,
		},
		{
			name:     "invalid - inactive and expired",
			active:   false,
			expiryAt: time.Now().Add(-24 * time.Hour),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &License{Active: tt.active, ExpiryAt: tt.expiryAt}
			if got := l.IsValid(); got != tt.want {
				t.Errorf("License.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicense_Regenerate(t *testing.T) {
	t.Run("regenerate with extension", func(t *testing.T) {
		originalExpiry := time.Now().Add(10 * 24 * time.Hour)
		l := &License{
			Key:      "old-key",
			Active:   false,
			ExpiryAt: originalExpiry,
		}

		err := l.Regenerate(30 * 24 * time.Hour)
		if err != nil {
			t.Fatalf("License.Regenerate() error = %v", err)
		}

		if l.Key == "old-key" {
			t.Error("License.Regenerate() did not change the key")
		}

		if !l.Active {
			t.Error("License.Regenerate() did not activate the license")
		}

		expectedExpiry := originalExpiry.Add(30 * 24 * time.Hour)
		if l.ExpiryAt.Before(expectedExpiry.Add(-time.Minute)) || l.ExpiryAt.After(expectedExpiry.Add(time.Minute)) {
			t.Errorf("License.Regenerate() ExpiryAt = %v, want approximately %v", l.ExpiryAt, expectedExpiry)
		}
	})

	t.Run("regenerate expired license with extension", func(t *testing.T) {
		l := &License{
			Key:      "old-key",
			Active:   false,
			ExpiryAt: time.Now().Add(-10 * 24 * time.Hour), // Already expired
		}

		err := l.Regenerate(30 * 24 * time.Hour)
		if err != nil {
			t.Fatalf("License.Regenerate() error = %v", err)
		}

		// Should extend from now, not from the expired date
		expectedExpiry := time.Now().Add(30 * 24 * time.Hour)
		if l.ExpiryAt.Before(expectedExpiry.Add(-time.Minute)) || l.ExpiryAt.After(expectedExpiry.Add(time.Minute)) {
			t.Errorf("License.Regenerate() ExpiryAt = %v, want approximately %v", l.ExpiryAt, expectedExpiry)
		}
	})

	t.Run("regenerate without extension", func(t *testing.T) {
		originalExpiry := time.Now().Add(10 * 24 * time.Hour)
		l := &License{
			Key:      "old-key",
			Active:   true,
			ExpiryAt: originalExpiry,
		}

		err := l.Regenerate(0)
		if err != nil {
			t.Fatalf("License.Regenerate() error = %v", err)
		}

		if l.Key == "old-key" {
			t.Error("License.Regenerate() did not change the key")
		}

		// Expiry should remain unchanged
		if !l.ExpiryAt.Equal(originalExpiry) {
			t.Errorf("License.Regenerate() changed ExpiryAt when extension was 0")
		}
	})
}

func TestLicense_DaysUntilExpiry(t *testing.T) {
	tests := []struct {
		name     string
		expiryAt time.Time
		want     int
	}{
		{
			name:     "30 days remaining",
			expiryAt: time.Now().Add(30 * 24 * time.Hour),
			want:     30,
		},
		{
			name:     "1 day remaining",
			expiryAt: time.Now().Add(24 * time.Hour),
			want:     1,
		},
		{
			name:     "expired",
			expiryAt: time.Now().Add(-24 * time.Hour),
			want:     0,
		},
		{
			name:     "less than a day remaining",
			expiryAt: time.Now().Add(12 * time.Hour),
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &License{ExpiryAt: tt.expiryAt}
			got := l.DaysUntilExpiry()
			// Allow 1 day tolerance due to timing
			if got < tt.want-1 || got > tt.want+1 {
				t.Errorf("License.DaysUntilExpiry() = %v, want approximately %v", got, tt.want)
			}
		})
	}
}
