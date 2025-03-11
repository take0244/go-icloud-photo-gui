package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/take0244/go-icloud-photo-gui/aop"
)

func DeleteFiles(dirPath string) error {
	aop.Logger().Debug("DeleteFiles: " + dirPath)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dirPath, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	return nil
}
