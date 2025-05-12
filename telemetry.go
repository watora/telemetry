package telemetry

import (
	"gitlab.ldcu5z28my.com/backend/telemetry/log"
	"gitlab.ldcu5z28my.com/backend/telemetry/metrics"
)

type Config struct {
	AppName           string
	Version           string
	EndPoint          string
	CollectorEndPoint string
	UseMetrics        bool
	UseLogger         bool
}

func Init(fn func(cfg *Config)) {
	cfg := &Config{}
	fn(cfg)
	if cfg.UseMetrics {
		metrics.Init(cfg.AppName, cfg.Version, cfg.EndPoint)
	}
	if cfg.UseLogger {
		log.Init(cfg.AppName, cfg.Version, cfg.EndPoint)
	}
}
