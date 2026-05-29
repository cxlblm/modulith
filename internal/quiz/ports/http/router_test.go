package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/quiz/app"
)

func TestRegister_ProtectsAnswerRoutes(t *testing.T) {
	e := echo.New()
	Register(e, &app.Application{})

	req := httptest.NewRequest(http.MethodPost, "/quiz/contests/contest-1/answers", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
