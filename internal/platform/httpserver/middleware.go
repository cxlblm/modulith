package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"modular_monolith/internal/platform/logctx"
)

func requestContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			ctx := logctx.WithRequestID(req.Context(), requestID)
			c.SetRequest(req.WithContext(ctx))
			defer func() {
				cleanupMultipartForm(c.Request())
			}()

			return next(c)
		}
	}
}

func cleanupMultipartForm(req *http.Request) {
	if req == nil || req.MultipartForm == nil {
		return
	}
	_ = req.MultipartForm.RemoveAll()
}

func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogContentLength: true,
		LogResponseSize:  true,
		HandleError:      true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			ctx := c.Request().Context()
			if userUUID, ok := UserUUID(c); ok {
				ctx = logctx.WithUserUUID(ctx, userUUID)
			}

			attrs := requestLogAttrs(v)
			if v.Error == nil {
				c.Logger().LogAttrs(ctx, slog.LevelInfo, "REQUEST", attrs...)
				return nil
			}

			attrs = appendErrorAttrs(attrs, v.Error)
			c.Logger().LogAttrs(ctx, slog.LevelError, "REQUEST_ERROR", attrs...)
			return nil
		},
	})
}

func requestLogAttrs(v middleware.RequestLoggerValues) []slog.Attr {
	return []slog.Attr{
		slog.String("method", v.Method),
		slog.String("uri", v.URI),
		slog.Int("status", v.Status),
		slog.String("latency", v.Latency.String()),
		slog.String("host", v.Host),
		slog.String("bytes_in", v.ContentLength),
		slog.Int64("bytes_out", v.ResponseSize),
		slog.String("user_agent", v.UserAgent),
		slog.String("remote_ip", v.RemoteIP),
		slog.String("request_id", v.RequestID),
	}
}

func appendErrorAttrs(attrs []slog.Attr, err error) []slog.Attr {
	return append(attrs, slog.Any("error", err))
}

func recoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("panic recovered: %v", recovered)
				}
			}()
			return next(c)
		}
	}
}
