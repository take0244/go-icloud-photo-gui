package ifstorelocal

import (
	"net/http"
	"path/filepath"

	"github.com/cavaliergopher/grab/v3"
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

func (d *downloader) DownloadFileUrls(dir string, urls []usecase.FileUrl, workers int) error {
	requests := []*grab.Request{}
	for _, url := range urls {
		req, err := grab.NewRequest(dir, url.Url)
		req.NoResume = true
		if err != nil {
			return err
		}

		if url.Filename != "" {
			req.Filename = filepath.Join(dir, url.Filename)
		}

		requests = append(requests, req)
	}

	respch := d.client.DoBatch(workers, requests...)
	for resp := range respch {
		if err := resp.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (d *downloader) DownloadUrls(dir string, urls []string, workers int) error {
	requests := []usecase.FileUrl{}
	for _, url := range urls {
		requests = append(requests, usecase.FileUrl{
			Url:      url,
			Filename: "",
		})
	}

	return d.DownloadFileUrls(dir, requests, workers)
}
