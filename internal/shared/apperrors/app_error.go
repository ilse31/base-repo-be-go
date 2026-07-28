package apperrors

import (
	"fmt"
	"net/http"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	code       string
	message    string
	httpStatus int
	details    []FieldError
	err        error
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		code:       code,
		message:    message,
		httpStatus: httpStatus,
	}
}

func Wrap(err error, httpStatus int, code, message string) *AppError {
	return &AppError{
		code:       code,
		message:    message,
		httpStatus: httpStatus,
		err:        err,
	}
}

func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *AppError) Unwrap() error { return e.err }

func (e *AppError) Code() string          { return e.code }
func (e *AppError) Message() string       { return e.message }
func (e *AppError) HTTPStatus() int       { return e.httpStatus }
func (e *AppError) Details() []FieldError { return e.details }

func (e *AppError) WithDetail(field, message string) *AppError {
	e.details = append(e.details, FieldError{Field: field, Message: message})
	return e
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

func Validation(details []FieldError) *AppError {
	return &AppError{
		code:       CodeValidationError,
		message:    "validation failed",
		httpStatus: http.StatusBadRequest,
		details:    details,
	}
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

func Internal(message string) *AppError {
	return New(http.StatusInternalServerError, CodeInternalError, message)
}
