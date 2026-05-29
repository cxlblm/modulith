package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
	"modular_monolith/internal/quiz/domain/question"
)

type SubmitAnswer struct {
	ContestID string
	UserID    string
	Answers   []SubmittedAnswer
	Now       time.Time
}

type SubmittedAnswer struct {
	QuestionID string
	OptionID   string
	Text       string
}

type SubmitAnswerResult struct {
	Status           string `json:"status"`
	ProcessedCount   int    `json:"processed_count"`
	SkippedCount     int    `json:"skipped_count"`
	MissingCount     int    `json:"missing_count"`
	IncorrectCount   int    `json:"incorrect_count"`
	UsedRevivalCount int    `json:"used_revival_count"`
}

type SubmitAnswerHandler struct {
	Contests       ContestReader
	Participations ParticipationRepository
	RevivalCards   AnswerRevivalCards
}

func (h SubmitAnswerHandler) Handle(ctx context.Context, cmd SubmitAnswer) (SubmitAnswerResult, error) {
	if cmd.ContestID == "" || cmd.UserID == "" {
		return SubmitAnswerResult{}, ErrInvalidCommand
	}
	now := cmd.Now
	if now.IsZero() {
		now = time.Now()
	}
	c, err := h.Contests.FindByUUID(ctx, contest.ContestUUID(cmd.ContestID))
	if err != nil {
		return SubmitAnswerResult{}, err
	}
	if !c.IsPublished() || now.Before(c.StartTime()) || now.After(c.EndTime().Add(30*time.Second)) {
		return SubmitAnswerResult{}, ErrContestNotOpen
	}
	p, err := h.Participations.FindByContestAndUser(ctx, cmd.ContestID, cmd.UserID)
	if err != nil {
		if !errors.Is(err, ErrParticipationNotFound) {
			return SubmitAnswerResult{}, err
		}
		p, err = h.newParticipation(ctx, cmd.UserID, c)
		if err != nil {
			return SubmitAnswerResult{}, err
		}
	}
	result, err := h.submitDueAnswers(ctx, cmd.UserID, c, p, cmd.Answers, now)
	if err != nil {
		return SubmitAnswerResult{}, err
	}
	if result.ProcessedCount > 0 {
		if err := h.Participations.Save(ctx, p); err != nil {
			return SubmitAnswerResult{}, fmt.Errorf("save participation: %w", err)
		}
	}
	result.Status = string(p.Status())
	return result, nil
}

func (h SubmitAnswerHandler) newParticipation(ctx context.Context, userID string, c *contest.Contest) (*participation.Participation, error) {
	snapshots := c.Snapshots()
	refs := make([]participation.QuestionRef, 0, len(snapshots))
	for _, snapshot := range snapshots {
		refs = append(refs, participation.QuestionRef{ID: snapshot.QuestionID})
	}
	return participation.New(c.UUID().String(), userID, refs)
}

func (h SubmitAnswerHandler) submitDueAnswers(
	ctx context.Context,
	userID string,
	c *contest.Contest,
	p *participation.Participation,
	answers []SubmittedAnswer,
	now time.Time,
) (SubmitAnswerResult, error) {
	snapshots := c.Snapshots()
	dueQuestionIDs := dueQuestionSet(c, now)
	answersByQuestionID, skipped := firstDueAnswers(answers, snapshots, dueQuestionIDs)
	result := SubmitAnswerResult{SkippedCount: skipped}

	for _, snapshot := range snapshots {
		if !dueQuestionIDs[snapshot.QuestionID] {
			continue
		}
		if p.HasAnswered(snapshot.QuestionID) {
			result.SkippedCount++
			continue
		}
		if p.Status() != participation.StatusActive {
			result.SkippedCount++
			continue
		}

		answer, ok := answersByQuestionID[snapshot.QuestionID]
		missing := !ok || !hasUsableAnswer(snapshot, answer)
		correct := false
		if missing {
			result.MissingCount++
		} else {
			correct = grade(snapshot, answer)
			if !correct {
				result.IncorrectCount++
			}
		}

		usedRevival := false
		if missing || !correct {
			consumed, err := h.RevivalCards.TryConsumeOne(ctx, userID)
			if err != nil {
				return SubmitAnswerResult{}, fmt.Errorf("try consume revival card: %w", err)
			}
			usedRevival = consumed
			if usedRevival {
				result.UsedRevivalCount++
			}
		}
		if _, err := p.Submit(snapshot.QuestionID, correct, usedRevival); err != nil {
			return SubmitAnswerResult{}, err
		}
		result.ProcessedCount++
	}
	return result, nil
}

func dueQuestionSet(c *contest.Contest, now time.Time) map[string]bool {
	due := make(map[string]bool, len(c.Snapshots()))
	for _, schedule := range c.Timeline() {
		if !now.Before(schedule.StartTime) {
			due[schedule.QuestionID] = true
		}
	}
	return due
}

func firstDueAnswers(
	answers []SubmittedAnswer,
	snapshots []contest.QuestionSnapshot,
	dueQuestionIDs map[string]bool,
) (map[string]SubmittedAnswer, int) {
	knownQuestionIDs := make(map[string]bool, len(snapshots))
	for _, snapshot := range snapshots {
		knownQuestionIDs[snapshot.QuestionID] = true
	}
	out := make(map[string]SubmittedAnswer, len(answers))
	skipped := 0
	for _, answer := range answers {
		if !knownQuestionIDs[answer.QuestionID] {
			skipped++
			continue
		}
		if !dueQuestionIDs[answer.QuestionID] {
			skipped++
			continue
		}
		if _, exists := out[answer.QuestionID]; exists {
			skipped++
			continue
		}
		out[answer.QuestionID] = answer
	}
	return out, skipped
}

func hasUsableAnswer(snapshot contest.QuestionSnapshot, answer SubmittedAnswer) bool {
	switch question.Type(snapshot.Type) {
	case question.TypeChoice:
		return answer.OptionID != ""
	case question.TypeBlank:
		return answer.Text != ""
	default:
		return false
	}
}

func grade(snapshot contest.QuestionSnapshot, answer SubmittedAnswer) bool {
	q := question.Rehydrate(
		question.QuestionUUID(snapshot.QuestionID),
		question.Type(snapshot.Type),
		snapshot.Prompt,
		questionOptionsFromSnapshots(snapshot.Options),
		snapshot.CorrectOptionID,
		snapshot.AcceptedAnswers,
		question.Material{
			Kind:        question.MaterialKind(snapshot.Material.Kind),
			AudioURL:    snapshot.Material.AudioURL,
			PassageText: snapshot.Material.PassageText,
		},
	)
	return q.Grade(question.Answer{OptionID: answer.OptionID, Text: answer.Text})
}

func questionOptionsFromSnapshots(options []contest.OptionSnapshot) []question.Option {
	out := make([]question.Option, 0, len(options))
	for _, option := range options {
		out = append(out, question.Option{ID: option.ID, Text: option.Text})
	}
	return out
}
