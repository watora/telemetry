package log

import (
	"context"
	"github.com/watora/telemetry/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithCtx 要带traceId的话需要先调这个
func WithCtx(logger *zap.Logger, ctx context.Context) *zap.Logger {
	if !config.Global.Init {
		return logger
	}
	return logger.With(zap.Any("context", ctx))
}

// WithCtxDefault 使用默认logger
func WithCtxDefault(ctx context.Context) *zap.Logger {
	if !config.Global.Init {
		return nil
	}
	return defaultLogger.With(zap.Any("context", ctx))
}

// 全局方法
func ctxLog(ctx context.Context, level zapcore.Level, message string, fields ...zap.Field) {
	if !config.Global.Init {
		return
	}
	// 传context可以自动取traceId
	fields = append(fields, zap.Any("context", ctx))
	defaultLogger.WithOptions(zap.AddCallerSkip(2)).Log(level, message, fields...)
}

func CtxInfo(ctx context.Context, message string, fields ...zap.Field) {
	ctxLog(ctx, zap.InfoLevel, message, fields...)
}

func CtxError(ctx context.Context, message string, fields ...zap.Field) {
	ctxLog(ctx, zap.ErrorLevel, message, fields...)
}

func CtxWarn(ctx context.Context, message string, fields ...zap.Field) {
	ctxLog(ctx, zap.WarnLevel, message, fields...)
}

func CtxDebug(ctx context.Context, message string, fields ...zap.Field) {
	ctxLog(ctx, zap.DebugLevel, message, fields...)
}
