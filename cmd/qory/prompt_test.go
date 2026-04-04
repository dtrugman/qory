package biz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "qory-test-*.txt")
	require.NoError(t, err)

	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return f.Name()
}

func Test_buildUserPrompt_TextOnly(t *testing.T) {
	result := buildUserPrompt([]string{"how", "are", "you"})
	assert.Equal(t, "how are you", result)
}

func Test_buildUserPrompt_SingleText(t *testing.T) {
	result := buildUserPrompt([]string{"hello"})
	assert.Equal(t, "hello", result)
}

func Test_buildUserPrompt_Empty(t *testing.T) {
	result := buildUserPrompt([]string{})
	assert.Equal(t, "", result)
}

func Test_buildUserPrompt_SingleFile(t *testing.T) {
	path := writeTemp(t, "file content")
	result := buildUserPrompt([]string{path})
	assert.Equal(t, "file content", result)
}

func Test_buildUserPrompt_TextThenFile(t *testing.T) {
	path := writeTemp(t, "file content")
	result := buildUserPrompt([]string{"explain", path})
	assert.Equal(t, "explain\nfile content", result)
}

func Test_buildUserPrompt_FileThenText(t *testing.T) {
	path := writeTemp(t, "file content")
	result := buildUserPrompt([]string{path, "summarize", "briefly"})
	assert.Equal(t, "file content\nsummarize briefly", result)
}

func Test_buildUserPrompt_MultipleFiles(t *testing.T) {
	path1 := writeTemp(t, "first file")
	path2 := writeTemp(t, "second file")
	result := buildUserPrompt([]string{path1, path2})
	assert.Equal(t, "first file\nsecond file", result)
}

func Test_buildUserPrompt_TextBetweenFiles(t *testing.T) {
	path1 := writeTemp(t, "file one")
	path2 := writeTemp(t, "file two")
	result := buildUserPrompt([]string{path1, "compare", "these", path2})
	assert.Equal(t, "file one\ncompare these\nfile two", result)
}

func Test_buildUserPrompt_NonExistentPathTreatedAsText(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "does-not-exist.txt")
	result := buildUserPrompt([]string{"review", nonExistent})
	assert.Equal(t, "review "+nonExistent, result)
}
