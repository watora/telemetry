package log

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var defaultLogger *zap.Logger
var env string

// Init 直接导出otel日志到collector
func Init(appName string, version string, _env string, endPoint string) {
	hostName, _ := os.Hostname()
	// Create resource
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(version),
			semconv.ServiceInstanceID(hostName),
		))
	env = _env
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
	initDefaultLogger()
}

func newLoggerProvider(res *resource.Resource, endPoint string) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(endPoint))
	if err != nil {
		return nil, err
	}
	level := minsev.SeverityInfo
	if env == "local" {
		level = minsev.SeverityDebug
	}
	processor := minsev.NewLogProcessor(log.NewBatchProcessor(exporter), level)
	provider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)
	return provider, nil
}

// ZapBridge 使zap导出otel格式日志
func ZapBridge(logger *zap.Logger) *zap.Logger {
	otelCore := otelzap.NewCore("telemetry", otelzap.WithLoggerProvider(global.GetLoggerProvider()))
	return zap.New(zapcore.NewTee(
		otelCore,
		logger.Core(),
	), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).
		With(zap.String("env", env))
}

// 初始化默认logger 输出到collector和stderr
func initDefaultLogger() {
	level := zapcore.InfoLevel
	if env == "local" {
		level = zapcore.DebugLevel
	}
	otelCore := otelzap.NewCore("telemetry", otelzap.WithLoggerProvider(global.GetLoggerProvider()))
	stdCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stderr),
		level,
	)
	defaultLogger = zap.New(zapcore.NewTee(
		otelCore,
		stdCore,
	), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).
		With(zap.String("env", env))
}
