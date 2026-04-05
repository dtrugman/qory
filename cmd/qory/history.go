package main

import (
	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/dtrugman/qory/lib/editor"
	"github.com/spf13/cobra"
)

func newHistoryCmd(q *biz.Qory) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "history",
		Short:        "Browse chat history",
		Long:         `Browse chat history interactively. Navigate sessions with ↑/↓ and press enter to continue one in your editor.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			selected, err := ShowHistoryMenu(q)
			if err != nil {
				return err
			}

			if selected == "" {
				return nil
			}

			editorName, _, err := q.GetConfig().Editor()
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

			return q.QuerySession(selected, []string{content})
		},
	}
	return cmd
}
