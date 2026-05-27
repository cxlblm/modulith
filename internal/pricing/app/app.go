package app

import "modular_monolith/internal/pricing/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	CalculateOrderPricing command.CalculateOrderPricingHandler
}
