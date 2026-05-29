package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"modular_monolith/internal/quiz/app/query"
	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/question"
)

type ReadModel struct {
	db *gorm.DB
}

func NewReadModel(db *gorm.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (r *ReadModel) ListQuestions(ctx context.Context) ([]query.QuestionDTO, error) {
	questions, err := NewQuestionRepository(r.db).List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]query.QuestionDTO, 0, len(questions))
	for _, q := range questions {
		material := q.Material()
		out = append(out, query.QuestionDTO{
			ID:      q.UUID().String(),
			Type:    string(q.Type()),
			Prompt:  q.Prompt(),
			Options: queryOptions(q.Options()),
			Material: query.MaterialDTO{
				Kind:        string(material.Kind),
				AudioURL:    material.AudioURL,
				PassageText: material.PassageText,
			},
		})
	}
	return out, nil
}

func (r *ReadModel) GetArena(ctx context.Context, contestID string) (query.ArenaDTO, error) {
	c, err := NewContestRepository(r.db).FindByUUID(ctx, contest.ContestUUID(contestID))
	if err != nil {
		return query.ArenaDTO{}, err
	}
	timeline := c.Timeline()
	schedules := make(map[string]contest.Schedule, len(timeline))
	for _, schedule := range timeline {
		schedules[schedule.QuestionID] = schedule
	}
	questions := make([]query.ArenaQuestionDTO, 0, len(c.Snapshots()))
	for _, snapshot := range c.Snapshots() {
		schedule, ok := schedules[snapshot.QuestionID]
		if !ok {
			return query.ArenaDTO{}, fmt.Errorf("missing schedule for question %s", snapshot.QuestionID)
		}
		questions = append(questions, query.ArenaQuestionDTO{
			ID:      snapshot.QuestionID,
			Type:    snapshot.Type,
			Prompt:  snapshot.Prompt,
			Options: queryOptionSnapshots(snapshot.Options),
			Material: query.MaterialDTO{
				Kind:        snapshot.Material.Kind,
				AudioURL:    snapshot.Material.AudioURL,
				PassageText: snapshot.Material.PassageText,
			},
			StartTime: schedule.StartTime,
			EndTime:   schedule.EndTime,
		})
	}
	return query.ArenaDTO{
		ContestID:          c.UUID().String(),
		Title:              c.Title(),
		Status:             string(c.Status()),
		StartTime:          c.StartTime(),
		EndTime:            c.EndTime(),
		PerQuestionSeconds: c.PerQuestionSeconds(),
		Questions:          questions,
	}, nil
}

func queryOptions(options []question.Option) []query.QuestionOptionDTO {
	out := make([]query.QuestionOptionDTO, 0, len(options))
	for _, option := range options {
		out = append(out, query.QuestionOptionDTO{ID: option.ID, Text: option.Text})
	}
	return out
}

func queryOptionSnapshots(options []contest.OptionSnapshot) []query.QuestionOptionDTO {
	out := make([]query.QuestionOptionDTO, 0, len(options))
	for _, option := range options {
		out = append(out, query.QuestionOptionDTO{ID: option.ID, Text: option.Text})
	}
	return out
}
