package log

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/watora/telemetry/config"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type LogxWriter struct {
	logger    log.Logger
	callDepth int
}

func (w *LogxWriter) Alert(v any) {
	w.Emit(v, log.SeverityWarn)
}

func (w *LogxWriter) Close() error {
	return nil
}

func (w *LogxWriter) Debug(v any, fields ...logx.LogField) {
	w.Emit(v, log.SeverityDebug, fields...)
}

func (w *LogxWriter) Error(v any, fields ...logx.LogField) {
	w.Emit(v, log.SeverityError, fields...)
}

func (w *LogxWriter) Info(v any, fields ...logx.LogField) {
	w.Emit(v, log.SeverityInfo, fields...)
}

func (w *LogxWriter) Severe(v any) {
	w.Emit(v, log.SeverityFatal)
}

func (w *LogxWriter) Slow(v any, fields ...logx.LogField) {
	w.Emit(v, log.SeverityWarn, fields...)
}

func (w *LogxWriter) Stack(v any) {
	w.Emit(v, log.SeverityError)
}

func (w *LogxWriter) Stat(v any, fields ...logx.LogField) {
}

func (w *LogxWriter) Emit(v any, level log.Severity, fields ...logx.LogField) {
	r := log.Record{}
	r.SetTimestamp(time.Now())
	r.SetBody(log.StringValue(v.(string)))
	r.SetSeverity(level)
	r.SetSeverityText(strings.ToLower(level.String()))

	pcs := make([]uintptr, 20)
	runtime.Callers(w.callDepth, pcs)
	if pcs[0] > 0 {
		frames := runtime.CallersFrames(pcs)
		frame, more := frames.Next()
		r.AddAttributes(
			log.String(string(semconv.CodeFilepathKey), frame.File),
			log.Int(string(semconv.CodeLineNumberKey), frame.Line),
			log.String(string(semconv.CodeFunctionKey), frame.Function),
		)
		var stack string
		if level >= log.SeverityError {
			for more {
				stack += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
				frame, more = frames.Next()
			}
			// 捕捉调用栈
			r.AddAttributes(log.String(string(semconv.CodeStacktraceKey), stack))
		}
	}

	ctx := context.Background()
	for _, field := range fields {
		if c, isCtx := field.Value.(context.Context); isCtx {
			ctx = c
			continue
		}
		rv := reflect.ValueOf(field.Value)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			continue
		}
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		actualValue := rv.Interface()
		switch actualValue.(type) {
		case string:
			r.AddAttributes(log.String(field.Key, field.Value.(string)))
		case int:
			r.AddAttributes(log.Int(field.Key, field.Value.(int)))
		case int64:
			r.AddAttributes(log.Int64(field.Key, field.Value.(int64)))
		case int32:
			r.AddAttributes(log.Int(field.Key, int(field.Value.(int32))))
		case float64:
			r.AddAttributes(log.Float64(field.Key, field.Value.(float64)))
		case float32:
			r.AddAttributes(log.Float64(field.Key, float64(field.Value.(float32))))
		default:
			d, _ := json.Marshal(field.Value)
			r.AddAttributes(log.String(field.Key, string(d)))
		}
	}
	r.AddAttributes(log.String("env", config.Global.Env))
	w.logger.Emit(ctx, r)
}
