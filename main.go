package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/aop"
	infraicloud "github.com/take0244/go-icloud-photo-gui/infrastructure/icloud"
	infraui "github.com/take0244/go-icloud-photo-gui/infrastructure/ui"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	ConfigFile struct {
		MaxParallel   int
		OauthClientId string
	}
)

const README = "# write app_config.json OauthClientId"

var (
	homeDir, _ = os.UserHomeDir()
	appDir     = filepath.Join(homeDir, ".goicloudgui")
	logFile    = filepath.Join(appDir, "log.txt")
	configFile = filepath.Join(appDir, "app_config.json")
	readmeFile = filepath.Join(appDir, "README.md")
)

func init() {
	if err := os.MkdirAll(appDir, 0777); err != nil {
		panic(err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	aop.InitLogger(io.Writer(file), slog.LevelError)
}

func main() {
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		if err = os.WriteFile(configFile, util.MustMarshal(ConfigFile{OauthClientId: "changeit", MaxParallel: 1}), 0777); err != nil {
			panic(err)
		}

		if err = os.WriteFile(readmeFile, []byte(README), 0777); err != nil {
			panic(err)
		}

		fmt.Println("see", configFile, "change oauthClientId")
		open.Start(appDir)
		os.Exit(0)
		return
	}

	configByts, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var config ConfigFile
	err = json.Unmarshal(configByts, &config)
	if err != nil {
		panic(err)

	}

	if config.OauthClientId == "changeit" || config.OauthClientId == "" {
		fmt.Println("see", configFile, "change oauthClientId")
		open.Start(appDir)
		os.Exit(0)
		return
	}

	icloud := infraicloud.NewICloud(
		config.OauthClientId,
		infraicloud.MetaDirPathOption(appDir),
	)

	ucase := usecase.NewUseCase(icloud)

	if err := infraui.Run(ucase); err != nil {
		panic(err)
	}
}
