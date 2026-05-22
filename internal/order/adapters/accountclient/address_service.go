package accountclient

import (
	"context"

	accountmod "modular_monolith/internal/account/ports/module"
	ordercmd "modular_monolith/internal/order/app/command"
)

type AddressService struct {
	account accountmod.AccountModule
}

func NewAddressService(account accountmod.AccountModule) *AddressService {
	return &AddressService{account: account}
}

func (s *AddressService) GetAddress(ctx context.Context, userID string, addressID string) (ordercmd.AddressInfo, error) {
	dto, err := s.account.GetAddress(ctx, userID, addressID)
	if err != nil {
		return ordercmd.AddressInfo{}, err
	}
	return ordercmd.AddressInfo{
		ID:       dto.ID,
		UserID:   dto.UserID,
		Receiver: dto.Receiver,
		Phone:    dto.Phone,
		City:     dto.City,
		Detail:   dto.Detail,
	}, nil
}
