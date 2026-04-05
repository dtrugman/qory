package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHistorySize = 50
	DefaultEditor      = "vi"
)

// Config is the application configuration layer. It wraps FileStorage and
// provides typed accessors with defaults, env-var resolution, and validation.
// All getters return an Origin indicating where the value came from.
type Config struct {
	storage *FileStorage
}

func NewConfig(userDir string) (*Config, error) {
	storage, err := NewFileStorage(userDir)
	if err != nil {
		return nil, err
	}
	return &Config{storage: storage}, nil
}

func (c *Config) GetConfigSubdir(name string) (string, error) {
	return c.storage.GetConfigSubdir(name)
}

// Editor returns the editor to use. Resolution order:
//  1. $VISUAL environment variable
//  2. $EDITOR environment variable
//  3. Stored config value
//  4. Built-in default ("vi")
func (c *Config) Editor() (string, Origin, error) {
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual, OriginEnv, nil
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed, OriginEnv, nil
	}
	v, err := c.storage.Get(Editor)
	if err != nil {
		return "", OriginNotSet, err
	}
	if v != nil && *v != "" {
		return *v, OriginUser, nil
	}
	return DefaultEditor, OriginDefault, nil
}

func (c *Config) SetEditor(value string) error {
	return c.storage.Set(Editor, value)
}

func (c *Config) UnsetEditor() error {
	return c.storage.Unset(Editor)
}

// HistorySize returns the number of unnamed sessions to retain.
// Falls back to DefaultHistorySize when not configured.
func (c *Config) HistorySize() (int, Origin, error) {
	v, err := c.storage.Get(HistorySize)
	if err != nil {
		return 0, OriginNotSet, err
	}
	if v == nil {
		return DefaultHistorySize, OriginDefault, nil
	}
	size, err := strconv.Atoi(*v)
	if err != nil {
		return 0, OriginUser, fmt.Errorf("invalid history size %q: %w", *v, err)
	}
	return size, OriginUser, nil
}

func (c *Config) SetHistorySize(value string) error {
	size, err := strconv.Atoi(value)
	if err != nil || size <= 0 {
		return fmt.Errorf("invalid history size %q: must be a positive integer", value)
	}
	return c.storage.Set(HistorySize, value)
}

func (c *Config) UnsetHistorySize() error {
	return c.storage.Unset(HistorySize)
}

func (c *Config) Mode() (string, Origin, error) {
	return c.getNoDefault(Mode)
}

func (c *Config) SetMode(value string) error {
	switch value {
	case ModeNew, ModeLast:
		// valid
	default:
		return fmt.Errorf("invalid mode %q", value)
	}
	return c.storage.Set(Mode, value)
}

func (c *Config) UnsetMode() error {
	return c.storage.Unset(Mode)
}

func (c *Config) APIKey() (string, Origin, error) {
	return c.getNoDefault(APIKey)
}

func (c *Config) SetAPIKey(value string) error {
	return c.storage.Set(APIKey, value)
}

func (c *Config) UnsetAPIKey() error {
	return c.storage.Unset(APIKey)
}

func (c *Config) BaseURL() (string, Origin, error) {
	return c.getNoDefault(BaseURL)
}

func (c *Config) SetBaseURL(value string) error {
	if !strings.HasSuffix(value, "/") {
		value = value + "/"
	}
	return c.storage.Set(BaseURL, value)
}

func (c *Config) UnsetBaseURL() error {
	return c.storage.Unset(BaseURL)
}

func (c *Config) Model() (string, Origin, error) {
	return c.getNoDefault(Model)
}

func (c *Config) SetModel(value string) error {
	return c.storage.Set(Model, value)
}

func (c *Config) UnsetModel() error {
	return c.storage.Unset(Model)
}

func (c *Config) Prompt() (string, Origin, error) {
	return c.getNoDefault(Prompt)
}

func (c *Config) SetPrompt(value string) error {
	return c.storage.Set(Prompt, value)
}

func (c *Config) UnsetPrompt() error {
	return c.storage.Unset(Prompt)
}

// getNoDefault reads a value from storage.
// Returns ("", OriginNotSet, nil) when the key has not been set.
func (c *Config) getNoDefault(key string) (string, Origin, error) {
	v, err := c.storage.Get(key)
	if err != nil {
		return "", OriginNotSet, err
	}
	if v == nil {
		return "", OriginNotSet, nil
	}
	return *v, OriginUser, nil
}
