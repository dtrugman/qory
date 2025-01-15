package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/model"
)

func readFile(filepath string) (*string, error) {
	content, err := os.ReadFile(filepath)
	if err == nil {
		contentStr := string(content)
		return &contentStr, nil
	}

	if os.IsNotExist(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

func buildPrompt(args []string) string {
	var promptBuilder strings.Builder

	for i, arg := range args {
		if i == 0 {
			continue
		}

		bytes, err := os.ReadFile(arg)
		if err == nil {
			promptBuilder.Write(bytes)
			promptBuilder.WriteString("\n")
		} else {
			promptBuilder.WriteString(arg)
			promptBuilder.WriteString("\n")
		}
	}

	return promptBuilder.String()
}

func runQuery(args []string, client model.Client, conf config.Config) error {
	modelName, err := conf.Get(config.Model)
	if modelName == nil {
		return fmt.Errorf("model is not set")
	} else if err != nil {
		return fmt.Errorf("get base URL failed: %w", err)
	}

	systemPrompt, err := conf.Get(config.Prompt)
	if err != nil {
		return fmt.Errorf("get base URL failed: %w", err)
	}

	prompt := buildPrompt(args)
	client.Query(*modelName, systemPrompt, prompt)
	return nil
}
