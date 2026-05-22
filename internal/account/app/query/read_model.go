package query

import "context"

type ReadModel interface {
	GetUser(ctx context.Context, userID string) (UserDTO, error)
	ListAddresses(ctx context.Context, userID string) ([]AddressDTO, error)
	GetAddress(ctx context.Context, userID string, addressID string) (AddressDTO, error)
}
