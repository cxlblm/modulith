package command

import (
	"context"
	"fmt"
)

type TryConsumeRevivalCard struct {
	UserID string
}

type TryConsumeRevivalCardResult struct {
	Consumed bool
}

type TryConsumeRevivalCardHandler struct {
	RevivalCards RevivalCardRepository
}

func (h TryConsumeRevivalCardHandler) Handle(ctx context.Context, cmd TryConsumeRevivalCard) (TryConsumeRevivalCardResult, error) {
	if cmd.UserID == "" {
		return TryConsumeRevivalCardResult{}, ErrInvalidCommand
	}
	consumed, err := h.RevivalCards.TryConsumeOne(ctx, cmd.UserID)
	if err != nil {
		return TryConsumeRevivalCardResult{}, fmt.Errorf("try consume revival card: %w", err)
	}
	return TryConsumeRevivalCardResult{Consumed: consumed}, nil
}
