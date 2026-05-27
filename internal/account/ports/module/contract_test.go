package module

import (
	"context"
	"errors"
	"testing"

	"modular_monolith/internal/account/app"
	"modular_monolith/internal/account/app/query"
	"modular_monolith/internal/account/domain/user"
)

type accountModuleReadModel struct {
	dto query.UserDTO
	err error
}

func (m accountModuleReadModel) GetUser(context.Context, string) (query.UserDTO, error) {
	return m.dto, m.err
}

func (m accountModuleReadModel) ListAddresses(context.Context, string) ([]query.AddressDTO, error) {
	return nil, nil
}

func (m accountModuleReadModel) GetAddress(context.Context, string, string) (query.AddressDTO, error) {
	return query.AddressDTO{}, nil
}

func TestAccountModule_EnsureUserActiveDelegatesToQuery(t *testing.T) {
	mod := NewAccountModule(&app.Application{
		Queries: app.Queries{
			EnsureUserActive: query.EnsureUserActiveHandler{ReadModel: accountModuleReadModel{
				dto: query.UserDTO{ID: "user-1", Status: string(user.StatusDisabled)},
			}},
		},
	})

	err := mod.EnsureUserActive(context.Background(), "user-1")

	if !errors.Is(err, user.ErrUserDisabled) {
		t.Fatalf("EnsureUserActive() error = %v, want ErrUserDisabled", err)
	}
}
