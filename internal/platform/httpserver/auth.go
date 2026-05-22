package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	HeaderUserUUID  = "X-User-UUID"
	authUserUUIDKey = "auth_user_uuid"
)

func RequireUserAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userUUID := c.Request().Header.Get(HeaderUserUUID)
			if userUUID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			c.Set(authUserUUIDKey, userUUID)

			return next(c)
		}
	}
}

func UserUUID(c *echo.Context) (string, bool) {
	userUUID, ok := c.Get(authUserUUIDKey).(string)
	return userUUID, ok && userUUID != ""
}
