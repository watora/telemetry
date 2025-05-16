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
	fn(cfg)
	cfg.AppName = strings.ReplaceAll(cfg.AppName, "-", "_")
	cfg.HostName, _ = os.Hostname()
	if cfg.UseMetrics {
		metrics.Init()
	}
	if cfg.UseLogger {
		log.Init()
		trace.Init()
	}
}
