package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	Photo struct {
		ID          string
		CheckSum    string
		DownloadUrl string
		Filename    string
		FileSize    float64
	}
	ICloudService interface {
		Login(ctx context.Context, username, password string) (bool, error)
		Code2fa(ctx context.Context, code string) error
		GetAllPhotos(ctx context.Context) []Photo
		MakeDownloadUrlByPhotos(ctx context.Context, photos []Photo) (string, error)
	}
	FileUrl struct {
		Url      string
		Filename string
		FileSize float64
	}
	Downloader interface {
		DownloadFileUrls(ctx context.Context, dir string, urls []FileUrl, workers int) error
	}
)

type (
	LoginResult struct {
		Required2fa bool
	}
	UseCase interface {
		Login(ctx context.Context, username, password string) (*LoginResult, error)
		Code2fa(ctx context.Context, code string) error
		DownloadAllPhotos(ctx context.Context, dir string) error
	}
	useCase struct {
		iCloudService ICloudService
		downloader    Downloader
	}
)

func NewUseCase(rep1 ICloudService, downloader Downloader) UseCase {
	return &useCase{
		iCloudService: rep1,
		downloader:    downloader,
	}
}

func (u *useCase) Login(ctx context.Context, username string, password string) (*LoginResult, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	required2fa, err := u.iCloudService.Login(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &LoginResult{Required2fa: required2fa}, nil
}

func (u *useCase) Code2fa(ctx context.Context, code string) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	return u.iCloudService.Code2fa(ctx, code)
}

func (u *useCase) DownloadAllPhotos(ctx context.Context, dir string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]uintptr, 10)
			n := runtime.Callers(2, buf)
			frames := runtime.CallersFrames(buf[:n])
			result := ""
			for frame, more := frames.Next(); more; frame, more = frames.Next() {
				result += fmt.Sprintf("  - %s\n    %s:%d\n", frame.Function, frame.File, frame.Line)
			}

			slog.ErrorContext(ctx, "Panic",
				slog.Any("🔥 Panic Recovered:", r),
				slog.String("📌 Stack Trace:", result),
			)
			err = errors.New(result)
		}
	}()
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)
	p, okProgress := appctx.Progress(ctx)
	config := appctx.Config(ctx)

	if okProgress {
		p.SetPhase("CHECK_FILES", 1)
	}
	photos, duplicatePhotos := u.splitDuplicateCheckSum(u.iCloudService.GetAllPhotos(ctx))
	chunkedPhotos := util.ChunkSlice(photos, 1000)

	// 被りをダウンロード
	if okProgress {
		p.SetPhase("DOWNLOAD_DUPLICATE", float64(len(duplicatePhotos)))
	}
	slog.InfoContext(ctx, "Start Duplicate")
	depRequests := []FileUrl{}
	for i, photo := range duplicatePhotos {
		slog.InfoContext(ctx, "Duplicate Index", slog.Int("index", i))
		depRequests = append(depRequests, FileUrl{
			Url:      photo.DownloadUrl,
			Filename: photo.Filename,
			FileSize: photo.FileSize,
		})
	}
	if err := u.downloader.DownloadFileUrls(ctx, dir, depRequests, config.MaxParallel); err != nil {
		return err
	}

	// 全てダウンロード
	if okProgress {
		p.SetPhase("DOWNLOAD_ZIP", float64(len(chunkedPhotos)))
	}
	slog.InfoContext(ctx, "Start Zip")
	for i, chunked := range util.ChunkSlice(chunkedPhotos, config.MaxParallel) {
		slog.InfoContext(ctx, "Zip Index", slog.Int("index", i))
		requests := []FileUrl{}
		for _, v := range chunked {
			url, err := u.iCloudService.MakeDownloadUrlByPhotos(ctx, v)
			if err != nil {
				return err
			}

			req := FileUrl{Url: url, FileSize: 0}
			for _, fs := range v {
				req.FileSize += fs.FileSize
			}
			requests = append(requests, req)

		}
		if err := u.downloader.DownloadFileUrls(ctx, dir, requests, config.MaxParallel); err != nil {
			return err
		}
	}

	return nil
}

func (u *useCase) splitDuplicateCheckSum(photos []Photo) ([]Photo, []Photo) {
	var (
		original  []Photo
		duplicate []Photo
		dupMap    = map[string]struct{}{}
	)

	for _, p := range photos {
		_, exits := dupMap[p.CheckSum]
		if exits {
			duplicate = append(duplicate, p)
		} else {
			original = append(original, p)
			dupMap[p.CheckSum] = struct{}{}
		}
	}

	return original, duplicate
}
