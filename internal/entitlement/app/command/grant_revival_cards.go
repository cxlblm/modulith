package command

import (
	"context"
	"fmt"
)

type GrantRevivalCards struct {
	UserID string
	Count  int
}

type GrantRevivalCardsHandler struct {
	RevivalCards RevivalCardRepository
}

func (h GrantRevivalCardsHandler) Handle(ctx context.Context, cmd GrantRevivalCards) error {
	if cmd.UserID == "" || cmd.Count <= 0 {
		return ErrInvalidCommand
	}
	if err := h.RevivalCards.Grant(ctx, cmd.UserID, cmd.Count); err != nil {
		return fmt.Errorf("grant revival cards: %w", err)
	}
	return nil
}
