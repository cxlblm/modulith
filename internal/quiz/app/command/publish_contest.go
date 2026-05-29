package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/quiz/domain/contest"
)

type PublishContest struct {
	ContestID string
}

type PublishContestHandler struct {
	Contests ContestRepository
}

func (h PublishContestHandler) Handle(ctx context.Context, cmd PublishContest) error {
	c, err := h.Contests.FindByUUID(ctx, contest.ContestUUID(cmd.ContestID))
	if err != nil {
		return err
	}
	if err := c.Publish(); err != nil {
		return err
	}
	if err := h.Contests.Save(ctx, c); err != nil {
		return fmt.Errorf("save contest: %w", err)
	}
	return nil
}
