package log

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"reflect"
	"runtime"
	"runtime/debug"
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

	pc, _, _, ok := runtime.Caller(w.callDepth)
	if ok {
		frames := runtime.CallersFrames([]uintptr{pc})
		frame, _ := frames.Next()
		r.AddAttributes(
			log.String(string(semconv.CodeFilepathKey), frame.File),
			log.Int(string(semconv.CodeLineNumberKey), frame.Line),
			log.String(string(semconv.CodeFunctionKey), frame.Function),
		)
	}

	if level >= log.SeverityError {
		r.AddAttributes(log.String(string(semconv.CodeStacktraceKey), string(debug.Stack())))
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
	w.logger.Emit(ctx, r)
}
