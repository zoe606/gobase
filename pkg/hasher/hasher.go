// Package hasher provides password hashing utilities using bcrypt.
// Following KISS principle, this is a simple function-based package
// without unnecessary interfaces since bcrypt is the standard choice.
package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost parameter.
// Higher values are more secure but slower.
const DefaultCost = bcrypt.DefaultCost

// Hash generates a bcrypt hash from a password.
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// HashWithCost generates a bcrypt hash with a custom cost parameter.
func HashWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Check compares a password with a hash and returns true if they match.
func Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckWithError compares a password with a hash and returns an error if they don't match.
func CheckWithError(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
