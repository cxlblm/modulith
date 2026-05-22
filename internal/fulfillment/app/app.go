package app

import (
	"modular_monolith/internal/fulfillment/app/command"
	"modular_monolith/internal/fulfillment/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateShipment command.CreateShipmentHandler
	SendShipment   command.SendShipmentHandler
}

type Queries struct {
	ListShipments query.ListShipmentsHandler
}
