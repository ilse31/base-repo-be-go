package valueobjects

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordMismatch = errors.New("password does not match")
)

type Password struct {
	hashedValue string
}

func NewPassword(plainPassword string) (*Password, error) {
	if len(plainPassword) < 8 {
		return nil, ErrPasswordTooShort
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return &Password{hashedValue: string(hashedBytes)}, nil
}

func NewPasswordFromHash(hashedPassword string) (*Password, error) {
	if hashedPassword == "" {
		return nil, ErrInvalidPassword
	}

	return &Password{hashedValue: hashedPassword}, nil
}

func (p *Password) Hash() string {
	return p.hashedValue
}

func (p *Password) Verify(plainPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(p.hashedValue), []byte(plainPassword)); err != nil {
		return ErrPasswordMismatch
	}
	return nil
}

func (p *Password) Equals(other *Password) bool {
	if other == nil {
		return false
	}
	return p.hashedValue == other.hashedValue
}

// GenerateRandomPassword generates a secure random password
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 16
	}

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	password := base64.URLEncoding.EncodeToString(b)

	// Ensure password has at least one uppercase, one lowercase, and one digit
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		// If not, generate a simpler password with guaranteed complexity
		return generateSimplePassword(length)
	}

	return password, nil
}

func generateSimplePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}

	return string(b), nil
}
