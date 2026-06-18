---
status: accepted
---

# Store interfaces are not retained; concrete structs + in-memory SQLite tests

Each store package (`product`, `work`, `cart`, `order`) previously defined an
interface with a single SQLite implementation. Those interfaces are removed:
constructors return concrete `*SQLite…Store` structs and consumers hold them
directly. The interfaces bought no leverage — one adapter each — and their
implementations encode SQLite-specific semantics (e.g. the `AddItem` clamp
`DO UPDATE SET quantity = MIN(quantity + 1, ?)`, the `GetItemDetails` JOIN to
`products`) that a hand-written fake could only mirror, not run.

Stores are tested by exercising the real SQL against a throwaway `:memory:`
SQLite database (`modernc.org/sqlite`, pure Go, fast), migrated per test. This
is the closest-to-production test surface and removes any fake-vs-real drift.

## Considered options

- **Add a second adapter (in-memory fake) to justify the seam.** Rejected: the
  fake would have to re-implement SQLite invariants by hand; tests would pass
  against behaviour that differs from production.
- **Keep single-implementation interfaces as-is.** Rejected: premature
  abstraction — idiomatic Go accepts interfaces at the consumer when a real
  need exists, not speculatively at the producer.

## Consequences

- A genuine future need for substitution (e.g. a Postgres adapter, or a fast
  fake for a hot handler path) reintroduces a *small* interface declared in the
  **consumer** package (`handler`), listing only the methods that consumer
  uses — not a producer-side interface mirroring the whole struct.
- Architecture reviews should not re-suggest "add adapters to justify the store
  seams." That trade-off is settled here.
