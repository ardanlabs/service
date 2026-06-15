---
name: layered-architecture-types
description: Enforce primitive-at-edges / strong-types-in-Business layering and the toBus/fromBusResponse/toDB converter pattern. Use when writing, editing, or auditing Go files under app/*, business/domain/*, or .../stores/*db.
---

# Layered Architecture Types

Data crosses three layers. **Primitive types live at the edges (API JSON and DB
rows); strong types from `business/types` live only in the Business layer.** Every
crossing goes through a named conversion function — never assign across a boundary
directly, and never let a strong type appear in an API request/response struct or a
DB row struct.

```
API (app/*)            Business (business/domain/*)        Storage (.../stores/*db)
primitive types  ──►   strong types (business/types/*)    ──►   native DB types
request struct    toBus<Type>          model type        toDB<Type>     db<Type> row
response struct   ◄── fromBus<Type>Response   ◄── toBus<Type>  ◄──
                      (App layer)                  (Storage layer)
```

## When to Apply

- Writing or editing API request/response types under `app/*`
- Writing or editing Business model types under `business/domain/*/model.go`
- Writing or editing DB row structs and store code under `.../stores/*db`
- Auditing any of the above for layering conformance

## User controls the types (MUST)

The user owns every type choice this skill touches — field types, DB column
representations, converter signatures, and any wrapper types.

- Before committing changes or writing a final review summary, state the concrete
  types you intend to use at each affected boundary (request/response field, DB row
  field, converter return) and get the user's explicit confirmation. Do not finalize
  on an assumed default.
- **Suggest existing types first.** When proposing a type, prefer one that already
  exists: a strong type from one of the `business/types/*` subpackages for the
  Business layer, or a type already defined locally in the package(s) you are
  editing. Only propose a brand-new type when nothing existing fits, and say why.
- Call out anything noteworthy for them to decide: pointer vs value, `sql.Null*` vs
  wrapper, `json.RawMessage` vs a typed struct, and how NULL/empty is represented.
- If the user picks a type that diverges from the patterns here, follow their
  choice and adapt the converters to it.

## Type boundaries (MUST)

- **API request/response structs** (`app/*`): primitives only — `string`, `int`,
  `bool`, `json.RawMessage`, `time.Time`, and slices/structs of these. Never a
  `business/types/*` strong type.
- **Business model types** (`business/domain/*/model.go`): strong types for
  IDs, enums, and classifications.
- **DB row structs** (`.../stores/*db`): native/SQL types only — `string`,
  `sql.Null*`, `json.RawMessage`, `time.Time`. Never a `business/types` strong
  type, **including validated enums — store them as `string`.**

## Package imports (MUST)

- **No cross-domain Business imports.** A Business domain package
  (`business/domain/<x>bus`) must not import another Business domain package
  (`.../<y>bus`). Compose across domains in the App layer, not by reaching sideways
  in Business.
- **No cross-domain App imports.** An App domain package (`app/domain/<x>app`)
  must not import another App domain package (`.../<y>app`).
- **An App package may import several Business domain packages.** This is the
  intended place to combine domains — e.g. an app handler importing both
  `<x>bus` and `<y>bus` to assemble a response is allowed and expected.
- `business/types/*` and `foundation/*` are shared leaf packages: any layer may
  import them, and they must not import App or Business domain packages.

## Foundation types are wrapped, not used directly (MUST)

Foundation types (`foundation/*`) must not appear directly in `business/types/*`
aggregate types or in Business domain models. Define a `business/types` wrapper
type and convert via `toFoundation<T>` / `fromFoundation<T>` at the
foundation-client boundary. A wrapper may store the foundation value in an
unexported field and delegate (un)marshaling to it; the public type the aggregate
references is the wrapper, never the foundation type.

## Pointer types (avoid)

Prefer non-pointer types at every layer. Reach for a pointer only when there is no
non-pointer way to express the requirement, and say why in the proposal.

- For nullable columns, prefer value types that already model absence: `sql.Null*`
  for scalars, and for nullable `JSONB` a small `sql.Scanner`/`driver.Valuer`
  wrapper around `json.RawMessage` that scans NULL into the empty value — not
  `*json.RawMessage`, and not a `COALESCE(col, CAST('null' AS jsonb))` patch that
  silently turns NULL into the JSON scalar `null` and defeats `omitempty`.
- Before reaching for a wrapper or a pointer, consider whether the data's shape can
  change so the situation cannot arise at all — and offer that to the user as an
  alternative, since it is their call:
  - **Storage:** a `NOT NULL DEFAULT '{}'::jsonb` (or `DEFAULT 'null'::jsonb`) column
    never yields a NULL to scan, removing the need for any special handling. This is
    a schema/migration change.
  - **App-layer incoming request:** if a request field's absence is the only reason a
    pointer/wrapper is being considered, prefer making the field required (validate
    it in `toBus<Type>`) or defaulting it to a zero value, so the field stays a plain
    primitive instead of `*T`.
  - Note the trade-off either way: a NOT NULL default or a forced default collapses
    "unset" and "empty" into one value, which may or may not be acceptable for the
    field.
- Do not introduce a pointer field, parameter, or return value to signal
  optionality when a zero value, `sql.Null*`, or an explicit "is set" flag does the
  job.
- If a pointer genuinely is the only option, surface that to the user as part of
  the type confirmation below before writing it.

## Converter functions (MUST exist per type, both directions)

| Direction          | Layer pair         | Name                    | Returns |
|--------------------|--------------------|-------------------------|---------|
| primitive → strong | App → Business     | `toBus<Type>`           | `(busInput, error)` — parse + validate, accumulate `errs.FieldErrors` |
| strong → primitive | Business → App     | `fromBus<Type>Response` | response struct — convert each strong field with `.String()` etc. |
| strong → native    | Business → Storage | `toDB<Type>`            | db row struct |
| native → strong    | Storage → Business | `toBus<Type>`           | `(busType, error)` — parse via the relevant `business/types/*` `Parse*` |

Rules for these functions:

- **All parsing and validation happens in the App-layer `toBus<Type>`.** Parse each
  primitive into its strong type; on failure `fieldErrors.Add(field, err)` and
  return `errs.FieldErrors`. The request struct carries only strings — do not put
  strong types in it and do not validate strong types there.
- `fromBus<Type>Response` converts explicitly (`id.String()`, `cls.String()`).
  Never rely on a strong type's `MarshalJSON` to hide a leak — the field type in
  the response struct must already be primitive.
- Storage `toBus<Type>` validates native values back into strong types and returns
  an error (e.g. `<subpkg>.Parse<Type>(row.ID)`); `toDB<Type>` calls `.String()` to
  flatten strong types into natives.
- One converter pair per persisted type, per CRUD op that needs it (Create / Get /
  List / Update / Delete).

## Constructor naming: Parse / MustParse only (MUST)

Strong types in `business/types/*` (IDs, enums, classifications, names, sizes) MUST
expose their public constructors as `Parse<Type>` and `MustParse<Type>` — never
`New<Type>` or `MustNew<Type>`.

- `Parse<Type>(s string) (<Type>, error)` — parses + validates a primitive into the
  strong type, returning an error on invalid input. This is what converters call
  (e.g. `<subpkg>.Parse<Type>`).
- `MustParse<Type>(s string) <Type>` — panics on invalid input. Reserve for tests
  and package-level vars with known-good constants, never for request-derived data.
- **Reject `New*` / `MustNew*` for these types.** If you find them, rename to
  `Parse*` / `MustParse*` and update call sites. This applies to every public
  constructor of a strong type, including typed (non-string) variants — e.g.
  `New<Type>FromUUID(uuid.UUID)` becomes `Parse<Type>FromUUID(uuid.UUID)`. Only keep a
  `New*` form when the user gives a concrete, convincing reason; when in doubt, ask
  the user rather than introducing or retaining a `New*` constructor.

## Example — App layer (both directions)

```go
// Request carries only primitives.
type CreateWidgetRequest struct {
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"` // NOT a strong OwnerID type
}

// toBus parses + validates here.
func toBusCreateWidget(req CreateWidgetRequest) (widgetbus.CreateWidgetInput, error) {
	var fieldErrors errs.FieldErrors

	ownerID, err := types.ParseOwnerID(req.OwnerID)
	if err != nil {
		fieldErrors.Add("owner_id", err)
	}

	if len(fieldErrors) > 0 {
		return widgetbus.CreateWidgetInput{}, fieldErrors
	}

	return widgetbus.CreateWidgetInput{
		Name:    strings.TrimSpace(req.Name),
		OwnerID: ownerID,
	}, nil
}

// fromBus converts strong -> primitive explicitly.
func fromBusCreateWidgetResponse(w widgetbus.Widget) CreateWidgetResponse {
	return CreateWidgetResponse{
		ID:      w.ID.String(),
		OwnerID: w.OwnerID.String(),
		Name:    w.Name,
	}
}
```

## Example — Storage layer (both directions)

```go
type dbWidget struct {
	ID      string `db:"id"`       // native, not a strong WidgetID type
	OwnerID string `db:"owner_id"`
}

func toDBWidget(w widgetbus.Widget) dbWidget {
	return dbWidget{
		ID:      w.ID.String(),
		OwnerID: w.OwnerID.String(),
	}
}

func toBusWidget(row dbWidget) (widgetbus.Widget, error) {
	id, err := types.ParseWidgetID(row.ID)
	if err != nil {
		return widgetbus.Widget{}, fmt.Errorf("parse widget id: %w", err)
	}
	// ... parse remaining native values into strong types ...
	return widgetbus.Widget{ID: id /* ... */}, nil
}
```

Names above are illustrative — do not confuse them with real codebase symbols.

## Conformance checklist

Before finishing work on any layer, confirm:

- [ ] No `business/types` strong type appears in an API request/response struct.
- [ ] No `business/types` strong type appears in a DB row struct (validated enums
      stored as `string`).
- [ ] Each crossing has its named function: `toBus<Type>`, `fromBus<Type>Response`,
      `toDB<Type>`, and storage `toBus<Type>`.
- [ ] App-layer `toBus<Type>` parses + validates and returns `errs.FieldErrors`.
- [ ] `fromBus<Type>Response` converts every strong field explicitly (no reliance
      on `MarshalJSON`).
- [ ] Storage `toBus<Type>` returns an error and parses natives into strong types.
- [ ] Strong-type constructors are named `Parse<Type>` / `MustParse<Type>`, not
      `New*` / `MustNew*` (unless the user gave a concrete reason to keep a `New*`).
- [ ] No pointer type was introduced where a value type, `sql.Null*`, or "is set"
      flag would do.
- [ ] Proposed types reused an existing `business/types/*` or local package type
      where one fit, rather than defining a new type.
- [ ] No Business domain package imports another Business domain package, and no App
      domain package imports another App domain package.
- [ ] No `foundation/*` type appears directly in a `business/types` aggregate or a
      Business domain model; each is wrapped with `toFoundation`/`fromFoundation`
      converters.
- [ ] The user confirmed the concrete types at each affected boundary before this
      work was finalized.

## When NOT to Apply

- Internal helpers and code that never crosses a layer boundary.
- Foundation packages (`foundation/*`) and `business/types/*` themselves.
- Pure intra-package refactors that don't move data between App, Business, or
  Storage.
