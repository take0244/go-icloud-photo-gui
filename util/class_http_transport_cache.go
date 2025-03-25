package util

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
)

type cacheTransport struct {
	Transport http.RoundTripper
	dir       string
}

func NewCacheTransport(child http.RoundTripper, _dir string) http.RoundTripper {
	if err := os.MkdirAll(_dir, 0777); err != nil {
		panic(err)
	}
	return &cacheTransport{
		Transport: child,
		dir:       _dir,
	}
}

func (t *cacheTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if !Cache(ctx) {
		return t.Transport.RoundTrip(req)
	}

	hash, hashedBody, err := HashBody(req.Body)
	if err != nil {
		return &http.Response{}, err
	}

	req.Body = hashedBody
	key := req.URL.String() + hash + req.Method
	filename := filepath.Join(t.dir, Hash(key))

	if file, err := os.Open(filename); err == nil {
		defer file.Close()
		respBytes := []byte{}
		err := gob.NewDecoder(file).Decode(&respBytes)
		if err == nil {
			return http.ReadResponse(bufio.NewReader(bytes.NewReader(respBytes)), req)
		}
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if !HttpCheck2XX(resp) {
		return resp, err
	}

	file, err := os.Create(filename)
	if err != nil {
		return resp, err
	}
	defer file.Close()

	byts, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return resp, err
	}

	if err := gob.NewEncoder(file).Encode(byts); err != nil {
		return resp, err
	}

	return resp, err
}

type contextKey string

const cacheKey contextKey = "cache"

func WithCache(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, cacheKey, enabled)
}

func Cache(ctx context.Context) bool {
	enabled, ok := ctx.Value(cacheKey).(bool)
	if !ok {
		return false
	}

	return enabled
}
