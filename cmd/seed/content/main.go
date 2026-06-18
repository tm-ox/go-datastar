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

	pages := []struct {
		key, metaTitle, metaDesc, title, tagline, body string
	}{
		{
			key:       "home",
			metaTitle: "Home",
			metaDesc:  "A reference implementation built with Go, templ, Datastar, and Tailwind v4.",
			title:     "go-datastar",
			tagline:   "Server-driven UI without the JavaScript framework.",
		},
		{
			key:       "about",
			metaTitle: "About",
			metaDesc:  "About description.",
			title:     "About",
			body:      "Brief bio here.",
		},
	}

	for _, p := range pages {
		_, err := database.Exec(
			`INSERT OR IGNORE INTO site_pages (key, meta_title, meta_description, title, tagline, body) VALUES (?, ?, ?, ?, ?, ?)`,
			p.key, p.metaTitle, p.metaDesc, p.title, p.tagline, p.body,
		)
		if err != nil {
			log.Fatalf("failed to seed page %s: %v", p.key, err)
		}
	}

	sections := []struct {
		pageKey, sectionKey, title, tagline string
		cols, sortOrder                     int
	}{
		{pageKey: "home", sectionKey: "stack", title: "The Stack", cols: 0, sortOrder: 1},
		{pageKey: "home", sectionKey: "patterns", title: "Patterns Implemented", cols: 2, sortOrder: 2},
		{pageKey: "home", sectionKey: "explore", title: "Explore", cols: 0, sortOrder: 3},
	}

	sectionIDs := map[string]int64{}
	for _, s := range sections {
		res, err := database.Exec(
			`INSERT OR IGNORE INTO site_sections (page_key, section_key, title, tagline, cols, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
			s.pageKey, s.sectionKey, s.title, s.tagline, s.cols, s.sortOrder,
		)
		if err != nil {
			log.Fatalf("failed to seed section %s: %v", s.sectionKey, err)
		}
		id, _ := res.LastInsertId()
		sectionIDs[s.sectionKey] = id
	}

	type card struct {
		sectionKey, title, description, href, icon, buttonText, buttonHref, buttonTarget string
		sortOrder                                                                        int
	}

	cards := []card{
		{
			sectionKey: "stack", sortOrder: 1,
			title:       "Server-Driven UI",
			icon:        "bug",
			description: "Datastar replaces client-side state with SSE streams from the Go server. No virtual DOM, no hydration - the server owns the UI.",
			buttonText:  "Datastar docs", buttonHref: "https://data-star.dev", buttonTarget: "_blank",
		},
		{
			sectionKey: "stack", sortOrder: 2,
			title:       "Type-Safe Templates",
			icon:        "bug",
			description: "Templ compiles HTML components to Go functions. Template errors are caught at build time, not in the browser.",
			buttonText:  "Templ docs", buttonHref: "https://templ.guide", buttonTarget: "_blank",
		},
		{
			sectionKey: "stack", sortOrder: 3,
			title:       "No Framework",
			icon:        "bug",
			description: "Routing, middleware, and handlers use net/http stdlib only. No Gin, no Echo - just Go.",
			buttonText:  "Source", buttonHref: "https://github.com/tm-ox/go-datastar", buttonTarget: "_blank",
		},
		{
			sectionKey: "stack", sortOrder: 4,
			title:       "CSS-First Styling",
			icon:        "bug",
			description: "Tailwind v4 with no config file. Theme tokens defined in CSS, not JavaScript.",
			buttonText:  "Tailwind v4 docs", buttonHref: "https://tailwindcss.com/blog/tailwindcss-v4", buttonTarget: "_blank",
		},
		{
			sectionKey: "patterns", sortOrder: 1,
			title:       "SSE Filtering",
			icon:        "bug",
			description: "Live product and work filtering via Datastar SSE - no page reload, server patches only the changed fragment.",
		},
		{
			sectionKey: "patterns", sortOrder: 2,
			title:       "Pagination",
			icon:        "bug",
			description: "Cursor-based pagination with disabled-state buttons, total count, and SSE-compatible partial rendering.",
		},
		{
			sectionKey: "patterns", sortOrder: 3,
			title:       "SQLite + Interfaces",
			icon:        "bug",
			description: "ProductStore interface with SQLiteProductStore implementation - swappable backend, dynamic WHERE clause building.",
		},
		{
			sectionKey: "patterns", sortOrder: 4,
			title:       "Content from YAML",
			icon:        "bug",
			description: "Site copy and work entries loaded from embedded YAML at startup - no CMS, no database round-trips for static content.",
		},
		{
			sectionKey: "explore", sortOrder: 1,
			title:       "Work",
			icon:        "bug",
			description: "Filterable work entries loaded from YAML. Demonstrates SSE partial rendering and tag-based filtering.",
			buttonText:  "View work", buttonHref: "/work",
		},
		{
			sectionKey: "explore", sortOrder: 2,
			title:       "Shop",
			icon:        "bug",
			description: "Product catalogue backed by SQLite. Filter by category and stock, paginate results - all server-driven.",
			buttonText:  "View shop", buttonHref: "/shop",
		},
	}

	for _, c := range cards {
		sectionID := sectionIDs[c.sectionKey]
		_, err := database.Exec(
			`INSERT OR IGNORE INTO site_cards (section_id, title, description, href, icon, sort_order, button_text, button_href, button_target) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			sectionID, c.title, c.description, c.href, c.icon, c.sortOrder, c.buttonText, c.buttonHref, c.buttonTarget,
		)
		if err != nil {
			log.Fatalf("failed to seed card %q: %v", c.title, err)
		}
	}

	log.Println("content seed complete")
}
