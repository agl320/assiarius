package cmd

import (
	"assiarius/internal/poll"
	"time"

	"github.com/spf13/cobra"
)

func pollCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [screener] [interval]",
		Short: "Poll a Finviz screener",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			interval := 15 * time.Second
			if len(args) == 2 {
				parsed, err := time.ParseDuration(args[1])
				if err != nil {
					return err
				}
				interval = parsed
			}

			return poll.StartPoller(ctx, args[0], interval)
		},
	}

	return cmd
}