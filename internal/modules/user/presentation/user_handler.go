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

// CreateUser godoc
// @Summary Create a new user
// @Description Creates a new user record
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "Create User Payload"
// @Success 201 {object} response.Body "user created successfully"
// @Failure 400 {object} apperrors.AppError "validation / bad request error"
// @Failure 401 {object} apperrors.AppError "unauthorized"
// @Router /users [post]
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

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieves a user details by unique ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.Body "user details"
// @Failure 400 {object} apperrors.AppError "bad request"
// @Failure 404 {object} apperrors.AppError "user not found"
// @Router /users/{id} [get]
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

// UpdateUser godoc
// @Summary Update user
// @Description Updates user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Update User Payload"
// @Success 200 {object} response.Body "user updated successfully"
// @Failure 400 {object} apperrors.AppError "bad request"
// @Failure 404 {object} apperrors.AppError "user not found"
// @Router /users/{id} [put]
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

// DeleteUser godoc
// @Summary Delete user
// @Description Removes user record by ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 "no content"
// @Failure 400 {object} apperrors.AppError "bad request"
// @Failure 404 {object} apperrors.AppError "user not found"
// @Router /users/{id} [delete]
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

// ListUsers godoc
// @Summary List users
// @Description Returns paginated list of users
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Body "list of users"
// @Router /users [get]
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
