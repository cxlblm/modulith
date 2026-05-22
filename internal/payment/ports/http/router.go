package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/payment/app"
	"modular_monolith/internal/payment/app/command"
	"modular_monolith/internal/payment/app/query"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	e.GET("/payments", h.listPayments)
	e.POST("/payments/:payment_id/confirm", h.confirmPayment)
}

func (h *Handler) listPayments(c *echo.Context) error {
	result, err := h.app.Queries.ListPayments.Handle(c.Request().Context(), query.ListPayments{OrderID: c.QueryParam("order_id")})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) confirmPayment(c *echo.Context) error {
	if err := h.app.Commands.ConfirmPayment.Handle(c.Request().Context(), command.ConfirmPayment{PaymentID: c.Param("payment_id")}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
