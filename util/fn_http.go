package util

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"

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

func MustRequest(ctx context.Context, method, url string, body io.Reader, headers any) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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

	if !HttpCheck2XX(resp) {
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

	result, err := Unmarshal[T](body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func FileNameFromResponse(resp *http.Response) string {
	contentDisp := resp.Header.Get("Content-Disposition")
	if contentDisp == "" {
		return ""
	}

	re := regexp.MustCompile(`filename\*?=["']?(?:UTF-8'')?([^"';]+)["']?`)
	matches := re.FindStringSubmatch(contentDisp)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func HttpCheck2XX(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func CurlCommand(req *http.Request) (string, error) {
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

func HashBody(body io.ReadCloser) (string, io.ReadCloser, error) {
	if body == nil {
		return "", nil, nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return "", nil, err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), io.NopCloser(strings.NewReader(string(data))), nil
}
