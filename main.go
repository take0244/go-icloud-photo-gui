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
	oauthClientId := "changeit"
	icloud := ificloud.NewICloud(
		oauthClientId,
		ificloud.MetaDirPathOption(logDir),
	)
	ucase := usecase.NewUseCase(icloud)
	ui.Run(ucase)
}
