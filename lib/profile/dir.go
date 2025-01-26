package profile

import (
	"fmt"
	"os"
	"runtime"
)

const (
	osWin         = "windows"
	envWinAppData = "APPDATA"

	dirUnixConfig = ".config"
)

func getUserDirWindows() (string, error) {
	appData, found := os.LookupEnv(envWinAppData)
	if !found {
		return "", fmt.Errorf("Env var [%s] not defined", envWinAppData)
	}

	return appData, nil
}

func GetUserDir() (string, error) {
	if runtime.GOOS == osWin {
		return getUserDirWindows()
	} else {
		return os.UserHomeDir()
	}
}
