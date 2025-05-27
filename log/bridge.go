package log

import (
	"github.com/watora/telemetry/config"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogxBridge 使logx导出otel日志
func LogxBridge() {
	if !config.Global.Init {
		return
	}
	logProvider := global.GetLoggerProvider()
	logx.AddWriter(&LogxWriter{
		logger:    logProvider.Logger("telemetry_logx"),
		callDepth: 5,
	})
}

// ZapBridge 使zap导出otel日志
func ZapBridge(logger *zap.Logger) *zap.Logger {
	if !config.Global.Init {
		return logger
	}
	otelCore := otelzap.NewCore("telemetry_zap", otelzap.WithLoggerProvider(global.GetLoggerProvider()))
	return zap.New(zapcore.NewTee(
		otelCore,
		logger.Core(),
	), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).
		With(zap.String("env", config.Global.Env))
}
