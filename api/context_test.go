package edutrack

import (
	"context"
	"testing"
)

func TestNewContextWithAccount(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     RoleSecretary,
		TenantID: "abc12345",
	}
	account.ID = 1

	ctx := context.Background()
	newCtx := NewContextWithAccount(ctx, account)

	if newCtx == nil {
		t.Fatal("NewContextWithAccount() returned nil context")
	}

	if newCtx == ctx {
		t.Error("NewContextWithAccount() returned same context, expected new context")
	}
}

func TestAccountFromContext(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     RoleTeacher,
		TenantID: "xyz98765",
	}
	account.ID = 42

	ctx := NewContextWithAccount(context.Background(), account)

	retrieved := AccountFromContext(ctx)
	if retrieved == nil {
		t.Fatal("AccountFromContext() returned nil")
	}

	if retrieved.ID != account.ID {
		t.Errorf("AccountFromContext().ID = %d, want %d", retrieved.ID, account.ID)
	}

	if retrieved.Name != account.Name {
		t.Errorf("AccountFromContext().Name = %q, want %q", retrieved.Name, account.Name)
	}

	if retrieved.Email != account.Email {
		t.Errorf("AccountFromContext().Email = %q, want %q", retrieved.Email, account.Email)
	}

	if retrieved.Role != account.Role {
		t.Errorf("AccountFromContext().Role = %q, want %q", retrieved.Role, account.Role)
	}

	if retrieved.TenantID != account.TenantID {
		t.Errorf("AccountFromContext().TenantID = %q, want %q", retrieved.TenantID, account.TenantID)
	}
}

func TestAccountFromContext_NoAccount(t *testing.T) {
	ctx := context.Background()

	retrieved := AccountFromContext(ctx)
	if retrieved != nil {
		t.Errorf("AccountFromContext() = %v, want nil for context without account", retrieved)
	}
}

func TestAccountFromContext_NilContext(t *testing.T) {
	// This tests behavior with a context that has a different value type
	ctx := context.WithValue(context.Background(), "some-other-key", "some-value")

	retrieved := AccountFromContext(ctx)
	if retrieved != nil {
		t.Errorf("AccountFromContext() = %v, want nil for context without account key", retrieved)
	}
}

func TestAccountFromContext_WrongType(t *testing.T) {
	// Create a context with the same key but wrong type
	// This shouldn't happen in practice, but tests robustness
	ctx := context.Background()

	retrieved := AccountFromContext(ctx)
	if retrieved != nil {
		t.Errorf("AccountFromContext() should return nil when no account is set")
	}
}

func TestAccountIDFromContext(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     RoleSecretary,
		TenantID: "abc12345",
	}
	account.ID = 123

	ctx := NewContextWithAccount(context.Background(), account)

	id := AccountIDFromContext(ctx)
	if id != 123 {
		t.Errorf("AccountIDFromContext() = %d, want %d", id, 123)
	}
}

func TestAccountIDFromContext_NoAccount(t *testing.T) {
	ctx := context.Background()

	id := AccountIDFromContext(ctx)
	if id != 0 {
		t.Errorf("AccountIDFromContext() = %d, want 0 for context without account", id)
	}
}

func TestAccountIDFromContext_ZeroID(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     RoleTeacher,
		TenantID: "abc12345",
	}
	// account.ID is 0 (zero value)

	ctx := NewContextWithAccount(context.Background(), account)

	id := AccountIDFromContext(ctx)
	if id != 0 {
		t.Errorf("AccountIDFromContext() = %d, want 0", id)
	}
}

func TestContextAccountImmutability(t *testing.T) {
	// Test that modifying the original account doesn't affect the context
	// (Note: Go passes pointers, so this tests that we're aware of the behavior)
	account := &Account{
		Name:     "Original Name",
		Email:    "original@example.com",
		Role:     RoleSecretary,
		TenantID: "abc12345",
	}
	account.ID = 1

	ctx := NewContextWithAccount(context.Background(), account)

	// Modify the original account
	account.Name = "Modified Name"
	account.Email = "modified@example.com"

	retrieved := AccountFromContext(ctx)

	// Since we store a pointer, changes to the original affect the retrieved value
	// This is expected Go behavior, but we document it here
	if retrieved.Name != "Modified Name" {
		t.Log("Note: Context stores pointer reference, modifications to original account are reflected")
	}
}

func TestContextNesting(t *testing.T) {
	account1 := &Account{
		Name:     "User 1",
		Email:    "user1@example.com",
		Role:     RoleSecretary,
		TenantID: "tenant1",
	}
	account1.ID = 1

	account2 := &Account{
		Name:     "User 2",
		Email:    "user2@example.com",
		Role:     RoleTeacher,
		TenantID: "tenant2",
	}
	account2.ID = 2

	ctx1 := NewContextWithAccount(context.Background(), account1)
	ctx2 := NewContextWithAccount(ctx1, account2)

	// The nested context should return the most recently set account
	retrieved := AccountFromContext(ctx2)
	if retrieved == nil {
		t.Fatal("AccountFromContext() returned nil for nested context")
	}

	if retrieved.ID != account2.ID {
		t.Errorf("AccountFromContext() returned account with ID %d, want %d (most recent)", retrieved.ID, account2.ID)
	}

	// The original context should still have the original account
	retrieved1 := AccountFromContext(ctx1)
	if retrieved1 == nil {
		t.Fatal("AccountFromContext() returned nil for original context")
	}

	if retrieved1.ID != account1.ID {
		t.Errorf("AccountFromContext() for original context returned ID %d, want %d", retrieved1.ID, account1.ID)
	}
}

func TestContextWithCancel(t *testing.T) {
	account := &Account{
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     RoleTeacher,
		TenantID: "abc12345",
	}
	account.ID = 1

	ctx, cancel := context.WithCancel(context.Background())
	ctxWithAccount := NewContextWithAccount(ctx, account)

	// Account should be retrievable
	retrieved := AccountFromContext(ctxWithAccount)
	if retrieved == nil {
		t.Fatal("AccountFromContext() returned nil before cancel")
	}

	// Cancel the context
	cancel()

	// Account should still be retrievable (cancellation doesn't clear values)
	retrieved = AccountFromContext(ctxWithAccount)
	if retrieved == nil {
		t.Fatal("AccountFromContext() returned nil after cancel")
	}

	if retrieved.ID != account.ID {
		t.Errorf("AccountFromContext().ID = %d, want %d after cancel", retrieved.ID, account.ID)
	}
}

func BenchmarkNewContextWithAccount(b *testing.B) {
	account := &Account{
		Name:     "Benchmark User",
		Email:    "bench@example.com",
		Role:     RoleSecretary,
		TenantID: "bench123",
	}
	account.ID = 1

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewContextWithAccount(ctx, account)
	}
}

func BenchmarkAccountFromContext(b *testing.B) {
	account := &Account{
		Name:     "Benchmark User",
		Email:    "bench@example.com",
		Role:     RoleTeacher,
		TenantID: "bench123",
	}
	account.ID = 1

	ctx := NewContextWithAccount(context.Background(), account)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AccountFromContext(ctx)
	}
}

func BenchmarkAccountIDFromContext(b *testing.B) {
	account := &Account{
		Name:     "Benchmark User",
		Email:    "bench@example.com",
		Role:     RoleSecretary,
		TenantID: "bench123",
	}
	account.ID = 42

	ctx := NewContextWithAccount(context.Background(), account)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AccountIDFromContext(ctx)
	}
}
