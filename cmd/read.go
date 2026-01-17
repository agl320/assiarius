package cmd

import (
	"github.com/spf13/cobra"
	"assiarius/internal/read"
)

func readCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [link]",
		Short: "Read news from a given link",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return read.ReadNewsFromLink(args[0])
		},
	}

	return cmd
}