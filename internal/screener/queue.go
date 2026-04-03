package screener

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"assiarius/internal/llm"
)

type queuedTicker struct {
	ticker      string
	volumeText  string
	volumeValue float64
	firstSeen   time.Time
	lastSeen    time.Time
}

// LLMQueue collects ticker work items, dedupes them, and lets a single consumer
// process them at a controlled rate.
//
// Priority is based on volume (higher first) plus a recency bonus that decays
// over time. The recency term helps newly-enqueued items bubble up, while the
// volume term provides stable ordering.
//
// This is intentionally simple (O(n) selection) since screener sizes are
// typically small.
type LLMQueue struct {
	mu     sync.Mutex
	wakeCh chan struct{}
	tasks  map[string]*queuedTicker
	client llm.Client

	minInterval time.Duration
	decayHalf   time.Duration
	lastCallAt  time.Time
}

func NewLLMQueue(client llm.Client, minInterval time.Duration) *LLMQueue {
	q := &LLMQueue{
		tasks:       map[string]*queuedTicker{},
		client:      client,
		wakeCh:      make(chan struct{}, 1),
		minInterval: minInterval,
		decayHalf:   90 * time.Second,
	}
	return q
}

// Enqueue adds a ticker if it is not already pending; if it is, it updates its
// last-seen time and keeps the highest known volume.
func (q *LLMQueue) Enqueue(ticker string, volumeText string) {
	// Ticker clean-up (if needed)
	ticker = strings.TrimSpace(strings.ToUpper(ticker))
	if ticker == "" {
		return
	}

	volumeValue := parseVolumeNumber(volumeText)
	now := time.Now()

	// Mutex lock to safely access the queue state (for multiple goroutines)
	q.mu.Lock()
	defer q.mu.Unlock()

	if existing, ok := q.tasks[ticker]; ok {
		existing.lastSeen = now
		if volumeValue > existing.volumeValue {
			existing.volumeValue = volumeValue
			existing.volumeText = volumeText
		}
		q.signal()
		return
	}

	q.tasks[ticker] = &queuedTicker{
		ticker:      ticker,
		volumeText:  volumeText,
		volumeValue: volumeValue,
		firstSeen:   now,
		lastSeen:    now,
	}
	q.signal()
}

func (q *LLMQueue) signal() {
	select {
	case q.wakeCh <- struct{}{}:
	default:
		// already signaled
	}
}

func (q *LLMQueue) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ticker, volumeText, ok := q.dequeueNext(ctx)
		if !ok {
			return
		}

		if !q.rateLimit(ctx) {
			return
		}

		// Consume the task outside of the lock.
		_, _ = GetNewsForTickerWithVolume(ctx, ticker, volumeText, q.client)
		q.mu.Lock()
		q.lastCallAt = time.Now()
		q.mu.Unlock()
	}
}

func (q *LLMQueue) rateLimit(ctx context.Context) bool {
	if q.minInterval <= 0 {
		return true
	}

	q.mu.Lock()
	since := time.Since(q.lastCallAt)
	q.mu.Unlock()

	wait := q.minInterval - since
	if wait <= 0 {
		return true
	}

	t := time.NewTimer(wait)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (q *LLMQueue) dequeueNext(ctx context.Context) (ticker string, volumeText string, ok bool) {
	for {
		q.mu.Lock()
		if len(q.tasks) > 0 {
			break
		}
		q.mu.Unlock()

		select {
		case <-ctx.Done():
			return "", "", false
		case <-q.wakeCh:
			// try again
		}
	}

	now := time.Now()
	bestScore := math.Inf(-1)
	var best *queuedTicker

	for _, t := range q.tasks {
		s := q.score(now, t)
		if s > bestScore {
			bestScore = s
			best = t
		}
	}

	if best == nil {
		q.mu.Unlock()
		return "", "", false
	}

	delete(q.tasks, best.ticker)
	q.mu.Unlock()
	return best.ticker, best.volumeText, true
}

func (q *LLMQueue) score(now time.Time, t *queuedTicker) float64 {
	// Volume score: stable ordering by size (log scaled).
	vol := 0.0
	if t.volumeValue > 0 {
		vol = math.Log1p(t.volumeValue)
	}

	// Recency bonus: decays toward 0 as time since lastSeen grows.
	age := now.Sub(t.lastSeen)
	recency := 0.0
	if q.decayHalf > 0 {
		// 1.0 at age=0, 0.5 at age=half-life, etc.
		recency = math.Exp(-math.Ln2 * age.Seconds() / q.decayHalf.Seconds())
	}

	return vol + recency
}

// GetQueueMinInterval reads GEMINI_MIN_INTERVAL (duration string) or returns the default.
func GetQueueMinInterval() time.Duration {
	// Conservative default. You can lower it if your API limits allow.
	defaultInterval := 2 * time.Second

	raw := strings.TrimSpace(os.Getenv("GEMINI_MIN_INTERVAL"))
	if raw == "" {
		return defaultInterval
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return defaultInterval
	}
	return parsed
}

// parseVolumeNumber parses values like "52.1M", "1.2B", "950K", "12,345".
// Returns 0 for unknown/empty.
func parseVolumeNumber(s string) float64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0
	}

	s = strings.ReplaceAll(s, ",", "")
	mult := 1.0

	last := s[len(s)-1]
	switch last {
	case 'K':
		mult = 1_000
		s = s[:len(s)-1]
	case 'M':
		mult = 1_000_000
		s = s[:len(s)-1]
	case 'B':
		mult = 1_000_000_000
		s = s[:len(s)-1]
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v * mult
}
