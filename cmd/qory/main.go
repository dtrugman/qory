package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dtrugman/qory/lib/config"
)

const (
        appName = "Qory"

        argConfig = "--config"

        argAPIKey = "api-key"
        argBaseURL = "base-url"
        argModel = "model"

        argGet = "get"
        argSet = "set"
        argUnset = "unset"
)

var (
        ErrorBadArguments = errors.New("bad arguments")
)

func usage(arg0 string) {
        fmt.Printf("%s: A language model in your terminal\n", appName)
        fmt.Printf("\n")
        fmt.Printf("Usage:  %s [input]...\n", arg0)
        fmt.Printf("        %s --config [options]\n", arg0)
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
        fmt.Printf("To see the configuration options, please run:\n")
        fmt.Printf("    > %s --config\n", arg0)
        fmt.Printf("\n")
}

func usageConfigAPIKey(arg0 string) {
        fmt.Printf("API key configuration\n")
        fmt.Printf("\n")
        fmt.Printf("Usage:  %s %s %s set [value]\n", arg0, argConfig, argBaseURL)
        fmt.Printf("        %s %s %s unset\n", arg0, argConfig, argBaseURL)
        fmt.Printf("        %s %s %s get\n", arg0, argConfig, argBaseURL)
        fmt.Printf("\n")
        fmt.Printf("Configure the API key to use when sending requests to the model.\n")
        fmt.Printf("When the value is not set, it is read from the OPENAI_API_KEY env var.\n")
        fmt.Printf("\n")
}

func usageConfigBaseURL(arg0 string) {
        fmt.Printf("Base URL configuration\n")
        fmt.Printf("\n")
        fmt.Printf("Usage:  %s %s %s set [value]\n", arg0, argConfig, argBaseURL)
        fmt.Printf("        %s %s %s unset\n", arg0, argConfig, argBaseURL)
        fmt.Printf("        %s %s %s get\n", arg0, argConfig, argBaseURL)
        fmt.Printf("\n")
        fmt.Printf("The base URL to use when sending requests to the model.\n")
        fmt.Printf("When the value is not set, it is read from the OPENAI_BASE_URL env var.\n")
        fmt.Printf("\n")
}

func usageConfigModel(arg0 string) {
        fmt.Printf("Model configuration\n")
        fmt.Printf("\n")
        fmt.Printf("Usage:  %s %s %s set [value]\n", arg0, argConfig, argModel)
        fmt.Printf("        %s %s %s unset\n", arg0, argConfig, argModel)
        fmt.Printf("        %s %s %s get\n", arg0, argConfig, argModel)
        fmt.Printf("\n")
        fmt.Printf("The ChatGPT model to use, e.g. gpt-4o\n")
        fmt.Printf("\n")
}

func usageConfig(arg0 string) {
        fmt.Printf("%s configuration\n", appName)
        fmt.Printf("\n")
        fmt.Printf("Usage:  %s %s [param]\n", arg0, argConfig)
        fmt.Printf("\n")
        fmt.Printf("        %-10s    Configure API key to use\n", argAPIKey)
        fmt.Printf("        %-10s    Configure the base URL of the model\n", argBaseURL)
        fmt.Printf("        %-10s    Configure the model to use\n", argModel)
        fmt.Printf("\n")
}

func validateNothing(value string) error {
        return nil
}

func validateBaseURL(value string) error {
        if !strings.HasSuffix(value, "/") {
                return fmt.Errorf("must end with a '/'")
        } else {
                return nil
        }
}

func promptUserInput(prompt string) (string, error) {
        fmt.Print(prompt)

        reader := bufio.NewReader(os.Stdin)
        input, err := reader.ReadString('\n')
        if err != nil {
            return "", err
        }

        return strings.TrimSpace(input), nil
}

func runConfigKey(
        args []string,
        conf config.Config,
        key string,
        desc string,
        valueValidator func(string) error,
        keyUsage func(string),
) error {
        if len(args) < 4 {
                keyUsage(args[0])
                return ErrorBadArguments
        }
        op := args[3]

        if op == argGet {
                if value, err := conf.Get(key); err != nil {
                        return err
                } else if value == nil {
                        fmt.Printf("No value for %s\n", desc)
                        return nil
                } else {
                        fmt.Printf("%s\n", *value)
                        return nil
                }
        }

        if op == argUnset {
                if err := conf.Unset(key); err != nil {
                        return err
                } else {
                        fmt.Printf("Successfuly unset %s\n", desc)
                        return nil
                }
        }

        if op == argSet {
                var err error
                var value string
                if len(args) != 5 {
                        value, err = promptUserInput("Enter value: ")
                } else {
                        value, err = args[4], nil
                }

                if err != nil {
                        return err
                }

                if err := valueValidator(value); err != nil {
                        return err
                }

                if err := conf.Set(key, value); err != nil {
                        return err
                } else {
                        fmt.Printf("Successfuly set %s\n", desc)
                        return nil
                }
        }

        keyUsage(args[0])
        return ErrorBadArguments
}

func runConfig(args []string, conf config.Config) error {
        if len(args) < 3 {
                usageConfig(args[0])
                return ErrorBadArguments
        }
        key := args[2]

        if key == argAPIKey {
                return runConfigKey(
                    args, conf, config.APIKey, "API key",
                    validateNothing, usageConfigAPIKey)
        } else if key == argBaseURL {
                return runConfigKey(
                    args, conf, config.BaseURL, "base URL",
                    validateBaseURL, usageConfigBaseURL)
        } else if key == argModel {
                return runConfigKey(
                    args, conf, config.Model, "model",
                    validateNothing, usageConfigModel)
        } else {
                usageConfig(args[0])
                return ErrorBadArguments
        }
}

func run(args []string) error {
        if len(args) < 2 {
                usage(args[0])
                return ErrorBadArguments
        }
        action := args[1]

        conf, err := config.NewManager()
        if err != nil {
                return err
        }

        if action == argConfig {
                return runConfig(args, conf)
        } else { // no action, an implicit query
                return runQuery(args, conf)
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
