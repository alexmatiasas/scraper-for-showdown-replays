package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/client"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/scraper"
)

func main() {
	// 1. Parse flags
	format := flag.String("format", "gen9ou", "battle format gen9vgc2024")
	limit := flag.Int("limit", 5, "max number of replays to download")
	workers := flag.Int("workers", 5, "number of parallel workers")
	delay := flag.Duration("delay", 500*time.Millisecond, "delay between requests")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout per HTTP request")
	flag.Parse()

	// 2. Context with cancelation with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 3. Create client and scraper
	c := client.NewClient(*timeout)
	s := scraper.New(c, scraper.Config{
		Workers: *workers,
		Delay:   *delay,
	})

	// 4. Execute
	log.Printf("Starting scraper: format=%s limit=%d workers=%d", *format, *limit, *workers)
	if err := s.Run(ctx, *format, *limit); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
