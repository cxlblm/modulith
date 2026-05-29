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
	ContestID  string
	UserID     string
	QuestionID string
	OptionID   string
	Text       string
	Now        time.Time
}

type SubmitAnswerResult struct {
	QuestionID  string `json:"question_id"`
	Correct     bool   `json:"correct"`
	UsedRevival bool   `json:"used_revival"`
	Status      string `json:"status"`
}

type SubmitAnswerHandler struct {
	Contests       ContestReader
	Participations ParticipationRepository
	RevivalCards   AnswerRevivalCards
}

func (h SubmitAnswerHandler) Handle(ctx context.Context, cmd SubmitAnswer) (SubmitAnswerResult, error) {
	if cmd.ContestID == "" || cmd.UserID == "" || cmd.QuestionID == "" {
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
	if !c.IsPublished() || now.Before(c.StartTime()) || now.After(c.EndTime()) {
		return SubmitAnswerResult{}, ErrContestNotOpen
	}
	snapshot, ok := c.Snapshot(cmd.QuestionID)
	if !ok {
		return SubmitAnswerResult{}, ErrInvalidCommand
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
	correct := grade(snapshot, cmd)
	if err := p.CanSubmit(cmd.QuestionID); err != nil {
		return SubmitAnswerResult{}, err
	}
	usedRevival := false
	if !correct {
		consumed, err := h.RevivalCards.TryConsumeOne(ctx, cmd.UserID)
		if err != nil {
			return SubmitAnswerResult{}, fmt.Errorf("try consume revival card: %w", err)
		}
		usedRevival = consumed
	}
	outcome, err := p.Submit(cmd.QuestionID, correct, usedRevival)
	if err != nil {
		return SubmitAnswerResult{}, err
	}
	if err := h.Participations.Save(ctx, p); err != nil {
		return SubmitAnswerResult{}, fmt.Errorf("save participation: %w", err)
	}
	return SubmitAnswerResult{
		QuestionID:  outcome.QuestionID,
		Correct:     outcome.Correct,
		UsedRevival: outcome.UsedRevival,
		Status:      string(outcome.Status),
	}, nil
}

func (h SubmitAnswerHandler) newParticipation(ctx context.Context, userID string, c *contest.Contest) (*participation.Participation, error) {
	snapshots := c.Snapshots()
	refs := make([]participation.QuestionRef, 0, len(snapshots))
	for _, snapshot := range snapshots {
		refs = append(refs, participation.QuestionRef{ID: snapshot.QuestionID})
	}
	return participation.New(c.UUID().String(), userID, refs)
}

func grade(snapshot contest.QuestionSnapshot, cmd SubmitAnswer) bool {
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
	return q.Grade(question.Answer{OptionID: cmd.OptionID, Text: cmd.Text})
}

func questionOptionsFromSnapshots(options []contest.OptionSnapshot) []question.Option {
	out := make([]question.Option, 0, len(options))
	for _, option := range options {
		out = append(out, question.Option{ID: option.ID, Text: option.Text})
	}
	return out
}
