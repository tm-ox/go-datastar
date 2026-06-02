package main

import (
	"log"

	"github.com/tm-ox/go-datastar/internal/db"
)

func main() {
	database, err := db.Open("./data.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	products := []struct {
		name, description, category, slug, image string
		price                                     int
	}{
		{"Test Product", "A test description", "prints", "test-product", "https://picsum.photos/seed/test-product/400/300", 1999},
		{"Risograph Print No. 1", "Two-colour risograph print, A3 format.", "prints", "risograph-print-1", "https://picsum.photos/seed/risograph-print-1/400/300", 3500},
		{"Poster — Grid Series", "Limited edition grid study poster.", "prints", "poster-grid-series", "https://picsum.photos/seed/poster-grid-series/400/300", 2800},
		{"Tote Bag", "Heavy canvas tote with screen printed logo.", "apparel", "tote-bag", "https://picsum.photos/seed/tote-bag/400/300", 2200},
		{"Zine Vol. 1", "First edition zine, hand-numbered, 40 pages.", "zines", "zine-vol-1", "https://picsum.photos/seed/zine-vol-1/400/300", 1500},
		{"Sticker Pack", "Set of 6 die-cut vinyl stickers.", "accessories", "sticker-pack", "https://picsum.photos/seed/sticker-pack/400/300", 800},
	}

	for _, p := range products {
		_, err := database.Exec(
			`INSERT OR IGNORE INTO products (name, description, price, category, slug, image) VALUES (?, ?, ?, ?, ?, ?)`,
			p.name, p.description, p.price, p.category, p.slug, p.image,
		)
		if err != nil {
			log.Fatalf("failed to seed product %s: %v", p.slug, err)
		}
	}

	log.Println("seed complete")
}
