package query

import (
	"context"
	"errors"
	"testing"

	"modular_monolith/internal/account/domain/user"
)

type ensureUserActiveReadModel struct {
	dto UserDTO
	err error
}

func (m ensureUserActiveReadModel) GetUser(context.Context, string) (UserDTO, error) {
	return m.dto, m.err
}

func (m ensureUserActiveReadModel) ListAddresses(context.Context, string) ([]AddressDTO, error) {
	return nil, nil
}

func (m ensureUserActiveReadModel) GetAddress(context.Context, string, string) (AddressDTO, error) {
	return AddressDTO{}, nil
}

func TestEnsureUserActiveHandler_ActiveUserPasses(t *testing.T) {
	h := EnsureUserActiveHandler{ReadModel: ensureUserActiveReadModel{
		dto: UserDTO{ID: "user-1", Status: string(user.StatusActive)},
	}}

	if err := h.Handle(context.Background(), EnsureUserActive{UserID: "user-1"}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestEnsureUserActiveHandler_DisabledUserReturnsError(t *testing.T) {
	h := EnsureUserActiveHandler{ReadModel: ensureUserActiveReadModel{
		dto: UserDTO{ID: "user-1", Status: string(user.StatusDisabled)},
	}}

	err := h.Handle(context.Background(), EnsureUserActive{UserID: "user-1"})

	if !errors.Is(err, user.ErrUserDisabled) {
		t.Fatalf("Handle() error = %v, want ErrUserDisabled", err)
	}
}

func TestEnsureUserActiveHandler_UserNotFoundPassesThrough(t *testing.T) {
	h := EnsureUserActiveHandler{ReadModel: ensureUserActiveReadModel{err: user.ErrUserNotFound}}

	err := h.Handle(context.Background(), EnsureUserActive{UserID: "missing-user"})

	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("Handle() error = %v, want ErrUserNotFound", err)
	}
}
