package metrics

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"os"
	"strings"
)

var meter api.Meter
var counterMap map[string]api.Int64Counter
var timerMap map[string]api.Int64Histogram
var gaugeMap map[string]api.Int64Gauge
var prefix string
var hostName string
var env string
var version string

// Init 初始化 通过收集器进行收集
func Init(appName string, _version string, _env string, endPoint string) {
	counterMap = make(map[string]api.Int64Counter)
	timerMap = make(map[string]api.Int64Histogram)
	gaugeMap = make(map[string]api.Int64Gauge)
	prefix = strings.ReplaceAll(appName, "-", "_")
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
	provider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exporter)))
	meter = provider.Meter(appName)
}
