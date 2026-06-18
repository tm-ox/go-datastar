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
    db.go                    — SQLite connection; WAL + busy_timeout on file DBs
    migrate.go               — schema migrations, run at startup
  store/                     — one package; concrete stores over *sql.DB, no interfaces (see docs/adr/0001)
    cart.go                  — CartStore, CartSummary, Summary; shared itemDetails/subtotal helpers
    product.go               — ProductStore, Product, ProductQuery; List/Filter/Get/Create/Update/Delete/UpdateStock
    work.go                  — WorkStore, Work (with Images), WorkQuery; List/Filter/Get/Create/Update/Delete/Unique*
    order.go                 — OrderStore.Place (one tx: persist order + clear cart), ErrEmptyCart
    *_test.go                — store tests against in-memory SQLite
  render/
    render.go                — Page: full-page render, or Datastar SSE shell-patch — the page shell in one place
  handler/
    constants.go             — defaultLimit = 20
    site.go                  — SiteHandler: Index, About
    shop.go                  — ShopHandler: Index, Filter, Detail
    settings.go              — SettingsHandler: Work* and Shop* CRUD
    work.go                  — WorkHandler: Index, Filter, Detail
    cart.go                  — CartHandler: Add, Remove, Total, Drawer, UpdateQty, Checkout, PlaceOrder, Success
  middleware/
    cart.go                  — injects cart total into request context
    logging.go               — request logging middleware
views/
  layouts/
    base.templ               — base HTML layout (id="main" swap target, id="site-header" for nav patch)
    sub.templ                — SubLayout (unused)
  modules/                   — Navbar, TabBar, Hero, Card, ProductCard, Button, Icon, Footer, Pagination, Search
  pages/                     — *Content partials wrapped by render.Page; NotFound and CartSuccess are full pages
static/input.css             — Tailwind source (theme tokens, base styles, component classes)
```

## Architecture notes

- **Stores are concrete, not interfaces.** Each store is a struct over `*sql.DB`;
  swapping the backend means rewriting `internal/store`, not the callers. Tested
  against in-memory SQLite. See `docs/adr/0001`.
- **`render.Page` owns the page shell.** Handlers pass a `*Content` partial plus
  nav/path/meta; `render.Page` decides between a full `BaseLayout` render and a
  Datastar SSE shell-patch (`site-header` + `main`). Domain vocabulary lives in
  `CONTEXT.md`.

## Test

```bash
go test ./...
```

Store, render, and db tests run against a throwaway in-memory SQLite database — no
fixtures, no external services.

## SQLite Tables

| Table | Purpose |
|---|---|
| `products` | Shop products |
| `work` | Work portfolio entries |
| `work_images` | Work entry images (FK → work.id) |
| `site_pages` | Page-level copy (title, tagline, body) |
| `site_sections` | Page sections |
| `site_cards` | Section cards |
