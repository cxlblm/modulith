package query

import "context"

type ListAddresses struct {
	UserID string
}

type ListAddressesHandler struct {
	ReadModel ReadModel
}

func (h ListAddressesHandler) Handle(ctx context.Context, q ListAddresses) ([]AddressDTO, error) {
	return h.ReadModel.ListAddresses(ctx, q.UserID)
}

type GetAddress struct {
	UserID    string
	AddressID string
}

type GetAddressHandler struct {
	ReadModel ReadModel
}

func (h GetAddressHandler) Handle(ctx context.Context, q GetAddress) (AddressDTO, error) {
	return h.ReadModel.GetAddress(ctx, q.UserID, q.AddressID)
}
