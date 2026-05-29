package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/quiz/domain/question"
)

type CreateQuestion struct {
	Type            string
	Prompt          string
	Options         []QuestionOption
	CorrectOptionID string
	AcceptedAnswers []string
	Material        QuestionMaterial
}

type QuestionOption struct {
	ID   string
	Text string
}

type QuestionMaterial struct {
	Kind        string
	AudioURL    string
	PassageText string
}

type CreateQuestionResult struct {
	QuestionID string `json:"question_id"`
}

type CreateQuestionHandler struct {
	Questions QuestionRepository
}

func (h CreateQuestionHandler) Handle(ctx context.Context, cmd CreateQuestion) (CreateQuestionResult, error) {
	var q *question.Question
	var err error
	material := question.Material{Kind: question.MaterialKind(cmd.Material.Kind), AudioURL: cmd.Material.AudioURL, PassageText: cmd.Material.PassageText}
	switch question.Type(cmd.Type) {
	case question.TypeChoice:
		q, err = question.NewChoice(cmd.Prompt, questionOptions(cmd.Options), cmd.CorrectOptionID, material)
	case question.TypeBlank:
		q, err = question.NewBlank(cmd.Prompt, cmd.AcceptedAnswers, material)
	default:
		return CreateQuestionResult{}, question.ErrInvalidQuestion
	}
	if err != nil {
		return CreateQuestionResult{}, err
	}
	if err := h.Questions.Save(ctx, q); err != nil {
		return CreateQuestionResult{}, fmt.Errorf("save question: %w", err)
	}
	return CreateQuestionResult{QuestionID: q.UUID().String()}, nil
}

func questionOptions(options []QuestionOption) []question.Option {
	out := make([]question.Option, 0, len(options))
	for _, option := range options {
		out = append(out, question.Option{ID: option.ID, Text: option.Text})
	}
	return out
}
