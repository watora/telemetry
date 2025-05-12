package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Log(ctx context.Context, level zapcore.Level, message string, fields ...zap.Field) {
	traceId := ctx.Value("traceId")
	if traceId != nil {
		fields = append(fields, zap.String("traceId", traceId.(string)))
	}
	logger := bridgeLogger[defaultLoggerName]
	logger.Log(level, message, fields...)
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zap.InfoLevel, message, fields...)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zap.ErrorLevel, message, fields...)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zap.WarnLevel, message, fields...)
}

func Debug(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zap.DebugLevel, message, fields...)
}
