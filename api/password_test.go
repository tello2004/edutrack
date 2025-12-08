package edutrack

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!#$%^&*()"},
		{"empty password", ""},
		{"long password", "this-is-a-very-long-password-that-exceeds-normal-length-expectations-1234567890"},
		{"unicode password", "contraseña123日本語"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			if hash == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}

			// Hash should be different each time due to salt.
			hash2, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() second call error = %v", err)
			}

			if hash == hash2 {
				t.Error("HashPassword() returned same hash for same password (no salt?)")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name      string
		password  string
		hash      string
		wantError bool
	}{
		{"correct password", password, hash, false},
		{"wrong password", "wrongpassword", hash, true},
		{"empty password", "", hash, true},
		{"empty hash", password, "", true},
		{"invalid hash", password, "not-a-valid-hash", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckPassword() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPasswordMatches(t *testing.T) {
	password := "mySecurePassword!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"matching password", password, hash, true},
		{"non-matching password", "differentPassword", hash, false},
		{"empty password", "", hash, false},
		{"case sensitive", "MYSECUREPASSWORD!", hash, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PasswordMatches(tt.password, tt.hash); got != tt.want {
				t.Errorf("PasswordMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckPassword(password, hash)
	}
}
