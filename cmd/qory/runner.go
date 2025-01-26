package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/model"
	"github.com/dtrugman/qory/lib/session"
	"github.com/google/uuid"
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

func runQuery(args []string, client model.Client, sessions session.Manager, conf config.Config) error {
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
	response, err := client.Query(*modelName, systemPrompt, prompt)
	if err != nil {
		return nil // Error is reported inside atm
	}

	messages := make([]session.Message, 0)

	if systemPrompt != nil {
		messages = append(messages, session.Message{
			Role:    "system",
			Content: *systemPrompt,
		})
	}

	messages = append(messages, session.Message{
		Role:    "user",
		Content: prompt,
	})

	messages = append(messages, session.Message{
		Role:    "assistant",
		Content: response,
	})

	session := session.Session{
		Messages: messages,
	}

	id := uuid.New()
	sessions.Store(id.String(), session)

	return nil
}
