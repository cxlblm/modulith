package module

import (
	"context"

	"modular_monolith/internal/entitlement/app"
	"modular_monolith/internal/entitlement/app/command"
)

type EntitlementModule interface {
	TryConsumeRevivalCard(ctx context.Context, req TryConsumeRevivalCardRequest) (TryConsumeRevivalCardResult, error)
}

type entitlementModule struct {
	app *app.Application
}

func NewEntitlementModule(app *app.Application) EntitlementModule {
	return &entitlementModule{app: app}
}

func (m *entitlementModule) TryConsumeRevivalCard(
	ctx context.Context,
	req TryConsumeRevivalCardRequest,
) (TryConsumeRevivalCardResult, error) {
	result, err := m.app.Commands.TryConsumeRevivalCard.Handle(ctx, command.TryConsumeRevivalCard{UserID: req.UserID})
	if err != nil {
		return TryConsumeRevivalCardResult{}, err
	}
	return TryConsumeRevivalCardResult{Consumed: result.Consumed}, nil
}
