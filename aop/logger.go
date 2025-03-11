package aop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

var logFile = ""
var logger *slog.Logger

func init() {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		logDir := filepath.Join(homeDir, ".logs2")
		os.MkdirAll(logDir, 0777)
		logFile = filepath.Join(logDir, "log.txt")
		os.Remove(logFile)
		os.WriteFile(logFile, nil, 0777)
	}
	logger = slog.New(newCustomHandler())
}

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	_, file, line, _ := runtime.Caller(3)
	r.Message = fmt.Sprintf("%s (at %s:%d)", r.Message, file, line)

	return h.Handler.Handle(ctx, r)
}

func newCustomHandler() slog.Handler {
	opt := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handler := slog.NewTextHandler(os.Stdout, opt)
		return &customHandler{
			Handler: handler,
		}
	}

	handler := slog.NewTextHandler(io.Writer(file), opt)

	return &customHandler{
		Handler: handler,
	}
}

func Logger() *slog.Logger {
	return logger
}

func LogJSON(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("JSON 変換エラー:", err)
		return
	}
	fmt.Println(string(jsonBytes))
}
