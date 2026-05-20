# product stores

This directory shows how to organize a SQL store for the `productbus`
domain so it can support multiple SQL engines without copy-pasting the
entire store package per engine.

```
stores/
├── commondb/         ← engine-agnostic helpers (no SQL statements)
├── productpg/        ← PostgreSQL implementation
└── productsqlite/    ← SQLite implementation (for illustration only)
```

The running service is wired through `productpg`. `productsqlite` exists
purely as a second engine to make the seam visible; it is not exercised
by the test suite and the project does not pull in a SQLite driver.

## The pattern

A single domain `Storer` interface is satisfied by one engine package
per database. Every engine package is the same shape:

```go
type Store struct {
    log     *logger.Logger
    db      sqlx.ExtContext
    dialect dialect.Dialect   // engine-specific behavior
}
```

What lives where:

| Concern                              | Where it lives                             |
| ------------------------------------ | ------------------------------------------ |
| `Storer` interface                   | `productbus`                               |
| Database row struct + conversions    | `commondb` (`ProductDB`, `ToDBProduct`, …) |
| `WHERE` clause builder               | `commondb.ApplyFilter`                     |
| `ORDER BY` field map + clause        | `commondb.OrderByFields`, `OrderByClause`  |
| Engine-specific pagination clause    | `sqldb/dialect` (`Postgres`, `SQLite`)     |
| The SQL statements themselves        | The engine package (`productpg`, …)        |

The rule of thumb is: **share helpers, not function bodies.** Each
engine's `Create`, `Update`, `Query`, … methods read top-to-bottom as
"here is the SQL we run on this database, plus a delegation to commondb
for the parts that do not vary." When two engines need genuinely
different SQL for a method (`ON CONFLICT` vs `MERGE`, a full-text search
syntax, an extra hint), each engine just writes the method it needs —
the helpers do not get in the way.

## Adding a new engine

1. Create `stores/product<engine>/` with a `Store` that satisfies
   `productbus.Storer`. Copy `productpg` as a starting point.
2. Add a compile-time check: `var _ productbus.Storer = (*Store)(nil)`.
3. Plug the right `dialect.Dialect` into the constructor.
4. Replace the SQL strings with the dialect's preferred form. Keep the
   calls to `commondb.ApplyFilter`, `commondb.OrderByClause`, and the
   `commondb.ProductDB` mapping — those do not need to change.
5. If a single method needs deviating SQL on this engine, just write
   that method by hand. Do not push the deviation into `commondb`.

## What does not belong in commondb

- Full SQL statements. The visible shape of the queries should live in
  the engine package.
- Anything that varies between engines. If two engines need different
  behavior, the seam is either `dialect.Dialect` or the engine package
  itself — not a flag in `commondb`.
- Domain logic. `commondb` only translates between row shape and
  business model, builds portable SQL fragments, and validates ordering
  keys.

## Why keep the SQL visible per engine?

The cost of writing each engine's `Create`/`Update`/`Query` body is
small compared to the cost of debugging a clever shared template that
behaves differently on one engine. When a query misbehaves on
PostgreSQL, you open `productpg/productpg.go` and read the SQL the
database actually sees. Hiding statements behind a generator or a
builder defeats that property.
