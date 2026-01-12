package cmd

import (
	"assiarius/internal/poll"

	"github.com/spf13/cobra"
)

func pollCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [preset]",
		Short: "Poll a Finviz screener",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return poll.StartPoller(ctx, args[0])
		},
	}

	return cmd
}