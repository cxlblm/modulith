package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/platform/httpserver"
	"modular_monolith/internal/quiz/app"
	"modular_monolith/internal/quiz/app/command"
	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
	"modular_monolith/internal/quiz/domain/question"
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

func TestSubmitAnswer_BindsAnswersArray(t *testing.T) {
	start := time.Now().Add(-5 * time.Second)
	server := httpserver.New(httpserver.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := server.Echo()
	Register(e, &app.Application{
		Commands: app.Commands{
			SubmitAnswer: command.SubmitAnswerHandler{
				Contests:       &fakeContests{contest: publishedContest(t, start)},
				Participations: &fakeParticipations{},
				RevivalCards:   &fakeRevivalCards{},
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/quiz/contests/contest-1/answers",
		strings.NewReader(`{"answers":[{"question_id":"q1","option_id":"b"}]}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(httpserver.HeaderUserUUID, "user-1")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"processed_count":1`) {
		t.Fatalf("body = %s, want processed_count 1", rec.Body.String())
	}
}

func TestSubmitAnswer_RejectsMissingAnswersField(t *testing.T) {
	server := httpserver.New(httpserver.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := server.Echo()
	Register(e, &app.Application{})

	req := httptest.NewRequest(http.MethodPost, "/quiz/contests/contest-1/answers", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(httpserver.HeaderUserUUID, "user-1")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestSubmitAnswer_AllowsEmptyAnswersArray(t *testing.T) {
	start := time.Now().Add(-5 * time.Second)
	server := httpserver.New(httpserver.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e := server.Echo()
	Register(e, &app.Application{
		Commands: app.Commands{
			SubmitAnswer: command.SubmitAnswerHandler{
				Contests:       &fakeContests{contest: publishedContest(t, start)},
				Participations: &fakeParticipations{},
				RevivalCards:   &fakeRevivalCards{consumeReplies: []bool{true}},
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/quiz/contests/contest-1/answers", strings.NewReader(`{"answers":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(httpserver.HeaderUserUUID, "user-1")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"missing_count":1`) {
		t.Fatalf("body = %s, want missing_count 1", rec.Body.String())
	}
}

func publishedContest(t *testing.T, start time.Time) *contest.Contest {
	t.Helper()

	c, err := contest.NewDraft("daily", start, 30, []contest.QuestionSnapshot{{
		QuestionID:      "q1",
		Prompt:          "choose",
		Type:            string(question.TypeChoice),
		Options:         []contest.OptionSnapshot{{ID: "a", Text: "Alpha"}, {ID: "b", Text: "Bravo"}},
		CorrectOptionID: "b",
	}})
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}
	if err := c.Publish(); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	return c
}

type fakeContests struct {
	contest *contest.Contest
}

func (f *fakeContests) FindByUUID(context.Context, contest.ContestUUID) (*contest.Contest, error) {
	return f.contest, nil
}

func (f *fakeContests) Save(context.Context, *contest.Contest) error {
	return nil
}

type fakeParticipations struct {
	saved *participation.Participation
}

func (f *fakeParticipations) FindByContestAndUser(context.Context, string, string) (*participation.Participation, error) {
	return nil, command.ErrParticipationNotFound
}

func (f *fakeParticipations) Save(_ context.Context, p *participation.Participation) error {
	f.saved = p
	return nil
}

type fakeRevivalCards struct {
	consumeReplies []bool
}

func (f *fakeRevivalCards) TryConsumeOne(context.Context, string) (bool, error) {
	if len(f.consumeReplies) == 0 {
		return false, nil
	}
	consumed := f.consumeReplies[0]
	f.consumeReplies = f.consumeReplies[1:]
	return consumed, nil
}
