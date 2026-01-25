package storage

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDBCreatesSchema(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	db, err := InitDB(dbPath)
	if !assert.NoError(t, err) {
		return
	}
	defer db.Close()

	assert.NoError(t, insertComment(db))
}

func insertComment(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO comments (filename, comment) VALUES (?, ?)`, "file.jpg", "nice")
	return err
}
