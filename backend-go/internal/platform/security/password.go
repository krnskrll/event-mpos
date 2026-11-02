// Package security holds cryptographic helpers shared across domains (password hashing, etc.).
package security

import "golang.org/x/crypto/bcrypt"

// bcryptCost matches the Kotlin BCryptPasswordEncoder(12).
const bcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password (cost 12).
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword reports whether the plaintext matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
