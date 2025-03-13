package aop

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
)

var logger *slog.Logger

func InitLogger(writer io.Writer, level slog.Leveler) *slog.Logger {
	logger = slog.New(&customHandler{
		Handler: slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level}),
	})
	return logger
}

func Logger() *slog.Logger {
	if logger == nil {
		panic("not call aop.InitLogger")
	}

	return logger
}

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	_, file, line, _ := runtime.Caller(3)
	r.Message = fmt.Sprintf("%s (at %s:%d)", r.Message, file, line)

	return h.Handler.Handle(ctx, r)
}
