package presentation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/user/application"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
)

type mockUserRepo struct {
	users map[string]*userdomain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*userdomain.User)}
}

func (r *mockUserRepo) Create(ctx context.Context, user *userdomain.User) error {
	if user.ID == "" {
		user.ID = "generated-id"
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *mockUserRepo) Update(ctx context.Context, user *userdomain.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("not found")
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.users[id]; !ok {
		return errors.New("not found")
	}
	delete(r.users, id)
	return nil
}

func (r *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	res := make([]*userdomain.User, 0, len(r.users))
	for _, u := range r.users {
		res = append(res, u)
	}
	return res, nil
}

func TestUserHandler(t *testing.T) {
	v := validation.New()
	repo := newMockUserRepo()
	svc := application.NewUserService(repo)
	handler := NewUserHandler(svc, v)
	e := echo.New()

	t.Run("CreateUser invalid body & validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateUser(c)
		if err == nil {
			t.Error("expected error for invalid json")
		}

		// Validation error
		reqPayload, _ := json.Marshal(CreateUserRequest{Name: "", Email: "bad-email", Password: "123"})
		req2 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqPayload))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err = handler.CreateUser(c2)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("CreateUser Success", func(t *testing.T) {
		reqPayload, _ := json.Marshal(CreateUserRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "Password123!",
		})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateUser(c)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		// Found
		req := httptest.NewRequest(http.MethodGet, "/users/generated-id", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("generated-id")

		err := handler.GetUser(c)
		if err != nil {
			t.Fatalf("GetUser failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		// Missing param
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req, rec2)
		err = handler.GetUser(c2)
		if err == nil {
			t.Error("expected error for empty id param")
		}

		// Not found
		rec3 := httptest.NewRecorder()
		c3 := e.NewContext(req, rec3)
		c3.SetParamNames("id")
		c3.SetParamValues("not-found-id")
		err = handler.GetUser(c3)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		reqPayload, _ := json.Marshal(UpdateUserRequest{Name: "Alice Updated"})
		req := httptest.NewRequest(http.MethodPut, "/users/generated-id", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("generated-id")

		err := handler.UpdateUser(c)
		if err != nil {
			t.Fatalf("UpdateUser failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		// Missing ID
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req, rec2)
		err = handler.UpdateUser(c2)
		if err == nil {
			t.Error("expected error for empty id param")
		}

		// Invalid Body
		reqBad := httptest.NewRequest(http.MethodPut, "/users/generated-id", bytes.NewBufferString("invalid json"))
		reqBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec3 := httptest.NewRecorder()
		c3 := e.NewContext(reqBad, rec3)
		c3.SetParamNames("id")
		c3.SetParamValues("generated-id")
		err = handler.UpdateUser(c3)
		if err == nil {
			t.Error("expected error for invalid json body")
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?limit=5&offset=0", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ListUsers(c)
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/generated-id", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("generated-id")

		err := handler.DeleteUser(c)
		if err != nil {
			t.Fatalf("DeleteUser failed: %v", err)
		}
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}

		// Missing ID
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req, rec2)
		err = handler.DeleteUser(c2)
		if err == nil {
			t.Error("expected error for empty id param")
		}
	})
}

func TestParsePagination(t *testing.T) {
	limit, offset := parsePagination("20", "10")
	if limit != 20 || offset != 10 {
		t.Errorf("expected 20, 10, got %d, %d", limit, offset)
	}

	limit, offset = parsePagination("invalid", "invalid")
	if limit != 10 || offset != 0 {
		t.Errorf("expected 10, 0, got %d, %d", limit, offset)
	}
}
