package edutrack

import "context"

// contextKey represents an internal key for adding context fields.
// This is considered best practice as it prevents other packages from
// interfering with our context keys.
type contextKey int

// List of context keys.
// These are used to store request-scoped information.
const (
	// Stores the current logged in account in the context.
	accountContextKey = contextKey(iota + 1)
)

// NewContextWithAccount returns a new context with the given account.
func NewContextWithAccount(ctx context.Context, account *Account) context.Context {
	return context.WithValue(ctx, accountContextKey, account)
}

// AccountFromContext returns the current logged in account.
func AccountFromContext(ctx context.Context) *Account {
	account, _ := ctx.Value(accountContextKey).(*Account)
	return account
}

// AccountIDFromContext is a helper function that returns the ID of the current
// logged in account. Returns zero if no account is logged in.
func AccountIDFromContext(ctx context.Context) uint {
	if account := AccountFromContext(ctx); account != nil {
		return account.ID
	}
	return 0
}
