package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dtrugman/qory/lib/profile"
)

const (
	osWin         = "windows"
	envWinAppData = "APPDATA"

	dirDotQory    = ".qory"
	dirUnixConfig = ".config"
	dirQory       = "qory"
	dirPerm       = 0700
)

func getConfigDir() (string, error) {
	userDir, err := profile.GetUserDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(userDir, dirDotQory)

	if runtime.GOOS != osWin {
		oldDir := filepath.Join(userDir, dirUnixConfig, dirQory)
		if tryMigrateOldConfigDir(oldDir, configDir) {
			return configDir, nil
		}
	}

	return getOrCreateDir(configDir)
}

func getOrCreateDir(path string) (string, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return path, nil
		} else {
			return "", fmt.Errorf("already exists, but not a dir")
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(path, dirPerm); err != nil {
		return "", err
	}

	return path, nil
}

func tryMigrateOldConfigDir(oldDir, newDir string) bool {
	if stat, err := os.Stat(oldDir); err != nil || !stat.IsDir() {
		return false
	}

	if err := os.Rename(oldDir, newDir); err != nil {
		return false
	}

	return true
}
