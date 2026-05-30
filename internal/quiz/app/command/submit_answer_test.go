package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
	"modular_monolith/internal/quiz/domain/question"
)

func TestSubmitAnswerProcessesDueAnswersAndSkipsFutureAnswers(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	participations := &fakeParticipations{}
	revivals := &fakeRevivalCards{consumeReplies: []bool{true}}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: participations,
		RevivalCards:   revivals,
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers: []SubmittedAnswer{
			{QuestionID: "q1", OptionID: "b"},
			{QuestionID: "q2", OptionID: "x"},
			{QuestionID: "q3", OptionID: "z"},
		},
		Now: start.Add(35 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	assertResult(t, result, SubmitAnswerResult{
		Status:           string(participation.StatusActive),
		ProcessedCount:   2,
		SkippedCount:     1,
		IncorrectCount:   1,
		UsedRevivalCount: 1,
	})
	if revivals.consumeCalls != 1 {
		t.Fatalf("consume calls = %d, want 1", revivals.consumeCalls)
	}
	if got := len(participations.saved.Answers()); got != 2 {
		t.Fatalf("saved answers = %d, want 2", got)
	}
}

func TestSubmitAnswerTreatsMissingDueQuestionsAsRevivedFailures(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	participations := &fakeParticipations{}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: participations,
		RevivalCards:   &fakeRevivalCards{consumeReplies: []bool{true}},
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers:   []SubmittedAnswer{{QuestionID: "q1", OptionID: "b"}},
		Now:       start.Add(35 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	assertResult(t, result, SubmitAnswerResult{
		Status:           string(participation.StatusActive),
		ProcessedCount:   2,
		MissingCount:     1,
		UsedRevivalCount: 1,
	})

	answers := participations.saved.Answers()
	if answers[1].QuestionID != "q2" || answers[1].Correct || !answers[1].UsedRevival || !answers[1].Passed {
		t.Fatalf("missing answer record = %+v, want q2 incorrect revived pass", answers[1])
	}
}

func TestSubmitAnswerSkipsAnsweredDuplicateUnknownAndFutureQuestions(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	existing, err := participation.New("contest-1", "user-1", []participation.QuestionRef{{ID: "q1"}, {ID: "q2"}, {ID: "q3"}})
	if err != nil {
		t.Fatalf("participation.New() error = %v", err)
	}
	if _, err := existing.Submit("q1", true, false); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	participations := &fakeParticipations{existing: existing}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: participations,
		RevivalCards:   &fakeRevivalCards{consumeReplies: []bool{true}},
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers: []SubmittedAnswer{
			{QuestionID: "q1", OptionID: "b"},
			{QuestionID: "q2", OptionID: "x"},
			{QuestionID: "q2", OptionID: "y"},
			{QuestionID: "missing", OptionID: "a"},
			{QuestionID: "q3", OptionID: "z"},
		},
		Now: start.Add(35 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	assertResult(t, result, SubmitAnswerResult{
		Status:           string(participation.StatusActive),
		ProcessedCount:   1,
		SkippedCount:     4,
		IncorrectCount:   1,
		UsedRevivalCount: 1,
	})
}

func TestSubmitAnswerEliminatesWhenRevivalIsUnavailableAndSkipsRemainingDueQuestions(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	participations := &fakeParticipations{}
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: participations,
		RevivalCards:   &fakeRevivalCards{consumeReplies: []bool{false}},
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers: []SubmittedAnswer{
			{QuestionID: "q1", OptionID: "a"},
			{QuestionID: "q2", OptionID: "y"},
			{QuestionID: "q3", OptionID: "z"},
		},
		Now: start.Add(65 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	assertResult(t, result, SubmitAnswerResult{
		Status:         string(participation.StatusEliminated),
		ProcessedCount: 1,
		SkippedCount:   2,
		IncorrectCount: 1,
	})
	if got := len(participations.saved.Answers()); got != 1 {
		t.Fatalf("saved answers = %d, want 1", got)
	}
}

func TestSubmitAnswerAllowsEmptyAnswersToSettleMissingDueQuestions(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: &fakeParticipations{},
		RevivalCards:   &fakeRevivalCards{consumeReplies: []bool{true, true}},
	}

	result, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers:   []SubmittedAnswer{},
		Now:       start.Add(35 * time.Second),
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	assertResult(t, result, SubmitAnswerResult{
		Status:           string(participation.StatusActive),
		ProcessedCount:   2,
		MissingCount:     2,
		UsedRevivalCount: 2,
	})
}

func TestSubmitAnswerRejectsSubmissionAfterGracePeriod(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	handler := SubmitAnswerHandler{
		Contests:       &fakeContests{contest: publishedContest(t, start)},
		Participations: &fakeParticipations{},
		RevivalCards:   &fakeRevivalCards{},
	}

	_, err := handler.Handle(context.Background(), SubmitAnswer{
		ContestID: "contest-1",
		UserID:    "user-1",
		Answers:   []SubmittedAnswer{},
		Now:       start.Add(121 * time.Second),
	})
	if !errors.Is(err, ErrContestNotOpen) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrContestNotOpen)
	}
}

func assertResult(t *testing.T, got SubmitAnswerResult, want SubmitAnswerResult) {
	t.Helper()

	if got != want {
		t.Fatalf("result = %+v, want %+v", got, want)
	}
}

func publishedContest(t *testing.T, start time.Time) *contest.Contest {
	t.Helper()

	c, err := contest.NewDraft("daily", start, 30, []contest.QuestionSnapshot{
		{
			QuestionID:      "q1",
			Prompt:          "choose 1",
			Type:            string(question.TypeChoice),
			Options:         []contest.OptionSnapshot{{ID: "a", Text: "Alpha"}, {ID: "b", Text: "Bravo"}},
			CorrectOptionID: "b",
		},
		{
			QuestionID:      "q2",
			Prompt:          "choose 2",
			Type:            string(question.TypeChoice),
			Options:         []contest.OptionSnapshot{{ID: "x", Text: "Xray"}, {ID: "y", Text: "Yankee"}},
			CorrectOptionID: "y",
		},
		{
			QuestionID:      "q3",
			Prompt:          "choose 3",
			Type:            string(question.TypeChoice),
			Options:         []contest.OptionSnapshot{{ID: "w", Text: "Whiskey"}, {ID: "z", Text: "Zulu"}},
			CorrectOptionID: "z",
		},
	})
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
	consumeReplies []bool
	consumeCalls   int
}

func (f *fakeRevivalCards) TryConsumeOne(context.Context, string) (bool, error) {
	f.consumeCalls++
	if len(f.consumeReplies) == 0 {
		return false, nil
	}
	consumed := f.consumeReplies[0]
	f.consumeReplies = f.consumeReplies[1:]
	return consumed, nil
}
