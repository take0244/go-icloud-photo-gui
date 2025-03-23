package appctx

import (
	"context"

	"github.com/take0244/go-icloud-photo-gui/util"
)

const (
	requestIdKey contextKey = "requestId"
)

func WithRequestId(ctx context.Context) context.Context {
	id := RequestId(ctx)
	if id != "" {
		return ctx
	}

	return context.WithValue(ctx, requestIdKey, util.MustUUID())
}

func RequestId(ctx context.Context) string {
	id, ok := ctx.Value(requestIdKey).(string)
	if ok {
		return id
	}

	return ""
}
