package validation

import (
	"errors"
	"testing"
)

type sampleStruct struct {
	Name     string `validate:"required,min=2,max=10"`
	Email    string `validate:"email"`
	Role     string `validate:"oneof=admin user"`
	UUID     string `validate:"uuid"`
	Password string `validate:"password_complexity"`
}

func TestValidator(t *testing.T) {
	v := New()

	t.Run("valid struct", func(t *testing.T) {
		s := sampleStruct{
			Name:     "Alice",
			Email:    "alice@example.com",
			Role:     "admin",
			UUID:     "123e4567-e89b-12d3-a456-426614174000",
			Password: "Password123",
		}
		err := v.Struct(s)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid tags and ToAppError mapping", func(t *testing.T) {
		s := sampleStruct{
			Name:     "",
			Email:    "not-an-email",
			Role:     "invalid-role",
			UUID:     "not-a-uuid",
			Password: "short",
		}
		err := v.Struct(s)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		appErr := ToAppError(err)
		if appErr.HTTPStatus() != 400 {
			t.Errorf("expected status 400, got %d", appErr.HTTPStatus())
		}
		details := appErr.Details()
		if len(details) == 0 {
			t.Error("expected non-empty details")
		}
	})

	t.Run("min max password complexity validation error tags", func(t *testing.T) {
		s1 := sampleStruct{Name: "A"} // min
		appErr1 := ToAppError(v.Struct(s1))
		if len(appErr1.Details()) == 0 {
			t.Error("expected details for min tag")
		}

		s2 := sampleStruct{Name: "VeryLongNameExceedingMax"} // max
		appErr2 := ToAppError(v.Struct(s2))
		if len(appErr2.Details()) == 0 {
			t.Error("expected details for max tag")
		}

		s3 := sampleStruct{Password: "passwordwithoutuppercase1"}
		appErr3 := ToAppError(v.Struct(s3))
		if len(appErr3.Details()) == 0 {
			t.Error("expected details for password_complexity tag")
		}
	})

	t.Run("non validation error passed to ToAppError", func(t *testing.T) {
		rawErr := errors.New("raw system error")
		appErr := ToAppError(rawErr)
		if appErr.Message() != "raw system error" {
			t.Errorf("expected raw system error message, got %s", appErr.Message())
		}
	})

	t.Run("lowerFirst", func(t *testing.T) {
		if lowerFirst("") != "" {
			t.Errorf("expected empty string")
		}
		if lowerFirst("Name") != "name" {
			t.Errorf("expected name, got %s", lowerFirst("Name"))
		}
		if lowerFirst("alreadyLower") != "alreadyLower" {
			t.Errorf("expected alreadyLower, got %s", lowerFirst("alreadyLower"))
		}
	})
}
