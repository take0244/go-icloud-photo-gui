package ui

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/aop"
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
}

func Run(ucase usecase.UseCase) {
	app := &App{
		ucase: ucase,
		ctx:   context.Background(),
	}

	err := wails.Run(&options.App{
		Title:  "ui",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) LoginICloud(username, password string) string {
	fmt.Println(username, password)
	result, err := a.ucase.ICloudService().Login(username, password)
	if err != nil {
		aop.Logger().Error(err.Error())
		return util.MustJsonString(usecase.LoginResult{Required2fa: false})
	}

	return util.MustJsonString(result)
}

func (a *App) Code2fa(code string) string {
	if err := a.ucase.ICloudService().Code2fa(code); err != nil {
		aop.Logger().Error(err.Error())
		os.Exit(0)
	}

	return util.MustJsonString(true)
}

func (a *App) AllDownloadPhotos(path string) {
	if err := a.ucase.ICloudService().PhotoService().DownloadAllPhotos(path); err != nil {
		aop.Logger().Error(err.Error())
	} else {
		open.Start(path)
	}
}

func (a *App) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		panic(err)
	}

	return dir
}

func (a *App) Cancel() {
	os.Exit(0)
	// a.ucase.ICloudService().Clear()
}
