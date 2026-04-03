package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReadNonExistent(t *testing.T) {
	dir := t.TempDir()
	result, err := fileRead(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFileWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key")
	want := "hello"

	err := fileWrite(path, want)
	require.NoError(t, err)

	got, err := fileRead(path)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestFileWriteSetsPrivatePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key")

	err := fileWrite(path, "value")
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(modePrivate), info.Mode().Perm())
}

func TestFileDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key")

	err := fileWrite(path, "value")
	require.NoError(t, err)

	err = fileDelete(path)
	require.NoError(t, err)

	result, err := fileRead(path)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFileOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key")

	err := fileWrite(path, "first")
	require.NoError(t, err)

	err = fileWrite(path, "second")
	require.NoError(t, err)

	got, err := fileRead(path)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "second", *got)
}
