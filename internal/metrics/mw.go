package metrics

import (
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"

	"github.com/labstack/echo/v4"
)

func SetupEcho(e *echo.Echo) {
	e.Use(otelecho.Middleware("api-server",
		otelecho.WithSkipper(func(c echo.Context) bool {
			// don’t count scrapes / health
			return c.Path() == "/metrics" || c.Path() == "/health"
		}),
		otelecho.WithEchoMetricAttributeFn(func(c echo.Context) []attribute.KeyValue {
			return []attribute.KeyValue{
				attribute.String("echo.route", c.Path()),
			}
		}),
	))
}
