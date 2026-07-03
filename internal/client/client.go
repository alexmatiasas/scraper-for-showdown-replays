// Package client provides a simple interface for interacting with the Pokemon Showdown replay API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/models"
)

const (
	baseURL     string = "http://replay.pokemonshowdown.com"
	maxResults  int    = 51
	maxBodySize int    = 10 << 20
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Search(ctx context.Context, format models.Format, before int64) (*models.SearchResponse, error) {
	params := url.Values{"format": []string{string(format)}}

	if before > 0 {
		params.Set("before", strconv.FormatInt(before, 10))
	}

	u := (&url.URL{
		Scheme:   "http",
		Host:     "replay.pokemonshowdown.com",
		Path:     "/search.json",
		RawQuery: params.Encode(),
	}).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := limitJSONResponse(resp)
	if err != nil {
		return nil, err
	}

	var results []models.SearchResult
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &models.SearchResponse{
		Results: results,
		HasMore: len(results) == maxResults,
	}, nil
}

func (c *Client) FetchReplay(ctx context.Context, id string) (*models.Replay, error) {
	err := validateID(id)
	if err != nil {
		return nil, fmt.Errorf("id invalid: %w", err)
	}

	u := baseURL + "/" + id + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := limitJSONResponse(resp)
	if err != nil {
		return nil, err
	}

	var replay models.Replay
	if err := json.Unmarshal(body, &replay); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &replay, nil
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("empty replay id")
	}

	// this is the format from Showdown: {format}-{number}
	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		return fmt.Errorf("invalid replay id format: %s", id)
	}
	// Verify the last part is numeric

	if _, err := strconv.Atoi(parts[len(parts)-1]); err != nil {
		return fmt.Errorf("replay id must end with numeric value: %s", id)
	}

	return nil
}

func limitJSONResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBodySize)))
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	if len(body) >= maxBodySize {
		return nil, fmt.Errorf("response body exceeds %d bytes", maxBodySize)
	}

	return body, nil
}
