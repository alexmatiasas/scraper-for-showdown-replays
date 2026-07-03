package storage

import (
	"testing"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/models"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/parser"
)

func testBattle() *parser.BattleLog {
	return &parser.BattleLog{
		ID:       "gen9ou-12345",
		Format:   "[Gen 9] OU",
		Gen:      9,
		GameType: "singles",
		Rated:    true,
		Winner:   "Alice",
		Players: []parser.Player{
			{ID: "p1", Name: "Alice", Rating: 1500},
			{ID: "p2", Name: "Bob", Rating: 0},
		},
		Turns: []parser.Turn{
			{
				Number:    1,
				Timestamp: 1782964964,
				Events: []parser.Event{
					{Type: "switch", Pokemon: "p1a: Garchomp", Species: "Garchomp", HP: "100/100"},
					{Type: "move", Pokemon: "p1a: Garchomp", Move: "Earthquake", Target: "p2a: Flutter Mane"},
				},
			},
			{
				Number:    2,
				Timestamp: 1782964972,
				Events: []parser.Event{
					{Type: "faint", Pokemon: "p2a: Flutter Mane"},
				},
			},
		},
	}
}

func testReplay() *models.Replay {
	return &models.Replay{
		ID:         "gen9ou-12345",
		Format:     "[Gen 9] OU",
		UploadTime: 1782964951,
		Views:      42,
		Log:        "|turn|1\n|move|...",
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	// file::memory:?cache=shared solves deadlock
	store, err := New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestGetReplay(t *testing.T) {
	store := newTestStore(t)
	battle := testBattle()
	replay := testReplay()

	if err := store.SaveReplay(battle, replay); err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	got, err := store.GetReplay("gen9ou-12345")
	if err != nil {
		t.Fatalf("GetReplay: %v", err)
	}

	if got.ID != battle.ID {
		t.Errorf("ID = %q, want %q", got.ID, battle.ID)
	}
	if got.Gen != 9 {
		t.Errorf("Gen = %d, want 9", got.Gen)
	}
	if got.Winner != "Alice" {
		t.Errorf("Winner = %q, want %q", got.Winner, "Alice")
	}
	if len(got.Players) != 2 {
		t.Errorf("len(Players) = %d, want 2", len(got.Players))
	}
	if len(got.Turns) != 2 {
		t.Errorf("len(Turns) = %d, want 2", len(got.Turns))
	}
}

func TestNew(t *testing.T) {
	store, err := New(":memory:?cache=shared")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = store.Close() }()

	// Verify schema was applied by querying replays table
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM replays").Scan(&count)
	if err != nil {
		t.Fatalf("schema not applied: %v", err)
	}
	if count != 0 {
		t.Errorf("replays count = %d, want 0", count)
	}
}

func TestSaveReplay(t *testing.T) {
	store := newTestStore(t)
	battle := testBattle()
	replay := testReplay()

	err := store.SaveReplay(battle, replay)
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	// Verify data exists in DB
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM replays WHERE id = ?", "gen9ou-12345").Scan(&count)
	if err != nil {
		t.Fatalf("counting replays: %v", err)
	}
	if count != 1 {
		t.Errorf("replays count = %d, want 1", count)
	}
	err = store.db.QueryRow("SELECT COUNT(*) FROM players WHERE replay_id = ?", "gen9ou-12345").Scan(&count)
	if err != nil {
		t.Fatalf("counting players: %v", err)
	}
	if count != 2 {
		t.Errorf("players count = %d, want 2", count)
	}
	err = store.db.QueryRow("SELECT COUNT(*) FROM turns WHERE replay_id = ?", "gen9ou-12345").Scan(&count)
	if err != nil {
		t.Fatalf("counting turns: %v", err)
	}
	if count != 2 {
		t.Errorf("turns count = %d, want 2", count)
	}
	err = store.db.QueryRow("SELECT COUNT(*) FROM events WHERE turn_id IN (SELECT id FROM turns WHERE replay_id = ?)", "gen9ou-12345").Scan(&count)
	if err != nil {
		t.Fatalf("counting events: %v", err)
	}
	if count != 3 {
		t.Errorf("events count = %d, want 3", count)
	}
}

func TestGetReplay_Players(t *testing.T) {
	store := newTestStore(t)
	err := store.SaveReplay(testBattle(), testReplay())
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	got, err := store.GetReplay("gen9ou-12345")
	if err != nil {
		t.Fatalf("GetReplay: %v", err)
	}

	if len(got.Players) != 2 {
		t.Fatalf("len(Players) = %d, want 2", len(got.Players))
	}

	// Player 1
	if got.Players[0].ID != "p1" {
		t.Errorf("Players[0].ID = %q, want %q", got.Players[0].ID, "p1")
	}
	if got.Players[0].Name != "Alice" {
		t.Errorf("Players[0].Name = %q, want %q", got.Players[0].Name, "Alice")
	}
	if got.Players[0].Rating != 1500 {
		t.Errorf("Players[0].Rating = %d, want 1500", got.Players[0].Rating)
	}

	// Player 2
	if got.Players[1].ID != "p2" {
		t.Errorf("Players[1].ID = %q, want %q", got.Players[1].ID, "p2")
	}
	if got.Players[1].Name != "Bob" {
		t.Errorf("Players[1].Name = %q, want %q", got.Players[1].Name, "Bob")
	}
	if got.Players[1].Rating != 0 {
		t.Errorf("Players[1].Rating = %d, want 0", got.Players[1].Rating)
	}
}

func TestGetReplay_Turns(t *testing.T) {
	store := newTestStore(t)
	err := store.SaveReplay(testBattle(), testReplay())
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	got, err := store.GetReplay("gen9ou-12345")
	if err != nil {
		t.Fatalf("GetReplay: %v", err)
	}

	if len(got.Turns) != 2 {
		t.Fatalf("len(Turns) = %d, want 2", len(got.Turns))
	}

	// Turn 1
	if got.Turns[0].Number != 1 {
		t.Errorf("Turns[0].Number = %d, want 1", got.Turns[0].Number)
	}
	if got.Turns[0].Timestamp != 1782964964 {
		t.Errorf("Turns[0].Timestamp = %d, want 1782964964", got.Turns[0].Timestamp)
	}

	// Turn 2
	if got.Turns[1].Number != 2 {
		t.Errorf("Turns[1].Number = %d, want 2", got.Turns[1].Number)
	}
	if got.Turns[1].Timestamp != 1782964972 {
		t.Errorf("Turns[1].Timestamp = %d, want 1782964972", got.Turns[1].Timestamp)
	}
}

func TestGetReplay_Events(t *testing.T) {
	store := newTestStore(t)
	err := store.SaveReplay(testBattle(), testReplay())
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	got, err := store.GetReplay("gen9ou-12345")
	if err != nil {
		t.Fatalf("GetReplay: %v", err)
	}

	// Turn 1: 2 events
	if len(got.Turns[0].Events) != 2 {
		t.Fatalf("Turn 1 events = %d, want 2", len(got.Turns[0].Events))
	}
	if got.Turns[0].Events[0].Type != "switch" {
		t.Errorf("Turn 1 Event[0].Type = %q, want %q", got.Turns[0].Events[0].Type, "switch")
	}
	if got.Turns[0].Events[0].Pokemon != "p1a: Garchomp" {
		t.Errorf("Turn 1 Event[0].Pokemon = %q", got.Turns[0].Events[0].Pokemon)
	}
	if got.Turns[0].Events[1].Type != "move" {
		t.Errorf("Turn 1 Event[1].Type = %q, want %q", got.Turns[0].Events[1].Type, "move")
	}
	if got.Turns[0].Events[1].Move != "Earthquake" {
		t.Errorf("Turn 1 Event[1].Move = %q, want %q", got.Turns[0].Events[1].Move, "Earthquake")
	}

	// Turn 2: 1 event
	if len(got.Turns[1].Events) != 1 {
		t.Fatalf("Turn 2 events = %d, want 1", len(got.Turns[1].Events))
	}
	if got.Turns[1].Events[0].Type != "faint" {
		t.Errorf("Turn 2 Event[0].Type = %q, want %q", got.Turns[1].Events[0].Type, "faint")
	}
	if got.Turns[1].Events[0].Pokemon != "p2a: Flutter Mane" {
		t.Errorf("Turn 2 Event[0].Pokemon = %q", got.Turns[1].Events[0].Pokemon)
	}
}

func TestGetReplay_NotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetReplay("nonexistent-id")
	if err == nil {
		t.Error("GetReplay with nonexistent ID should return error")
	}
}

func TestSaveReplay_Overwrite(t *testing.T) {
	store := newTestStore(t)
	replay := testReplay()

	// Save with Alice as winner
	battle1 := testBattle()
	battle1.Winner = "Alice"
	err := store.SaveReplay(battle1, replay)
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	// Overwrite with Bob as winner
	battle2 := testBattle()
	battle2.Winner = "Bob"
	err = store.SaveReplay(battle2, replay)
	if err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	got, err := store.GetReplay("gen9ou-12345")
	if err != nil {
		t.Fatalf("GetReplay: %v", err)
	}
	if got.Winner != "Bob" {
		t.Errorf("Winner = %q, want %q (should be overwritten)", got.Winner, "Bob")
	}
}

func TestClose(t *testing.T) {
	store, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = store.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}
