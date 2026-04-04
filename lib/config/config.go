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
	Mode    = "mode"
	Editor  = "editor"
)

const ( // Valid values for Mode
	ModeNew  = "new"
	ModeLast = "last"
)

type Config struct {
	dir string
}

func NewConfig(userDir string) (*Config, error) {
	configDir, err := getConfigDir(userDir)
	if err != nil {
		return nil, err
	}

	return &Config{
		dir: configDir,
	}, nil
}

func (c *Config) valueGet(filename string) (*string, error) {
	path := filepath.Join(c.dir, filename)
	return fileRead(path)
}

func (c *Config) valueSet(filename string, value string) error {
	path := filepath.Join(c.dir, filename)
	return fileWrite(path, value)
}

func (c *Config) valueUnset(filename string) error {
	path := filepath.Join(c.dir, filename)
	return fileDelete(path)
}

func (c *Config) GetConfigSubdir(name string) (string, error) {
	path := filepath.Join(c.dir, name)
	return getOrCreateDir(path)
}

func (c *Config) Get(key string) (*string, error) {
	return c.valueGet(key)
}

func (c *Config) Set(key string, value string) error {
	return c.valueSet(key, value)
}

func (c *Config) Unset(key string) error {
	return c.valueUnset(key)
}
