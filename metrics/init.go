package metrics

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"os"
)

var meter api.Meter
var counterMap map[string]api.Int64Counter
var timerMap map[string]api.Int64Histogram
var gaugeMap map[string]api.Int64Gauge
var appName string
var hostName string
var env string
var version string

// Init 初始化 通过收集器进行收集
func Init(_appName string, _version string, _env string, endPoint string) {
	counterMap = make(map[string]api.Int64Counter)
	timerMap = make(map[string]api.Int64Histogram)
	gaugeMap = make(map[string]api.Int64Gauge)
	appName = _appName
	hostName, _ = os.Hostname()
	env = _env
	version = _version
	exporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endPoint),
	)
	if err != nil {
		panic(fmt.Sprintf("init exporter: %v", err))
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Kind: metric.InstrumentKindHistogram,
			},
			metric.Stream{
				// 调整桶的精度
				Aggregation: metric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{0, 1, 2, 5, 10, 20, 30, 50, 75, 100, 250, 500, 1000, 2500, 5000, 10000},
				},
			},
		)),
	)
	meter = provider.Meter(_appName)
}
