package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/model"
	"github.com/dtrugman/qory/lib/session"
)

const (
	appName = "Qory"

	argConfig       = "--config"
	argConfigShort  = "-c"
	argVersion      = "--version"
	argVersionShort = "-v"
	argHistory      = "--history"
	argHistoryShort = "-h"
	argSession      = "--session"
	argSessionShort = "-s"
	argLastSession  = "^"

	argAPIKey  = "api-key"
	argBaseURL = "base-url"
	argModel   = "model"
	argPrompt  = "prompt"

	argGet   = "get"
	argSet   = "set"
	argUnset = "unset"

	historyLength = 10

	sessionUnnamedLimit = 10
)

var (
	version = "dev"

	ErrorBadArguments = errors.New("bad arguments")
)

func usage(arg0 string) {
	fmt.Printf("%s: A language model in your terminal\n", appName)
	fmt.Printf("\n")
	fmt.Printf("Usage:  %s [%s|%s session-id] <input...>\n", arg0, argSessionShort, argSession)
	fmt.Printf("        %s %s|%s\n", arg0, argVersionShort, argVersion)
	fmt.Printf("        %s %s|%s [session-id]\n", arg0, argHistoryShort, argHistory)
	fmt.Printf("        %s %s|%s [options]\n", arg0, argConfigShort, argConfig)
	fmt.Printf("\n")
	fmt.Printf("%s is a tool for accessing language models directly from your CLI\n", appName)
	fmt.Printf("allowing you to specify free-form queries and any local file as context\n")
	fmt.Printf("\n")
	fmt.Printf("The simplest query, would look like this:\n")
	fmt.Printf("    > %s \"Please create a basic OpenAPI yaml template\"\n", arg0)
	fmt.Printf("\n")
	fmt.Printf("A query with an attached local file as input, would look like this:\n")
	fmt.Printf("    > %s \"Please add a health check to my OpenAPI spec\" openapi.yaml\n", arg0)
	fmt.Printf("\n")
	fmt.Printf("A query that creates a named session looks like this:\n")
	fmt.Printf("    > %s %s spec \"Please add a health check to my OpenAPI spec\" openapi.yaml\n", arg0, argSession)
	fmt.Printf("    ... some output from model ...\n")
	fmt.Printf("    > %s %s spec \"Please define a new parameter for the body\"\n", arg0, argSession)
	fmt.Printf("\n")
	fmt.Printf("Follow up on last query\n")
	fmt.Printf("    > %s %s \"Please use argparse for the arguments\"\n", arg0, argLastSession)
	fmt.Printf("\n")
	fmt.Printf("To see your last queries, just run:\n")
	fmt.Printf("    > %s %s [session-id]\n", arg0, argHistory)
	fmt.Printf("\n")
	fmt.Printf("To see the configuration options, please run:\n")
	fmt.Printf("    > %s %s\n", arg0, argConfig)
	fmt.Printf("\n")
}

func usageConfigSetUnsetGet(arg0 string, key string, title string, extra []string) {
	fmt.Printf("%s\n", title)
	fmt.Printf("\n")
	fmt.Printf("Usage:  %s %s %s set [value]\n", arg0, argConfig, key)
	fmt.Printf("        %s %s %s unset\n", arg0, argConfig, key)
	fmt.Printf("        %s %s %s get\n", arg0, argConfig, key)
	fmt.Printf("\n")
	for _, line := range extra {
		fmt.Printf("%s\n", line)
	}
	fmt.Printf("\n")
}

func usageConfigAPIKey(arg0 string) {
	title := "API key configuration"
	extra := []string{
		"Configure the API key to use when sending requests to the model.",
		"When the value is not set, it is read from the OPENAI_API_KEY env var.",
	}
	usageConfigSetUnsetGet(arg0, argAPIKey, title, extra)
}

func usageConfigBaseURL(arg0 string) {
	title := "Base URL configuration"
	extra := []string{
		"The base URL to use when sending requests to the model.",
		"When the value is not set, it is read from the OPENAI_BASE_URL env var.",
	}
	usageConfigSetUnsetGet(arg0, argBaseURL, title, extra)
}

func usageConfigModel(arg0 string) {
	title := "Model configuration"
	extra := []string{"The ChatGPT model to use, e.g. gpt-4o"}
	usageConfigSetUnsetGet(arg0, argModel, title, extra)
}

func usageConfigPrompt(arg0 string) {
	title := "Persistent prompt configuration"
	extra := []string{
		"A system prompt to add to all your requests.",
		"For example, \"Do not explain, just provide the essence of the request\"",
	}
	usageConfigSetUnsetGet(arg0, argPrompt, title, extra)
}

func usageConfig(arg0 string) {
	fmt.Printf("%s configuration\n", appName)
	fmt.Printf("\n")
	fmt.Printf("Usage:  %s %s [param]\n", arg0, argConfig)
	fmt.Printf("\n")
	fmt.Printf("    %-10s    Configure API key to use\n", argAPIKey)
	fmt.Printf("    %-10s    Configure the base URL of the model\n", argBaseURL)
	fmt.Printf("    %-10s    Configure the model to use\n", argModel)
	fmt.Printf("    %-10s    Configure a persistent system prompt\n", argPrompt)
	fmt.Printf("\n")
}

