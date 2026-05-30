package command

import (
	"context"
)

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
