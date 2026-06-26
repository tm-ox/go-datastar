# go-datastar

A learning project exploring Go, [Datastar](https://data-star.dev), [templ](https://templ.guide), and Tailwind v4.

## Stack

- **Go 1.26** — `net/http` stdlib router, no framework
- **templ** — type-safe HTML templates compiled to Go
- **Datastar v1.0.1** — SSE-based reactivity (server-driven UI)
- **Tailwind v4** — CSS-first, no config file
- **SQLite** — `modernc.org/sqlite`, pure Go, no CGo
- **Shopify Storefront API** — products sourced via GraphQL (`genqlient`)

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
go run ./cmd/seed/work/
go run ./cmd/seed/work_images/
go run ./cmd/seed/content/
```

> Must be run in order — `work_images` depends on work IDs existing.
> Products are sourced from Shopify — no seed step required.

## Dev

Copy `.env.example` to `.env` and fill in values:

```bash
cp .env.example .env
```

```
ADMIN_PASSWORD=yourpassword
SESSION_SECRET=<output of: openssl rand -hex 32>
SHOPIFY_STORE_DOMAIN=your-store.myshopify.com
SHOPIFY_STOREFRONT_TOKEN=your-storefront-access-token
SMTP_HOST=mail.spacemail.com
SMTP_PORT=587
SMTP_USER=hello@tmox.net
SMTP_PASS=yourpassword
```

> `SHOPIFY_STORE_DOMAIN` and `SHOPIFY_STOREFRONT_TOKEN` are required — the server will not start without them.
> `SMTP_*` vars are required for the contact form. Values with special characters must be single-quoted in `.env`.

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
| GET | `/context` | `site.Context` |
| GET | `/dashboard` | `dashboard.Index` |
| GET | `/dashboard/stream` | `dashboard.Stream` (long-lived Datastar SSE) |
| GET | `/work` | `work.Index` |
| GET | `/work/{slug}` | `work.Detail` |
| GET | `/work/filter` | `work.Filter` (Datastar SSE) |
| GET | `/shop` | `shop.Index` |
| GET | `/shop/{slug}` | `shop.Detail` |
| GET | `/shop/filter` | `shop.Filter` (Datastar SSE) |
| POST | `/cart/add` | `cart.Add` (Datastar SSE) |
| GET | `/cart/total` | `cart.Total` (Datastar SSE) |
| GET | `/cart/drawer` | `cart.Drawer` (Datastar SSE) |
| POST | `/cart/remove` | `cart.Remove` (Datastar SSE) |
| GET | `/cart` | `cart.Checkout` |
| POST | `/cart/qty` | `cart.UpdateQty` (Datastar SSE) |
| POST | `/cart/checkout` | `cart.PlaceOrder` |
| GET | `/cart/success` | `cart.Success` |
| POST | `/contact` | `contact.Send` (Datastar SSE) |
| POST | `/login` | `auth.Login` (Datastar SSE) |
| GET | `/logout` | `auth.Logout` |
| GET | `/settings` | redirect → `/settings/work` |
| GET | `/settings/work` | `settings.Work` (requires auth) |
| GET | `/settings/work/filter` | `settings.WorkFilter` (Datastar SSE, requires auth) |
| GET | `/settings/work/form` | `settings.WorkForm` (Datastar SSE, requires auth) |
| POST | `/settings/work/create` | `settings.WorkCreate` (Datastar SSE, requires auth) |
| POST | `/settings/work/update` | `settings.WorkUpdate` (Datastar SSE, requires auth) |
| POST | `/settings/work/delete` | `settings.WorkDelete` (Datastar SSE, requires auth) |

## Structure

```
cmd/
  main.go                    — server wiring, route registration, startup
  seed/
    work/main.go             — 10 work entries (INSERT OR IGNORE)
    work_images/main.go      — 86 images across 10 entries
    content/main.go          — site_pages, site_sections, site_cards
internal/
  content/
    site.go                  — site types (HomePage, AboutPage, Section, Card) + Load() + LoadContext() (goldmark → HTML)
    content.yaml             — home + about copy
    CONTEXT.md               — embedded markdown, rendered to HTML via LoadContext()
  db/
    db.go                    — SQLite connection; WAL + busy_timeout on file DBs
    migrate.go               — schema migrations, run at startup
  shopify/
    store.go                 — ProductStore: List, Filter (client-side search/filter), GetByHandle, FilterMeta
    queries.graphql          — ListProducts, GetProduct, GetCollectionFilters queries
    schema.graphql           — Shopify Storefront API schema (2025-01)
    genqlient.yaml           — genqlient config; run go generate ./internal/shopify/... to regenerate
    generated.go             — genqlient-generated typed GraphQL client (do not edit)
  store/                     — one package; concrete stores over *sql.DB, no interfaces (see docs/adr/0001)
    cart.go                  — CartStore, CartSummary, Summary; shared itemDetails/subtotal helpers
    product.go               — Product, ProductQuery domain types (products sourced from Shopify, not SQLite)
    work.go                  — WorkStore, Work (with Images), WorkQuery; List/Filter/Get/Create/Update/Delete/Unique*
    order.go                 — OrderStore.Place (one tx: persist order + clear cart), ErrEmptyCart
    *_test.go                — store tests against in-memory SQLite
  render/
    render.go                — Page: full-page render, or Datastar SSE shell-patch — the page shell in one place
  stream/                    — in-memory realtime pipeline for the Dashboard (no SQLite; see CONTEXT.md "Live data")
    stream.go                — Event + ParseEvent (Wikimedia recentchange payload → Event)
    hub.go                   — Hub: fan-out one Source to many subscribers; bounded buffer, drop-on-full, recent-event ring
    source.go                — Source: single upstream connection to stream.wikimedia.org; feeds Aggregator + Hub
    aggregator.go            — Aggregator + Stats: rolling counters, per-wiki top-N, per-second sparkline buckets; read via Snapshot
    *_test.go                — ParseEvent, Hub, Aggregator unit tests (fixtures + fake channels, no network)
  handler/
    constants.go             — defaultLimit = 9
    auth.go                  — AuthHandler: Login, Logout; HMAC-signed session cookie
    contact.go               — ContactHandler: Send; honeypot, validation, SMTP delivery via Spacemail
    site.go                  — SiteHandler: Index, Context
    dashboard.go             — DashboardHandler: Index (skeleton), Stream (long-lived SSE: feed per-event + tiles/charts on 1s ticker)
    shop.go                  — ShopHandler: Index, Filter, Detail; productReader consumer-side interface
    settings.go              — SettingsHandler: Work CRUD
    work.go                  — WorkHandler: Index, Filter, Detail
    cart.go                  — CartHandler: Add (Shopify stock check), Remove, Total, Drawer, UpdateQty, Checkout, PlaceOrder, Success
  middleware/
    auth.go                  — RequireAuth: HMAC cookie verification; opens login modal on Datastar requests, redirects on full-page
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

- **Products are Shopify-sourced.** `internal/shopify.ProductStore` fetches products from the Shopify Storefront API via `genqlient`-generated GraphQL queries. `ShopHandler` depends on a small consumer-side `productReader` interface; `CartHandler` uses a `productStockChecker` interface — both satisfied by `shopify.ProductStore`. Filtering (type, vendor, in-stock, search) is applied client-side in Go after fetching up to 250 products, avoiding Shopify `ProductFilter` serialisation quirks with zero-value booleans.
- **Stores are concrete, not interfaces.** SQLite stores (`WorkStore`, `CartStore`, `OrderStore`) are structs over `*sql.DB`; swapping the backend means rewriting `internal/store`, not the callers. Tested against in-memory SQLite. See `docs/adr/0001`.
- **`render.Page` owns the page shell.** Handlers pass a `*Content` partial plus nav/path/meta; `render.Page` decides between a full `BaseLayout` render and a Datastar SSE shell-patch (`site-header` + `main`). Domain vocabulary lives in `CONTEXT.md`.
- **The Dashboard pushes, it doesn't pull.** `internal/stream` is a realtime pipeline: one `Source` holds a single upstream connection to `stream.wikimedia.org`, fans every `Event` out through an in-memory `Hub` to many browser SSE subscribers (bounded buffer, drop-on-full so one slow client can't stall the rest), and feeds an `Aggregator` of rolling stats read via `Snapshot`. `/dashboard/stream` holds a long-lived SSE connection — opened by `data-init` and torn down by `requestCancellation: 'cleanup'` on SPA navigate-away — patching a live feed per event and the tiles/charts on a 1s ticker. Everything is in-memory; state resets on restart, nothing is stored.

## Test

```bash
go test ./...
```

Store, render, and db tests run against a throwaway in-memory SQLite database — no fixtures, no external services.

## SQLite Tables

| Table | Purpose |
|---|---|
| `work` | Work portfolio entries |
| `work_images` | Work entry images (FK → work.id) |
| `cart_items` | Active cart line items (keyed by cart_id cookie) |
| `orders` | Placed orders |
| `order_items` | Order line items (FK → orders.id) |
| `site_pages` | Page-level copy (title, tagline, body) |
| `site_sections` | Page sections |
| `site_cards` | Section cards |
