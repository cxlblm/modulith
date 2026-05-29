package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"modular_monolith/internal/quiz/app/command"
	"modular_monolith/internal/quiz/domain/participation"
)

type ParticipationRepository struct {
	db *gorm.DB
}

func NewParticipationRepository(db *gorm.DB) *ParticipationRepository {
	return &ParticipationRepository{db: db}
}

func (r *ParticipationRepository) FindByContestAndUser(ctx context.Context, contestID string, userID string) (*participation.Participation, error) {
	var model ParticipationModel
	if err := r.db.WithContext(ctx).First(&model, "contest_uuid = ? AND user_uuid = ?", contestID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, command.ErrParticipationNotFound
		}
		return nil, fmt.Errorf("find participation: %w", err)
	}
	var answerModels []ParticipationAnswerModel
	if err := r.db.WithContext(ctx).Where("participation_uuid = ?", model.UUID).Order("id").Find(&answerModels).Error; err != nil {
		return nil, fmt.Errorf("find participation answers: %w", err)
	}
	return toParticipation(model, answerModels)
}

func (r *ParticipationRepository) Save(ctx context.Context, p *participation.Participation) error {
	model, answers, err := participationModels(p)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uuid"}},
			DoUpdates: clause.AssignmentColumns([]string{"revival_cards", "status", "questions_json", "updated_at"}),
		}).Create(&model).Error; err != nil {
			return fmt.Errorf("save participation: %w", err)
		}
		if err := tx.Where("participation_uuid = ?", p.UUID().String()).Delete(&ParticipationAnswerModel{}).Error; err != nil {
			return fmt.Errorf("delete participation answers: %w", err)
		}
		for _, answer := range answers {
			if err := tx.Create(&answer).Error; err != nil {
				return fmt.Errorf("create participation answer: %w", err)
			}
		}
		return nil
	})
}

func participationModels(p *participation.Participation) (ParticipationModel, []ParticipationAnswerModel, error) {
	questionsJSON, err := json.Marshal(p.Questions())
	if err != nil {
		return ParticipationModel{}, nil, fmt.Errorf("encode participation questions: %w", err)
	}
	model := ParticipationModel{
		UUID:          p.UUID().String(),
		ContestUUID:   p.ContestID(),
		UserUUID:      p.UserID(),
		QuestionsJSON: string(questionsJSON),
		RevivalCards:  p.RevivalCards(),
		Status:        string(p.Status()),
	}
	answers := make([]ParticipationAnswerModel, 0, len(p.Answers()))
	for _, answer := range p.Answers() {
		answers = append(answers, ParticipationAnswerModel{
			ParticipationUUID: p.UUID().String(),
			QuestionUUID:      answer.QuestionID,
			Correct:           answer.Correct,
			UsedRevival:       answer.UsedRevival,
			Passed:            answer.Passed,
		})
	}
	return model, answers, nil
}

func toParticipation(model ParticipationModel, answerModels []ParticipationAnswerModel) (*participation.Participation, error) {
	var questions []participation.QuestionRef
	if err := json.Unmarshal([]byte(model.QuestionsJSON), &questions); err != nil {
		return nil, fmt.Errorf("decode participation questions: %w", err)
	}
	answers := make([]participation.AnswerRecord, 0, len(answerModels))
	for _, answer := range answerModels {
		answers = append(answers, participation.AnswerRecord{
			QuestionID:  answer.QuestionUUID,
			Correct:     answer.Correct,
			UsedRevival: answer.UsedRevival,
			Passed:      answer.Passed,
		})
	}
	return participation.Rehydrate(
		participation.ParticipationUUID(model.UUID),
		model.ContestUUID,
		model.UserUUID,
		questions,
		model.RevivalCards,
		participation.Status(model.Status),
		answers,
	), nil
}
