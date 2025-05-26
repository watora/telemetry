package telemetry

import (
	"context"
	"fmt"
	"github.com/aceld/zinx/ziface"
	redisV6 "github.com/go-redis/redis"
	"github.com/go-redis/redis/v8"
	"github.com/watora/telemetry/config"
	"github.com/watora/telemetry/metrics"
	"github.com/watora/telemetry/trace"
	"github.com/zeromicro/go-zero/rest"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// InstrumentGORM 仪表化gorm
func InstrumentGORM(db *gorm.DB) {
	if !config.Global.Init {
		return
	}
	before := func(db *gorm.DB) {
		start := time.Now().UnixMilli()
		db.Set("metrics.start", start)
	}
	after := func(command string) func(db *gorm.DB) {
		return func(db *gorm.DB) {
			if db.Statement == nil || db.Statement.Schema == nil {
				return
			}
			end := time.Now().UnixMilli()
			if v, ok := db.Get("metrics.start"); ok {
				start := v.(int64)
				ctx := context.Background()
				if db.Statement.Context != nil {
					ctx = db.Statement.Context
				}
				attr := []attribute.KeyValue{
					{Key: "table", Value: attribute.StringValue(db.Statement.Table)},
					{Key: "success", Value: attribute.BoolValue(db.Statement.Error == nil)},
					{Key: "command", Value: attribute.StringValue(command)},
					{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
					{Key: "env", Value: attribute.StringValue(config.Global.Env)},
					{Key: "driver", Value: attribute.StringValue(db.Dialector.Name())},
					{Key: "version", Value: attribute.StringValue(config.Global.Version)},
				}
				metrics.EmitTime(ctx, "gorm_duration", end-start, attr...)
				metrics.EmitCount(ctx, "gorm_count", 1, attr...)
			}
		}
	}
	// register callback
	_ = db.Callback().Create().Before("*").Register("metrics.create.before", before)
	_ = db.Callback().Create().After("*").Register("metrics.create.after", after("create"))
	_ = db.Callback().Query().Before("*").Register("metrics.query.before", before)
	_ = db.Callback().Query().After("*").Register("metrics.query.after", after("query"))
	_ = db.Callback().Update().Before("*").Register("metrics.update.before", before)
	_ = db.Callback().Update().After("*").Register("metrics.update.after", after("update"))
	_ = db.Callback().Delete().Before("*").Register("metrics.delete.before", before)
	_ = db.Callback().Delete().After("*").Register("metrics.delete.after", after("delete"))
}

// InstrumentGoZero 仪表化gozero
func InstrumentGoZero(server *rest.Server) {
	if !config.Global.Init {
		return
	}
	//add middleware
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UnixMilli()
			wl := &metrics.WriteLogger{ResponseWriter: w}
			newCtx, span := trace.Tracer.Start(r.Context(), "http_request")
			defer span.End()
			r = r.WithContext(newCtx)
			next(wl, r)
			if strings.HasPrefix(r.URL.Path, "/swagger") ||
				strings.HasPrefix(r.URL.Path, "/metrics") {
				return
			}
			end := time.Now().UnixMilli()
			attr := []attribute.KeyValue{
				{Key: "path", Value: attribute.StringValue(r.URL.Path)},
				{Key: "method", Value: attribute.StringValue(r.Method)},
				{Key: "status_code", Value: attribute.IntValue(wl.StatusCode)},
				{Key: "success", Value: attribute.BoolValue(wl.StatusCode < 400)},
				{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
				{Key: "env", Value: attribute.StringValue(config.Global.Env)},
				{Key: "version", Value: attribute.StringValue(config.Global.Version)},
			}
			metrics.EmitTime(newCtx, "http_duration", end-start, attr...)
			metrics.EmitCount(newCtx, "http_count", 1, attr...)
		}
	})
}

