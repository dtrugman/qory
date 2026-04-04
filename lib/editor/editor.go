package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Edit opens the given editor binary on a temp file, waits for the editor to
// exit, and returns the trimmed content. Returns ("", nil) if the user saved
// an empty file.
func Edit(editorBin string) (string, error) {
	bytes, err := Open(editorBin)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(bytes)) == "" {
		return "", nil
	}
	return string(bytes), nil
}

// Open opens the given editor binary for editing and returns the result.
func Open(editorBin string) ([]byte, error) {
	path, cleanup, err := createEditFile()
	if err != nil {
		return nil, fmt.Errorf("create edit file: %w", err)
	}
	defer cleanup()

	cmd := exec.Command(editorBin, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run editor: %w", err)
	}

	return os.ReadFile(path)
}
