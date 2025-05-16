package telemetry

import (
	"github.com/watora/telemetry/config"
	"github.com/watora/telemetry/log"
	"github.com/watora/telemetry/metrics"
	"github.com/watora/telemetry/trace"
	"os"
	"strings"
)

var hostName string
var env string
var version string

func Init(fn func(cfg *config.Config)) {
	cfg := config.Global
	fn(cfg)
	cfg.AppName = strings.ReplaceAll(cfg.AppName, "-", "_")
	hostName, _ = os.Hostname()
	env = cfg.Env
	version = cfg.Version
	if cfg.UseMetrics {
		metrics.Init()
	}
	if cfg.UseLogger {
		log.Init()
		trace.Init()
	}
}
