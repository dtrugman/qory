package main

import (
	"fmt"

	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/spf13/cobra"
)

func newConfigCmd(q *biz.Qory) *cobra.Command {
	const noLong = ""

	cmdAPIKey := newConfigKeyCmd("api-key",
		"API key for the model provider", noLong,
		q.ConfigGetAPIKey, q.ConfigSetAPIKey, q.ConfigUnsetAPIKey,
		promptUserInput,
	)

	cmdBaseURL := newConfigKeyCmd("base-url",
		"Base URL for the model provider", noLong,
		q.ConfigGetBaseURL, q.ConfigSetBaseURL, q.ConfigUnsetBaseURL, promptUserInput,
	)

	cmdModel := newConfigKeyCmd("model",
		"Model to use for queries", noLong,
		q.ConfigGetModel, q.ConfigSetModel, q.ConfigUnsetModel,
		func() (string, error) {
			models, err := q.AvailableModels()
			if err != nil {
				return "", err
			}
			return promptModel(models)
		},
	)

	cmdPrompt := newConfigKeyCmd("prompt",
		"Persistent system prompt prepended to every new session", noLong,
		q.ConfigGetPrompt, q.ConfigSetPrompt, q.ConfigUnsetPrompt,
		promptUserInput,
	)

	cmdMode := newConfigKeyCmd(
		"mode",
		`Controls the default session behavior ("new" or "last")`,
		`Controls the default session behavior when no session flag is provided:

  new   Start a fresh session each time (default)
  last  Automatically continue the most recent session

Use --new or --last on individual queries to override the configured mode.`,
		q.ConfigGetMode, q.ConfigSetMode, q.ConfigUnsetMode,
		func() (string, error) {
			return promptFromList([]string{"new", "last"})
		},
	)

	cmdEditor := newConfigKeyCmd("editor",
		`Editor to open when no input is provided (default "vi")`,
		`Controls which editor is opened when qory is run without any input arguments.

The editor is resolved in the following order:
  1. This config value (if set)
  2. The $VISUAL environment variable
  3. The $EDITOR environment variable
  4. "vi" (built-in default)`,
		q.ConfigGetEditor, q.ConfigSetEditor, q.ConfigUnsetEditor,
		promptUserInput,
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
	cmd.AddCommand(
		cmdAPIKey,
		cmdBaseURL,
		cmdPrompt,
		cmdModel,
		cmdMode,
		cmdEditor,
	)

	return cmd
}

// newConfigKeyCmd builds a subcommand for a single config key with get/set/unset children.
func newConfigKeyCmd(
	use string,
	short string,
	long string,
	getter func() (*string, error),
	setter func(string) error,
	unsetter func() error,
	prompter func() (string, error),
) *cobra.Command {
	cmd := &cobra.Command{Use: use, Short: short, Long: long}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get",
			Short: "Print the current value",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				value, err := getter()
				if err != nil {
					return err
				}

				if value == nil {
					fmt.Println("No value")
				} else {
					fmt.Println(*value)
				}

				return nil
			},
		},

		&cobra.Command{
			Use:   "set [value]",
			Short: "Store a new value (prompts interactively if omitted)",
			Args:  cobra.MaximumNArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				var err error
				var value string
				if len(args) == 1 {
					value = args[0]
				} else {
					value, err = prompter()
				}

				if err != nil {
					return err
				}

				return setter(value)
			},
		},

		&cobra.Command{
			Use:   "unset",
			Short: "Remove the stored value",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				return unsetter()
			},
		},
	)
	return cmd
}
