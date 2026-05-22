package app

import (
	"modular_monolith/internal/account/app/command"
	"modular_monolith/internal/account/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateUser    command.CreateUserHandler
	AddAddress    command.AddAddressHandler
	UpdateAddress command.UpdateAddressHandler
	DeleteAddress command.DeleteAddressHandler
}

type Queries struct {
	GetUser       query.GetUserHandler
	ListAddresses query.ListAddressesHandler
	GetAddress    query.GetAddressHandler
}
