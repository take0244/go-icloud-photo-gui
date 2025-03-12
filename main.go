package main

import (
	"os"
	"path/filepath"

	ificloud "github.com/take0244/go-icloud-photo-gui/infrastructure/icloud"
	"github.com/take0244/go-icloud-photo-gui/infrastructure/ui"
	"github.com/take0244/go-icloud-photo-gui/usecase"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".logs2")

	byts, err := os.ReadFile(filepath.Join(logDir, "OAUTH_CLIENT_ID"))
	if err != nil {
		panic(err)
	}
	icloud := ificloud.NewICloud(
		string(byts),
		ificloud.MetaDirPathOption(logDir),
	)
	ucase := usecase.NewUseCase(icloud)
	ui.Run(ucase)
}
