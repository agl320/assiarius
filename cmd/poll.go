package cmd

import (
	"assiarius/internal/poll"
	"assiarius/internal/screener"
	"assiarius/internal/webhook"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func pollCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [screenerURL] [window]",
		Short: "Poll a Finviz screener URL for recent news",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Default
			interval := 5 * time.Minute

			// Ensure poll window can be mapped to finviz screener
			if len(args) == 2 {
				parsed, err := parsePollWindow(args[1])
				if err != nil {
					return err
				}
				interval = parsed
			}

			// Validate screener URL (news window will be injected/overwritten by poller).
			if _, err := screener.ValidateFinvizScreenerURL(args[0]); err != nil {
				return err
			}

			log.Printf("poll: starting window=%s url=%s", interval, args[0])

			// Start poller
			results, err := poll.StartPoller(ctx, args[0], interval, app.LLM)
			if err != nil {
				return err
			}

			fmt.Println("Poll started...")
			for {
				select {
				// Channel closed 
				case <-ctx.Done():
					return nil
				// Results
				case res, ok := <-results:
					if !ok {
						return nil
					}
					msg := screener.FormatScreenResult(res)
					fmt.Println(msg)
					log.Printf("poll: result ticker=%s link=%s", strings.TrimSpace(res.Ticker), strings.TrimSpace(res.Latest.Link))
					if err := webhook.NotifyDiscord(ctx, msg); err != nil {
						fmt.Fprintln(os.Stderr, err)
						log.Printf("poll: webhook error: %v", err)
					}
				}
			}
		},
	}

	return cmd
}

func parsePollWindow(raw string) (time.Duration, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")

	switch s {
	case "5m", "5min", "5mins", "last5min":
		return 5 * time.Minute, nil
	case "30m", "30min", "30mins", "last30min":
		return 30 * time.Minute, nil
	case "1h", "hour", "lasthour", "60m":
		return 1 * time.Hour, nil
	case "24h", "24hour", "24hours", "1d", "day":
		return 24 * time.Hour, nil
	case "7d", "7day", "7days", "week":
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid poll window %q: use one of 5m, 30m, 1h, 24h, 7d", raw)
	}
}