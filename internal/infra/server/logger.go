package server

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func requestLogger() echo.MiddlewareFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("remote_ip", v.RemoteIP),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "request", attrs...)
			} else {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "request", attrs...)
			}
			return nil
		},
	})
}
