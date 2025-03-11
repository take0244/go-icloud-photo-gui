package main

import (
	"os"
	"path/filepath"

	ificloud "github.com/take0244/go-icloud-photo-gui/infrastructure/icloud"
	"github.com/take0244/go-icloud-photo-gui/infrastructure/ui"
	"github.com/take0244/go-icloud-photo-gui/usecase"
)

func main() {
	icloud := ificloud.NewICloud(
		ificloud.MetaDirPathOption(filepath.Join(os.TempDir(), "goicloud")),
	)
	ucase := usecase.NewUseCase(icloud)
	ui.Run(ucase)
}
