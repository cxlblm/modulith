package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/shop/app"
	"modular_monolith/internal/shop/app/command"
	"modular_monolith/internal/shop/app/query"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	e.POST("/products", h.createProduct)
	e.GET("/products", h.listProducts)
	e.GET("/products/:product_id", h.getProduct)
}

func (h *Handler) createProduct(c *echo.Context) error {
	var req struct {
		Name       string `json:"name" validate:"required"`
		PriceCents int64  `json:"price_cents" validate:"required,gt=0"`
		Stock      int    `json:"stock" validate:"gte=0"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	result, err := h.app.Commands.CreateProduct.Handle(c.Request().Context(), command.CreateProduct{Name: req.Name, PriceCents: req.PriceCents, Stock: req.Stock})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) listProducts(c *echo.Context) error {
	result, err := h.app.Queries.ListProduct.Handle(c.Request().Context(), query.ListProducts{})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) getProduct(c *echo.Context) error {
	result, err := h.app.Queries.GetProduct.Handle(c.Request().Context(), query.GetProduct{ProductID: c.Param("product_id")})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}
