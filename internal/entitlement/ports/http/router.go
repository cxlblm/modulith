package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/entitlement/app"
	"modular_monolith/internal/entitlement/app/command"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	e.POST("/entitlements/revival-cards/grant", h.grantRevivalCards)
}

func (h *Handler) grantRevivalCards(c *echo.Context) error {
	var req struct {
		UserID string `json:"user_id" validate:"required"`
		Count  int    `json:"count" validate:"required,gt=0"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	if err := h.app.Commands.GrantRevivalCards.Handle(c.Request().Context(), command.GrantRevivalCards{
		UserID: req.UserID,
		Count:  req.Count,
	}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
