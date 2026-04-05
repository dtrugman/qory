package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const appName = "Qory"

var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s version %s\n", appName, version)
			return nil
		},
	}
}
