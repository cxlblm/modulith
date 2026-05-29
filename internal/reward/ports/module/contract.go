package module

import (
	"context"

	"modular_monolith/internal/reward/app"
	"modular_monolith/internal/reward/app/command"
)

type RewardModule interface {
	ClaimContestReward(ctx context.Context, req ClaimContestRewardRequest) (ClaimContestRewardResult, error)
}

type rewardModule struct {
	app *app.Application
}

func NewRewardModule(app *app.Application) RewardModule {
	return &rewardModule{app: app}
}

func (m *rewardModule) ClaimContestReward(ctx context.Context, req ClaimContestRewardRequest) (ClaimContestRewardResult, error) {
	result, err := m.app.Commands.ClaimContestReward.Handle(ctx, command.ClaimContestReward{ContestID: req.ContestID, UserID: req.UserID})
	if err != nil {
		return ClaimContestRewardResult{}, err
	}
	return ClaimContestRewardResult{Claimed: result.Claimed, AlreadyClaimed: result.AlreadyClaimed}, nil
}
