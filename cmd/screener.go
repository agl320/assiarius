package cmd

import (
	"assiarius/internal/screener"

	"github.com/spf13/cobra"
)

func screenerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen [preset]",
		Short: "Run a single Finviz screener preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return screener.RunScreen(args[0])
		},
	}

	return cmd
}