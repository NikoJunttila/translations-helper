package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitTracer(ctx context.Context, endpoint string) (func(context.Context) error, error) {
	// 1. Create OTLP gRPC exporter
	// endpoint should be something like "alloy-host:4317"
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// 2. Define resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("translations"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 3. Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// 4. Return shutdown function
	return tp.Shutdown, nil
}

func StartMetrics(addr string) error {
	reg := prom.NewRegistry()

	// OTel -> Prometheus exporter writes into THIS registry.
	exp, err := prometheus.New(prometheus.WithRegisterer(reg))
	if err != nil {
		return err
	}

	// MeterProvider: exporter as Reader + runtime producer for scheduler histograms.
	otel.SetMeterProvider(
		sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(exp),
			sdkmetric.WithResource(
				resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceName("translations"),
				),
			),
		),
	)

	// Start Go runtime metric collection (GC / memstats etc).
	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(5 * time.Second),
	); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	go func() {
		slog.Info("metrics listening", "addr", addr, "path", "/metrics")
		_ = http.ListenAndServe(addr, mux)
	}()

	return nil
}
