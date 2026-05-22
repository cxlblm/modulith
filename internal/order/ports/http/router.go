package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/order/app"
	"modular_monolith/internal/order/app/command"
	"modular_monolith/internal/order/app/query"
	platformhttp "modular_monolith/internal/platform/httpserver"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	g := e.Group("/orders", platformhttp.RequireUserAuth())
	g.POST("", h.placeOrder)
	g.GET("/:order_id", h.getOrder)
	g.GET("", h.listOrders)
}

func (h *Handler) placeOrder(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	var req struct {
		UserID    string `json:"user_id"`
		AddressID string `json:"address_id" validate:"required"`
		Items     []struct {
			ProductID string `json:"product_id" validate:"required"`
			Qty       int    `json:"qty" validate:"required,gt=0"`
		} `json:"items" validate:"required,min=1,dive"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	if req.UserID != "" && req.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	items := make([]command.PlaceOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, command.PlaceOrderItem{ProductID: item.ProductID, Qty: item.Qty})
	}
	result, err := h.app.Commands.PlaceOrder.Handle(c.Request().Context(), command.PlaceOrder{UserID: userID, AddressID: req.AddressID, Items: items})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) getOrder(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	result, err := h.app.Queries.GetOrder.Handle(c.Request().Context(), query.GetOrder{OrderID: c.Param("order_id")})
	if err != nil {
		return err
	}
	if result.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) listOrders(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if queryUserID := c.QueryParam("user_id"); queryUserID != "" && queryUserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	result, err := h.app.Queries.ListOrders.Handle(c.Request().Context(), query.ListOrders{UserID: userID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func currentUserID(c *echo.Context) (string, error) {
	userID, ok := platformhttp.UserUUID(c)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	return userID, nil
}
