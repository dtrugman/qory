package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/spf13/cobra"
)

func newHistoryCmd(q *biz.Qory) *cobra.Command {
	return &cobra.Command{
		Use:   "history [session-id]",
		Short: "Show chat history",
		Long: `Show chat history.

Run without a session ID to see snippets of the most recent sessions.
Specify a session ID to print the full message transcript.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				sess, err := q.HistorySession(args[0])
				if err != nil {
					return err
				}
				for _, m := range sess.Messages {
					fmt.Printf("=== %s ===\n", strings.ToUpper(string(m.Role)))
					fmt.Printf("%s\n", m.Content)
				}
				return nil
			}

			previews, err := q.HistoryAll()
			if err != nil {
				return err
			}
			for _, preview := range previews {
				fmt.Printf("=== %s (%s) ===\n", preview.Name, preview.UpdatedAt.Format(time.RFC822))
				snippet, _ := strings.CutSuffix(preview.Snippet, "\n")
				fmt.Printf("%s\n", snippet)
			}
			return nil
		},
	}
}
