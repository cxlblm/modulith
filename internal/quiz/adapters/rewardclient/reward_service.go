package rewardclient

import (
	"context"

	"modular_monolith/internal/quiz/app/command"
	rewardmod "modular_monolith/internal/reward/ports/module"
)

type RewardService struct {
	rewards rewardmod.RewardModule
}

func NewRewardService(rewards rewardmod.RewardModule) *RewardService {
	return &RewardService{rewards: rewards}
}

func (s *RewardService) ClaimContestReward(ctx context.Context, contestID string, userID string) (command.RewardClaimResult, error) {
	result, err := s.rewards.ClaimContestReward(ctx, rewardmod.ClaimContestRewardRequest{ContestID: contestID, UserID: userID})
	if err != nil {
		return command.RewardClaimResult{}, err
	}
	return command.RewardClaimResult{Claimed: result.Claimed, AlreadyClaimed: result.AlreadyClaimed}, nil
}
