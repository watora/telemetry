package telemetry

import (
	"github.com/watora/telemetry/config"
	"github.com/watora/telemetry/log"
	"github.com/watora/telemetry/metrics"
	"github.com/watora/telemetry/trace"
	"os"
	"strings"
)

func Init(fn func(cfg *config.Config)) {
	cfg := config.Global
	cfg.HostName, _ = os.Hostname()
	fn(cfg)
	cfg.AppName = strings.ReplaceAll(cfg.AppName, "-", "_")
	if cfg.UseMetrics {
		metrics.Init()
	}
	if cfg.UseLogger {
		log.Init()
		trace.Init()
	}
}
