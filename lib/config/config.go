package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
        envWinAppData = "APPDATA"

        dirUnixConfig = ".config"
        dirQory = "qory"

        modePrivate = 0600 // Owner can RW, others have no access
)

const ( // Public configuration values
        APIKey = "api_key"
        BaseURL = "base_url"
        Model = "model"
)

type config struct {
        dir string
}

type Config interface {
        Get(key string) (*string, error)
        Set(key string, value string) error
        Unset(key string) error
}

func NewManager() (Config, error) {
        configDir, err := ensureUserConfigDir()
        if err != nil {
            return nil, err
        }

        return &config{
                dir: configDir,
        }, nil
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

func ensureUserConfigDir() (string, error) {
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

func fileRead(filepath string) (*string, error) {
        content, err := os.ReadFile(filepath)
        if err == nil {
            contentStr := string(content)
            return &contentStr, nil
        }

        if os.IsNotExist(err) {
            return nil, nil
        } else {
            return nil, err
        }
}

func (c *config) valueGet(filename string) (*string, error) {
    path := filepath.Join(c.dir, filename)
    return fileRead(path)
}

func fileWrite(filepath string, value string) error {
        return os.WriteFile(filepath, []byte(value), modePrivate)
}

func (c *config) valueSet(filename string, value string) error {
    path := filepath.Join(c.dir, filename)
    return fileWrite(path, value)
}

func fileDelete(filepath string) error {
        return os.Remove(filepath)
}

func (c *config) valueUnset(filename string) error {
    path := filepath.Join(c.dir, filename)
    return fileDelete(path)
}

func (c *config) Get(key string) (*string, error) {
    return c.valueGet(key)
}

func (c *config) Set(key string, value string) error {
    return c.valueSet(key, value)
}

func (c *config) Unset(key string) error {
    return c.valueUnset(key)
}
