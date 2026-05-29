package command

import (
	"context"
	"errors"
	"fmt"

	"modular_monolith/internal/quiz/domain/participation"
)

type ClaimReward struct {
	ContestID string
	UserID    string
}

type ClaimRewardResult struct {
	Claimed        bool `json:"claimed"`
	AlreadyClaimed bool `json:"already_claimed"`
}

type ClaimRewardHandler struct {
	Participations ParticipationRepository
	Rewards        RewardService
}

func (h ClaimRewardHandler) Handle(ctx context.Context, cmd ClaimReward) (ClaimRewardResult, error) {
	if cmd.ContestID == "" || cmd.UserID == "" {
		return ClaimRewardResult{}, ErrInvalidCommand
	}
	p, err := h.Participations.FindByContestAndUser(ctx, cmd.ContestID, cmd.UserID)
	if err != nil {
		if errors.Is(err, ErrParticipationNotFound) {
			return ClaimRewardResult{}, ErrRewardNotAllowed
		}
		return ClaimRewardResult{}, err
	}
	if p.Status() != participation.StatusPassed {
		return ClaimRewardResult{}, ErrRewardNotAllowed
	}
	result, err := h.Rewards.ClaimContestReward(ctx, cmd.ContestID, cmd.UserID)
	if err != nil {
		return ClaimRewardResult{}, fmt.Errorf("claim contest reward: %w", err)
	}
	return ClaimRewardResult{Claimed: result.Claimed, AlreadyClaimed: result.AlreadyClaimed}, nil
}
