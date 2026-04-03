package main

import "github.com/spf13/cobra"

func newRootCmd(q *Qory) *cobra.Command {
	var sessionID string
	var last bool

	cmd := &cobra.Command{
		Use:   "qory <input...>",
		Short: "A language model in your terminal",
		Long: `Qory is a tool for accessing language models directly from your CLI,
allowing you to specify free-form queries and any local file as context.

Examples:
  qory "Please create a basic OpenAPI yaml template"
  qory "Please add a health check to my OpenAPI spec" openapi.yaml
  qory --session spec "Please add a health check to my OpenAPI spec" openapi.yaml
  qory --session spec "Please define a new parameter for the body"
  qory --last "Please use argparse for the arguments"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if last {
				return q.QueryLast(args)
			}
			if sessionID != "" {
				return q.QuerySession(sessionID, args)
			}
			return q.QueryNew(args)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session name to continue")
	cmd.Flags().BoolVarP(&last, "last", "l", false, "Continue the last session")

	return cmd
}
