package presentation

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/user/application"
	"github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
	"github.com/ilse31/base-repo-be-go/internal/shared/response"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
)

type UserHandler struct {
	userService *application.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService *application.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   v,
	}
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.userService.CreateUser(c.Request().Context(), user); err != nil {
		return err
	}

	return response.Created(c, "user created successfully", user)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return apperrors.BadRequest("user id is required")
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return response.OK(c, user)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return apperrors.BadRequest("user id is required")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	user.Name = req.Name

	if err := h.userService.UpdateUser(c.Request().Context(), user); err != nil {
		return err
	}

	return response.OKWithMessage(c, "user updated successfully", user)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return apperrors.BadRequest("user id is required")
	}

	if err := h.userService.DeleteUser(c.Request().Context(), id); err != nil {
		return err
	}

	return response.NoContent(c)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	limit, offset := parsePagination(c.QueryParam("limit"), c.QueryParam("offset"))

	users, err := h.userService.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return err
	}

	return response.List(c, users, &response.Meta{
		Limit:  limit,
		Offset: offset,
	})
}

func parsePagination(limitStr, offsetStr string) (int, int) {
	limit, offset := 10, 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
