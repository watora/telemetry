package trace

import (
	"context"
	"fmt"
	"github.com/go-logr/stdr"
	"github.com/watora/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func Init() {
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

	tracer = provider.Tracer(config.Global.AppName)
}

func StartTrace(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

type noopWriter struct {
}

func (w *noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
