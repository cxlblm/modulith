package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/platform/logctx"
)

func TestRequireUserAuth_UnauthorizedWithoutUserUUID(t *testing.T) {
	e := echo.New()
	e.GET("/protected", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, RequireUserAuth())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireUserAuth_AddsUserUUIDToEchoContext(t *testing.T) {
	e := echo.New()
	e.GET("/protected", func(c *echo.Context) error {
		userUUID, ok := UserUUID(c)
		if !ok {
			t.Fatal("UserUUID ok = false, want true")
		}
		if userUUID != testUserUUID {
			t.Fatalf("UserUUID = %q, want %q", userUUID, testUserUUID)
		}
		for _, attr := range logctx.Attrs(c.Request().Context()) {
			if attr.Key == "user_uuid" {
				t.Fatalf("request context user_uuid present after auth middleware: %v", attr)
			}
		}
		return c.NoContent(http.StatusNoContent)
	}, RequireUserAuth())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderUserUUID, testUserUUID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}
