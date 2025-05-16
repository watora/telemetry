package trace

import (
	"fmt"
	"github.com/go-logr/stdr"
	"github.com/watora/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

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

	Tracer = provider.Tracer(config.Global.AppName)
}

type noopWriter struct {
}

func (w *noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
