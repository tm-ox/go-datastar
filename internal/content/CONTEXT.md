## Context

> A proof of concept site built with Go, Templ, Datastar and Tailwind: a realtime **Dashboard**
> of public Wikimedia activity, a public **Work** portfolio and a **Shop** of **Products**.
> The UI is server-driven over Datastar SSE — both
> request-response (filters, cart, CRUD) and long-lived server push (the
> Dashboard stream). This file fixes the domain language so modules and
> conversations use one word per concept.

### Entities

The domain nouns — the things that are stored and reasoned about.

**Work**:
A portfolio entry — a project shown on the site, with images, type, client,
year, and tools. Note: **Work** is also the name of a Surface (below). The
entity is the thing; the Surface is the page that lists it.
_Avoid_: project, portfolio item, entry

**Product**:
An item for sale in the Shop, with price (stored as integer cents), stock, and
category.
_Avoid_: item, SKU, listing

**Cart**:
A guest's in-progress selection of Products, keyed by a `cart_id` cookie. Holds
line items and computes its own subtotal and count.
_Avoid_: basket, bag

**Order**:
A Cart turned into a placed purchase — persisted with its total, after which the
Cart is cleared.
_Avoid_: purchase, transaction, checkout (checkout is the _act_, not the thing)

**Site content**:
Editorial copy for the Home and About pages, loaded from embedded YAML rather
than the database.
_Avoid_: CMS, pages-table

### Surfaces

The places a visitor goes — sections and pages. A Surface is not an Entity:
the **Shop** Surface lists **Product** entities; the **Work** Surface lists
**Work** entities (same word, different layer).

**Shop**, **Work**, **Settings**, **About**, **Home**, **Dashboard** — the top-level sections.

**Dashboard**:
The Surface showing live Wikimedia activity, streamed over a long-lived SSE
connection. Unlike every other Surface it has no stored Entity — its content is
ephemeral, held in memory (see Live data) and pushed by the server, not fetched
on request.

**Checkout**:
The Surface (and the act) of reviewing the Cart before placing an Order. Not a
stored thing — don't model it as an entity.

**Drawer**:
The slide-out panel showing Cart contents. A purely presentational element —
belongs in the views layer only, never in store or handler vocabulary.

### Live data

In-memory, ephemeral domain nouns for the Dashboard's realtime pipeline — none
are stored or persisted (no SQLite, no `store`). They live in `internal/stream`.

**Event**:
One parsed upstream change from the Wikimedia `recentchange` feed — title, user,
wiki, bot flag, byte delta, timestamp. The unit the whole pipeline moves.
_Avoid_: edit (an Event may be a non-edit change), record, message

**Source**:
The single upstream client that holds one connection to the feed and emits
Events. Pluggable — a future poller (e.g. crypto prices) is just another Source.
_Avoid_: feed, client, producer

**Hub**:
The in-memory fan-out — one Source broadcasts each Event to many subscribers
(one per connected browser), with a bounded buffer and drop-on-full so a slow
client can't stall the rest. Keeps a small ring buffer of recent Events to seed
a new subscriber's first paint.
_Avoid_: broker, bus, pubsub

**Aggregator**:
Owns the rolling statistics derived from the Event stream (counters, per-wiki
top-N, per-second buckets). Fed reliably — never drops — and read via Snapshot.
_Avoid_: stats, metrics (those name its output, not the owner)

**Snapshot**:
An immutable, point-in-time copy of the Aggregator's stats, used to render the
tiles and charts. A value, not a live view.
_Avoid_: state, view

### Naming rule

One axis, applied consistently:

- **Handlers are named by Surface/feature**: `SiteHandler`, `ShopHandler`,
  `WorkHandler`, `SettingsHandler`, `CartHandler`.
- **Stores and types are named by Entity**: `ProductStore`, `WorkStore`,
  `CartStore`, `OrderStore`; `Product`, `Work`, `Cart`, `Order`.

So `ShopHandler` (Surface) drives a `ProductStore` (Entity). `WorkHandler` +
`WorkStore` only _look_ like an exception — that's the Work word overlap, not a
broken rule.

The **Dashboard** Surface (`DashboardHandler`) is the one Surface with no Entity
store — it reads from the in-memory `Hub` and `Aggregator` (Live data) instead.
Stream types are named by **role** (`Source`, `Hub`, `Aggregator`), not Entity,
because nothing there is stored.

Handler methods name the **domain action**, not the render target: `UpdateQty`,
not `DrawerUpdateQty`. A method split only by which DOM region it patches is a
render concern leaking into handler vocabulary.
