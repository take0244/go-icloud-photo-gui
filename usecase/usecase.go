package usecase

import (
	"context"

	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	Photo struct {
		ID          string
		CheckSum    string
		DownloadUrl string
		Filename    string
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
	}
	Downloader interface {
		DownloadUrls(ctx context.Context, dir string, urls []string, workers int) error
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
		return nil, err
	}

	return &LoginResult{Required2fa: required2fa}, nil
}

func (u *useCase) Code2fa(ctx context.Context, code string) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	return u.iCloudService.Code2fa(ctx, code)
}

func (u *useCase) DownloadAllPhotos(ctx context.Context, dir string) error {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	config := appctx.Config(ctx)
	photos, duplicatePhotos := u.splitDuplicateCheckSum(u.iCloudService.GetAllPhotos(ctx))

	reqs := []FileUrl{}
	for _, photo := range duplicatePhotos {
		reqs = append(reqs, FileUrl{
			Url:      photo.DownloadUrl,
			Filename: photo.Filename,
		})
		u.downloader.DownloadFileUrls(ctx, dir, reqs, config.MaxParallel)
	}

	for _, chunked := range util.ChunkSlice(util.ChunkSlice(photos, 1000), config.MaxParallel) {
		urls := []string{}
		for _, v := range chunked {
			url, err := u.iCloudService.MakeDownloadUrlByPhotos(ctx, v)
			if err != nil {
				return err
			}
			urls = append(urls, url)
		}

		u.downloader.DownloadUrls(ctx, dir, urls, config.MaxParallel)
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
