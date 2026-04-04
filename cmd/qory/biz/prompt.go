package biz

import (
	"os"
	"strings"
)

// buildUserPrompt converts a list of CLI inputs into a single prompt string.
// Each element is treated as a file path first; if the file can be read its
// contents become a separate newline-delimited part. Consecutive plain-text
// arguments are joined with spaces into a single part.
//
// Examples:
//
//	["how", "are", "you"]        → "how are you"
//	["explain", "file.txt"]      → "explain\n<file contents>"
//	["file1.txt", "file2.txt"]   → "<file1 contents>\n<file2 contents>"
func buildUserPrompt(inputs []string) string {
	var parts []string
	var textTokens []string

	flushText := func() {
		if len(textTokens) > 0 {
			parts = append(parts, strings.Join(textTokens, " "))
			textTokens = nil
		}
	}

	for _, arg := range inputs {
		bytes, err := os.ReadFile(arg)
		if err == nil {
			flushText()
			parts = append(parts, string(bytes))
		} else {
			textTokens = append(textTokens, arg)
		}
	}
	flushText()

	return strings.Join(parts, "\n")
}
