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
	revivals := &fakeRevivalCards{consumeReply: true}
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
	if revivals.consumeCalls != 1 {
		t.Fatalf("consume calls = %d, want 1", revivals.consumeCalls)
	}
	if participations.saved == nil {
		t.Fatal("participation was not saved")
	}
}

func TestSubmitAnswerEliminatesWhenRevivalIsNotConsumed(t *testing.T) {
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

	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: c},
		Participations: &fakeParticipations{},
		RevivalCards:   &fakeRevivalCards{consumeReply: false},
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
	if result.Status != string(participation.StatusEliminated) || result.UsedRevival || result.Correct {
		t.Fatalf("result = %+v, want eliminated without revival", result)
	}
}

func TestSubmitAnswerDoesNotConsumeRevivalForCorrectAnswer(t *testing.T) {
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
	revivals := &fakeRevivalCards{consumeReply: true}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: c},
		Participations: &fakeParticipations{},
		RevivalCards:   revivals,
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID:  "contest-1",
		UserID:     "user-1",
		QuestionID: "q1",
		OptionID:   "b",
		Now:        start.Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if result.Status != string(participation.StatusPassed) || result.UsedRevival || !result.Correct {
		t.Fatalf("result = %+v, want passed without revival", result)
	}
	if revivals.consumeCalls != 0 {
		t.Fatalf("consume calls = %d, want 0", revivals.consumeCalls)
	}
}

func TestSubmitAnswerDoesNotConsumeRevivalForAlreadyAnsweredQuestion(t *testing.T) {
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
	p, err := participation.New("contest-1", "user-1", []participation.QuestionRef{{ID: "q1"}})
	if err != nil {
		t.Fatalf("participation.New() error = %v", err)
	}
	if _, err := p.Submit("q1", false, true); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	participations := &fakeParticipations{existing: p}
	revivals := &fakeRevivalCards{consumeReply: true}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: c},
		Participations: participations,
		RevivalCards:   revivals,
	}

	_, err = handler.Handle(context.Background(), SubmitAnswer{
		ContestID:  "contest-1",
		UserID:     "user-1",
		QuestionID: "q1",
		OptionID:   "a",
		Now:        start.Add(10 * time.Second),
	})
	if err == nil {
		t.Fatal("Handle() error = nil, want already answered error")
	}
	if revivals.consumeCalls != 0 {
		t.Fatalf("consume calls = %d, want 0", revivals.consumeCalls)
	}
}

type fakeContests struct {
	contest *contest.Contest
}

func (f *fakeContests) FindByUUID(context.Context, contest.ContestUUID) (*contest.Contest, error) {
	return f.contest, nil
}

type fakeParticipations struct {
	existing *participation.Participation
	saved    *participation.Participation
}

func (f *fakeParticipations) FindByContestAndUser(context.Context, string, string) (*participation.Participation, error) {
	if f.existing != nil {
		return f.existing, nil
	}
	return nil, ErrParticipationNotFound
}

func (f *fakeParticipations) Save(_ context.Context, p *participation.Participation) error {
	f.saved = p
	return nil
}

type fakeRevivalCards struct {
	consumeReply bool
	consumeCalls int
}

func (f *fakeRevivalCards) TryConsumeOne(context.Context, string) (bool, error) {
	f.consumeCalls++
	return f.consumeReply, nil
}
