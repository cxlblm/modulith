package app

import (
	"modular_monolith/internal/payment/app/command"
	"modular_monolith/internal/payment/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	InitializePayment command.InitializePaymentHandler
	ConfirmPayment    command.ConfirmPaymentHandler
}

type Queries struct {
	ListPayments query.ListPaymentsHandler
}
