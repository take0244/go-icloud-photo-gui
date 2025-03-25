package ifstorelocal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	downloader struct {
		client *grab.Client
	}
)

func NewDownloader() *downloader {
	return &downloader{
		client: &grab.Client{
			UserAgent: util.UserAgent,
			HTTPClient: &http.Client{
				Transport: util.NewLoggingTransport(&http.Transport{
					Proxy: http.ProxyFromEnvironment,
				}),
			},
		},
	}
}

func (d *downloader) DownloadFileUrls(ctx context.Context, dir string, urls []usecase.FileUrl, workers int) error {
	requests := []*grab.Request{}
	for _, url := range urls {
		req, err := grab.NewRequest(dir, url.Url)
		req.NoResume = true
		if err != nil {
			return fmt.Errorf("missing request grab%w", err)
		}

		if url.Filename != "" {
			req.Filename = filepath.Join(dir, url.Filename)
		}

		requests = append(requests, req)
	}

	progressIds := util.GenerateUniqKeys(len(requests))
	respch := d.client.DoBatch(workers, requests...)
	p, ok := appctx.Progress(ctx)
	if ok {
		t := time.NewTicker(100 * time.Millisecond)
		t2 := time.NewTicker(2 * time.Second)
		defer t.Stop()
		defer t2.Stop()
		responses := []*grab.Response{}
		countNil := func() int {
			var count = 0
			for _, r := range responses {
				if r == nil {
					count++
				}
			}
			return count
		}

		for countNil() < len(requests) {
			select {
			case resp := <-respch:
				if resp != nil {
					responses = append(responses, resp)
				}
			case <-t2.C:
				slog.InfoContext(ctx, "countNil", slog.Int("cnt", countNil()), slog.Int("requests.len", len(requests)))
			case <-t.C:
				for i, resp := range responses {
					if resp == nil {
						p.Count(progressIds[i], 1)
						continue
					}
					if !resp.IsComplete() {
						fmt.Println(resp.Size())
						fmt.Println(resp.Bytes())
						fmt.Println(resp.BytesComplete())
						fmt.Println(resp.BytesPerSecond())
						p.Count(progressIds[i], resp.Progress())
						continue
					}
					if err := resp.Err(); err != nil {
						return fmt.Errorf("grab download error: %w", err)
					}

					responses[i] = nil
					p.Count(progressIds[i], 1)
				}
			}
		}
	} else {
		for resp := range respch {
			if err := resp.Err(); err != nil {
				return err
			}
		}
	}

	return nil
}
