package metrics

import (
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetrics(addr string) error {
	reg := prom.NewRegistry()

	// OTel -> Prometheus exporter writes into THIS registry.
	exp, err := prometheus.New(prometheus.WithRegisterer(reg))
	if err != nil {
		return err
	}

	// MeterProvider: exporter as Reader + runtime producer for scheduler histograms.
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exp),
	)
	otel.SetMeterProvider(provider)

	// Start Go runtime metric collection (GC / memstats etc).
	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(5 * time.Second),
	); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	go func() {
		log.Printf("metrics listening on %s/metrics", addr)
		_ = http.ListenAndServe(addr, mux)
	}()

	return nil
}
