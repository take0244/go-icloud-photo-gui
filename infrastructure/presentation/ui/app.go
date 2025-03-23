package infraui

import (
	"context"
	"embed"
	"log/slog"
	"os"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx   context.Context
	ucase usecase.UseCase
	id    string
}

func Run(ucase usecase.UseCase) error {
	app := &App{
		ucase: ucase,
		ctx:   appctx.NewAppContext(),
	}

	err := wails.Run(&options.App{
		Title:  "iCloud Photos Downloader",
		Width:  1024 / 2,
		Height: 768 / 2,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []any{
			app,
		},
	})

	return err
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) before(id string) {
	a.ctx = util.ContextChain(
		a.ctx,
		appctx.WithRequestId,
		appctx.WithCacheCookies,
		appctx.WithCacheConfig,
	)
	if id != "" {
		a.id = id
		a.ctx = appctx.WithUser(a.ctx, appctx.ContextUser{ID: id})
	} else {
		a.ctx = appctx.WithUser(a.ctx, appctx.ContextUser{ID: a.id})
	}
}

func (a *App) LoginICloud(username, password string) string {
	a.before(util.Hash(username + password))

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer util.RecoverFromPanic()

	result, err := a.ucase.Login(a.ctx, username, password)
	if err != nil {
		slog.ErrorContext(a.ctx, err.Error())
		return util.MustJsonString(map[string]any{"error": true})
	}

	return util.MustJsonString(result)
}

func (a *App) Code2fa(code string) string {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer util.RecoverFromPanic()

	if err := a.ucase.Code2fa(a.ctx, code); err != nil {
		slog.ErrorContext(a.ctx, err.Error())
		return util.MustJsonString(false)
	}

	return util.MustJsonString(true)
}

func (a *App) AllDownloadPhotos(path string) string {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer util.RecoverFromPanic()

	if err := a.ucase.DownloadAllPhotos(a.ctx, path); err != nil {
		slog.ErrorContext(a.ctx, err.Error())
		return "失敗しました。"
	} else {
		open.Start(path)
		return ""
	}
}

func (a *App) SelectDirectory() string {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer util.RecoverFromPanic()

	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		panic(err)
	}

	return dir
}

func (a *App) Cancel() {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer util.RecoverFromPanic()

	os.Exit(0)
}
