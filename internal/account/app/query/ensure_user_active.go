package query

import (
	"context"

	"modular_monolith/internal/account/domain/user"
)

type EnsureUserActive struct {
	UserID string
}

type EnsureUserActiveHandler struct {
	ReadModel ReadModel
}

func (h EnsureUserActiveHandler) Handle(ctx context.Context, q EnsureUserActive) error {
	dto, err := h.ReadModel.GetUser(ctx, q.UserID)
	if err != nil {
		return err
	}
	if user.Status(dto.Status) == user.StatusDisabled {
		return user.ErrUserDisabled
	}
	return nil
}
