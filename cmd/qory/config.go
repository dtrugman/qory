package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd(q *Qory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
	cmd.AddCommand(
		newConfigKeyCmd("api-key", "API key for the model provider",
			q.ConfigGetAPIKey, q.ConfigSetAPIKey, q.ConfigUnsetAPIKey, promptUserInput),
		newConfigKeyCmd("base-url", "Base URL for the model provider",
			q.ConfigGetBaseURL, q.ConfigSetBaseURL, q.ConfigUnsetBaseURL, promptUserInput),
		newConfigKeyCmd("model", "Model to use for queries",
			q.ConfigGetModel, q.ConfigSetModel, q.ConfigUnsetModel, func() (string, error) {
				models, err := q.AvailableModels()
				if err != nil {
					return "", err
				}
				return promptModel(models)
			}),
		newConfigKeyCmd("prompt", "Persistent system prompt prepended to every new session",
			q.ConfigGetPrompt, q.ConfigSetPrompt, q.ConfigUnsetPrompt, promptUserInput),
	)
	return cmd
}

// newConfigKeyCmd builds a subcommand for a single config key with get/set/unset children.
func newConfigKeyCmd(
	use string,
	short string,
	getter func() (*string, error),
	setter func(string) error,
	unsetter func() error,
	prompter func() (string, error),
) *cobra.Command {
	cmd := &cobra.Command{Use: use, Short: short}
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
