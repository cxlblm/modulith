package command

import "context"

type ClaimContestReward struct {
	ContestID string
	UserID    string
}

type ClaimContestRewardResult struct {
	Claimed        bool
	AlreadyClaimed bool
}

type ClaimContestRewardHandler struct {
	Claims ClaimRepository
}

func (h ClaimContestRewardHandler) Handle(ctx context.Context, cmd ClaimContestReward) (ClaimContestRewardResult, error) {
	if cmd.ContestID == "" || cmd.UserID == "" {
		return ClaimContestRewardResult{}, ErrInvalidRewardClaim
	}
	already, err := h.Claims.Claim(ctx, cmd.ContestID, cmd.UserID)
	if err != nil {
		return ClaimContestRewardResult{}, err
	}
	return ClaimContestRewardResult{Claimed: true, AlreadyClaimed: already}, nil
}
