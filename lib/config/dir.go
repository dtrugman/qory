package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	envWinAppData = "APPDATA"
	dirUnixConfig = ".config"
	dirQory       = "qory"
)

// getConfigDir returns the configuration directory, creating it if it doesn't exist
func getConfigDir() (string, error) {
	userConfigDir, err := getUserConfigDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(userConfigDir, dirQory)

	_, err = os.Stat(configDir)
	if err == nil {
		return configDir, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

func getUserConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv(envWinAppData)
		return appData, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, dirUnixConfig)
	return configDir, nil
}
