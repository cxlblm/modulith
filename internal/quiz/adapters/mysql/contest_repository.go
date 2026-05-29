package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"modular_monolith/internal/quiz/domain/contest"
)

type ContestRepository struct {
	db *gorm.DB
}

func NewContestRepository(db *gorm.DB) *ContestRepository {
	return &ContestRepository{db: db}
}

func (r *ContestRepository) Save(ctx context.Context, c *contest.Contest) error {
	model, snapshots, err := contestModels(c)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uuid"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "start_time", "per_question_seconds", "status", "updated_at"}),
		}).Create(&model).Error; err != nil {
			return fmt.Errorf("save contest: %w", err)
		}
		if err := tx.Where("contest_uuid = ?", c.UUID().String()).Delete(&ContestQuestionModel{}).Error; err != nil {
			return fmt.Errorf("delete contest snapshots: %w", err)
		}
		for _, snapshot := range snapshots {
			if err := tx.Create(&snapshot).Error; err != nil {
				return fmt.Errorf("create contest snapshot: %w", err)
			}
		}
		return nil
	})
}

func (r *ContestRepository) FindByUUID(ctx context.Context, uuid contest.ContestUUID) (*contest.Contest, error) {
	var model ContestModel
	if err := r.db.WithContext(ctx).First(&model, "uuid = ?", uuid.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contest.NewContestNotFound(err)
		}
		return nil, fmt.Errorf("find contest: %w", err)
	}
	var snapshotModels []ContestQuestionModel
	if err := r.db.WithContext(ctx).Where("contest_uuid = ?", model.UUID).Order("position").Find(&snapshotModels).Error; err != nil {
		return nil, fmt.Errorf("find contest snapshots: %w", err)
	}
	snapshots, err := toContestSnapshots(snapshotModels)
	if err != nil {
		return nil, err
	}
	return contest.Rehydrate(
		contest.ContestUUID(model.UUID),
		model.Title,
		model.StartTime,
		model.PerQuestionSeconds,
		contest.Status(model.Status),
		snapshots,
	), nil
}

func contestModels(c *contest.Contest) (ContestModel, []ContestQuestionModel, error) {
	model := ContestModel{
		UUID:               c.UUID().String(),
		Title:              c.Title(),
		StartTime:          c.StartTime(),
		PerQuestionSeconds: c.PerQuestionSeconds(),
		Status:             string(c.Status()),
	}
	snapshots := c.Snapshots()
	out := make([]ContestQuestionModel, 0, len(snapshots))
	for i, snapshot := range snapshots {
		optionsJSON, err := json.Marshal(snapshot.Options)
		if err != nil {
			return ContestModel{}, nil, fmt.Errorf("encode contest options: %w", err)
		}
		answersJSON, err := json.Marshal(snapshot.AcceptedAnswers)
		if err != nil {
			return ContestModel{}, nil, fmt.Errorf("encode contest answers: %w", err)
		}
		out = append(out, ContestQuestionModel{
			ContestUUID:         c.UUID().String(),
			Position:            i,
			QuestionUUID:        snapshot.QuestionID,
			Prompt:              snapshot.Prompt,
			Type:                snapshot.Type,
			MaterialKind:        snapshot.Material.Kind,
			AudioURL:            snapshot.Material.AudioURL,
			PassageText:         snapshot.Material.PassageText,
			CorrectOptionID:     snapshot.CorrectOptionID,
			AcceptedAnswersJSON: string(answersJSON),
			OptionsJSON:         string(optionsJSON),
		})
	}
	return model, out, nil
}

func toContestSnapshots(models []ContestQuestionModel) ([]contest.QuestionSnapshot, error) {
	out := make([]contest.QuestionSnapshot, 0, len(models))
	for _, model := range models {
		var options []contest.OptionSnapshot
		if model.OptionsJSON != "" {
			if err := json.Unmarshal([]byte(model.OptionsJSON), &options); err != nil {
				return nil, fmt.Errorf("decode contest options: %w", err)
			}
		}
		var answers []string
		if model.AcceptedAnswersJSON != "" {
			if err := json.Unmarshal([]byte(model.AcceptedAnswersJSON), &answers); err != nil {
				return nil, fmt.Errorf("decode contest answers: %w", err)
			}
		}
		out = append(out, contest.QuestionSnapshot{
			QuestionID:      model.QuestionUUID,
			Prompt:          model.Prompt,
			Type:            model.Type,
			Options:         options,
			CorrectOptionID: model.CorrectOptionID,
			AcceptedAnswers: answers,
			Material: contest.MaterialSnapshot{
				Kind:        model.MaterialKind,
				AudioURL:    model.AudioURL,
				PassageText: model.PassageText,
			},
		})
	}
	return out, nil
}
