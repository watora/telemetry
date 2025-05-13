package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CtxLog(ctx context.Context, level zapcore.Level, message string, fields ...zap.Field) {
	// 传context可以自动取traceId
	fields = append(fields, zap.Any("context", ctx))
	fields = append(fields, zap.String("env", env))
	defaultLogger.Log(level, message, fields...)
}

func CtxInfo(ctx context.Context, message string, fields ...zap.Field) {
	CtxLog(ctx, zap.InfoLevel, message, fields...)
}

func CtxError(ctx context.Context, message string, fields ...zap.Field) {
	CtxLog(ctx, zap.ErrorLevel, message, fields...)
}

func CtxWarn(ctx context.Context, message string, fields ...zap.Field) {
	CtxLog(ctx, zap.WarnLevel, message, fields...)
}

func CtxDebug(ctx context.Context, message string, fields ...zap.Field) {
	CtxLog(ctx, zap.DebugLevel, message, fields...)
}
