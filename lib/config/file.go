package config

import (
	"os"
)

func fileRead(filepath string) (*string, error) {
	content, err := os.ReadFile(filepath)
	if err == nil {
		contentStr := string(content)
		return &contentStr, nil
	}

	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, err
}

func fileWrite(filepath string, value string) error {
	return os.WriteFile(filepath, []byte(value), modePrivate)
}

func fileDelete(filepath string) error {
	return os.Remove(filepath)
}
