package metrics

import (
	"context"
	"fmt"
	"github.com/watora/telemetry/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"sync"
	"time"
)

var meter api.Meter
var counterMap sync.Map // api.Int64Counter
var timerMap sync.Map   // api.Int64Histogram
var gaugeMap sync.Map   // api.Int64Gauge

// Init 初始化 通过收集器进行收集
func Init() {
	exporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(config.Global.MetricsEndPoint),
	)
	if err != nil {
		panic(fmt.Sprintf("init exporter: %v", err))
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(14*time.Second))), //14s导出一次数据
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
	meter = provider.Meter(config.Global.AppName)
}
