package command

import (
	"context"
	"testing"
	"time"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
	"modular_monolith/internal/quiz/domain/question"
)

func TestSubmitAnswerCreatesParticipationAndUsesRevival(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
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

	participations := &fakeParticipations{}
	revivals := &fakeRevivalCards{cards: map[string]int{"user-1": 1}}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: c},
		Participations: participations,
		RevivalCards:   revivals,
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID:  "contest-1",
		UserID:     "user-1",
		QuestionID: "q1",
		OptionID:   "a",
		Now:        start.Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Status != string(participation.StatusPassed) || !result.UsedRevival || result.Correct {
		t.Fatalf("result = %+v, want passed with revival after wrong answer", result)
	}
	if revivals.cards["user-1"] != 0 {
		t.Fatalf("revival cards = %d, want 0", revivals.cards["user-1"])
	}
	if participations.saved == nil {
		t.Fatal("participation was not saved")
	}
}

type fakeContests struct {
	contest *contest.Contest
}

func (f *fakeContests) FindByUUID(context.Context, contest.ContestUUID) (*contest.Contest, error) {
	return f.contest, nil
}

type fakeParticipations struct {
	saved *participation.Participation
}

func (f *fakeParticipations) FindByContestAndUser(context.Context, string, string) (*participation.Participation, error) {
	return nil, ErrParticipationNotFound
}

func (f *fakeParticipations) Save(_ context.Context, p *participation.Participation) error {
	f.saved = p
	return nil
}

type fakeRevivalCards struct {
	cards map[string]int
}

func (f *fakeRevivalCards) Balance(_ context.Context, userID string) (int, error) {
	return f.cards[userID], nil
}

func (f *fakeRevivalCards) ConsumeOne(_ context.Context, userID string) error {
	f.cards[userID]--
	return nil
}
