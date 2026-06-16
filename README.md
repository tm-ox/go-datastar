# go-datastar

A learning project exploring Go, [Datastar](https://data-star.dev), [templ](https://templ.guide), and Tailwind v4.

## Stack

- **Go 1.26** — `net/http` stdlib router, no framework
- **templ** — type-safe HTML templates compiled to Go
- **Datastar v1.0.1** — SSE-based reactivity (server-driven UI)
- **Tailwind v4** — CSS-first, no config file
- **SQLite** — `modernc.org/sqlite`, pure Go, no CGo

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

Populate the database. Run each command once after first startup — safe to re-run, uses `INSERT OR IGNORE`.

```bash
go run ./cmd/seed/products/
go run ./cmd/seed/work/
go run ./cmd/seed/work_images/
go run ./cmd/seed/content/
```

> Must be run in order — `work_images` depends on work IDs existing.

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

If air serves a stale binary after changes: `rm tmp/server && make dev`.

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
| GET | `/settings/work/filter` | `settings.WorkFilter` (Datastar SSE) |
| GET | `/settings/work/form` | `settings.WorkForm` (Datastar SSE) |
| POST | `/settings/work/create` | `settings.WorkCreate` (Datastar SSE) |
| POST | `/settings/work/update` | `settings.WorkUpdate` (Datastar SSE) |
| POST | `/settings/work/delete` | `settings.WorkDelete` (Datastar SSE) |
| GET | `/settings/shop` | `settings.Shop` |
| GET | `/settings/shop/filter` | `settings.ShopFilter` (Datastar SSE) |
| POST | `/settings/shop/stock` | `settings.ShopStock` (Datastar SSE) |
| GET | `/settings/shop/products/form` | `settings.ShopProductForm` (Datastar SSE) |
| POST | `/settings/shop/products/create` | `settings.ShopProductCreate` (Datastar SSE) |
| POST | `/settings/shop/products/update` | `settings.ShopProductUpdate` (Datastar SSE) |
| POST | `/settings/shop/products/delete` | `settings.ShopProductDelete` (Datastar SSE) |

## Structure

```
cmd/
  main.go                    — server wiring, route registration, startup
  seed/
    products/main.go         — 36 products (INSERT OR IGNORE)
    work/main.go             — 10 work entries (INSERT OR IGNORE)
    work_images/main.go      — 86 images across 10 entries
    content/main.go          — site_pages, site_sections, site_cards
internal/
  content/
    site.go                  — site types (HomePage, AboutPage, Section, Card) + Load()
    content.yaml             — home + about copy
  db/
    db.go                    — SQLite connection (modernc.org/sqlite)
    migrate.go               — schema migrations, run at startup
  store/
    product/
      product.go             — Product struct, ProductStore interface
      sqlite.go              — SQLiteProductStore: List, Filter, GetBySlug, GetByID, UpdateStock, Create, Update, Delete
    work/
      work.go                — Work struct (with Images []WorkImage), WorkStore interface
      sqlite.go              — SQLiteWorkStore: List, Filter, GetBySlug, GetByID, Create, Update, Delete, UniqueTypes/Clients/Years/Tools
  handler/
    constants.go             — defaultLimit = 20
    site.go                  — SiteHandler: Index, About
    shop.go                  — ShopHandler: Index, Filter, Detail
    settings.go              — SettingsHandler: Work, WorkFilter, WorkForm, WorkCreate, WorkUpdate, WorkDelete, Shop, ShopFilter, ShopStock, ShopProductForm, ShopProductCreate, ShopProductUpdate, ShopProductDelete
    work.go                  — WorkHandler: Index, Filter, Detail
  middleware/
    logging.go               — request logging middleware
views/
  layouts/
    base.templ               — base HTML layout (id="main" swap target, id="site-header" for nav patch)
    sub.templ                — SubLayout (unused — settings moved to BaseLayout + inline subnav)
  modules/                   — Navbar, TabBar, Hero, Card, CardProduct, Button, Icon, Footer, Pagination, Search
  pages/                     — page templates; each page has a full wrapper + *Content partial for SPA swap
static/input.css             — Tailwind source (theme tokens, base styles, component classes)
```

## SQLite Tables

| Table | Purpose |
|---|---|
| `products` | Shop products |
| `work` | Work portfolio entries |
| `work_images` | Work entry images (FK → work.id) |
| `site_pages` | Page-level copy (title, tagline, body) |
| `site_sections` | Page sections |
| `site_cards` | Section cards |
