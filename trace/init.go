package trace

import (
	"fmt"
	"github.com/go-logr/stdr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func InitTracer(appName string) {
	stdr.SetVerbosity(5)

	exp, err := stdouttrace.New(stdouttrace.WithWriter(&noopWriter{}))
	if err != nil {
		panic(fmt.Sprintf("init tracer err: %v", err))
	}
	processor := sdktrace.NewBatchSpanProcessor(exp)
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(processor),
	)
	otel.SetTracerProvider(provider)

	Tracer = provider.Tracer(appName)
}

type noopWriter struct {
}

func (w *noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
