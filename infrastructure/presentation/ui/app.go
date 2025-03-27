package infraui

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx   context.Context
	ucase usecase.UseCase
	id    string
}

func Run(ucase usecase.UseCase) error {
	icon, _ := os.ReadFile("./build/appicon.png")
	app := &App{
		ucase: ucase,
		ctx:   appctx.NewAppContext(),
	}

	err := wails.Run(&options.App{
		Title:    "iCloud Photos Downloader",
		LogLevel: logger.ERROR,
		Width:    1024 / 2,
		Height:   768 / 2,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			routines := runtime.NumGoroutine()
			slog.InfoContext(ctx, "routines", slog.Int("count", routines))
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Icon: icon,
			},
		},
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
	defer panicTrace(a.ctx)

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
	defer panicTrace(a.ctx)

	if err := a.ucase.Code2fa(a.ctx, code); err != nil {
		slog.ErrorContext(a.ctx, err.Error())
		return util.MustJsonString(false)
	}

	return util.MustJsonString(true)
}

func (a *App) AllDownloadPhotos(path string) string {
	a.before("")
	ticker := time.NewTicker(time.Second)
	a.ctx = appctx.WithProgress(a.ctx)
	appctx.AppTrace(a.ctx)
	p, ok := appctx.Progress(a.ctx)

	defer func() {
		panicTrace(a.ctx)
		p.Close()
		appctx.DeferAppTrace(a.ctx)
		ticker.Stop()
		wailsruntime.EventsEmit(a.ctx, "app_progressEvent", 1)
	}()

	if ok {
		valueCh := p.Value()
		go func() {
			for v := range valueCh {
				wailsruntime.EventsEmit(a.ctx, "app_progressEvent", util.MustJsonString(map[string]any{
					"value": v,
					"phase": p.Phase(),
				}))
			}
		}()
	}

	if err := a.ucase.DownloadAllPhotos(a.ctx, path); err != nil {
		slog.ErrorContext(a.ctx, err.Error())
		return "失敗しました。(" + err.Error() + ")"
	}

	open.Start(path)
	return ""
}

func (a *App) SelectDirectory() string {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer panicTrace(a.ctx)

	dir, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{})
	if err != nil {
		panic(err)
	}

	return dir
}

func (a *App) Cancel() {
	a.before("")

	appctx.AppTrace(a.ctx)
	defer appctx.DeferAppTrace(a.ctx)
	defer panicTrace(a.ctx)

	wailsruntime.Quit(a.ctx)
}

func panicTrace(ctx context.Context) {
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
	}
}
