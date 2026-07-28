package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
)

func TestResponseHelpers(t *testing.T) {
	e := echo.New()

	t.Run("Success / OK / Created / OKWithMessage", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		err := OK(c, map[string]string{"foo": "bar"})
		if err != nil {
			t.Fatalf("OK failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var b Body
		_ = json.Unmarshal(rec.Body.Bytes(), &b)
		if !b.Success || b.Message != "success" {
			t.Errorf("unexpected body: %+v", b)
		}
	})

	t.Run("Created", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodPost, "/", nil), rec)

		err := Created(c, "resource created", map[string]int{"id": 1})
		if err != nil {
			t.Fatalf("Created failed: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("OKWithMessage", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		err := OKWithMessage(c, "custom msg", nil)
		if err != nil {
			t.Fatalf("OKWithMessage failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Message", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		err := Message(c, http.StatusAccepted, "accepted")
		if err != nil {
			t.Fatalf("Message failed: %v", err)
		}
		if rec.Code != http.StatusAccepted {
			t.Errorf("expected 202, got %d", rec.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		meta := &Meta{Limit: 10, Offset: 0, Total: 100}
		err := List(c, []string{"a", "b"}, meta)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var b Body
		_ = json.Unmarshal(rec.Body.Bytes(), &b)
		if b.Meta == nil || b.Meta.Total != 100 {
			t.Errorf("unexpected meta: %+v", b.Meta)
		}
	})

	t.Run("NoContent", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodDelete, "/", nil), rec)

		err := NoContent(c)
		if err != nil {
			t.Fatalf("NoContent failed: %v", err)
		}
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("Error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		appErr := apperrors.NotFound("item not found").WithDetail("id", "123")
		err := Error(c, appErr)
		if err != nil {
			t.Fatalf("Error failed: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}

		var b Body
		_ = json.Unmarshal(rec.Body.Bytes(), &b)
		if b.Success || b.Error == nil || b.Error.Code != apperrors.CodeNotFound {
			t.Errorf("unexpected error body: %+v", b)
		}
	})
}
