package command

import (
	"context"
	"fmt"
	"time"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/question"
)

type CreateContest struct {
	Title              string
	StartTime          time.Time
	PerQuestionSeconds int
	QuestionIDs        []string
}

type CreateContestResult struct {
	ContestID string `json:"contest_id"`
}

type CreateContestHandler struct {
	Contests  contest.Repository
	Questions question.Repository
}

func (h CreateContestHandler) Handle(ctx context.Context, cmd CreateContest) (CreateContestResult, error) {
	snapshots, err := h.snapshots(ctx, cmd.QuestionIDs)
	if err != nil {
		return CreateContestResult{}, err
	}
	c, err := contest.NewDraft(cmd.Title, cmd.StartTime, cmd.PerQuestionSeconds, snapshots)
	if err != nil {
		return CreateContestResult{}, err
	}
	if err := h.Contests.Save(ctx, c); err != nil {
		return CreateContestResult{}, fmt.Errorf("save contest: %w", err)
	}
	return CreateContestResult{ContestID: c.UUID().String()}, nil
}

func (h CreateContestHandler) snapshots(ctx context.Context, ids []string) ([]contest.QuestionSnapshot, error) {
	snapshots := make([]contest.QuestionSnapshot, 0, len(ids))
	for _, id := range ids {
		q, err := h.Questions.FindByUUID(ctx, question.QuestionUUID(id))
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, questionSnapshot(q))
	}
	return snapshots, nil
}

func questionSnapshot(q *question.Question) contest.QuestionSnapshot {
	material := q.Material()
	options := q.Options()
	return contest.QuestionSnapshot{
		QuestionID:      q.UUID().String(),
		Prompt:          q.Prompt(),
		Type:            string(q.Type()),
		Options:         optionSnapshots(options),
		CorrectOptionID: q.CorrectOptionID(),
		AcceptedAnswers: q.AcceptedAnswers(),
		Material: contest.MaterialSnapshot{
			Kind:        string(material.Kind),
			AudioURL:    material.AudioURL,
			PassageText: material.PassageText,
		},
	}
}

func optionSnapshots(options []question.Option) []contest.OptionSnapshot {
	out := make([]contest.OptionSnapshot, 0, len(options))
	for _, option := range options {
		out = append(out, contest.OptionSnapshot{ID: option.ID, Text: option.Text})
	}
	return out
}
