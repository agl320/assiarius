package cmd

import (
	"assiarius/internal/poll"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func pollCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [screener] [intervalSeconds]",
		Short: "Poll a Finviz screener",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Default
			interval := 15 * time.Second
			if len(args) == 2 {
				// More flexible than parseInt()
				if seconds, err := strconv.Atoi(args[1]); err == nil {
					if seconds <= 0 {
						return fmt.Errorf("intervalSeconds must be a positive integer, got %d", seconds)
					}
					interval = time.Duration(seconds) * time.Second
				} else {
					parsed, err := time.ParseDuration(args[1])
					if err != nil {
						return fmt.Errorf("invalid interval %q: use an integer number of seconds (e.g. 10) or a duration (e.g. 10s, 1m): %w", args[1], err)
					}
					if parsed <= 0 {
						return fmt.Errorf("interval must be > 0, got %s", parsed)
					}
					interval = parsed
				}
			}

			return poll.StartPoller(ctx, args[0], interval)
		},
	}

	return cmd
}