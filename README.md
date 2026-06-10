# go-datastar

A learning project exploring Go, [Datastar](https://data-star.dev), [templ](https://templ.guide), and Tailwind v4.

## Stack

- **Go 1.26** — `net/http` stdlib router, no framework
- **templ** — type-safe HTML templates compiled to Go
- **Datastar v1.0.1** — SSE-based reactivity (server-driven UI)
- **Tailwind v4** — CSS-first, no config file
- **SQLite** — `modernc.org/sqlite`, pure Go, no CGo
- **YAML** — content stored as plain YAML files

## Prerequisites

```bash
go install github.com/a-h/templ/cmd/templ@latest
curl -fsSL https://bun.sh/install | bash
```

## Setup

```bash
bun install
templ generate
bun run build:css
```

## Seed

Populate the database with sample products:

```bash
go run cmd/seed/main.go
```

Run once after first startup. Safe to re-run — uses `INSERT OR IGNORE`.

## Dev

```bash
make dev
```

Runs four processes concurrently: Go server (`:8081`), templ watcher, Tailwind watcher, browser-sync proxy (`:3000`).

| Layer | Port | Purpose |
|---|---|---|
| Go server | 8081 | App |
| browser-sync | 3000 | CSS injection without full reload |

Open `http://localhost:3000`.

## Routes

| Method | Path | Handler |
|---|---|---|
| GET | `/` | `site.Index` |
| GET | `/about` | `site.About` |
| GET | `/work` | `work.Index` |
| GET | `/work/{slug}` | `work.Detail` |
| GET | `/work/filter` | `work.Filter` (Datastar SSE) |
| GET | `/shop` | `shop.Index` |
| GET | `/shop/{slug}` | `shop.Detail` |
| GET | `/shop/filter` | `shop.Filter` (Datastar SSE) |
| GET | `/settings` | redirect → `/settings/work` |
| GET | `/settings/work` | `settings.Work` |
| GET | `/settings/shop` | `settings.Shop` |
| GET | `/settings/shop/filter` | `settings.ShopFilter` (Datastar SSE) |
| POST | `/settings/shop/stock` | `settings.ShopStock` (Datastar SSE) |

## Structure

```
cmd/
  main.go                  — server wiring, route registration, startup
  seed/main.go             — development seed script (36 products)
internal/
  content/
    site.go                — site types (HomePage, AboutPage, Section, Card) + Load()
    work.go                — WorkEntry, LoadWork(), FilterWork(), SortWork(), PaginateWork()
    content.yaml           — home + about data
    work/*.yaml            — one file per work entry
  db/
    db.go                  — SQLite connection (modernc.org/sqlite)
    migrate.go             — schema migrations, run at startup
  store/product/
    product.go             — Product struct, ProductStore interface
    sqlite.go              — SQLiteProductStore: list, filter, detail, UpdateStock
  handler/
    site.go                — Index, About handlers
    shop.go                — ShopHandler — product listing, filtering, detail
    settings.go            — SettingsHandler — settings pages, shop inventory management
    work.go                — WorkHandler — work listing, filtering, detail
  middleware/
    logging.go             — request logging middleware
views/
  layouts/
    base.templ             — base HTML layout
    sub.templ              — SubLayout — settings subnav via TabBar
  modules/                 — Navbar, TabBar, Hero, Card, CardProduct, Button, Icon, Footer, Pagination, Search
  pages/                   — page templates + SSE partials
static/input.css           — Tailwind source (theme tokens, base styles, component classes)
```
