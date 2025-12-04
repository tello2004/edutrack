package edutrack

import "golang.org/x/crypto/bcrypt"

// DefaultCost is the default bcrypt cost factor.
const DefaultCost = bcrypt.DefaultCost

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plain text password with a hashed password.
// Returns nil if the password matches, or an error if it doesn't.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// PasswordMatches is a convenience function that returns true if the
// plain text password matches the hashed password.
func PasswordMatches(password, hash string) bool {
	return CheckPassword(password, hash) == nil
}
