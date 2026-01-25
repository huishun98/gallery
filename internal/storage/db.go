package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	dataSourceName := "file:" + dbPath + "?_busy_timeout=5000&_journal_mode=WAL"
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Optional but recommended for SQLite
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Verify the connection is valid
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Ensure schema is created
	if err := ensureSchema(db); err != nil {
		return nil, err
	}

	return db, err
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			comment TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			removed_at DATETIME
		)
	`)
	return err
}
