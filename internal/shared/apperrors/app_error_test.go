package apperrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Methods(t *testing.T) {
	errBase := errors.New("underlying DB error")
	appErr := Wrap(errBase, http.StatusInternalServerError, CodeInternalError, "something went wrong")

	if appErr.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", appErr.HTTPStatus())
	}
	if appErr.Code() != CodeInternalError {
		t.Errorf("expected code %s, got %s", CodeInternalError, appErr.Code())
	}
	if appErr.Message() != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", appErr.Message())
	}
	if appErr.Unwrap() != errBase {
		t.Errorf("expected unwrapped error to match, got %v", appErr.Unwrap())
	}
	if appErr.Error() != "something went wrong: underlying DB error" {
		t.Errorf("unexpected Error() string: %s", appErr.Error())
	}

	// Without underlying error
	appErr2 := New(http.StatusBadRequest, CodeBadRequest, "bad request input")
	if appErr2.Error() != "bad request input" {
		t.Errorf("unexpected Error() string: %s", appErr2.Error())
	}
	if appErr2.Unwrap() != nil {
		t.Errorf("expected nil unwrapped error, got %v", appErr2.Unwrap())
	}
}

func TestAppError_WithDetail(t *testing.T) {
	appErr := New(http.StatusBadRequest, CodeValidationError, "validation failed")
	appErr.WithDetail("email", "email is required")

	details := appErr.Details()
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	if details[0].Field != "email" || details[0].Message != "email is required" {
		t.Errorf("unexpected FieldError: %+v", details[0])
	}
}

func TestAppError_Constructors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantStatus int
		wantCode   string
	}{
		{"BadRequest", BadRequest("invalid"), http.StatusBadRequest, CodeBadRequest},
		{"Validation", Validation([]FieldError{{Field: "f", Message: "m"}}), http.StatusBadRequest, CodeValidationError},
		{"Unauthorized", Unauthorized("unauth"), http.StatusUnauthorized, CodeUnauthorized},
		{"Forbidden", Forbidden("forbidden"), http.StatusForbidden, CodeForbidden},
		{"NotFound", NotFound("not found"), http.StatusNotFound, CodeNotFound},
		{"Conflict", Conflict("conflict"), http.StatusConflict, CodeConflict},
		{"Internal", Internal("internal"), http.StatusInternalServerError, CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTPStatus() != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, tt.err.HTTPStatus())
			}
			if tt.err.Code() != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, tt.err.Code())
			}
		})
	}
}
