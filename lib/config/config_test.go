package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- FileStorage tests ----

func newTestStorage(t *testing.T) *FileStorage {
	t.Helper()
	dir := t.TempDir()
	s, err := NewFileStorage(dir)
	require.NoError(t, err)
	return s
}

func TestFileStorage_GetMissing(t *testing.T) {
	s := newTestStorage(t)
	result, err := s.Get(APIKey)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFileStorage_SetAndGet(t *testing.T) {
	s := newTestStorage(t)
	want := "sk-test-key"
	require.NoError(t, s.Set(APIKey, want))
	got, err := s.Get(APIKey)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestFileStorage_SetAndUnset(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.Set(Model, "gpt-4"))
	require.NoError(t, s.Unset(Model))
	result, err := s.Get(Model)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFileStorage_SetOverwrite(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.Set(BaseURL, "https://old.example.com"))
	want := "https://new.example.com"
	require.NoError(t, s.Set(BaseURL, want))
	got, err := s.Get(BaseURL)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestFileStorage_KeysAreIndependent(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.Set(APIKey, "key-value"))
	require.NoError(t, s.Set(Model, "model-value"))

	apiKey, err := s.Get(APIKey)
	require.NoError(t, err)
	assert.Equal(t, "key-value", *apiKey)

	model, err := s.Get(Model)
	require.NoError(t, err)
	assert.Equal(t, "model-value", *model)
}

func TestFileStorage_GetConfigSubdir(t *testing.T) {
	s := newTestStorage(t)
	path, err := s.GetConfigSubdir("sessions")
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	path2, err := s.GetConfigSubdir("sessions")
	require.NoError(t, err)
	assert.Equal(t, path, path2)
}

// ---- Config tests ----

func newTestConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	c, err := NewConfig(dir)
	require.NoError(t, err)
	return c
}

func TestConfig_GetConfigSubdir(t *testing.T) {
	c := newTestConfig(t)
	path, err := c.GetConfigSubdir("sessions")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestConfig_Editor_Default(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, DefaultEditor, val)
	assert.Equal(t, OriginDefault, origin)
}

func TestConfig_Editor_StoredValue(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	require.NoError(t, c.SetEditor("nvim"))
	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, "nvim", val)
	assert.Equal(t, OriginUser, origin)
}

func TestConfig_Editor_VisualEnvVar(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "emacs")
	t.Setenv("EDITOR", "nano")

	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, "emacs", val)
	assert.Equal(t, OriginEnv, origin)
}

func TestConfig_Editor_EditorEnvVar(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")

	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, "nano", val)
	assert.Equal(t, OriginEnv, origin)
}

func TestConfig_Editor_EnvVarTakesPrecedenceOverStored(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "emacs")
	require.NoError(t, c.SetEditor("nvim"))

	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, "emacs", val)
	assert.Equal(t, OriginEnv, origin)
}

func TestConfig_Editor_UnsetRestoresDefault(t *testing.T) {
	c := newTestConfig(t)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	require.NoError(t, c.SetEditor("nvim"))
	require.NoError(t, c.UnsetEditor())
	val, origin, err := c.Editor()
	require.NoError(t, err)
	assert.Equal(t, DefaultEditor, val)
	assert.Equal(t, OriginDefault, origin)
}

func TestConfig_HistorySize_Default(t *testing.T) {
	c := newTestConfig(t)
	size, origin, err := c.HistorySize()
	require.NoError(t, err)
	assert.Equal(t, DefaultHistorySize, size)
	assert.Equal(t, OriginDefault, origin)
}

func TestConfig_HistorySize_StoredValue(t *testing.T) {
	c := newTestConfig(t)
	require.NoError(t, c.SetHistorySize("100"))
	size, origin, err := c.HistorySize()
	require.NoError(t, err)
	assert.Equal(t, 100, size)
	assert.Equal(t, OriginUser, origin)
}

func TestConfig_HistorySize_UnsetRestoresDefault(t *testing.T) {
	c := newTestConfig(t)
	require.NoError(t, c.SetHistorySize("100"))
	require.NoError(t, c.UnsetHistorySize())
	size, origin, err := c.HistorySize()
	require.NoError(t, err)
	assert.Equal(t, DefaultHistorySize, size)
	assert.Equal(t, OriginDefault, origin)
}

func TestConfig_SetHistorySize_RejectsZero(t *testing.T) {
	c := newTestConfig(t)
	assert.Error(t, c.SetHistorySize("0"))
}

func TestConfig_SetHistorySize_RejectsNegative(t *testing.T) {
	c := newTestConfig(t)
	assert.Error(t, c.SetHistorySize("-1"))
}

func TestConfig_SetHistorySize_RejectsNonInteger(t *testing.T) {
	c := newTestConfig(t)
	assert.Error(t, c.SetHistorySize("abc"))
}

func TestConfig_SetMode_AcceptsValid(t *testing.T) {
	for _, v := range []string{ModeNew, ModeLast} {
		t.Run(v, func(t *testing.T) {
			c := newTestConfig(t)
			require.NoError(t, c.SetMode(v))
			got, origin, err := c.Mode()
			require.NoError(t, err)
			assert.Equal(t, v, got)
			assert.Equal(t, OriginUser, origin)
		})
	}
}

func TestConfig_SetMode_RejectsInvalid(t *testing.T) {
	c := newTestConfig(t)
	assert.Error(t, c.SetMode("invalid"))
}

func TestConfig_Mode_NotSetReturnsNotSet(t *testing.T) {
	c := newTestConfig(t)
	val, origin, err := c.Mode()
	require.NoError(t, err)
	assert.Equal(t, "", val)
	assert.Equal(t, OriginNotSet, origin)
}

func TestConfig_SetBaseURL_NormalizesTrailingSlash(t *testing.T) {
	c := newTestConfig(t)
	require.NoError(t, c.SetBaseURL("https://api.example.com"))
	got, _, err := c.BaseURL()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/", got)
}

func TestConfig_SetBaseURL_PreservesExistingTrailingSlash(t *testing.T) {
	c := newTestConfig(t)
	require.NoError(t, c.SetBaseURL("https://api.example.com/"))
	got, _, err := c.BaseURL()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/", got)
}

func TestConfig_APIKey_NotSet(t *testing.T) {
	c := newTestConfig(t)
	val, origin, err := c.APIKey()
	require.NoError(t, err)
	assert.Equal(t, "", val)
	assert.Equal(t, OriginNotSet, origin)
}

func TestConfig_APIKey_Set(t *testing.T) {
	c := newTestConfig(t)
	require.NoError(t, c.SetAPIKey("sk-test"))
	val, origin, err := c.APIKey()
	require.NoError(t, err)
	assert.Equal(t, "sk-test", val)
	assert.Equal(t, OriginUser, origin)
}
