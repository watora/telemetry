package metrics

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type redisHook struct {
	meterBefore func(ctx context.Context) context.Context
	meterAfter  func(ctx context.Context, cmd string)
}

func (hook *redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return hook.meterBefore(ctx), nil
}

func (hook *redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	hook.meterAfter(ctx, cmd.Name())
	return nil
}

func (hook *redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return hook.meterBefore(ctx), nil
}

func (hook *redisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	hook.meterAfter(ctx, "pipeline")
	return nil
}
