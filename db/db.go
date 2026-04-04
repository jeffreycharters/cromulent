package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err = applyPragmas(); err != nil {
		return fmt.Errorf("pragmas: %w", err)
	}

	return migrate()
}

func applyPragmas() error {
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA foreign_keys=ON`,
		`PRAGMA busy_timeout=5000`,
	}
	for _, p := range pragmas {
		if _, err := DB.Exec(p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}
