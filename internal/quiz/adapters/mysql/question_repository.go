package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/quiz/domain/question"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Save(ctx context.Context, q *question.Question) error {
	model, options, err := questionModels(q)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return fmt.Errorf("create question: %w", err)
		}
		for _, option := range options {
			if err := tx.Create(&option).Error; err != nil {
				return fmt.Errorf("create question option: %w", err)
			}
		}
		return nil
	})
}

func (r *QuestionRepository) FindByUUID(ctx context.Context, uuid question.QuestionUUID) (*question.Question, error) {
	var model QuestionModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, question.NewQuestionNotFound(err)
		}
		return nil, fmt.Errorf("find question: %w", err)
	}
	var options []QuestionOptionModel
	if err := r.db.WithContext(ctx).Where("question_uuid = ?", model.UUID).Order("id").Find(&options).Error; err != nil {
		return nil, fmt.Errorf("find question options: %w", err)
	}
	return toQuestion(model, options)
}

func (r *QuestionRepository) List(ctx context.Context) ([]*question.Question, error) {
	var models []QuestionModel
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	if len(models) == 0 {
		return []*question.Question{}, nil
	}
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.UUID)
	}
	var optionModels []QuestionOptionModel
	if err := r.db.WithContext(ctx).Where("question_uuid IN ?", ids).Order("question_uuid, id").Find(&optionModels).Error; err != nil {
		return nil, fmt.Errorf("list question options: %w", err)
	}
	optionsByQuestion := make(map[string][]QuestionOptionModel, len(models))
	for _, option := range optionModels {
		optionsByQuestion[option.QuestionUUID] = append(optionsByQuestion[option.QuestionUUID], option)
	}
	out := make([]*question.Question, 0, len(models))
	for _, model := range models {
		q, err := toQuestion(model, optionsByQuestion[model.UUID])
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, nil
}

func questionModels(q *question.Question) (QuestionModel, []QuestionOptionModel, error) {
	answers, err := json.Marshal(q.AcceptedAnswers())
	if err != nil {
		return QuestionModel{}, nil, fmt.Errorf("encode accepted answers: %w", err)
	}
	material := q.Material()
	model := QuestionModel{
		UUID:                q.UUID().String(),
		Type:                string(q.Type()),
		Prompt:              q.Prompt(),
		MaterialKind:        string(material.Kind),
		AudioURL:            material.AudioURL,
		PassageText:         material.PassageText,
		CorrectOptionID:     q.CorrectOptionID(),
		AcceptedAnswersJSON: string(answers),
	}
	options := make([]QuestionOptionModel, 0, len(q.Options()))
	for _, option := range q.Options() {
		options = append(options, QuestionOptionModel{QuestionUUID: q.UUID().String(), OptionID: option.ID, Text: option.Text})
	}
	return model, options, nil
}

func toQuestion(model QuestionModel, optionModels []QuestionOptionModel) (*question.Question, error) {
	var answers []string
	if model.AcceptedAnswersJSON != "" {
		if err := json.Unmarshal([]byte(model.AcceptedAnswersJSON), &answers); err != nil {
			return nil, fmt.Errorf("decode accepted answers: %w", err)
		}
	}
	options := make([]question.Option, 0, len(optionModels))
	for _, option := range optionModels {
		options = append(options, question.Option{ID: option.OptionID, Text: option.Text})
	}
	return question.Rehydrate(
		question.QuestionUUID(model.UUID),
		question.Type(model.Type),
		model.Prompt,
		options,
		model.CorrectOptionID,
		answers,
		question.Material{Kind: question.MaterialKind(model.MaterialKind), AudioURL: model.AudioURL, PassageText: model.PassageText},
	), nil
}
