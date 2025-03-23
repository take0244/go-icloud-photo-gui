package util

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type loggingTransport struct {
	Transport http.RoundTripper
}

func NewLoggingTransport(child http.RoundTripper) http.RoundTripper {
	return &loggingTransport{
		Transport: child,
	}
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	isDebug := slog.Default().Enabled(req.Context(), slog.LevelDebug)

	if isDebug {
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		curl, err := CurlCommand(req)
		if err == nil {
			slog.DebugContext(ctx, "Debug-Request", slog.String("curl", curl))
		}
	}
	slog.InfoContext(ctx, "Request",
		slog.String("url", req.URL.String()),
		slog.String("method", req.Method))

	start := time.Now().UnixMilli()
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, err
	}
	end := time.Now().UnixMilli()

	slog.InfoContext(ctx, "Response",
		slog.String("status", resp.Status),
		slog.Int64("time", end-start))

	if isDebug {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		slog.DebugContext(ctx, "Debug-Response",
			slog.Group("headers", headersString(resp.Header)...),
			slog.String("body", TruncateString(string(bodyBytes), 300)))
	}

	return resp, nil
}

func headersString(header http.Header) []any {
	result := []any{}
	for key, values := range header {
		result = append(result, key+":"+strings.Join(values, ","))
	}
	return result
}

func headersStringSlice(header http.Header) []string {
	result := []string{}
	for key, values := range header {
		result = append(result, key+":"+strings.Join(values, ","))
	}
	return result
}
