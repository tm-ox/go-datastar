package db

import "database/sql"

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS carts (
      		id INTEGER PRIMARY KEY AUTOINCREMENT,
      		cart_id TEXT UNIQUE NOT NULL,
      		user_id INTEGER,   -- nullable, future auth
      		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS cart_items (
        	id INTEGER PRIMARY KEY AUTOINCREMENT,
        	cart_id TEXT NOT NULL REFERENCES carts(cart_id),
         	product_handle TEXT NOT NULL,
         	name TEXT NOT NULL,
          	price INTEGER NOT NULL,
        	quantity INTEGER NOT NULL DEFAULT 1,
        	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        	UNIQUE(cart_id, product_handle)
        );
        CREATE TABLE IF NOT EXISTS orders (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            cart_id TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'pending',
            total INTEGER NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS order_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            order_id INTEGER NOT NULL REFERENCES orders(id),
            product_handle TEXT NOT NULL,
            name TEXT NOT NULL,
            price INTEGER NOT NULL,
            quantity INTEGER NOT NULL
        );
		CREATE TABLE IF NOT EXISTS work (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT UNIQUE NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			title TEXT NOT NULL,
			type TEXT NOT NULL,
			client TEXT NOT NULL,
			year INTEGER NOT NULL,
			tools TEXT NOT NULL,
			description TEXT,
			website TEXT,
			link TEXT,
			cover_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS work_images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			work_id INTEGER NOT NULL REFERENCES work(id),
			url TEXT NOT NULL,
			alt TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS site_pages (
			key TEXT PRIMARY KEY,
			meta_title TEXT,
			meta_description TEXT,
			title TEXT,
			tagline TEXT,
			body TEXT
		);
		CREATE TABLE IF NOT EXISTS site_sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			page_key TEXT NOT NULL,
			section_key TEXT NOT NULL,
			title TEXT,
			tagline TEXT,
			cols INTEGER NOT NULL DEFAULT 0,
			sort_order INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS site_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			section_id INTEGER,
			title TEXT,
			description TEXT,
			href TEXT,
			icon TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			button_text TEXT,
			button_href TEXT,
			button_target TEXT,
			UNIQUE(section_id, title)
		);
	`)
	return err
}
