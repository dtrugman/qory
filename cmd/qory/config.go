package main

import (
	"fmt"

	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/dtrugman/qory/lib/config"
	"github.com/spf13/cobra"
)

func newConfigCmd(q *biz.Qory) *cobra.Command {
	conf := q.GetConfig()

	cmdAPIKey := newConfigKeyCmd("api-key",
		"API key for the model provider", "",
		conf.APIKey, conf.SetAPIKey, conf.UnsetAPIKey,
		promptUserInput,
	)

	cmdBaseURL := newConfigKeyCmd("base-url",
		"Base URL for the model provider", "",
		conf.BaseURL, conf.SetBaseURL, conf.UnsetBaseURL,
		promptUserInput,
	)

	cmdModel := newConfigKeyCmd("model",
		"Model to use for queries", "",
		conf.Model, conf.SetModel, conf.UnsetModel,
		func() (string, error) {
			models, err := q.AvailableModels()
			if err != nil {
				return "", err
			}
			return promptModel(models)
		},
	)

	cmdPrompt := newConfigKeyCmd("prompt",
		"Persistent system prompt prepended to every new session", "",
		conf.Prompt, conf.SetPrompt, conf.UnsetPrompt,
		promptUserInput,
	)

	cmdMode := newConfigKeyCmd(
		"mode",
		`Controls the default session behavior ("new" or "last")`,
		`Controls the default session behavior when no session flag is provided:

  new   Start a fresh session each time (default)
  last  Automatically continue the most recent session

Use --new or --last on individual queries to override the configured mode.`,
		conf.Mode, conf.SetMode, conf.UnsetMode,
		func() (string, error) {
			return promptFromList([]string{"new", "last"})
		},
	)

	cmdEditor := newConfigKeyCmd("editor",
		`Editor to open when no input is provided (default "vi")`,
		`Controls which editor is opened when qory is run without any input arguments.

The editor is resolved in the following order:
  1. The $VISUAL environment variable
  2. The $EDITOR environment variable
  3. This config value (if set)
  4. "vi" (built-in default)`,
		conf.Editor, conf.SetEditor, conf.UnsetEditor,
		promptUserInput,
	)

	getHistorySizeStr := func() (string, config.Origin, error) {
		size, origin, err := conf.HistorySize()
		if err != nil {
			return "", origin, err
		}
		return fmt.Sprintf("%d", size), origin, nil
	}

	cmdHistorySize := newConfigKeyCmd(
		"history-size",
		fmt.Sprintf("Number of unnamed sessions to keep (default %d)", config.DefaultHistorySize),
		`Controls how many unnamed (auto-generated) sessions are retained on disk.

When a new query is completed, sessions beyond this limit are deleted oldest-first.
If the limit is smaller than the current number of stored sessions, no immediate
cleanup occurs — the excess sessions are removed the next time a new query is run.`,
		getHistorySizeStr, conf.SetHistorySize, conf.UnsetHistorySize,
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
		cmdHistorySize,
	)

	return cmd
}

// newConfigKeyCmd builds a subcommand for a single config key with get/set/unset children.
func newConfigKeyCmd(
	use string,
	short string,
	long string,
	getter func() (string, config.Origin, error),
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
				value, origin, err := getter()
				if err != nil {
					return err
				}

				if origin == config.OriginNotSet {
					fmt.Println(origin)
				} else {
					fmt.Printf("%s  [%s]\n", value, origin)
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
