package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrateFiles embed.FS

// RunGlobalOnly runs only global migrations
func RunMigrations(db *sql.DB) error {
	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	// Use default table name for global migrations
	goose.SetTableName("goose_db_version")
	goose.SetBaseFS(migrateFiles)
	return goose.Up(db, ".")
}
