package entitlementclient

import (
	"context"

	entitlementmod "modular_monolith/internal/entitlement/ports/module"
	"modular_monolith/internal/quiz/app/command"
)

type RevivalCardsService struct {
	entitlements entitlementmod.EntitlementModule
}

func NewRevivalCardsService(entitlements entitlementmod.EntitlementModule) *RevivalCardsService {
	return &RevivalCardsService{entitlements: entitlements}
}

func (s *RevivalCardsService) TryConsumeOne(ctx context.Context, userID string) (bool, error) {
	result, err := s.entitlements.TryConsumeRevivalCard(ctx, entitlementmod.TryConsumeRevivalCardRequest{UserID: userID})
	if err != nil {
		return false, err
	}
	return result.Consumed, nil
}

var _ command.AnswerRevivalCards = (*RevivalCardsService)(nil)
