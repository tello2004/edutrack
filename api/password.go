package edutrack

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost factor.
const DefaultCost = bcrypt.DefaultCost

// MaxBcryptLength is the maximum password length bcrypt can handle.
const MaxBcryptLength = 72

// preparePassword pre-hashes passwords that exceed bcrypt's 72-byte limit
// using SHA-256. This ensures any length password can be used securely.
func preparePassword(password string) []byte {
	if len(password) <= MaxBcryptLength {
		return []byte(password)
	}
	// For passwords longer than 72 bytes, hash with SHA-256 first
	// and encode as base64 (which produces a 44-byte string).
	hash := sha256.Sum256([]byte(password))
	encoded := base64.StdEncoding.EncodeToString(hash[:])
	return []byte(encoded)
}

// HashPassword hashes a plain text password using bcrypt.
// Passwords longer than 72 bytes are pre-hashed with SHA-256.
func HashPassword(password string) (string, error) {
	prepared := preparePassword(password)
	bytes, err := bcrypt.GenerateFromPassword(prepared, DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plain text password with a hashed password.
// Returns nil if the password matches, or an error if it doesn't.
func CheckPassword(password, hash string) error {
	prepared := preparePassword(password)
	return bcrypt.CompareHashAndPassword([]byte(hash), prepared)
}

// PasswordMatches is a convenience function that returns true if the
// plain text password matches the hashed password.
func PasswordMatches(password, hash string) bool {
	return CheckPassword(password, hash) == nil
}
