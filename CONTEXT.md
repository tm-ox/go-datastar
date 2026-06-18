# go-datastar

A portfolio-and-shop site: a public **Work** portfolio and a **Shop** of
**Products**, with a server-driven UI (Datastar over SSE). This file fixes the
domain language so modules and conversations use one word per concept.

## Entities

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
_Avoid_: purchase, transaction, checkout (checkout is the *act*, not the thing)

**Site content**:
Editorial copy for the Home and About pages, loaded from embedded YAML rather
than the database.
_Avoid_: CMS, pages-table

## Surfaces

The places a visitor goes — sections and pages. A Surface is not an Entity:
the **Shop** Surface lists **Product** entities; the **Work** Surface lists
**Work** entities (same word, different layer).

**Shop**, **Work**, **Settings**, **About**, **Home** — the top-level sections.

**Checkout**:
The Surface (and the act) of reviewing the Cart before placing an Order. Not a
stored thing — don't model it as an entity.

**Drawer**:
The slide-out panel showing Cart contents. A purely presentational element —
belongs in the views layer only, never in store or handler vocabulary.

## Naming rule

One axis, applied consistently:

- **Handlers are named by Surface/feature**: `SiteHandler`, `ShopHandler`,
  `WorkHandler`, `SettingsHandler`, `CartHandler`.
- **Stores and types are named by Entity**: `ProductStore`, `WorkStore`,
  `CartStore`, `OrderStore`; `Product`, `Work`, `Cart`, `Order`.

So `ShopHandler` (Surface) drives a `ProductStore` (Entity). `WorkHandler` +
`WorkStore` only *look* like an exception — that's the Work word overlap, not a
broken rule.

Handler methods name the **domain action**, not the render target: `UpdateQty`,
not `DrawerUpdateQty`. A method split only by which DOM region it patches is a
render concern leaking into handler vocabulary.
