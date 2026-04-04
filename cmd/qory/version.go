package main

import (
	"fmt"

	"github.com/dtrugman/qory/cmd/qory/biz"
	"github.com/spf13/cobra"
)

func newVersionCmd(q *biz.Qory) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(q.Version())
			return nil
		},
	}
}
