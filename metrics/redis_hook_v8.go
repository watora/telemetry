package metrics

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisHook struct {
	MeterBefore func(ctx context.Context) context.Context
	MeterAfter  func(ctx context.Context, cmd string)
}

func (hook *RedisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return hook.MeterBefore(ctx), nil
}

func (hook *RedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	hook.MeterAfter(ctx, cmd.Name())
	return nil
}

func (hook *RedisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return hook.MeterBefore(ctx), nil
}

func (hook *RedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	hook.MeterAfter(ctx, "pipeline")
	return nil
}
