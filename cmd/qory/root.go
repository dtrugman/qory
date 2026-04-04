package main

import (
	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/dtrugman/qory/lib/editor"
	"github.com/spf13/cobra"
)

func newRootCmd(q *biz.Qory) *cobra.Command {
	var sessionID string
	var last bool
	var new_ bool

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
  qory --last "Please use argparse for the arguments"
  qory --new "Start fresh regardless of configured mode"`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				editorName, err := q.GetEditor()
				if err != nil {
					return err
				}
				content, err := editor.Edit(editorName)
				if err != nil {
					return err
				}
				if content == "" {
					return nil
				}
				args = []string{content}
			}
			if new_ {
				return q.QueryNew(args)
			}
			if last {
				return q.QueryLast(args)
			}
			if sessionID != "" {
				return q.QuerySession(sessionID, args)
			}
			return q.QueryDefault(args)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session name to continue")
	cmd.Flags().BoolVarP(&last, "last", "l", false, "Continue the last session")
	cmd.Flags().BoolVarP(&new_, "new", "n", false, "Start a new session")
	cmd.MarkFlagsMutuallyExclusive("new", "last")
	cmd.MarkFlagsMutuallyExclusive("new", "session")
	cmd.MarkFlagsMutuallyExclusive("last", "session")

	return cmd
}
