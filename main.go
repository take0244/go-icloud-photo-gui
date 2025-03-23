package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/take0244/go-icloud-photo-gui/appctx"
	ifstorelocal "github.com/take0244/go-icloud-photo-gui/infrastructure/datastore/local"
	infraui "github.com/take0244/go-icloud-photo-gui/infrastructure/presentation/ui"
	infraicloud "github.com/take0244/go-icloud-photo-gui/infrastructure/repository/icloud"
	"github.com/take0244/go-icloud-photo-gui/usecase"
)

var isDev = os.Getenv("DEVELOPMENT") == "true"

func init() {
	homeDir, _ := os.UserHomeDir()
	appDir := filepath.Join(homeDir, ".goicloudgui")

	appctx.InitConfig(appDir)
	appctx.InitCookies(appDir)
	if isDev {
		appctx.InitLogger(os.Stdout, slog.LevelInfo)
	} else {
		file, err := os.OpenFile(filepath.Join(appDir, "log.txt"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			panic(err)
		}
		appctx.InitLogger(file, slog.LevelInfo)
	}
}

func main() {
	icloud := infraicloud.NewICloud()
	downloader := ifstorelocal.NewDownloader()

	ucase := usecase.NewUseCase(icloud, downloader)

	if err := infraui.Run(ucase); err != nil {
		panic(err)
	}
}
