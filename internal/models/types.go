// Package models defines the core data structures for the scraper pipeline
package models

type SearchResult struct {
	ID         string   `json:"id"`
	Format     string   `json:"format"`
	Players    []string `json:"players"`
	Rating     *int     `json:"rating"`
	UploadTime int64    `json:"uploadtime"`
}

type SearchResponse struct {
	Results []SearchResult
	HasMore bool
}

type Replay struct {
	ID         string   `json:"id"`
	Format     string   `json:"format"`
	FormatID   string   `json:"formatid"`
	Players    []string `json:"players"`
	Log        string   `json:"log"`
	Rating     *int     `json:"rating"`
	UploadTime int64    `json:"uploadtime"`
	Views      int      `json:"views"`
	Private    int      `json:"private"`
}
