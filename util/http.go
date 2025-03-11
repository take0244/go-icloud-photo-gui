package util

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/take0244/go-icloud-photo-gui/aop"
)

type LoggingTransport struct {
	Transport http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	headersString := func(header http.Header) []any {
		result := []any{}
		for key, values := range header {
			result = append(result, key+":"+strings.Join(values, ","))
		}
		return result
	}
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	aop.Logger().Info("Request",
		slog.String("url", req.URL.String()),
		slog.String("method", req.Method),
	)
	curl, err := GetCurlCommand(req)
	if err == nil {
		aop.Logger().Debug("Debug-Request",
			slog.String("curl", curl),
		)
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, _ = io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	aop.Logger().Info("Response",
		slog.String("status", resp.Status),
	)
	aop.Logger().Debug("Debug-Response",
		slog.Group("headers", headersString(resp.Header)...),
		slog.String("body", string(bodyBytes)),
	)

	return resp, nil
}

func HttpClientWithJarCookie() *http.Client {
	return &http.Client{
		Timeout:   50 * time.Second,
		Transport: &LoggingTransport{Transport: http.DefaultTransport},
		Jar:       NewPersistentCookieJar(),
	}
}

func MustParseUrl(urlString string, queries map[string]string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	q := u.Query()
	for k, v := range queries {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func MustRequest(method, url string, body io.Reader, headers any) *http.Request {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}

	if headers != nil {
		if strHeader, ok := headers.(map[string]string); ok {
			for key, value := range strHeader {
				req.Header.Set(key, value)
			}
		} else if strHeaders, ok := headers.(map[string][]string); ok {
			for key, values := range strHeaders {
				for _, hv := range values {
					req.Header.Add(key, hv)
				}
			}
		} else {
			panic("unknown headers type")
		}
	}
	return req
}

func GetCurlCommand(req *http.Request) (string, error) {
	if req.URL == nil {
		return "", fmt.Errorf("invalid request: req.URL is nil")
	}

	command := []string{"curl", "-X", strconv.Quote(req.Method), strconv.Quote(req.URL.String()), "--compressed"}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return "", fmt.Errorf("error reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) > 0 {
			command = append(command, "-d", string(body))
		}
	}

	for key, values := range req.Header {
		for _, value := range values {
			command = append(command, "-H", strconv.Quote(fmt.Sprintf("%s: %s", key, value)))
		}
	}

	return strings.Join(command, " "), nil
}

func HttpDoJSON[T any](client *http.Client, req *http.Request) (*T, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}
	defer resp.Body.Close()

	if !HttpCheck2XX(resp) {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &result, nil
}

func HttpDoGzipJSON[T any](client *http.Client, req *http.Request) (*T, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "gzip" {
		return nil, errors.New("missing encoding type")
	}

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to new reader: %w", err)
	}
	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadAll: %w", err)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

func HttpDownloadAndUnzip(url, destDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download ZIP file: %w", err)
	}
	defer resp.Body.Close()

	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read ZIP data: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create ZIP reader: %w", err)
	}

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, file := range zipReader.File {
		destFilePath := filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destFilePath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		dirPath := filepath.Dir(destFilePath)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		destFile, err := os.Create(destFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer destFile.Close()

		zipFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open ZIP file: %w", err)
		}
		defer zipFile.Close()

		if _, err := io.Copy(destFile, zipFile); err != nil {
			return fmt.Errorf("failed to copy file contents: %w", err)
		}
	}

	return nil
}

func HttpCheck2XX(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
