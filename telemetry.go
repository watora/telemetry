package telemetry

import (
	"github.com/watora/telemetry/log"
	"github.com/watora/telemetry/metrics"
	"github.com/watora/telemetry/trace"
	"os"
	"strings"
)

type Config struct {
	AppName         string
	Version         string
	MetricsEndPoint string
	LogEndPoint     string
	UseMetrics      bool
	UseLogger       bool
	Env             string
}

func Init(fn func(cfg *Config)) {
	cfg := &Config{}
	fn(cfg)
	cfg.AppName = strings.ReplaceAll(cfg.AppName, "-", "_")
	_ = os.Setenv("OTEL_SERVICE_NAME", cfg.AppName)
	if cfg.UseMetrics {
		metrics.Init(cfg.AppName, cfg.Version, cfg.Env, cfg.MetricsEndPoint)
	}
	if cfg.UseLogger {
		log.Init(cfg.AppName, cfg.Version, cfg.Env, cfg.LogEndPoint)
		trace.InitTracer(cfg.AppName)
	}
}
