# go-datastar

A learning project exploring Go, [Datastar](https://data-star.dev), [templ](https://templ.guide), and Tailwind v4.

## Stack

- **Go 1.26** — `net/http` stdlib router, no framework
- **templ** — type-safe HTML templates compiled to Go
- **Datastar v1.0.1** — SSE-based reactivity (server-driven UI)
- **Tailwind v4** — CSS-first, no config file
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

## Dev

```bash
make dev
```

Runs four processes concurrently: Go server (`:8081`), templ watcher, Tailwind watcher, browser-sync proxy (`:3000`).

| Layer | Port | Purpose |
|---|---|---|
| Go server | 8081 | App |
| templ proxy | 7331 | Full reload on `.templ` changes |
| browser-sync | 3000 | CSS injection without full reload |

Open `http://localhost:3000`.

## Structure

```
cmd/main.go                  — server wiring, route registration, startup
internal/
  content/
    site.go                  — site types (HomePage, AboutPage) + Load()
    work.go                  — WorkEntry, LoadWork(), filter/sort helpers
    content.yaml             — home + about data
    work/*.yaml              — one file per work entry
  handler/
    site.go                  — Index, About handlers
    work.go                  — WorkIndex, WorkDetail, Filter handlers
  middleware/
    logging.go               — request logging middleware
views/
  layouts/base.templ         — base HTML layout
  modules/                   — shared components (navbar, hero, card)
  pages/                     — page templates
static/input.css             — Tailwind source (theme tokens, base styles)
```
