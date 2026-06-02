package db

import "database/sql"

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			price REAL NOT NULL,
			category TEXT,
			slug TEXT UNIQUE NOT NULL,
			image TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}
