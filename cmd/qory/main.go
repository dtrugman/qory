package main

import (
	"fmt"
	"os"

	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/model"
	"github.com/dtrugman/qory/lib/profile"
	"github.com/dtrugman/qory/lib/session"
)

func buildClient(conf biz.Config) (*model.Client, error) {
	apiKeyStr, _, err := conf.APIKey()
	if err != nil {
		return nil, fmt.Errorf("get API key failed: %w", err)
	}

	baseURLStr, _, err := conf.BaseURL()
	if err != nil {
		return nil, fmt.Errorf("get base URL failed: %w", err)
	}

	var apiKey *string
	if apiKeyStr != "" {
		apiKey = &apiKeyStr
	}

	var baseURL *string
	if baseURLStr != "" {
		baseURL = &baseURLStr
	}

	client := model.NewClient(apiKey, baseURL)
	return client, nil
}

func buildSessionManager(conf biz.Config) (*session.Manager, error) {
	dir, err := conf.GetConfigSubdir(session.SessionsDirName)
	if err != nil {
		return nil, err
	}

	manager := session.NewManager(dir)
	return manager, nil
}

func buildQory() (*biz.Qory, error) {
	userDir, err := profile.GetUserDir()
	if err != nil {
		return nil, fmt.Errorf("profile: %w", err)
	}

	conf, err := config.NewConfig(userDir)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	client, err := buildClient(conf)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	sm, err := buildSessionManager(conf)
	if err != nil {
		return nil, fmt.Errorf("session manager: %w", err)
	}

	q := biz.NewQory(conf, client, sm)
	return q, nil
}

func main() {
	q, err := buildQory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := newRootCmd(q)
	root.AddCommand(
		newVersionCmd(q),
		newHistoryCmd(q),
		newConfigCmd(q),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
