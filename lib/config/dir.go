package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirDotQory = ".qory"
	dirPerm    = 0700
)

func getConfigDir(userDir string) (string, error) {
	dir := filepath.Join(userDir, dirDotQory)
	return getOrCreateDir(dir)
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
