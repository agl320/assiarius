package poll

import (
	"assiarius/internal/screener"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [preset]",
		Short: "Poll a Finviz screener",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			startPoller(ctx, args[0])
			return nil
		},
	}

	return cmd
}

type ScreenerResult struct {
	Ticker string
	Price  float64
}

func startPoller(ctx context.Context, screen string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	fmt.Println("Poll started...")
	for {
		// wait for channel
		<-ticker.C
		err := screener.RunScreen(screen)
		if err != nil {
			return
		}
	}
}
