package infraui

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
	"github.com/wailsapp/wails/v2/pkg/application"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type app struct {
	ctx   context.Context
	ucase usecase.UseCase
	id    string
}

func NewApp(ucase usecase.UseCase) *app {
	return &app{
		ucase: ucase,
		ctx:   appctx.NewAppContext(),
	}
}

func (a *app) attachMenu(wailsApp *application.Application) {
	config := appctx.Config(a.ctx)
	appMenu := menu.NewMenu()

	appMenu.Append(menu.AppMenu())

	appMenu.Append(menu.EditMenu())

	settingMenu := appMenu.AddSubmenu("ParallelCount")
	for i := range 8 {
		c := i + 1
		settingMenu.AddRadio(strconv.Itoa(c), config.MaxParallel == c, nil, func(cd *menu.CallbackData) {
			for j, item := range settingMenu.Items {
				_c := j + 1
				if c == _c {
					appctx.PeekConfig(a.ctx, func(cf *appctx.ConfigFile) { cf.MaxParallel = _c })
				}
				item.SetChecked(c == _c)
			}
			wailsApp.SetApplicationMenu(appMenu)
		})
	}

	wailsApp.SetApplicationMenu(appMenu)
}

func (a *app) Run() error {
	icon, _ := os.ReadFile("./build/appicon.png")
	myapp := application.NewWithOptions(&options.App{
		Title:            "iCloud Photos Downloader",
		LogLevel:         logger.ERROR,
		Width:            1024 / 2,
		Height:           768 / 2,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		Mac:              &mac.Options{About: &mac.AboutInfo{Icon: icon}},
		Bind:             []any{a},
		OnStartup: func(ctx context.Context) {
			a.ctx = ctx
		},
		OnShutdown: func(ctx context.Context) {
			routines := runtime.NumGoroutine()
			slog.InfoContext(ctx, "routines", slog.Int("count", routines))
		},
	})

	a.attachMenu(myapp)

	return myapp.Run()
}

func (a *app) before(id string) {
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

func (a *app) LoginICloud(username, password string) string {
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

func (a *app) Code2fa(code string) string {
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

func (a *app) AllDownloadPhotos(path string) string {
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

func (a *app) SelectDirectory() string {
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

func (a *app) Cancel() {
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
