package auth

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"regexp"
	"unicode"
)

const (
	// DefaultStretches matches Rails production Devise config.
	DefaultStretches = 20

	// MinPasswordLength matches Devise config: config.password_length = 14..128
	MinPasswordLength = 14
	MaxPasswordLength = 128
)

// DigestPassword computes the authlogic_sha512 hash.
// Algorithm: digest = password + salt; 20.times { digest = SHA512.hexdigest(digest) }
func DigestPassword(password, salt string, stretches int) string {
	digest := password + salt
	for range stretches {
		hash := sha512.Sum512([]byte(digest))
		digest = hex.EncodeToString(hash[:])
	}
	return digest
}

// VerifyPassword performs constant-time comparison of a password against a stored hash.
func VerifyPassword(password, encryptedPassword, salt string, stretches int) bool {
	computed := DigestPassword(password, salt, stretches)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(encryptedPassword)) == 1
}

// PasswordComplexityError describes which complexity rules a password violates.
type PasswordComplexityError struct {
	TooShort  bool
	TooLong   bool
	NoDigit   bool
	NoLower   bool
	NoUpper   bool
	NoSymbol  bool
}

func (e *PasswordComplexityError) Error() string {
	return "password does not meet complexity requirements"
}

// HasErrors returns true if any complexity rule was violated.
func (e *PasswordComplexityError) HasErrors() bool {
	return e.TooShort || e.TooLong || e.NoDigit || e.NoLower || e.NoUpper || e.NoSymbol
}

// Messages returns human-readable error messages.
func (e *PasswordComplexityError) Messages() []string {
	var msgs []string
	if e.TooShort {
		msgs = append(msgs, fmt.Sprintf("must be at least %d characters", MinPasswordLength))
	}
	if e.TooLong {
		msgs = append(msgs, fmt.Sprintf("must be at most %d characters", MaxPasswordLength))
	}
	if e.NoDigit {
		msgs = append(msgs, "must contain at least one digit")
	}
	if e.NoLower {
		msgs = append(msgs, "must contain at least one lowercase letter")
	}
	if e.NoUpper {
		msgs = append(msgs, "must contain at least one uppercase letter")
	}
	if e.NoSymbol {
		msgs = append(msgs, "must contain at least one symbol")
	}
	return msgs
}

var symbolRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

// ValidatePasswordComplexity checks password against Devise-security rules.
func ValidatePasswordComplexity(password string) *PasswordComplexityError {
	err := &PasswordComplexityError{}

	if len(password) < MinPasswordLength {
		err.TooShort = true
	}
	if len(password) > MaxPasswordLength {
		err.TooLong = true
	}

	var hasDigit, hasLower, hasUpper bool
	for _, ch := range password {
		switch {
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		}
	}

	err.NoDigit = !hasDigit
	err.NoLower = !hasLower
	err.NoUpper = !hasUpper
	err.NoSymbol = !symbolRegex.MatchString(password)

	return err
}
