package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/account/domain/user"
)

type CreateUser struct {
	Name  string
	Email string
}

type CreateUserResult struct {
	UserID string
}

type CreateUserHandler struct {
	Users user.Repository
}

func (h CreateUserHandler) Handle(ctx context.Context, cmd CreateUser) (CreateUserResult, error) {
	u, err := user.NewUser(cmd.Name, cmd.Email)
	if err != nil {
		return CreateUserResult{}, err
	}
	if err := h.Users.Save(ctx, u); err != nil {
		return CreateUserResult{}, fmt.Errorf("save user: %w", err)
	}
	return CreateUserResult{UserID: u.UUID().String()}, nil
}