// InstrumentZinx 仪表化zinx
func InstrumentZinx(server ziface.IServer) {
	if !config.Global.Init {
		return
	}
	server.Use(func(request ziface.IRequest) {
		start := time.Now().UnixMilli()
		request.RouterSlicesNext()
		end := time.Now().UnixMilli()
		attr := []attribute.KeyValue{
			{Key: "msg_id", Value: attribute.StringValue(fmt.Sprintf("%v", request.GetMsgID()))},
			{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
			{Key: "env", Value: attribute.StringValue(config.Global.Env)},
			{Key: "version", Value: attribute.StringValue(config.Global.Version)},
		}
		ctx := context.Background()
		metrics.EmitTime(ctx, "zinx_duration", end-start, attr...)
		metrics.EmitCount(ctx, "zinx_count", 1, attr...)
	})
	// 记录连接数
	var connected int64
	server.SetOnConnStart(func(connection ziface.IConnection) {
		attr := []attribute.KeyValue{
			{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
			{Key: "env", Value: attribute.StringValue(config.Global.Env)},
			{Key: "version", Value: attribute.StringValue(config.Global.Version)},
		}
		atomic.AddInt64(&connected, 1)
		metrics.EmitGauge(connection.Context(), "zinx_live", atomic.LoadInt64(&connected), attr...)
	})
	server.SetOnConnStop(func(connection ziface.IConnection) {
		attr := []attribute.KeyValue{
			{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
			{Key: "env", Value: attribute.StringValue(config.Global.Env)},
			{Key: "version", Value: attribute.StringValue(config.Global.Version)},
		}
		atomic.AddInt64(&connected, -1)
		metrics.EmitGauge(connection.Context(), "zinx_live", atomic.LoadInt64(&connected), attr...)
	})
}

// InstrumentRedisV8 仪表化redis，必须是v8的连接
func InstrumentRedisV8(client *redis.ClusterClient) {
	if !config.Global.Init {
		return
	}
	client.AddHook(&metrics.RedisHook{
		MeterBefore: func(ctx context.Context) context.Context {
			start := time.Now().UnixMilli()
			if ctx == nil {
				ctx = context.Background()
			}
			return context.WithValue(ctx, "metrics.before", start)
		},
		MeterAfter: func(ctx context.Context, cmd string) {
			if ctx == nil {
				return
			}
			start := ctx.Value("metrics.before")
			if start == nil {
				return
			}
			end := time.Now().UnixMilli()
			attr := []attribute.KeyValue{
				{Key: "cmd", Value: attribute.StringValue(cmd)},
				{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
				{Key: "env", Value: attribute.StringValue(config.Global.Env)},
				{Key: "version", Value: attribute.StringValue(config.Global.Version)},
			}
			metrics.EmitTime(ctx, "redis_v8_duration", end-start.(int64), attr...)
			metrics.EmitCount(ctx, "redis_v8_count", 1, attr...)
		},
	})
}

// InstrumentRedis 仪表化redis，必须是v6的连接
func InstrumentRedis(client *redisV6.ClusterClient) {
	if !config.Global.Init {
		return
	}
	// 替换process
	client.WrapProcess(func(oldProcess func(redisV6.Cmder) error) func(redisV6.Cmder) error {
		return func(cmder redisV6.Cmder) error {
			start := time.Now().UnixMilli()
			err := oldProcess(cmder)
			attr := []attribute.KeyValue{
				{Key: "cmd", Value: attribute.StringValue(cmder.Name())},
				{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
				{Key: "env", Value: attribute.StringValue(config.Global.Env)},
				{Key: "version", Value: attribute.StringValue(config.Global.Version)},
			}
			ctx := context.Background()
			metrics.EmitTime(ctx, "redis_v6_duration", time.Now().UnixMilli()-start, attr...)
			metrics.EmitCount(ctx, "redis_v6_count", 1, attr...)
			return err
		}
	})
	// pipeline
	client.WrapProcessPipeline(func(oldProcess func([]redisV6.Cmder) error) func([]redisV6.Cmder) error {
		return func(cmders []redisV6.Cmder) error {
			start := time.Now().UnixMilli()
			err := oldProcess(cmders)
			attr := []attribute.KeyValue{
				{Key: "cmd", Value: attribute.StringValue("pipeline")},
				{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
				{Key: "env", Value: attribute.StringValue(config.Global.Env)},
				{Key: "version", Value: attribute.StringValue(config.Global.Version)},
			}
			ctx := context.Background()
			metrics.EmitTime(ctx, "redis_v6_duration", time.Now().UnixMilli()-start, attr...)
			metrics.EmitCount(ctx, "redis_v6_count", 1, attr...)
			return err
		}
	})
}

// InstrumentMongo 仪表化mongo
func InstrumentMongo(options *options.ClientOptions) *options.ClientOptions {
	if !config.Global.Init {
		return options
	}
	emit := func(ctx context.Context, command string, success bool, duration time.Duration) {
		attr := []attribute.KeyValue{
			{Key: "cmd", Value: attribute.StringValue(command)},
			{Key: "host", Value: attribute.StringValue(config.Global.HostName)},
			{Key: "env", Value: attribute.StringValue(config.Global.Env)},
			{Key: "version", Value: attribute.StringValue(config.Global.Version)},
			{Key: "success", Value: attribute.BoolValue(success)},
		}
		metrics.EmitTime(ctx, "mongo_duration", duration.Milliseconds(), attr...)
		metrics.EmitCount(ctx, "mongo_count", 1, attr...)
	}
	monitor := &event.CommandMonitor{
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
			emit(ctx, succeededEvent.CommandName, true, succeededEvent.Duration)
		},
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
			emit(ctx, failedEvent.CommandName, true, failedEvent.Duration)
		},
	}
	return options.SetMonitor(monitor)
}
