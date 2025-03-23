package appctx

import (
	"context"
)

const processKey contextKey = "process"

func WithProgress(ctx context.Context) context.Context {
	pchan := make(chan int, 10)
	return context.WithValue(ctx, processKey, pchan)
}

func Progress(ctx context.Context) (chan int, bool) {
	p, ok := ctx.Value(processKey).(chan int)
	return p, ok
}
