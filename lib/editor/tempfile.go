package editor

import (
	"fmt"
	"os"
	"path/filepath"
)

// createEditFile returns the path the editor should open and a cleanup
// function. The caller must invoke cleanup after reading back the edited
// content.
func createEditFile() (string, func(), error) {
	dir, err := os.MkdirTemp("", "editor-scratch-*")
	if err != nil {
		return "", nil, fmt.Errorf("mkdirtemp: %w", err)
	}

	// Ensure owner-only access regardless of umask.
	if err := os.Chmod(dir, 0700); err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("chmod tmpdir: %w", err)
	}

	path := filepath.Join(dir, "edit")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("create tmpfile: %w", err)
	}
	f.Close()

	cleanup := func() { os.RemoveAll(dir) }
	return path, cleanup, nil
}
