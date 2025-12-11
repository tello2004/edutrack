package edutrack

import (
	"testing"
)

func TestAccount_IsSecretary(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "secretary role",
			role: RoleSecretary,
			want: true,
		},
		{
			name: "teacher role",
			role: RoleTeacher,
			want: false,
		},
		{
			name: "empty role",
			role: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Account{Role: tt.role}
			if got := a.IsSecretary(); got != tt.want {
				t.Errorf("Account.IsSecretary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_IsTeacher(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "teacher role",
			role: RoleTeacher,
			want: true,
		},
		{
			name: "secretary role",
			role: RoleSecretary,
			want: false,
		},
		{
			name: "empty role",
			role: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Account{Role: tt.role}
			if got := a.IsTeacher(); got != tt.want {
				t.Errorf("Account.IsTeacher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_Constants(t *testing.T) {
	// Ensure role constants have expected string values
	if RoleSecretary != "secretary" {
		t.Errorf("RoleSecretary = %q, want %q", RoleSecretary, "secretary")
	}

	if RoleTeacher != "teacher" {
		t.Errorf("RoleTeacher = %q, want %q", RoleTeacher, "teacher")
	}
}

func TestAccount_Fields(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     RoleSecretary,
		Active:   true,
		TenantID: "abc12345",
	}

	if account.Name != "Test User" {
		t.Errorf("Account.Name = %q, want %q", account.Name, "Test User")
	}

	if account.Email != "test@example.com" {
		t.Errorf("Account.Email = %q, want %q", account.Email, "test@example.com")
	}

	if account.Password != "hashedpassword" {
		t.Errorf("Account.Password = %q, want %q", account.Password, "hashedpassword")
	}

	if account.Role != RoleSecretary {
		t.Errorf("Account.Role = %q, want %q", account.Role, RoleSecretary)
	}

	if !account.Active {
		t.Error("Account.Active = false, want true")
	}

	if account.TenantID != "abc12345" {
		t.Errorf("Account.TenantID = %q, want %q", account.TenantID, "abc12345")
	}
}

func TestAccount_RoleMutualExclusivity(t *testing.T) {
	// An account cannot be both secretary and teacher
	secretaryAccount := &Account{Role: RoleSecretary}
	teacherAccount := &Account{Role: RoleTeacher}

	// Secretary should be secretary but not teacher
	if !secretaryAccount.IsSecretary() {
		t.Error("Secretary account should return true for IsSecretary()")
	}
	if secretaryAccount.IsTeacher() {
		t.Error("Secretary account should return false for IsTeacher()")
	}

	// Teacher should be teacher but not secretary
	if !teacherAccount.IsTeacher() {
		t.Error("Teacher account should return true for IsTeacher()")
	}
	if teacherAccount.IsSecretary() {
		t.Error("Teacher account should return false for IsSecretary()")
	}
}

func TestAccount_DefaultValues(t *testing.T) {
	// Test that zero-value account has expected defaults
	account := &Account{}

	if account.Name != "" {
		t.Errorf("Account.Name default = %q, want empty string", account.Name)
	}

	if account.Email != "" {
		t.Errorf("Account.Email default = %q, want empty string", account.Email)
	}

	if account.Role != "" {
		t.Errorf("Account.Role default = %q, want empty string", account.Role)
	}

	if account.Active {
		t.Error("Account.Active default = true, want false (Go zero value)")
	}

	if account.TenantID != "" {
		t.Errorf("Account.TenantID default = %q, want empty string", account.TenantID)
	}

	// Zero-value account should be neither secretary nor teacher
	if account.IsSecretary() {
		t.Error("Zero-value account should not be secretary")
	}
	if account.IsTeacher() {
		t.Error("Zero-value account should not be teacher")
	}
}
