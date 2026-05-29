package mysql

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/question"
)

func TestRepositoriesPersistQuestionContestAndArena(t *testing.T) {
	ctx := context.Background()
	db := openQuizRepositoryTestDB(t)
	questions := NewQuestionRepository(db)
	contests := NewContestRepository(db)
	readModel := NewReadModel(db)
	q, err := question.NewChoice("choose", []question.Option{
		{ID: "a", Text: "Alpha"},
		{ID: "b", Text: "Bravo"},
	}, "b", question.Material{Kind: question.MaterialListening, AudioURL: "https://example.test/audio.mp3"})
	if err != nil {
		t.Fatalf("NewChoice() error = %v", err)
	}
	if err := questions.Save(ctx, q); err != nil {
		t.Fatalf("Save(question) error = %v", err)
	}

	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	c, err := contest.NewDraft("daily", start, 30, []contest.QuestionSnapshot{{
		QuestionID:      q.UUID().String(),
		Prompt:          q.Prompt(),
		Type:            string(q.Type()),
		Options:         []contest.OptionSnapshot{{ID: "a", Text: "Alpha"}, {ID: "b", Text: "Bravo"}},
		CorrectOptionID: q.CorrectOptionID(),
		Material:        contest.MaterialSnapshot{Kind: string(question.MaterialListening), AudioURL: "https://example.test/audio.mp3"},
	}})
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}
	if err := c.Publish(); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if err := contests.Save(ctx, c); err != nil {
		t.Fatalf("Save(contest) error = %v", err)
	}

	arena, err := readModel.GetArena(ctx, c.UUID().String())
	if err != nil {
		t.Fatalf("GetArena() error = %v", err)
	}
	if arena.ContestID != c.UUID().String() || len(arena.Questions) != 1 {
		t.Fatalf("arena = %+v, want persisted contest with one question", arena)
	}
	if arena.Questions[0].StartTime != start || arena.Questions[0].EndTime != start.Add(30*time.Second) {
		t.Fatalf("question schedule = %s-%s, want %s-%s", arena.Questions[0].StartTime, arena.Questions[0].EndTime, start, start.Add(30*time.Second))
	}
}

func openQuizRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(Models()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}
