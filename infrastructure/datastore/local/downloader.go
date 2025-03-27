package ifstorelocal

import (
	"context"
	"fmt"
	"log/slog"
	"math"
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
	requests, err := d.cnvGrabRequest(dir, urls)
	if err != nil {
		return err
	}
	progressIds := util.GenerateUniqKeys(len(requests))
	respch := d.client.DoBatch(workers, requests...)

	p, ok := appctx.Progress(ctx)
	if !ok {
		for resp := range respch {
			if err := resp.Err(); err != nil {
				return err
			}
		}
		return nil
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
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
				slog.InfoContext(ctx, "Loaded Response", slog.String("filename", resp.Filename))
				responses = append(responses, resp)
			}
		case <-t.C:
			for i, resp := range responses {
				if resp == nil {
					continue
				}
				select {
				case <-resp.Done:
					if err := resp.Err(); err != nil {
						return fmt.Errorf("grab download error: %w", err)
					}
					responses[i] = nil
					p.Count(progressIds[i], 1)
				default:
					fileProgress := float64(0)
					if urls[i].FileSize != 0 {
						fileProgress = math.Min(float64(resp.BytesComplete())/(urls[i].FileSize*0.95), 0.999999)
					}
					p.Count(progressIds[i], math.Max(fileProgress, resp.Progress()))
				}
			}
		}
	}

	return nil
}

func (d downloader) cnvGrabRequest(dir string, files []usecase.FileUrl) ([]*grab.Request, error) {
	requests := []*grab.Request{}
	for _, url := range files {
		req, err := grab.NewRequest(dir, url.Url)
		req.NoResume = true
		if err != nil {
			return nil, fmt.Errorf("missing request grab%w", err)
		}

		if url.Filename != "" {
			req.Filename = filepath.Join(dir, url.Filename)
		}

		requests = append(requests, req)
	}

	return requests, nil
}
