package log

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

func LogxWithCtx(ctx context.Context) logx.Logger {
	return logx.WithContext(ctx).WithFields(logx.LogField{Key: "context", Value: ctx})
}
