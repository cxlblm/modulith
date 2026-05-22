package command

import (
	"context"
	"fmt"

	"modular_monolith/internal/account/domain/user"
)

type AddAddress struct {
	UserID   string
	Receiver string
	Phone    string
	City     string
	Detail   string
}

type AddAddressResult struct {
	AddressID string
}

type AddAddressHandler struct {
	Users user.Repository
}

func (h AddAddressHandler) Handle(ctx context.Context, cmd AddAddress) (AddAddressResult, error) {
	u, err := h.Users.FindByUUID(ctx, user.UserUUID(cmd.UserID))
	if err != nil {
		return AddAddressResult{}, err
	}
	address, err := u.AddAddress(cmd.Receiver, cmd.Phone, cmd.City, cmd.Detail)
	if err != nil {
		return AddAddressResult{}, err
	}
	if err := h.Users.SaveAddress(ctx, address); err != nil {
		return AddAddressResult{}, fmt.Errorf("save address: %w", err)
	}
	return AddAddressResult{AddressID: address.UUID().String()}, nil
}

type UpdateAddress struct {
	UserID    string
	AddressID string
	Receiver  string
	Phone     string
	City      string
	Detail    string
}

type UpdateAddressHandler struct {
	Users user.Repository
}

func (h UpdateAddressHandler) Handle(ctx context.Context, cmd UpdateAddress) error {
	address, err := user.NewAddress(user.UserUUID(cmd.UserID), cmd.Receiver, cmd.Phone, cmd.City, cmd.Detail)
	if err != nil {
		return err
	}
	address = user.RehydrateAddress(user.AddressUUID(cmd.AddressID), user.UserUUID(cmd.UserID), address.Receiver(), address.Phone(), address.City(), address.Detail())
	if err := h.Users.UpdateAddress(ctx, address); err != nil {
		return fmt.Errorf("update address: %w", err)
	}
	return nil
}

type DeleteAddress struct {
	UserID    string
	AddressID string
}

type DeleteAddressHandler struct {
	Users user.Repository
}

func (h DeleteAddressHandler) Handle(ctx context.Context, cmd DeleteAddress) error {
	if err := h.Users.DeleteAddress(ctx, user.UserUUID(cmd.UserID), user.AddressUUID(cmd.AddressID)); err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	return nil
}
