package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
)

type SubmitAnswer struct {
	ContestID string
	UserID    string
	Answers   []SubmittedAnswer
	Now       time.Time
}

type SubmittedAnswer struct {
	QuestionID string
	OptionID   string
	Text       string
}

type SubmitAnswerResult struct {
	Status           string `json:"status"`
	ProcessedCount   int    `json:"processed_count"`
	SkippedCount     int    `json:"skipped_count"`
	MissingCount     int    `json:"missing_count"`
	IncorrectCount   int    `json:"incorrect_count"`
	UsedRevivalCount int    `json:"used_revival_count"`
}

type SubmitAnswerHandler struct {
	Contests       ContestReader
	Participations ParticipationRepository
	RevivalCards   AnswerRevivalCards
}

func (h SubmitAnswerHandler) Handle(ctx context.Context, cmd SubmitAnswer) (SubmitAnswerResult, error) {
	if cmd.ContestID == "" || cmd.UserID == "" {
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
	if !c.CanSubmitAt(now, 30*time.Second) {
		return SubmitAnswerResult{}, ErrContestNotOpen
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
	processor := submitAnswerProcessor{
		revivalCards:  h.RevivalCards,
		userID:        cmd.UserID,
		contest:       c,
		participation: p,
		answers:       cmd.Answers,
		now:           now,
	}
	result, err := processor.process(ctx)
	if err != nil {
		return SubmitAnswerResult{}, err
	}
	if result.ProcessedCount > 0 {
		if err := h.Participations.Save(ctx, p); err != nil {
			return SubmitAnswerResult{}, fmt.Errorf("save participation: %w", err)
		}
	}
	result.Status = string(p.Status())
	return result, nil
}

func (h SubmitAnswerHandler) newParticipation(ctx context.Context, userID string, c *contest.Contest) (*participation.Participation, error) {
	snapshots := c.Snapshots()
	refs := make([]participation.QuestionRef, 0, len(snapshots))
	for _, snapshot := range snapshots {
		refs = append(refs, participation.QuestionRef{ID: snapshot.QuestionID})
	}
	return participation.New(c.UUID().String(), userID, refs)
}

type submitAnswerProcessor struct {
	revivalCards  AnswerRevivalCards
	userID        string
	contest       *contest.Contest
	participation *participation.Participation
	answers       []SubmittedAnswer
	now           time.Time
	result        SubmitAnswerResult
}

func (p *submitAnswerProcessor) process(ctx context.Context) (SubmitAnswerResult, error) {
	snapshots := p.contest.Snapshots()
	dueQuestionIDs := p.contest.DueQuestionIDs(p.now)
	answersByQuestionID, skipped := firstDueAnswers(p.answers, snapshots, dueQuestionIDs)
	p.result.SkippedCount = skipped

	for _, snapshot := range snapshots {
		if !dueQuestionIDs[snapshot.QuestionID] {
			continue
		}
		answer, ok := answersByQuestionID[snapshot.QuestionID]
		if err := p.processQuestion(ctx, snapshot, answer, ok); err != nil {
			return SubmitAnswerResult{}, err
		}
	}
	return p.result, nil
}

func (p *submitAnswerProcessor) processQuestion(
	ctx context.Context,
	snapshot contest.QuestionSnapshot,
	answer SubmittedAnswer,
	answerFound bool,
) error {
	if p.participation.HasAnswered(snapshot.QuestionID) {
		p.result.SkippedCount++
		return nil
	}
	if p.participation.Status() != participation.StatusActive {
		p.result.SkippedCount++
		return nil
	}

	evaluation := p.evaluate(snapshot, answer, answerFound)
	usedRevival := false
	if evaluation.Missing || !evaluation.Correct {
		consumed, err := p.tryRevive(ctx)
		if err != nil {
			return err
		}
		usedRevival = consumed
	}
	if _, err := p.participation.Submit(snapshot.QuestionID, evaluation.Correct, usedRevival); err != nil {
		return err
	}
	p.result.ProcessedCount++
	return nil
}

func (p *submitAnswerProcessor) evaluate(
	snapshot contest.QuestionSnapshot,
	answer SubmittedAnswer,
	answerFound bool,
) contest.AnswerEvaluation {
	evaluation := snapshot.EvaluateAnswer(contest.Answer{
		OptionID: answer.OptionID,
		Text:     answer.Text,
	}, answerFound)
	if evaluation.Missing {
		p.result.MissingCount++
		return evaluation
	}
	if !evaluation.Correct {
		p.result.IncorrectCount++
	}
	return evaluation
}

func (p *submitAnswerProcessor) tryRevive(ctx context.Context) (bool, error) {
	consumed, err := p.revivalCards.TryConsumeOne(ctx, p.userID)
	if err != nil {
		return false, fmt.Errorf("try consume revival card: %w", err)
	}
	if consumed {
		p.result.UsedRevivalCount++
	}
	return consumed, nil
}

func firstDueAnswers(
	answers []SubmittedAnswer,
	snapshots []contest.QuestionSnapshot,
	dueQuestionIDs map[string]bool,
) (map[string]SubmittedAnswer, int) {
	knownQuestionIDs := make(map[string]bool, len(snapshots))
	for _, snapshot := range snapshots {
		knownQuestionIDs[snapshot.QuestionID] = true
	}
	out := make(map[string]SubmittedAnswer, len(answers))
	skipped := 0
	for _, answer := range answers {
		if !knownQuestionIDs[answer.QuestionID] {
			skipped++
			continue
		}
		if !dueQuestionIDs[answer.QuestionID] {
			skipped++
			continue
		}
		if _, exists := out[answer.QuestionID]; exists {
			skipped++
			continue
		}
		out[answer.QuestionID] = answer
	}
	return out, skipped
}
