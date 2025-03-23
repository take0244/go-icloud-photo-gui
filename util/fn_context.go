package util

import "context"

func ContextChain(ctx context.Context, fs ...func(context.Context) context.Context) context.Context {
	for _, f := range fs {
		ctx = f(ctx)
	}
	return ctx
}
