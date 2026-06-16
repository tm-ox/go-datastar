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

	entries := []struct {
		slug, title, workType, client, tools, description, website, link, coverURL string
		year, sort_order                                                            int
	}{
		{
			slug: "logos", sort_order: 0, title: "Logo Design", workType: "Logo Design", client: "Various", year: 2024,
			tools:       "Adobe Illustrator,Inkscape",
			description: "Various logo designs.",
			website:     "", link: "",
			coverURL: "https://tmox.net/_astro/dark-house.BnYhrUjj_9FW7.webp",
		},
		{
			slug: "tg", sort_order: 3, title: "TripleGoal Platform UI", workType: "UI Design", client: "TripleGoal", year: 2023,
			tools:       "Figma",
			description: "UI design for corporate behaviour management platform TripleGoal, designed in Figma.",
			website:     "triplegoal.com", link: "https://triplegoal.com",
			coverURL: "https://tmox.net/_astro/tg-gm-login.D2fPD4ob_Z1FzXOr.webp",
		},
		{
			slug: "ntr", sort_order: 7, title: "Nags to Riches Catalogue", workType: "Print Design", client: "Nags to Riches", year: 2022,
			tools:       "InDesign",
			description: "Catalogue for Australian professional equine product company Nags to Riches.",
			website:     "nagstoriches.com.au", link: "https://www.nagstoriches.com.au",
			coverURL: "https://tmox.net/_astro/NTR001-Brochure-1.DEt378LL_1K1QhN.webp",
		},
		{
			slug: "oxen-free", sort_order: 2, title: "Oxen Free", workType: "ID Design", client: "Oxen Free", year: 2021,
			tools:       "Adobe Illustrator,InDesign",
			description: "Visual identity and selected collateral for Yogyakartan bar & kitchen.",
			website:     "oxenfree.net", link: "https://oxenfree.net",
			coverURL: "https://tmox.net/_astro/ox-logo-min.Yaeft7cE_1XtU4B.webp",
		},
		{
			slug: "redsky", sort_order: 4, title: "RedSky", workType: "Print Design", client: "RedSky", year: 2022,
			tools:       "InDesign",
			description: "Training cards for Australian corporate training firm Redsky.",
			website:     "redsky.com.au", link: "https://redsky.com.au",
			coverURL: "https://tmox.net/_astro/RedSky-01.C78wcXrD_Z1eFw1.webp",
		},
		{
			slug: "print", sort_order: 1, title: "Print Design", workType: "Print Design", client: "Various", year: 2023,
			tools:       "Adobe Illustrator,InDesign",
			description: "Various print designs including flyers, catalogues, and event collateral.",
			website:     "", link: "",
			coverURL: "https://tmox.net/_astro/compass-a4.-vzdV72V_2bRyxX.webp",
		},
		{
			slug: "saj", sort_order: 5, title: "Share A Jet", workType: "UI Design", client: "Share A Jet", year: 2023,
			tools:       "Penpot",
			description: "UI design for private jet charter company Share A Jet, designed in Penpot.",
			website:     "shareajet.vip", link: "https://www.shareajet.vip",
			coverURL: "https://tmox.net/_astro/saj-desktop-1.yzr4oNf7_1lGKVw.webp",
		},
		{
			slug: "fredst", sort_order: 6, title: "Fred St.", workType: "Web Design", client: "Fred St.", year: 2022,
			tools:       "Alpine.js,Netlify CMS,Nuxt Content,Tailwind CSS,Vue/Nuxt3",
			description: "Website for Australian landscape architecture firm.",
			website:     "fredst.com", link: "https://fredst.com",
			coverURL: "https://tmox.net/_astro/fredst-web-projects-2.DP1KU6Gd_21VsVN.webp",
		},
		{
			slug: "ynk", sort_order: 8, title: "Yes No Klub", workType: "Web Design", client: "Yes No Klub", year: 2021,
			tools:       "Alpine.js,Netlify CMS,Nuxt Content,Tailwind CSS,Vue/Nuxt3",
			description: "Website for music event organisation Yes No Klub.",
			website:     "yesnoklub.net", link: "https://yesnoklub.net",
			coverURL: "https://tmox.net/_astro/ynk-web-dark-home.CPv2KHFJ_1MRK7C.webp",
		},
		{
			slug: "mjo", sort_order: 9, title: "MJO Mortgage Solutions", workType: "Web Design", client: "MJO", year: 2022,
			tools:       "Alpine.js,Netlify CMS,Nuxt Content,Tailwind CSS,Vue/Nuxt3",
			description: "Website for Australian mortgage broker.",
			website:     "mjomortgagesolutions.com.au", link: "https://www.mjomortgagesolutions.com.au",
			coverURL: "https://tmox.net/_astro/mjo-web-landing.DLoFeW4D_16YoIb.webp",
		},
	}

	for _, w := range entries {
		_, err := database.Exec(
			`INSERT OR IGNORE INTO work (slug, sort_order, title, type, client, year, tools, description, website, link, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			w.slug, w.sort_order, w.title, w.workType, w.client, w.year, w.tools, w.description, w.website, w.link, w.coverURL,
		)
		if err != nil {
			log.Fatalf("failed to seed work %s: %v", w.slug, err)
		}
	}

	log.Println("work seed complete")
}
