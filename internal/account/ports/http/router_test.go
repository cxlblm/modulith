package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/account/app"
)

func TestRegister_ProtectsUserRoutes(t *testing.T) {
	e := echo.New()
	Register(e, &app.Application{})

	req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
