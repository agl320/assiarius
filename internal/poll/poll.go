package poll

import (
	"assiarius/internal/screener"
	"context"
	"fmt"
	"time"
)

type ScreenerResult struct {
	Ticker string
	Price  float64
}

func StartPoller(ctx context.Context, screen string) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	fmt.Println("Poll started...")
	for {
		// wait for channel
		<-ticker.C
		err := screener.RunScreen(screen)
		if err != nil {
			return err
		}
	}
}

