// Package storage provides a simple interface for storing and retrieving battle logs in a SQLite database.
package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/alexmatias/scraper-for-showdown-replays/internal/models"
	"github.com/alexmatias/scraper-for-showdown-replays/internal/parser"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// WAL mode for better concurrent read performance
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) SaveReplay(battle *parser.BattleLog, replay *models.Replay) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // If commit fails, automatic rollback

	// 1. INSERT replays
	_, err = tx.Exec(`
	INSERT OR REPLACE INTO replays (id, format, gen, gametype, rated, winner, upload_time, views, log_raw)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		replay.ID, battle.Format, battle.Gen, battle.GameType, battle.Rated, battle.Winner, replay.UploadTime, replay.Views, replay.Log,
	)
	if err != nil {
		return fmt.Errorf("inserting replay: %w", err)
	}

	// 2. INSERT players
	for _, p := range battle.Players {
		_, err = tx.Exec(`
		INSERT INTO players (replay_id, player_id, name, rating)
		VALUES (?, ?, ?, ?)`,
			replay.ID, p.ID, p.Name, p.Rating,
		)
		if err != nil {
			return fmt.Errorf("inserting player: %w", err)
		}
	}

	// 3. INSERT turns + events
	for _, turn := range battle.Turns {
		result, err := tx.Exec(`
		INSERT INTO turns (replay_id, turn_number, timestamp)
		VALUES (?, ?, ?)`,
			replay.ID, turn.Number, turn.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("inserting turn: %w", err)
		}

		turnID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("getting turn ID: %w", err)
		}

		// 4. INSERT events for this turn
		for _, event := range turn.Events {
			_, err = tx.Exec(`
			INSERT INTO events (turn_id, event_type, pokemon, target, move, hp, stat, amount, detail)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				turnID, event.Type, event.Pokemon, event.Target,
				event.Move, event.HP, event.Stat, event.Amount, event.Detail,
			)
			if err != nil {
				return fmt.Errorf("inserting event: %w", err)
			}

		}
	}
	return tx.Commit()
}

func (s *Store) GetReplay(id string) (*parser.BattleLog, error) {
	battle := &parser.BattleLog{}

	// 1. Get replay metadata
	row := s.db.QueryRow(`
		SELECT id, format, gen, gametype, rated, winner
		FROM replays WHERE id = ?`, id)

	err := row.Scan(&battle.ID, &battle.Format, &battle.Gen,
		&battle.GameType, &battle.Rated, &battle.Winner)
	if err != nil {
		return nil, fmt.Errorf("getting replay: %w", err)
	}

	// 2. Get players
	rows, err := s.db.Query(`
		SELECT player_id, name, rating
		FROM players WHERE replay_id = ?
	`, id)
	if err != nil {
		return nil, fmt.Errorf("querying players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p parser.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.Rating); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		battle.Players = append(battle.Players, p)
	}
	// not necessary to check error in local sqlite, but good practice.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating players: %w", err)
	}

	// 3. Get turns
	turnRows, err := s.db.Query(`
	SELECT id, turn_number, timestamp
	FROM turns WHERE replay_id = ?
	ORDER BY turn_number`, id)
	if err != nil {
		return nil, fmt.Errorf("querying turns: %w", err)
	}
	defer turnRows.Close()

	for turnRows.Next() {
		var turn parser.Turn
		var turnID int64
		if err := turnRows.Scan(&turnID, &turn.Number, &turn.Timestamp); err != nil {
			return nil, fmt.Errorf("scanning turn: %w", err)
		}

		// 4. Get events for this turn
		eventsRows, err := s.db.Query(`
		SELECT event_type, pokemon, target, move, hp, stat, amount, detail
		FROM events WHERE turn_id = ?`, turnID)
		if err != nil {
			return nil, fmt.Errorf("querying events: %w", err)
		}

		for eventsRows.Next() {
			var e parser.Event
			if err := eventsRows.Scan(&e.Type, &e.Pokemon, &e.Target,
				&e.Move, &e.HP, &e.Stat, &e.Amount, &e.Detail); err != nil {
				eventsRows.Close()
				return nil, fmt.Errorf("scanning event: %w", err)
			}
			turn.Events = append(turn.Events, e)
		}
		// this closing is manual (unlike to rows.Close())
		// due to the for loop to catch every event.
		eventsRows.Close()

		battle.Turns = append(battle.Turns, turn)
	}
	// not necessary (like in players), but good practice
	if err := turnRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating turns: %w", err)
	}

	return battle, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
