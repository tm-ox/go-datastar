package main

import (
	"fmt"
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
		name, description, category, slug string
		price, stock                      int
	}{
		{"Test Product", "A test description.", "prints", "test-product", 1999, 5},
		{"Risograph Print No. 1", "Two-colour risograph print, A3 format.", "prints", "risograph-print-1", 3500, 12},
		{"Poster — Grid Series", "Limited edition grid study poster.", "prints", "poster-grid-series", 2800, 0},
		{"Tote Bag", "Heavy canvas tote with screen printed logo.", "apparel", "tote-bag", 2200, 8},
		{"Zine Vol. 1", "First edition zine, hand-numbered, 40 pages.", "zines", "zine-vol-1", 1500, 3},
		{"Sticker Pack", "Set of 6 die-cut vinyl stickers.", "accessories", "sticker-pack", 800, 25},
		{"Risograph Print No. 2", "Three-colour risograph print, A2 format.", "prints", "risograph-print-2", 4500, 6},
		{"Enamel Pin — Logo", "Hard enamel pin, 30mm, gold plating.", "accessories", "enamel-pin-logo", 1200, 40},
		{"Zine Vol. 2", "Second edition zine, 52 pages, colour cover.", "zines", "zine-vol-2", 1800, 0},
		{"Poster — Type Series No. 1", "Typographic poster, A2, two colours.", "prints", "poster-type-1", 3200, 9},
		{"Crewneck Sweatshirt", "Heavyweight 400gsm crewneck, embroidered logo.", "apparel", "crewneck-sweatshirt", 7500, 15},
		{"Notebook — Grid", "A5 grid notebook, 120 pages, lay-flat binding.", "accessories", "notebook-grid", 1600, 20},
		{"Poster — Type Series No. 2", "Typographic poster, A1, single colour.", "prints", "poster-type-2", 3800, 0},
		{"Cap", "Six-panel cap, embroidered logo, adjustable strap.", "apparel", "cap", 4200, 11},
		{"Sticker Sheet", "A4 sheet of 12 assorted stickers.", "accessories", "sticker-sheet", 600, 50},
		{"Zine Vol. 3", "Third edition zine, perfect bound, 64 pages.", "zines", "zine-vol-3", 2200, 7},
		{"Risograph Print No. 3", "Four-colour risograph, A3, limited to 50.", "prints", "risograph-print-3", 5500, 4},
		{"Tote Bag — Natural", "Natural canvas tote, screen printed design.", "apparel", "tote-bag-natural", 2400, 0},
		{"Poster — Colour Study No. 1", "A3 colour field study, two-colour risograph.", "prints", "poster-colour-study-1", 2600, 8},
		{"Woven Patch", "Embroidered woven patch, iron-on backing, 80mm.", "accessories", "woven-patch", 900, 60},
		{"Zine Vol. 4", "Fourth edition, landscape format, 48 pages.", "zines", "zine-vol-4", 2000, 0},
		{"Hoodie", "Garment-dyed heavyweight hoodie, screen printed.", "apparel", "hoodie", 8900, 6},
		{"Risograph Print No. 4", "Two-colour A4 print, edition of 100.", "prints", "risograph-print-4", 2200, 14},
		{"Keyring", "Die-cast metal keyring, enamel fill.", "accessories", "keyring", 700, 35},
		{"Poster — Abstract No. 1", "Abstract composition, A2, screen printed.", "prints", "poster-abstract-1", 3400, 0},
		{"Long Sleeve Tee", "Heavyweight long sleeve, embroidered chest logo.", "apparel", "long-sleeve-tee", 5500, 9},
		{"Notebook — Ruled", "A5 ruled notebook, 120 pages, lay-flat binding.", "accessories", "notebook-ruled", 1600, 18},
		{"Zine Vol. 5", "Fifth edition, saddle-stitched, 32 pages.", "zines", "zine-vol-5", 1200, 22},
		{"Poster — Colour Study No. 2", "A2 colour field study, three-colour risograph.", "prints", "poster-colour-study-2", 3600, 5},
		{"Bucket Hat", "Cotton bucket hat, embroidered logo, unstructured.", "apparel", "bucket-hat", 4800, 0},
		{"Enamel Pin — Type", "Hard enamel pin, typographic design, 25mm.", "accessories", "enamel-pin-type", 1100, 45},
		{"Risograph Print No. 5", "Five-colour A3 print, edition of 30.", "prints", "risograph-print-5", 6500, 3},
		{"T-Shirt — Logo", "Medium-weight tee, screen printed front logo.", "apparel", "tshirt-logo", 4500, 20},
		{"Poster — Abstract No. 2", "Abstract composition, A1, two-colour screen print.", "prints", "poster-abstract-2", 4200, 7},
		{"Tote Bag — Black", "Black canvas tote, screen printed white logo.", "apparel", "tote-bag-black", 2400, 0},
	}

	for _, p := range products {
		image := fmt.Sprintf("https://picsum.photos/seed/%s/400/300", p.slug)
		_, err := database.Exec(
			`INSERT OR IGNORE INTO products (name, description, price, category, slug, image, stock) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			p.name, p.description, p.price, p.category, p.slug, image, p.stock,
		)
		if err != nil {
			log.Fatalf("failed to seed product %s: %v", p.slug, err)
		}
	}

	log.Println("products seed complete")
}
