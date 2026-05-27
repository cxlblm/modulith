package httpserver

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/platform/validation"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string                  `json:"code"`
	Reason  string                  `json:"reason,omitempty"`
	Message string                  `json:"message"`
	Fields  []validation.FieldError `json:"fields,omitempty"`
}

type statusCodeError interface {
	error
	StatusCode() int
}

type notFoundError interface {
	error
	NotFound()
}

type invalidError interface {
	error
	Invalid()
}

type conflictError interface {
	error
	Conflict()
}

type forbiddenError interface {
	error
	Forbidden()
}

type reasonedError interface {
	error
	Reason() string
}

func newHTTPErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if responseCommitted(c) {
			return
		}

		status, body := errorResponse(err)
		if status >= http.StatusInternalServerError {
			logger.ErrorContext(
				c.Request().Context(),
				"http request failed",
				"error",
				err,
				"method",
				c.Request().Method,
				"path",
				c.Request().URL.Path,
			)
		}

		if writeErr := c.JSON(status, body); writeErr != nil {
			logger.ErrorContext(c.Request().Context(), "write error response", "error", writeErr)
		}
	}
}

func errorResponse(err error) (int, ErrorResponse) {
	if validationErr, ok := errors.AsType[validation.Error](err); ok {
		return http.StatusBadRequest, ErrorResponse{
			Error: ErrorBody{
				Code:    "validation_failed",
				Message: "validation failed",
				Fields:  validationErr.Fields,
			},
		}
	}

	if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		return httpErr.Code, ErrorResponse{
			Error: ErrorBody{
				Code:    statusCodeName(httpErr.Code),
				Message: httpErr.Message,
			},
		}
	}

	if _, ok := errors.AsType[notFoundError](err); ok {
		return appErrorResponse(http.StatusNotFound, errorReason(err))
	}

	if _, ok := errors.AsType[invalidError](err); ok {
		return appErrorResponse(http.StatusBadRequest, errorReason(err))
	}

	if _, ok := errors.AsType[conflictError](err); ok {
		return appErrorResponse(http.StatusConflict, errorReason(err))
	}

	if _, ok := errors.AsType[forbiddenError](err); ok {
		return appErrorResponse(http.StatusForbidden, errorReason(err))
	}

	if statusErr, ok := errors.AsType[statusCodeError](err); ok {
		status := statusErr.StatusCode()
		if status != 0 {
			return status, ErrorResponse{
				Error: ErrorBody{
					Code:    statusCodeName(status),
					Message: http.StatusText(status),
				},
			}
		}
	}

	return http.StatusInternalServerError, ErrorResponse{
		Error: ErrorBody{
			Code:    "internal_error",
			Message: "internal server error",
		},
	}
}

func appErrorResponse(status int, reason string) (int, ErrorResponse) {
	return status, ErrorResponse{
		Error: ErrorBody{
			Code:    statusCodeName(status),
			Reason:  reason,
			Message: http.StatusText(status),
		},
	}
}

func errorReason(err error) string {
	if reasoned, ok := errors.AsType[reasonedError](err); ok {
		return reasoned.Reason()
	}
	return ""
}

func statusCodeName(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	default:
		if status >= http.StatusInternalServerError {
			return "internal_error"
		}
		return "http_error"
	}
}

func responseCommitted(c *echo.Context) bool {
	response, err := echo.UnwrapResponse(c.Response())
	return err == nil && response.Committed
}
