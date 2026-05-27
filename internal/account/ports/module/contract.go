package module

import (
	"context"

	"modular_monolith/internal/account/app"
	"modular_monolith/internal/account/app/query"
)

type AccountModule interface {
	GetAddress(ctx context.Context, userID string, addressID string) (AddressDTO, error)
	EnsureUserActive(ctx context.Context, userID string) error
}

type accountModule struct {
	app *app.Application
}

func NewAccountModule(app *app.Application) AccountModule {
	return &accountModule{app: app}
}

func (m *accountModule) GetAddress(ctx context.Context, userID string, addressID string) (AddressDTO, error) {
	dto, err := m.app.Queries.GetAddress.Handle(ctx, query.GetAddress{UserID: userID, AddressID: addressID})
	if err != nil {
		return AddressDTO{}, err
	}
	return AddressDTO{
		ID:       dto.ID,
		UserID:   dto.UserID,
		Receiver: dto.Receiver,
		Phone:    dto.Phone,
		City:     dto.City,
		Detail:   dto.Detail,
	}, nil
}

func (m *accountModule) EnsureUserActive(ctx context.Context, userID string) error {
	return m.app.Queries.EnsureUserActive.Handle(ctx, query.EnsureUserActive{UserID: userID})
}
