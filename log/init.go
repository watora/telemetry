package log

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
)

var bridgeLogger map[string]*zap.Logger

const defaultLoggerName = "_default"

// Init 直接导出otel日志到collector
func Init(appName string, version string, endPoint string) {
	bridgeLogger = make(map[string]*zap.Logger)
	// Create resource
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(version),
		))
	if err != nil {
		panic(fmt.Sprintf("init resource: %v", err))
	}
	// 新建provider
	loggerProvider, err := newLoggerProvider(res, endPoint)
	if err != nil {
		panic(fmt.Sprintf("init provider: %v", err))
	}
	// provider注册到全局
	global.SetLoggerProvider(loggerProvider)
	// init default logger
	ZapBridge(defaultLoggerName)
}

func newLoggerProvider(res *resource.Resource, endPoint string) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(endPoint))
	if err != nil {
		return nil, err
	}
	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)
	return provider, nil
}

// ZapBridge 使zap导出otel格式日志
func ZapBridge(name string) *zap.Logger {
	if logger, ok := bridgeLogger[name]; ok {
		return logger
	}
	bridgeLogger[name] = zap.New(otelzap.NewCore(name, otelzap.WithLoggerProvider(global.GetLoggerProvider())))
	return bridgeLogger[name]
}
