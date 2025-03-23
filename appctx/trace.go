package appctx

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"
)

func AppTrace(ctx context.Context) {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		slog.InfoContext(ctx, RequestId(ctx),
			slog.String("start-func", filepath.Base(runtime.FuncForPC(pc).Name())),
		)
	}
}

func DeferAppTrace(ctx context.Context) {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		slog.InfoContext(ctx, RequestId(ctx),
			slog.String("end-func", filepath.Base(runtime.FuncForPC(pc).Name())),
		)
	}
}
