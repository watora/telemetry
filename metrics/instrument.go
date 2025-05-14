package metrics

import (
	"context"
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/rest"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// InstrumentGORM 仪表化gorm
func InstrumentGORM(db *gorm.DB) {
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
					{Key: "__table", Value: attribute.StringValue(db.Statement.Table)},
					{Key: "__success", Value: attribute.BoolValue(db.Statement.Error == nil)},
					{Key: "__command", Value: attribute.StringValue(command)},
					{Key: "__host", Value: attribute.StringValue(hostName)},
					{Key: "__env", Value: attribute.StringValue(env)},
					{Key: "__driver", Value: attribute.StringValue(db.Dialector.Name())},
					{Key: "__version", Value: attribute.StringValue(version)},
				}
				EmitTime(ctx, "gorm", end-start, attr...)
				EmitCount(ctx, "gorm", 1, attr...)
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
	//add middleware
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UnixMilli()
			wl := &writeLogger{ResponseWriter: w}
			next(wl, r)
			if strings.HasPrefix(r.URL.Path, "/swagger") ||
				strings.HasPrefix(r.URL.Path, "/metrics") {
				return
			}
			end := time.Now().UnixMilli()
			attr := []attribute.KeyValue{
				{Key: "__path", Value: attribute.StringValue(r.URL.Path)},
				{Key: "__method", Value: attribute.StringValue(r.Method)},
				{Key: "__status_code", Value: attribute.IntValue(wl.statusCode)},
				{Key: "__success", Value: attribute.BoolValue(wl.statusCode < 400)},
				{Key: "__host", Value: attribute.StringValue(hostName)},
				{Key: "__env", Value: attribute.StringValue(env)},
				{Key: "__version", Value: attribute.StringValue(version)},
			}
			EmitTime(r.Context(), "http", end-start, attr...)
			EmitCount(r.Context(), "http", 1, attr...)
		}
	})
}

// InstrumentZinx 仪表化zinx
func InstrumentZinx(server ziface.IServer) {
	server.Use(func(request ziface.IRequest) {
		start := time.Now().UnixMilli()
		request.RouterSlicesNext()
		end := time.Now().UnixMilli()
		attr := []attribute.KeyValue{
			{Key: "__msg_id", Value: attribute.StringValue(fmt.Sprintf("%v", request.GetMsgID()))},
			{Key: "__host", Value: attribute.StringValue(hostName)},
			{Key: "__env", Value: attribute.StringValue(env)},
			{Key: "__version", Value: attribute.StringValue(version)},
		}
		ctx := context.Background()
		EmitTime(ctx, "zinx", end-start, attr...)
		EmitCount(ctx, "zinx", 1, attr...)
	})
	// 记录连接数
	var connected int64
	server.SetOnConnStart(func(connection ziface.IConnection) {
		attr := []attribute.KeyValue{
			{Key: "__host", Value: attribute.StringValue(hostName)},
			{Key: "__env", Value: attribute.StringValue(env)},
			{Key: "__version", Value: attribute.StringValue(version)},
		}
		atomic.AddInt64(&connected, 1)
		EmitGauge(connection.Context(), "zinx", atomic.LoadInt64(&connected), attr...)
	})
	server.SetOnConnStop(func(connection ziface.IConnection) {
		attr := []attribute.KeyValue{
			{Key: "__host", Value: attribute.StringValue(hostName)},
			{Key: "__env", Value: attribute.StringValue(env)},
			{Key: "__version", Value: attribute.StringValue(version)},
		}
		atomic.AddInt64(&connected, -1)
		EmitGauge(connection.Context(), "zinx", atomic.LoadInt64(&connected), attr...)
	})
}

// InstrumentRedisV8 仪表化redis，必须是v8的连接
func InstrumentRedisV8(client *redis.ClusterClient) {
	client.AddHook(&redisHook{
		meterBefore: func(ctx context.Context) context.Context {
			start := time.Now().UnixMilli()
			if ctx == nil {
				ctx = context.Background()
			}
			return context.WithValue(ctx, "metrics.before", start)
		},
		meterAfter: func(ctx context.Context, cmd string) {
			if ctx == nil {
				return
			}
			start := ctx.Value("metrics.before")
			if start == nil {
				return
			}
			end := time.Now().UnixMilli()
			attr := []attribute.KeyValue{
				{Key: "__cmd", Value: attribute.StringValue(cmd)},
				{Key: "__host", Value: attribute.StringValue(hostName)},
				{Key: "__env", Value: attribute.StringValue(env)},
				{Key: "__version", Value: attribute.StringValue(version)},
			}
			EmitTime(ctx, "redis", end-start.(int64), attr...)
			EmitCount(ctx, "redis", 1, attr...)
		},
	})
}
