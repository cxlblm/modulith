package app

import (
	"modular_monolith/internal/order/app/command"
	"modular_monolith/internal/order/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	PlaceOrder  command.PlaceOrderHandler
	MarkPaid    command.MarkPaidHandler
	MarkShipped command.MarkShippedHandler
}

type Queries struct {
	GetOrder   query.GetOrderHandler
	ListOrders query.ListOrdersHandler
}
