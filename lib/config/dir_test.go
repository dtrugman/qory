package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDirCreatesQorySubdir(t *testing.T) {
	userDir := t.TempDir()

	got, err := getConfigDir(userDir)
	require.NoError(t, err)

	want := filepath.Join(userDir, dirDotQory)
	assert.Equal(t, want, got)

	info, err := os.Stat(got)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetConfigDirIsIdempotent(t *testing.T) {
	userDir := t.TempDir()

	first, err := getConfigDir(userDir)
	require.NoError(t, err)

	second, err := getConfigDir(userDir)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestGetOrCreateDirCreatesNew(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "subdir")

	got, err := getOrCreateDir(path)
	require.NoError(t, err)
	assert.Equal(t, path, got)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetOrCreateDirExistingDir(t *testing.T) {
	path := t.TempDir()

	got, err := getOrCreateDir(path)
	require.NoError(t, err)
	assert.Equal(t, path, got)
}

func TestGetOrCreateDirFailsOnFile(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "notadir")

	err := os.WriteFile(path, []byte("x"), 0600)
	require.NoError(t, err)

	_, err = getOrCreateDir(path)
	assert.Error(t, err)
}
