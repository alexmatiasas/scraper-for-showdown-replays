package storage

import (
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string

// Migrate applies the embedded SQL schema to the database.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
