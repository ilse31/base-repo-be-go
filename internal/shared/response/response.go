package response

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
)

type Meta struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
}

type ErrorBody struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details []apperrors.FieldError `json:"details,omitempty"`
}

type Body struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    any        `json:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

func Success(c echo.Context, status int, message string, data any) error {
	return c.JSON(status, Body{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c echo.Context, message string, data any) error {
	return Success(c, http.StatusCreated, message, data)
}

func OK(c echo.Context, data any) error {
	return Success(c, http.StatusOK, "success", data)
}

func OKWithMessage(c echo.Context, message string, data any) error {
	return Success(c, http.StatusOK, message, data)
}

func Message(c echo.Context, status int, message string) error {
	return Success(c, status, message, nil)
}

func List(c echo.Context, items any, meta *Meta) error {
	return c.JSON(http.StatusOK, Body{
		Success: true,
		Message: "success",
		Data:    items,
		Meta:    meta,
	})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func Error(c echo.Context, appErr *apperrors.AppError) error {
	return c.JSON(appErr.HTTPStatus(), Body{
		Success: false,
		Error: &ErrorBody{
			Code:    appErr.Code(),
			Message: appErr.Message(),
			Details: appErr.Details(),
		},
	})
}
