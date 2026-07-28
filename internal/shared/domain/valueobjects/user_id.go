package valueobjects

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidUserID = errors.New("invalid user ID")
)

type UserID struct {
	value string
}

func NewUserID(value string) (*UserID, error) {
	if value == "" {
		return nil, ErrInvalidUserID
	}

	// Validate if it's a valid UUID
	if _, err := uuid.Parse(value); err != nil {
		return nil, ErrInvalidUserID
	}

	return &UserID{value: value}, nil
}

func NewUserIDFromUUID(id uuid.UUID) *UserID {
	return &UserID{value: id.String()}
}

func GenerateUserID() *UserID {
	return NewUserIDFromUUID(uuid.New())
}

func (u *UserID) String() string {
	return u.value
}

func (u *UserID) Equals(other *UserID) bool {
	if other == nil {
		return false
	}
	return u.value == other.value
}
