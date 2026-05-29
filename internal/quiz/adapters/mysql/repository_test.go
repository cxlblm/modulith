package mysql

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
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

func TestParticipationRepositorySaveCreatesThenUpdatesSnapshot(t *testing.T) {
	ctx := context.Background()
	db := openQuizRepositoryTestDB(t)
	participations := NewParticipationRepository(db)
	p, err := participation.New("contest-save", "user-save", []participation.QuestionRef{{ID: "q1"}, {ID: "q2"}})
	if err != nil {
		t.Fatalf("participation.New() error = %v", err)
	}

	if err := participations.Save(ctx, p); err != nil {
		t.Fatalf("Save(new participation) error = %v", err)
	}
	found, err := participations.FindByContestAndUser(ctx, "contest-save", "user-save")
	if err != nil {
		t.Fatalf("FindByContestAndUser() error = %v", err)
	}
	if found.UUID() != p.UUID() || found.Status() != participation.StatusActive || len(found.Answers()) != 0 {
		t.Fatalf("found participation = %+v answers=%+v, want active participation without answers", found, found.Answers())
	}

	if _, err := p.Submit("q1", true, false); err != nil {
		t.Fatalf("Submit(q1) error = %v", err)
	}
	if err := participations.Save(ctx, p); err != nil {
		t.Fatalf("Save(participation with one answer) error = %v", err)
	}
	if _, err := p.Submit("q2", false, true); err != nil {
		t.Fatalf("Submit(q2) error = %v", err)
	}
	if err := participations.Save(ctx, p); err != nil {
		t.Fatalf("Save(updated participation) error = %v", err)
	}

	var participationCount int64
	if err := db.Model(&ParticipationModel{}).Where("uuid = ?", p.UUID().String()).Count(&participationCount).Error; err != nil {
		t.Fatalf("count participations: %v", err)
	}
	if participationCount != 1 {
		t.Fatalf("participation count = %d, want 1", participationCount)
	}
	var answerCount int64
	if err := db.Model(&ParticipationAnswerModel{}).Where("participation_uuid = ?", p.UUID().String()).Count(&answerCount).Error; err != nil {
		t.Fatalf("count participation answers: %v", err)
	}
	if answerCount != 2 {
		t.Fatalf("answer count = %d, want 2", answerCount)
	}

	found, err = participations.FindByContestAndUser(ctx, "contest-save", "user-save")
	if err != nil {
		t.Fatalf("FindByContestAndUser(updated) error = %v", err)
	}
	answers := found.Answers()
	if found.Status() != participation.StatusPassed || len(answers) != 2 {
		t.Fatalf("found status=%s answers=%+v, want passed with two answers", found.Status(), answers)
	}
	if !answers[0].Correct || answers[0].UsedRevival || !answers[0].Passed {
		t.Fatalf("first answer = %+v, want correct pass without revival", answers[0])
	}
	if answers[1].Correct || !answers[1].UsedRevival || !answers[1].Passed {
		t.Fatalf("second answer = %+v, want incorrect revived pass", answers[1])
	}
}

func TestParticipationRepositorySaveUsesExistingParticipationOnDuplicateContestUser(t *testing.T) {
	ctx := context.Background()
	db := openQuizRepositoryTestDB(t)
	participations := NewParticipationRepository(db)
	existing, err := participation.New("contest-duplicate", "user-duplicate", []participation.QuestionRef{{ID: "q1"}})
	if err != nil {
		t.Fatalf("participation.New(existing) error = %v", err)
	}
	if err := participations.Save(ctx, existing); err != nil {
		t.Fatalf("Save(existing participation) error = %v", err)
	}

	duplicate, err := participation.New("contest-duplicate", "user-duplicate", []participation.QuestionRef{{ID: "q1"}})
	if err != nil {
		t.Fatalf("participation.New(duplicate) error = %v", err)
	}
	if _, err := duplicate.Submit("q1", true, false); err != nil {
		t.Fatalf("Submit(q1) error = %v", err)
	}
	if err := participations.Save(ctx, duplicate); err != nil {
		t.Fatalf("Save(duplicate participation) error = %v", err)
	}

	var participationCount int64
	if err := db.Model(&ParticipationModel{}).
		Where("contest_uuid = ? AND user_uuid = ?", "contest-duplicate", "user-duplicate").
		Count(&participationCount).Error; err != nil {
		t.Fatalf("count participations: %v", err)
	}
	if participationCount != 1 {
		t.Fatalf("participation count = %d, want 1", participationCount)
	}
	found, err := participations.FindByContestAndUser(ctx, "contest-duplicate", "user-duplicate")
	if err != nil {
		t.Fatalf("FindByContestAndUser() error = %v", err)
	}
	if found.UUID() != existing.UUID() {
		t.Fatalf("found UUID = %s, want existing UUID %s", found.UUID(), existing.UUID())
	}
	answers := found.Answers()
	if found.Status() != participation.StatusPassed || len(answers) != 1 || !answers[0].Correct {
		t.Fatalf("found status=%s answers=%+v, want passed with one correct answer", found.Status(), answers)
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
