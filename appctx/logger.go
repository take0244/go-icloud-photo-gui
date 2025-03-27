package appctx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
)

func InitLogger(writer io.Writer, level slog.Leveler) *slog.Logger {
	logger := slog.New(&customHandler{
		Handler: slog.NewTextHandler(
			writer,
			&slog.HandlerOptions{Level: level},
		),
	})

	slog.SetDefault(logger)

	return logger
}

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		r.Message = fmt.Sprintf("%s (at %s:%d)", r.Message, file, line)
	}

	return h.Handler.Handle(ctx, r)
}
