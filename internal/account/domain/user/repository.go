package user

import "context"

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByUUID(ctx context.Context, uuid UserUUID) (*User, error)
	SaveAddress(ctx context.Context, address Address) error
	UpdateAddress(ctx context.Context, address Address) error
	DeleteAddress(ctx context.Context, userUUID UserUUID, addressUUID AddressUUID) error
	FindAddress(ctx context.Context, userUUID UserUUID, addressUUID AddressUUID) (Address, error)
}
