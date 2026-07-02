package storage

import (
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string

func Migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
