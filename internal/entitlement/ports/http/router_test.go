package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/entitlement/app"
	"modular_monolith/internal/entitlement/app/command"
	"modular_monolith/internal/platform/httpserver"
)

func TestGrantRevivalCards(t *testing.T) {
	revivals := &fakeRevivalCards{}
	server := httpserver.New(httpserver.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := server.Echo()
	Register(e, &app.Application{
		Commands: app.Commands{
			GrantRevivalCards: command.GrantRevivalCardsHandler{RevivalCards: revivals},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/entitlements/revival-cards/grant", strings.NewReader(`{"user_id":"user-1","count":2}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if revivals.grantedUserID != "user-1" || revivals.grantedCount != 2 {
		t.Fatalf("grant = (%q, %d), want (user-1, 2)", revivals.grantedUserID, revivals.grantedCount)
	}
}

func TestGrantRevivalCards_RejectsInvalidCount(t *testing.T) {
	server := httpserver.New(httpserver.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := server.Echo()
	Register(e, &app.Application{})

	req := httptest.NewRequest(http.MethodPost, "/entitlements/revival-cards/grant", strings.NewReader(`{"user_id":"user-1","count":0}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type fakeRevivalCards struct {
	grantedUserID string
	grantedCount  int
}

func (f *fakeRevivalCards) Grant(_ context.Context, userID string, count int) error {
	f.grantedUserID = userID
	f.grantedCount = count
	return nil
}

func (f *fakeRevivalCards) TryConsumeOne(context.Context, string) (bool, error) {
	return false, nil
}
