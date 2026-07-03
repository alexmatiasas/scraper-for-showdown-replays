package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/client"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/models"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/scraper"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/storage"
)

func main() {
	// 1. Parse flags
	formatStr := flag.String("format", "gen9ou", "battle format gen9ou")
	limit := flag.Int("limit", 5, "max number of replays to download")
	workers := flag.Int("workers", 5, "number of parallel workers")
	delay := flag.Duration("delay", 500*time.Millisecond, "delay between requests")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout per HTTP request")
	dbPath := flag.String("db", "scraper.db", "path to SQLite database")
	flag.Parse()

	format, err := models.ValidFormat(*formatStr)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Context with cancelation with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 3. Create storage
	store, err := storage.New(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	// 4. Create client and scraper
	c := client.NewClient(*timeout)
	s := scraper.New(c, store, scraper.Config{
		Workers: *workers,
		Delay:   *delay,
	})

	// 5. Execute
	log.Printf("Starting scraper: format=%s limit=%d workers=%d", format, *limit, *workers)
	if err := s.Run(ctx, format, *limit); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
