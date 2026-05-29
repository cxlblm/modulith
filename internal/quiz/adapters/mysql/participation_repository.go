package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	platformmysql "modular_monolith/internal/platform/mysql"
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
	model, err := participationModel(p)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		participationUUID, err := saveParticipationModel(tx, model)
		if err != nil {
			return fmt.Errorf("save participation: %w", err)
		}
		if err := tx.Where("participation_uuid = ?", participationUUID).Delete(&ParticipationAnswerModel{}).Error; err != nil {
			return fmt.Errorf("delete participation answers: %w", err)
		}
		answers := participationAnswerModels(p, participationUUID)
		for _, answer := range answers {
			if err := tx.Create(&answer).Error; err != nil {
				return fmt.Errorf("create participation answer: %w", err)
			}
		}
		return nil
	})
}

func saveParticipationModel(tx *gorm.DB, model ParticipationModel) (string, error) {
	result := updateParticipationModel(tx, model.UUID, model)
	if result.Error != nil {
		return "", fmt.Errorf("update participation: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		return model.UUID, nil
	}
	if err := tx.Create(&model).Error; err != nil {
		if isParticipationDuplicateKey(err) {
			participationUUID, findErr := findParticipationUUID(tx, model)
			if findErr != nil {
				return "", findErr
			}
			if updateErr := updateParticipationModel(tx, participationUUID, model).Error; updateErr != nil {
				return "", fmt.Errorf("update participation after duplicate: %w", updateErr)
			}
			return participationUUID, nil
		}
		return "", fmt.Errorf("create participation: %w", err)
	}
	return model.UUID, nil
}

func updateParticipationModel(tx *gorm.DB, uuid string, model ParticipationModel) *gorm.DB {
	return tx.Model(&ParticipationModel{}).
		Where("uuid = ?", uuid).
		Updates(map[string]any{
			"status":         model.Status,
			"questions_json": model.QuestionsJSON,
		})
}

func findParticipationUUID(tx *gorm.DB, model ParticipationModel) (string, error) {
	var existing ParticipationModel
	err := tx.Select("uuid").
		Where("uuid = ?", model.UUID).
		Or("contest_uuid = ? AND user_uuid = ?", model.ContestUUID, model.UserUUID).
		First(&existing).Error
	if err != nil {
		return "", fmt.Errorf("find participation after duplicate: %w", err)
	}
	return existing.UUID, nil
}

func isParticipationDuplicateKey(err error) bool {
	return platformmysql.IsDuplicateKey(err) || strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func participationModel(p *participation.Participation) (ParticipationModel, error) {
	questionsJSON, err := json.Marshal(p.Questions())
	if err != nil {
		return ParticipationModel{}, fmt.Errorf("encode participation questions: %w", err)
	}
	return ParticipationModel{
		UUID:          p.UUID().String(),
		ContestUUID:   p.ContestID(),
		UserUUID:      p.UserID(),
		QuestionsJSON: string(questionsJSON),
		Status:        string(p.Status()),
	}, nil
}

func participationAnswerModels(p *participation.Participation, participationUUID string) []ParticipationAnswerModel {
	answers := make([]ParticipationAnswerModel, 0, len(p.Answers()))
	for _, answer := range p.Answers() {
		answers = append(answers, ParticipationAnswerModel{
			ParticipationUUID: participationUUID,
			QuestionUUID:      answer.QuestionID,
			Correct:           answer.Correct,
			UsedRevival:       answer.UsedRevival,
			Passed:            answer.Passed,
		})
	}
	return answers
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
		participation.Status(model.Status),
		answers,
	), nil
}
