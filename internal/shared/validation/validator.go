package validation

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/go-playground/validator/v10"

	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
)

func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("password_complexity", validatePasswordComplexity)
	return v
}

func validatePasswordComplexity(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func ToAppError(err error) *apperrors.AppError {
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		details := make([]apperrors.FieldError, 0, len(valErrs))
		for _, fe := range valErrs {
			details = append(details, apperrors.FieldError{
				Field:   lowerFirst(fe.Field()),
				Message: defaultMessage(fe),
			})
		}
		return apperrors.Validation(details)
	}
	return apperrors.BadRequest(err.Error())
}

func defaultMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "password_complexity":
		return fmt.Sprintf("%s must be at least 8 characters long and contain uppercase, lowercase, and a number", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	out := []rune(s)
	if out[0] >= 'A' && out[0] <= 'Z' {
		out[0] += 'a' - 'A'
	}
	return string(out)
}
