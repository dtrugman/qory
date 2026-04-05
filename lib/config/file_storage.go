package config

import (
	"path/filepath"
)

const (
	modePrivate = 0600 // Owner can RW, others have no access
)

const ( // Public configuration keys
	APIKey      = "api_key"
	BaseURL     = "base_url"
	Model       = "model"
	Prompt      = "prompt"
	Mode        = "mode"
	Editor      = "editor"
	HistorySize = "history_size"
)

const ( // Valid values for Mode
	ModeNew  = "new"
	ModeLast = "last"
)

// FileStorage persists each configuration value as a separate file under the
// application config directory.
type FileStorage struct {
	dir string
}

func NewFileStorage(userDir string) (*FileStorage, error) {
	configDir, err := getConfigDir(userDir)
	if err != nil {
		return nil, err
	}
	return &FileStorage{dir: configDir}, nil
}

func (s *FileStorage) GetConfigSubdir(name string) (string, error) {
	path := filepath.Join(s.dir, name)
	return getOrCreateDir(path)
}

func (s *FileStorage) Get(key string) (*string, error) {
	path := filepath.Join(s.dir, key)
	return fileRead(path)
}

func (s *FileStorage) Set(key string, value string) error {
	path := filepath.Join(s.dir, key)
	return fileWrite(path, value)
}

func (s *FileStorage) Unset(key string) error {
	path := filepath.Join(s.dir, key)
	return fileDelete(path)
}
