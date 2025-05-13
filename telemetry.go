package telemetry

import (
	"gitlab.ldcu5z28my.com/backend/telemetry/log"
	"gitlab.ldcu5z28my.com/backend/telemetry/metrics"
)

type Config struct {
	AppName    string
	Version    string
	EndPoint   string
	UseMetrics bool
	UseLogger  bool
	Env        string
}

func Init(fn func(cfg *Config)) {
	cfg := &Config{}
	fn(cfg)
	if cfg.UseMetrics {
		metrics.Init(cfg.AppName, cfg.Version, cfg.Env, cfg.EndPoint)
	}
	if cfg.UseLogger {
		log.Init(cfg.AppName, cfg.Version, cfg.Env, cfg.EndPoint)
	}
}
