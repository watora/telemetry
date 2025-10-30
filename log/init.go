package log

import (
	"context"
	"fmt"
	"github.com/watora/telemetry/config"
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

// Init 直接导出otel日志到collector
func Init() {
	res, err := buildResource(config.Global.AppName, config.Global.Version)
	if err != nil {
		panic(fmt.Sprintf("build resource error: %v", err))
	}
	// 新建provider
	loggerProvider, err := newLoggerProvider(res, config.Global.LogEndPoint)
	if err != nil {
		panic(fmt.Sprintf("init provider: %v", err))
	}
	// provider注册到全局
	global.SetLoggerProvider(loggerProvider)
	// init default logger
	defaultLogger = initLogger(loggerProvider)
}

// 新建resource
func buildResource(appName string, version string) (*resource.Resource, error) {
	hostName, _ := os.Hostname()
	// 新建resource
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(version),
			semconv.ServiceInstanceID(hostName),
		))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func newLoggerProvider(res *resource.Resource, endPoint string) (*log.LoggerProvider, error) {
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(endPoint))
	if err != nil {
		return nil, err
	}
	level := minsev.SeverityInfo
	if config.Global.Env == "local" {
		level = minsev.SeverityDebug
	}
	processor := minsev.NewLogProcessor(log.NewBatchProcessor(exporter), level)
	provider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)
	return provider, nil
}

// 初始化默认logger 输出到collector和stderr
func initLogger(loggerProvider *log.LoggerProvider) *zap.Logger {
	level := zapcore.InfoLevel
	if config.Global.Env == "local" {
		level = zapcore.DebugLevel
	}
	otelCore := otelzap.NewCore("telemetry_zap", otelzap.WithLoggerProvider(loggerProvider))
	stdCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stderr),
		level,
	)
	return zap.New(zapcore.NewTee(
		otelCore,
		stdCore,
	), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).
		With(zap.String("env", config.Global.Env))
}

// GetLogger 生成指定服务的logger
func GetLogger(appName string, version string) (*zap.Logger, error) {
	res, err := buildResource(appName, version)
	if err != nil {
		return nil, err
	}
	provider, err := newLoggerProvider(res, config.Global.LogEndPoint)
	if err != nil {
		return nil, err
	}
	return initLogger(provider), nil
}
