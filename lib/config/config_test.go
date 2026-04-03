package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()

	dir := t.TempDir()
	conf, err := NewConfig(dir)
	require.NoError(t, err)

	return conf
}

func TestConfigGetMissing(t *testing.T) {
	conf := newTestConfig(t)

	result, err := conf.Get(APIKey)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestConfigSetAndGet(t *testing.T) {
	conf := newTestConfig(t)
	want := "sk-test-key"

	err := conf.Set(APIKey, want)
	require.NoError(t, err)

	got, err := conf.Get(APIKey)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestConfigSetAndUnset(t *testing.T) {
	conf := newTestConfig(t)

	err := conf.Set(Model, "gpt-4")
	require.NoError(t, err)

	err = conf.Unset(Model)
	require.NoError(t, err)

	result, err := conf.Get(Model)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestConfigSetOverwrite(t *testing.T) {
	conf := newTestConfig(t)

	err := conf.Set(BaseURL, "https://old.example.com")
	require.NoError(t, err)

	want := "https://new.example.com"
	err = conf.Set(BaseURL, want)
	require.NoError(t, err)

	got, err := conf.Get(BaseURL)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestConfigKeysAreIndependent(t *testing.T) {
	conf := newTestConfig(t)

	err := conf.Set(APIKey, "key-value")
	require.NoError(t, err)

	err = conf.Set(Model, "model-value")
	require.NoError(t, err)

	apiKey, err := conf.Get(APIKey)
	require.NoError(t, err)
	assert.Equal(t, "key-value", *apiKey)

	model, err := conf.Get(Model)
	require.NoError(t, err)
	assert.Equal(t, "model-value", *model)
}

func TestConfigGetConfigSubdir(t *testing.T) {
	conf := newTestConfig(t)

	path, err := conf.GetConfigSubdir("sessions")
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	path2, err := conf.GetConfigSubdir("sessions")
	require.NoError(t, err)
	assert.Equal(t, path, path2)
}
