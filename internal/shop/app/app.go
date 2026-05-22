package app

import (
	"modular_monolith/internal/shop/app/command"
	"modular_monolith/internal/shop/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateProduct command.CreateProductHandler
	ReserveStock  command.ReserveStockHandler
}

type Queries struct {
	GetProduct  query.GetProductHandler
	ListProduct query.ListProductsHandler
}
