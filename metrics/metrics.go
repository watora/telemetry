package metrics

import (
	"context"
	"fmt"
	"github.com/watora/telemetry/config"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"golang.org/x/sync/singleflight"
)

var g singleflight.Group

func fillCommonAttr(attr []attribute.KeyValue) []attribute.KeyValue {
	keyMap := make(map[string]struct{}, len(attr))
	for _, item := range attr {
		keyMap[string(item.Key)] = struct{}{}
	}
	if _, ok := keyMap["env"]; !ok {
		attr = append(attr, attribute.String("env", config.Global.Env))
	}
	if _, ok := keyMap["version"]; !ok {
		attr = append(attr, attribute.String("version", config.Global.Version))
	}
	if _, ok := keyMap["host"]; !ok {
		attr = append(attr, attribute.String("host", config.Global.HostName))
	}
	if _, ok := keyMap["service.name"]; !ok {
		attr = append(attr, attribute.String("service.name", config.Global.AppName))
	}
	return attr
}

// EmitCount 计量次数
func EmitCount(ctx context.Context, name string, incr int64, attr ...attribute.KeyValue) {
	if !config.Global.Init {
		return
	}
	counter, err := getCounter(name)
	if err != nil {
		return
	}
	attr = fillCommonAttr(attr)
	counter.Add(ctx, incr, api.WithAttributes(attr...))
}

func getCounter(name string) (api.Int64Counter, error) {
	counter, err, _ := g.Do(fmt.Sprintf("counter_init_%v", name), func() (interface{}, error) {
		counter, ok := counterMap[name]
		if !ok {
			var err error
			counter, err = meter.Int64Counter(fmt.Sprintf("%v_%v", config.Global.AppName, name))
			if err != nil {
				return nil, err
			}
			counterMap[name] = counter
		}
		return counter, nil
	})
	if err != nil {
		return nil, err
	}
	return counter.(api.Int64Counter), nil
}

// EmitTime 计量时间
func EmitTime(ctx context.Context, name string, ms int64, attr ...attribute.KeyValue) {
	if !config.Global.Init {
		return
	}
	timer, err := getTimer(name)
	if err != nil {
		return
	}
	attr = fillCommonAttr(attr)
	timer.Record(ctx, ms, api.WithAttributes(attr...))
}

func getTimer(name string) (api.Int64Histogram, error) {
	timer, err, _ := g.Do(fmt.Sprintf("timer_init_%v", name), func() (interface{}, error) {
		timer, ok := timerMap[name]
		if !ok {
			var err error
			timer, err = meter.Int64Histogram(fmt.Sprintf("%v_%v", config.Global.AppName, name))
			if err != nil {
				return nil, err
			}
			timerMap[name] = timer
		}
		return timer, nil
	})
	if err != nil {
		return nil, err
	}
	return timer.(api.Int64Histogram), nil
}

// EmitGauge 记录当前值
func EmitGauge(ctx context.Context, name string, n int64, attr ...attribute.KeyValue) {
	if !config.Global.Init {
		return
	}
	gauge, err := getGauge(name)
	if err != nil {
		return
	}
	attr = fillCommonAttr(attr)
	gauge.Record(ctx, n, api.WithAttributes(attr...))
}

func getGauge(name string) (api.Int64Gauge, error) {
	gauge, err, _ := g.Do(fmt.Sprintf("gauge_init_%v", name), func() (interface{}, error) {
		gauge, ok := gaugeMap[name]
		if !ok {
			var err error
			gauge, err = meter.Int64Gauge(fmt.Sprintf("%v_%v", config.Global.AppName, name))
			if err != nil {
				return nil, err
			}
			gaugeMap[name] = gauge
		}
		return gauge, nil
	})
	if err != nil {
		return nil, err
	}
	return gauge.(api.Int64Gauge), nil
}
