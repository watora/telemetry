初始化
```golang
telemetry.Init(func(cfg *telemetry.Config) {
  cfg.Env = "local"
  cfg.Version = "1.0.0"
  cfg.AppName = "AppName"
  cfg.UseLogger = true
  cfg.UseMetrics = true
  cfg.LogEndPoint = "localhost:4318"   // collector的地址
  cfg.MetricsEndPoint = "localhost:4317"
})
```

log:
- 引入依赖
  - github.com/watora/telemetry/log
  - github.com/watora/telemetry/trace
- 对原来的zaplogger 使用zapbridge替换
  - logger = log.ZapBridge(log)
- 如果要记录traceId 先创建span 然后把ctx赋给logger
  - ctx, span := trace.Tracer.Start(context.Background(), "xxx")
  - logger = log.WithCtx(logger, ctx)
- 如果没有logger 可以用全局方法
  - log.CtxInfo(ctx, "xxxx")
 
metrics
- 引入依赖
  - github.com/watora/telemetry/metrics
- 使用
  - counter: metrics.EmitCount(ctx, "xxx", 1)
  - gauge: metrics.EmitGauge(ctx, "xxx", 1)
  - time: metrics.EmitTime(ctx, "xxx", time.Since(start).Milliseconds())
- 仪表化 zinx需开启新版路由 redis只支持v8
  - gorm: metrics.InstrumentGORM(db)
  - gozero: metrics.InstrumentGoZero(server)
  - zinx: metrics.InstrumentZinx(server)
  - redis: metrics.InstrumentRedisV8(cluster)