func runConfigKey(
	args []string,
	param Param,
	showUsage func(arg string),
) error {
	if len(args) < 4 {
		showUsage(args[0])
		return ErrorBadArguments
	}
	op := args[3]

	if op == argGet {
		return param.Get()
	}

	if op == argUnset {
		return param.Unset()
	}

	if op == argSet {
		var value *string = nil
		if len(args) >= 5 {
			value = &args[4]
		}

		return param.Set(value)
	}

	showUsage(args[0])
	return ErrorBadArguments
}

func runVersion() error {
	fmt.Printf("%s version %s\n", appName, version)
	return nil
}

func usageHistory(arg0 string) {
	fmt.Printf("%s history\n", appName)
	fmt.Printf("\n")
	fmt.Printf("Usage:  %s %s [session-id]\n", arg0, argHistory)
	fmt.Printf("\n")
	fmt.Printf("    Run without a session ID to see snippets of latest %d sessions\n", historyLength)
	fmt.Printf("    Or specify a session ID to see the entire session\n")
	fmt.Printf("\n")
}

func runHistorySpecific(sessionID string, sessionManager session.Manager) error {
	sess, err := sessionManager.Load(sessionID)
	if err != nil {
		return err
	}

	for _, m := range sess.Messages {
		fmt.Printf("=== %s ===\n", strings.ToUpper(string(m.Role)))
		fmt.Printf("%s\n", m.Content)
	}

	return nil
}

func runHistoryLast(sessionManager session.Manager) error {
	sessions, err := sessionManager.Enum(historyLength)
	if err != nil {
		return err
	}

	for id, preview := range sessions {
		fmt.Printf("=== %s (%s) ===\n", id, preview.UpdatedAt.Format(time.RFC822))

		// If there's a newline at the end, assume it's the end of the
		snippet, _ := strings.CutSuffix(preview.Snippet, "\n")
		fmt.Printf("%s\n", snippet)
	}

	return nil
}

func runHistory(args []string, sessionManager session.Manager) error {
	if len(args) > 3 {
		usageHistory(args[0])
		return ErrorBadArguments
	} else if len(args) == 3 {
		sessionID := args[2]
		return runHistorySpecific(sessionID, sessionManager)
	} else {
		return runHistoryLast(sessionManager)
	}
}

func runConfig(args []string, client model.Client, conf config.Config) error {
	if len(args) < 3 {
		usageConfig(args[0])
		return ErrorBadArguments
	}
	key := args[2]

	if key == argAPIKey {
		param := NewParam(conf, config.APIKey)
		return runConfigKey(args, param, usageConfigAPIKey)
	} else if key == argBaseURL {
		param := NewParamBaseURL(conf, config.BaseURL)
		return runConfigKey(args, param, usageConfigBaseURL)
	} else if key == argModel {
		param := NewParamModel(conf, config.Model, client)
		return runConfigKey(args, param, usageConfigModel)
	} else if key == argPrompt {
		param := NewParam(conf, config.Prompt)
		return runConfigKey(args, param, usageConfigPrompt)
	} else {
		usageConfig(args[0])
		return ErrorBadArguments
	}
}

func buildClient(conf config.Config) (model.Client, error) {
	apiKey, err := conf.Get(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("get API key failed: %w", err)
	}

	baseURL, err := conf.Get(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("get base URL failed: %w", err)
	}

	client := model.NewClient(apiKey, baseURL)
	return client, nil
}

func buildSessionManager(conf config.Config) (session.Manager, error) {
	dir, err := conf.GetConfigSubdir(session.SessionsDirName)
	if err != nil {
		return nil, err
	}

	return session.NewManager(dir)
}

func run(args []string) error {
	if len(args) < 2 {
		usage(args[0])
		return ErrorBadArguments
	}
	action := args[1]

	conf, err := config.NewConfig()
	if err != nil {
		return err
	}

	client, err := buildClient(conf)
	if err != nil {
		return err
	}

	sessionManager, err := buildSessionManager(conf)
	if err != nil {
		return err
	}

	if action == argVersion || action == argVersionShort {
		return runVersion()
	} else if action == argConfig || action == argConfigShort {
		return runConfig(args, client, conf)
	} else if action == argHistory || action == argHistoryShort {
		return runHistory(args, sessionManager)
	} else if action == argSession || action == argSessionShort {
		return runQueryWithSession(args, client, sessionManager, conf)
	} else if action == argLastSession {
		return runQueryWithLastSession(args, client, sessionManager, conf)
	} else { // an implicit query without a session
		return runQuery(args, client, sessionManager, conf)
	}
}

func main() {
	if err := run(os.Args); err != nil {
		if !errors.Is(err, ErrorBadArguments) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		os.Exit(1)
	}

	os.Exit(0)
}
