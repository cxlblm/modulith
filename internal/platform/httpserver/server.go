package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"modular_monolith/internal/platform/validation"
)

type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	cfg       Config
	echo      *echo.Echo
	http      *http.Server
	startedCh chan error
}

type HealthResponse struct {
	Status string `json:"status"`
}

func New(cfg Config, logger *slog.Logger) *Server {
	e := echo.New()
	e.Logger = logger
	e.Validator = validation.New()
	e.HTTPErrorHandler = newHTTPErrorHandler(logger)
	e.Use(middleware.RequestID())
	e.Use(requestContextMiddleware())
	e.Use(requestLogger())
	e.Use(recoverMiddleware())

	e.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	})

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      e,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		cfg:       cfg,
		echo:      e,
		http:      httpServer,
		startedCh: make(chan error, 1),
	}
}

func (s *Server) Echo() *echo.Echo {
	return s.echo
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		s.startedCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-s.startedCh:
		return normalizeServerError(err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return normalizeServerError(<-s.startedCh)
	}
}

func normalizeServerError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("run http server: %w", err)
}
