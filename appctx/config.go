package appctx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	AppleInfo struct {
		WebServiceSckdatabasewsUrl string
		AppleId                    string
	}
	ConfigFile struct {
		MaxParallel   int
		OauthClientId string
		AppleInfo     map[string]AppleInfo
	}
)

const (
	README               = "# write app_config.json OauthClientId"
	configKey contextKey = "config"
)

var (
	configDir                 = ""
	configFile                = "app_config.json"
	readmeFile                = "README.md"
	confCache     *ConfigFile = nil
	defaultConfig             = ConfigFile{
		OauthClientId: "changeit",
		MaxParallel:   3,
		AppleInfo:     map[string]AppleInfo{},
	}
)

func saveConf(config ConfigFile) {
	configFilePath := filepath.Join(configDir, configFile)
	if err := os.WriteFile(configFilePath, util.MustMarshal(config), 0777); err != nil {
		panic(err)
	}
	confCache = &config
}

func loadConf() ConfigFile {
	if confCache != nil {
		return *confCache
	}

	configFilePath := filepath.Join(configDir, configFile)
	configByts, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}

	c, err := util.Unmarshal[ConfigFile](configByts)
	if err != nil {
		panic(err)
	}

	return *c
}
func InitConfig(appDir string) {
	configDir = appDir
	if err := os.MkdirAll(configDir, 0777); err != nil {
		panic(err)
	}

	configFilePath := filepath.Join(configDir, configFile)
	readmeFilePath := filepath.Join(configDir, readmeFile)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		saveConf(defaultConfig)
		os.WriteFile(readmeFilePath, []byte(README), 0777)

		fmt.Println("see", configFilePath, "change oauthClientId")
		open.Start(appDir)
		os.Exit(0)
	}
}

func WithCacheConfig(ctx context.Context) context.Context {
	config := loadConf()
	return context.WithValue(ctx, configKey, config)
}

func PeekConfig(ctx context.Context, v func(*ConfigFile)) context.Context {
	config := Config(ctx)
	v(&config)
	saveConf(config)

	return context.WithValue(ctx, configKey, config)
}

func Config(ctx context.Context) ConfigFile {
	conf, ok := ctx.Value(configKey).(ConfigFile)
	if ok {
		return conf
	}

	return loadConf()
}

func CacheConfig(v func(*ConfigFile)) {
	PeekConfig(context.TODO(), v)
}
