初始化
- 引入依赖
  - github.com/watora/telemetry
  - github.com/watora/telemetry/config
  ```golang
  telemetry.Init(func(cfg *config.Config) {
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
  - logger = log.ZapBridge(logger)
- 如果要记录traceId 先创建span 然后把ctx赋给logger 用完的span必须end
  - ctx, span := trace.StartTrace(context.Background(), "xxx")
  - defer span.End()
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
