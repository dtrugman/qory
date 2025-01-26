package config

import (
	"path/filepath"
)

const (
	modePrivate = 0600 // Owner can RW, others have no access
)

const ( // Public configuration values
	APIKey  = "api_key"
	BaseURL = "base_url"
	Model   = "model"
	Prompt  = "prompt"
)

type config struct {
	dir string
}

type Config interface {
	GetConfigSubdir(name string) (string, error)

	Get(key string) (*string, error)
	Set(key string, value string) error
	Unset(key string) error
}

func NewConfig() (Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	return &config{
		dir: configDir,
	}, nil
}

func (c *config) valueGet(filename string) (*string, error) {
	path := filepath.Join(c.dir, filename)
	return fileRead(path)
}

func (c *config) valueSet(filename string, value string) error {
	path := filepath.Join(c.dir, filename)
	return fileWrite(path, value)
}

func (c *config) valueUnset(filename string) error {
	path := filepath.Join(c.dir, filename)
	return fileDelete(path)
}

func (c *config) GetConfigSubdir(name string) (string, error) {
	path := filepath.Join(c.dir, name)
	return getOrCreateDir(path)
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
