package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/account/app"
	"modular_monolith/internal/account/app/command"
	"modular_monolith/internal/account/app/query"
	platformhttp "modular_monolith/internal/platform/httpserver"
)

type Handler struct {
	app *app.Application
}

func NewRouter(app *app.Application) *echo.Group {
	e := echo.New()
	h := &Handler{app: app}
	auth := platformhttp.RequireUserAuth()
	e.POST("/users", h.createUser)
	e.GET("/users/:user_id", h.getUser, auth)
	e.POST("/users/:user_id/addresses", h.addAddress, auth)
	e.GET("/users/:user_id/addresses", h.listAddresses, auth)
	e.PUT("/users/:user_id/addresses/:address_id", h.updateAddress, auth)
	e.DELETE("/users/:user_id/addresses/:address_id", h.deleteAddress, auth)
	return e.Group("")
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	auth := platformhttp.RequireUserAuth()
	e.POST("/users", h.createUser)
	e.GET("/users/:user_id", h.getUser, auth)
	e.POST("/users/:user_id/addresses", h.addAddress, auth)
	e.GET("/users/:user_id/addresses", h.listAddresses, auth)
	e.PUT("/users/:user_id/addresses/:address_id", h.updateAddress, auth)
	e.DELETE("/users/:user_id/addresses/:address_id", h.deleteAddress, auth)
}

func (h *Handler) createUser(c *echo.Context) error {
	var req struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	result, err := h.app.Commands.CreateUser.Handle(c.Request().Context(), command.CreateUser{Name: req.Name, Email: req.Email})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) getUser(c *echo.Context) error {
	userID, err := pathUserID(c)
	if err != nil {
		return err
	}
	result, err := h.app.Queries.GetUser.Handle(c.Request().Context(), query.GetUser{UserID: userID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) addAddress(c *echo.Context) error {
	userID, err := pathUserID(c)
	if err != nil {
		return err
	}
	var req struct {
		Receiver string `json:"receiver" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		City     string `json:"city" validate:"required"`
		Detail   string `json:"detail" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	result, err := h.app.Commands.AddAddress.Handle(c.Request().Context(), command.AddAddress{
		UserID: userID, Receiver: req.Receiver, Phone: req.Phone, City: req.City, Detail: req.Detail,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) listAddresses(c *echo.Context) error {
	userID, err := pathUserID(c)
	if err != nil {
		return err
	}
	result, err := h.app.Queries.ListAddresses.Handle(c.Request().Context(), query.ListAddresses{UserID: userID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) updateAddress(c *echo.Context) error {
	userID, err := pathUserID(c)
	if err != nil {
		return err
	}
	var req struct {
		Receiver string `json:"receiver" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		City     string `json:"city" validate:"required"`
		Detail   string `json:"detail" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	if err := h.app.Commands.UpdateAddress.Handle(c.Request().Context(), command.UpdateAddress{
		UserID: userID, AddressID: c.Param("address_id"), Receiver: req.Receiver, Phone: req.Phone, City: req.City, Detail: req.Detail,
	}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) deleteAddress(c *echo.Context) error {
	userID, err := pathUserID(c)
	if err != nil {
		return err
	}
	if err := h.app.Commands.DeleteAddress.Handle(c.Request().Context(), command.DeleteAddress{UserID: userID, AddressID: c.Param("address_id")}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func pathUserID(c *echo.Context) (string, error) {
	userID, ok := platformhttp.UserUUID(c)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	if pathUserID := c.Param("user_id"); pathUserID != userID {
		return "", echo.NewHTTPError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	return userID, nil
}
