package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open connects to the SQLite database at path. For file-backed databases it
// enables WAL (readers don't block the single writer) and a busy_timeout so a
// contended write waits for the lock instead of failing with "database is
// locked". An in-memory path is left untouched — it has one connection and no
// on-disk journal.
func Open(path string) (*sql.DB, error) {
	dsn := path
	if path != ":memory:" {
		// URI form is required to attach connection pragmas.
		dsn = "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
