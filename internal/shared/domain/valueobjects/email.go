package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type Email struct {
	value string
}

func NewEmail(value string) (*Email, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, ErrInvalidEmail
	}

	if !emailRegex.MatchString(value) {
		return nil, ErrInvalidEmail
	}

	return &Email{value: strings.ToLower(value)}, nil
}

func (e *Email) String() string {
	return e.value
}

func (e *Email) Equals(other *Email) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}
