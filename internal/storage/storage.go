package storage

import (
	"os"
	"path/filepath"
	"strings"
)

func DataDir(appName string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exePath, _ = filepath.EvalSymlinks(exePath)

	// Case 1: Running inside a macOS .app bundle
	if strings.Contains(exePath, ".app/Contents/") {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		dir := filepath.Join(base, appName)
		return dir, os.MkdirAll(dir, 0755)
	}

	// Case 2: go run / local binary
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(wd, ".data")
	return dir, os.MkdirAll(dir, 0755)
}
