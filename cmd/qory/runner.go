package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/message"
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

func buildUserPrompt(args []string) string {
	var promptBuilder strings.Builder

	for _, arg := range args {
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

func usageQueryWithSession(arg0 string) {
	fmt.Printf("Usage:  %s %s|%s session-id <args...>\n", arg0, argSession, argSessionShort)
	fmt.Printf("\n")
}

func getSession(sessionManager session.Manager, sessionID string) (session.Session, error) {
	s, err := sessionManager.Load(sessionID)
	if err == nil {
		return s, nil
	}

	if !errors.Is(err, session.ErrNotFound) {
		return session.Session{}, err
	}

	return session.NewSession(), nil
}

func runQueryInner(
	args []string,
	client model.Client,
	sessionManager session.Manager,
	sessionID string,
	conf config.Config,
) error {
	sess, err := getSession(sessionManager, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	modelName, err := conf.Get(config.Model)
	if modelName == nil {
		return fmt.Errorf("model is not set")
	} else if err != nil {
		return fmt.Errorf("get base URL failed: %w", err)
	}

	if len(sess.Messages) == 0 {
		systemPrompt, err := conf.Get(config.Prompt)
		if err != nil {
			return fmt.Errorf("get base URL failed: %w", err)
		}

		if systemPrompt != nil {
			sess.AddMessage(message.NewSystemMessage(*systemPrompt))
		}
	}

	userPrompt := buildUserPrompt(args)
	sess.AddMessage(message.NewUserMessage(userPrompt))

	response, err := client.Query(*modelName, sess.Messages)
	if err != nil {
		return nil // Error is reported inside atm
	}

	sess.AddMessage(message.NewAssistantMessage(response))

	if err = sessionManager.Store(sessionID, sess); err != nil {
		fmt.Printf("Store session failed: %v", err)
	}
	return nil
}

func runQueryWithSession(
	args []string,
	client model.Client,
	sessionManager session.Manager,
	conf config.Config,
) error {
	if len(args) < 4 {
		usageQueryWithSession(args[0])
		return ErrorBadArguments
	}
	sessionID := args[2]
	args = args[3:] // Leave only relevant arguments

	return runQueryInner(args, client, sessionManager, sessionID, conf)
}

func runQuery(
	args []string,
	client model.Client,
	sessionManager session.Manager,
	conf config.Config,
) error {
	sessionID := uuid.New().String()
	args = args[1:] // Remove arg0
	return runQueryInner(args, client, sessionManager, sessionID, conf)
}
