package command

import (
	"context"

	"modular_monolith/internal/quiz/domain/contest"
	"modular_monolith/internal/quiz/domain/participation"
	"modular_monolith/internal/quiz/domain/question"
)

type ContestRepository interface {
	Save(ctx context.Context, c *contest.Contest) error
	FindByUUID(ctx context.Context, uuid contest.ContestUUID) (*contest.Contest, error)
}

type ContestReader interface {
	FindByUUID(ctx context.Context, uuid contest.ContestUUID) (*contest.Contest, error)
}

type QuestionRepository interface {
	Save(ctx context.Context, q *question.Question) error
	FindByUUID(ctx context.Context, uuid question.QuestionUUID) (*question.Question, error)
}

type ParticipationRepository interface {
	FindByContestAndUser(ctx context.Context, contestID string, userID string) (*participation.Participation, error)
	Save(ctx context.Context, p *participation.Participation) error
}

type AnswerRevivalCards interface {
	TryConsumeOne(ctx context.Context, userID string) (bool, error)
}

type RewardService interface {
	ClaimContestReward(ctx context.Context, contestID string, userID string) (RewardClaimResult, error)
}

type RewardClaimResult struct {
	Claimed        bool
	AlreadyClaimed bool
}
