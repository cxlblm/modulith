package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/fulfillment/app"
	"modular_monolith/internal/fulfillment/app/command"
	"modular_monolith/internal/fulfillment/app/query"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	e.GET("/shipments", h.listShipments)
	e.POST("/shipments/:shipment_id/send", h.sendShipment)
}

func (h *Handler) listShipments(c *echo.Context) error {
	result, err := h.app.Queries.ListShipments.Handle(c.Request().Context(), query.ListShipments{OrderID: c.QueryParam("order_id")})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) sendShipment(c *echo.Context) error {
	if err := h.app.Commands.SendShipment.Handle(c.Request().Context(), command.SendShipment{ShipmentID: c.Param("shipment_id")}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
