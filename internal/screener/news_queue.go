package screener

import (
	"context"
	"strings"
	"sync"
	"time"

	"assiarius/internal/llm"
	"assiarius/internal/scraper"
)

type NewsTask struct {
	Ticker string
	Item   scraper.NewsItem
}

// NewsQueue is a FIFO queue of news items to analyze.
//
// Minimal safety dedupe: while a news URL is queued or in-flight, it will not
// be enqueued again.
type NewsQueue struct {
	mu sync.Mutex

	wakeCh chan struct{}
	queue  []NewsTask
	seen   map[string]struct{} // link -> present in queue or in-flight

	client llm.Client

	resultsCh chan ScreenResult
	closeOnce sync.Once

	minInterval time.Duration
	lastCallAt  time.Time
}

func NewNewsQueue(client llm.Client, minInterval time.Duration) *NewsQueue {
	return &NewsQueue{
		// Buffered channel to allow signaling without blocking
		// Size 1 is enough since we only need to know "not empty" vs "empty"
		wakeCh:      make(chan struct{}, 1),
		// Pre-allocate queue slice to avoid frequent allocations; it will grow if needed
		queue:       make([]NewsTask, 0, 256),
		// Use a map for O(1) deduplication checks
		seen:        map[string]struct{}{},
		client:      client,
		resultsCh:   make(chan ScreenResult, 256),
		minInterval: minInterval,
	}
}

func (q *NewsQueue) Results() <-chan ScreenResult {
	if q == nil {
		return nil
	}
	return q.resultsCh
}

// Enqueue adds a NewsTask if it has a non-empty link and the link is not already
// queued/in-flight.
func (q *NewsQueue) Enqueue(t NewsTask) bool {
	if q == nil {
		return false
	}

	t.Ticker = strings.TrimSpace(strings.ToUpper(t.Ticker))
	if t.Ticker == "" {
		return false
	}
	t.Item.Link = strings.TrimSpace(t.Item.Link)
	if t.Item.Link == "" {
		return false
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.seen[t.Item.Link]; ok {
		return false
	}
	q.seen[t.Item.Link] = struct{}{}
	q.queue = append(q.queue, t)
	q.signal()
	return true
}

func (q *NewsQueue) signal() {
	select {
	case q.wakeCh <- struct{}{}:
	default:
		// already signaled
	}
}

func (q *NewsQueue) Start(ctx context.Context) {
	defer q.closeOnce.Do(func() { close(q.resultsCh) })

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		task, ok := q.dequeue(ctx)
		if !ok {
			return
		}

		if !q.rateLimit(ctx) {
			q.finish(task)
			return
		}

		res, err := AnalyzeNewsItem(ctx, task.Ticker, task.Item, TickerSignals{}, q.client)
		q.mu.Lock()
		q.lastCallAt = time.Now()
		q.mu.Unlock()

		if err == nil {
			select {
			case q.resultsCh <- res:
			default:
				// best-effort
			}
		}

		q.finish(task)
	}
}

func (q *NewsQueue) dequeue(ctx context.Context) (NewsTask, bool) {
	for {
		q.mu.Lock()
		if len(q.queue) > 0 {
			t := q.queue[0]
			q.queue = q.queue[1:]
			q.mu.Unlock()
			return t, true
		}
		q.mu.Unlock()

		select {
		case <-ctx.Done():
			return NewsTask{}, false
		case <-q.wakeCh:
			// retry
		}
	}
}

func (q *NewsQueue) finish(task NewsTask) {
	link := strings.TrimSpace(task.Item.Link)
	if link == "" {
		return
	}
	q.mu.Lock()
	delete(q.seen, link)
	q.mu.Unlock()
}

// Blocks until either the context is done or the rate limit allows another call.
func (q *NewsQueue) rateLimit(ctx context.Context) bool {
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
