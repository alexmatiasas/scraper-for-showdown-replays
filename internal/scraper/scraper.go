// Package scraper provides a simple interface for concurrent downloading and storage of Showdown replays and storing them in the database.
package scraper

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/client"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/models"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/parser"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/storage"
)

type Config struct {
	Workers int
	Delay   time.Duration
}

type Scraper struct {
	client *client.Client
	store  *storage.Store
	config Config
}

func New(c *client.Client, store *storage.Store, cfg Config) *Scraper {
	// Apply sane defaults so callers can pass a zero-value Config
	if cfg.Workers <= 0 {
		cfg.Workers = 5
	}
	if cfg.Delay <= 0 {
		cfg.Delay = 500 * time.Millisecond
	}

	return &Scraper{client: c, store: store, config: cfg}
}

// Run executes the complete pipeline: feed -> workers -> collector.
func (s *Scraper) Run(ctx context.Context, format models.Format, limit int) error {
	jobs := make(chan string, s.config.Workers)
	results := make(chan *models.Replay, s.config.Workers)

	var wg sync.WaitGroup

	// 1. Launch N workers
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.worker(ctx, workerID, jobs, results)
		}(i)
	}

	// 2. Launch feeder (goroutine that paginate the API and sends IDs)
	feederErr := make(chan error, 1)
	go func() {
		feederErr <- s.feedJobs(ctx, format, limit, jobs)
	}()

	// 3. Closer: waits for the workers to finish, close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Collector: reads results
	count := 0
	for replay := range results {
		count++
		log.Printf("[%d/%d] %s - %v", count, limit, replay.ID, replay.Players)

		battle, err := parser.Parse(replay.Log)
		if err != nil {
			log.Printf("error parsing %s: %v", replay.ID, err)
			continue
		}

		// Set metadata from API (parser doesn't have this)
		battle.ID = replay.ID
		battle.Format = replay.FormatID

		err = s.store.SaveReplay(battle, replay)
		if err != nil {
			log.Printf("error saving %s: %v", replay.ID, err)
			continue
		}
	}

	// 5. Verify errors from the feeder
	if err := <-feederErr; err != nil {
		return fmt.Errorf("feeder: %w", err)
	}

	log.Printf("Done: %d replays fetched", count)

	return nil
}

// feedJobs paginates the API Search and sends IDs to the channel.
func (s *Scraper) feedJobs(ctx context.Context, format models.Format, limit int, jobs chan<- string) error {
	defer close(jobs)

	var before int64 = 0
	fetched := 0

	for fetched < limit {
		// Verify cancelation before each request
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := s.client.Search(ctx, format, before)
		if err != nil {
			return fmt.Errorf("searching replays: %w", err)
		}

		for _, result := range resp.Results {
			if fetched >= limit {
				return nil
			}
			jobs <- result.ID
			fetched++
		}

		if !resp.HasMore {
			return nil
		}

		before = resp.Results[len(resp.Results)-1].UploadTime
		time.Sleep(s.config.Delay)
	}
	return nil
}

// worker is a goroutine that downloads replays from the channel.
func (s *Scraper) worker(ctx context.Context, id int, jobs <-chan string, results chan<- *models.Replay) {
	log.Printf("Worker %d started", id)

	for JobID := range jobs {
		// Rate limiting before each request
		time.Sleep(s.config.Delay)

		replay, err := s.client.FetchReplay(ctx, JobID)
		if err != nil {
			log.Printf("Worker %d: error fetching %s: %v", id, JobID, err)
			continue
		}

		results <- replay
	}
	log.Printf("Worker %d finished", id)
}
