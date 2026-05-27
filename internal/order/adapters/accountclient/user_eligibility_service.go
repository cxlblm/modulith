package accountclient

import (
	"context"

	accountmod "modular_monolith/internal/account/ports/module"
)

type UserEligibilityService struct {
	account accountmod.AccountModule
}

func NewUserEligibilityService(account accountmod.AccountModule) *UserEligibilityService {
	return &UserEligibilityService{account: account}
}

func (s *UserEligibilityService) EnsureCanPlaceOrder(ctx context.Context, userID string) error {
	return s.account.EnsureUserActive(ctx, userID)
}
