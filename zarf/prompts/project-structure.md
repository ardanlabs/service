# Ardan Labs Service Project Structure

This document describes the architecture and structure of the Ardan Labs Service project. It is intended to serve as a blueprint for creating new projects that follow the same design patterns and conventions.

## Overview

This is a Go-based microservice project that follows a strict layered architecture with clear dependency rules. The project is organized into four top-level packages that represent the architectural layers, plus a deployment/configuration layer.

```
project-root/
├── api/            # Binaries and entry points (services, tooling, frontends)
├── app/            # Application layer (request/response handling, routing, middleware)
├── business/       # Business logic layer (core domain logic, data access)
├── foundation/     # Foundational packages (reusable across projects)
├── zarf/           # Deployment and configuration (Docker, K8s, keys, prompts)
├── vendor/         # Vendored dependencies
├── go.mod
├── go.sum
└── makefile
```

## Dependency Rule

The dependency flow is strictly one-directional:

```
api → app → business → foundation
```

- **foundation** depends on nothing inside the project. These packages can be extracted and used in other projects.
- **business** depends only on **foundation**.
- **app** depends on **business** and **foundation**.
- **api** depends on **app**, **business**, and **foundation**.

No package may import from a layer above it. This rule is absolute.

---

## Layer 1: foundation/

The foundation layer contains small, generic, reusable packages that have no knowledge of the business domain. These packages could be moved to their own module or repository.

```
foundation/
├── docker/         # Docker container management for testing
├── keystore/       # RSA key management (loading from files, JSON)
├── logger/         # Structured logging (wraps slog patterns)
├── otel/           # OpenTelemetry tracing initialization and helpers
├── web/            # Minimal HTTP web framework
└── worker/         # Background worker/goroutine management
```

### foundation/web (The Web Framework)

This is a small, custom web framework built on top of Go's `net/http.ServeMux`. Key design decisions:

- **`web.App`**: The central type that wraps `http.ServeMux` with OpenTelemetry instrumentation, CORS support, and middleware chaining.
- **`web.HandlerFunc`**: Custom handler signature `func(ctx context.Context, r *http.Request) Encoder`. Handlers return an `Encoder` interface instead of writing to `http.ResponseWriter` directly. This forces a consistent response pattern.
- **`web.Encoder`**: Interface with `Encode() (data []byte, contentType string, err error)`. All response types (success models, errors) implement this interface.
- **`web.MidFunc`**: Middleware type `func(handler HandlerFunc) HandlerFunc`. Middleware wraps handlers using the decorator pattern.
- **`web.Respond`**: A single function handles writing all responses. It checks the returned `Encoder` for an `httpStatus` interface to determine the status code, defaults to 200, or uses 204 for nil responses.

Route registration supports two forms:
- `app.HandlerFunc(method, group, path, handler, ...middleware)` — applies both global and per-route middleware with OTEL tracing.
- `app.HandlerFuncNoMid(method, group, path, handler)` — no middleware, no tracing (used for health checks).

---

## Layer 2: business/

The business layer contains all core domain logic and data access. It has no knowledge of HTTP, request/response formats, or application-level concerns.

```
business/
├── domain/         # Domain packages (one per business entity)
│   ├── auditbus/       # Audit log business logic
│   ├── homebus/        # Home entity business logic
│   ├── productbus/     # Product entity business logic
│   ├── userbus/        # User entity business logic
│   └── vproductbus/    # View Product (read model) business logic
├── sdk/            # Shared business support packages
│   ├── dbtest/         # Database test helpers
│   ├── delegate/       # Cross-domain function calls (event-like)
│   ├── migrate/        # Database migration support (uses darwin)
│   ├── order/          # Query ordering support
│   ├── page/           # Query pagination support
│   ├── sqldb/          # Database connection and transaction support
│   └── unittest/       # Unit test helpers
└── types/          # Shared domain value types
    ├── domain/         # Domain enum type
    ├── home/           # Home address value type
    ├── money/          # Money value type
    ├── name/           # Name value type (with validation)
    ├── password/       # Password value type (with validation)
    ├── quantity/       # Quantity value type
    └── role/           # Role value type (ADMIN, USER, etc.)
```

### Domain Package Pattern (e.g., `business/domain/userbus/`)

Every domain package follows a consistent structure:

```
userbus/
├── extensions/         # Extension implementations
│   ├── useraudit/          # Audit extension (wraps business, logs actions)
│   └── userotel/           # OpenTelemetry extension (wraps business, adds tracing)
├── stores/             # Storage implementations
│   ├── usercache/          # Cache-layer store (wraps another Storer)
│   └── userdb/             # Database store (PostgreSQL implementation)
├── event.go            # Delegate event definitions
├── filter.go           # QueryFilter struct for query filtering
├── model.go            # Domain models (User, NewUser, UpdateUser)
├── order.go            # Sort ordering field constants
├── testutil.go         # Test utilities for this domain
├── userbus.go          # Core business logic (Business struct, CRUD methods)
└── userbus_test.go     # Business-level tests
```

**Key files explained:**

- **`userbus.go`** — Contains the `Business` struct and core CRUD methods (Create, Update, Delete, Query, QueryByID, etc.). Defines the `Storer` interface for data access and the `ExtBusiness` interface for the extension pattern. The `NewBusiness` constructor returns `ExtBusiness` and wraps itself in any provided extensions.

- **`model.go`** — Defines the core domain models using strong types from `business/types/`:
  - `User` — The full entity model (what gets stored/retrieved)
  - `NewUser` — Input model for creation
  - `UpdateUser` — Input model for updates (pointer fields for optional/partial updates)

- **`filter.go`** — Defines `QueryFilter` struct with pointer fields for optional filter criteria.

- **`order.go`** — Defines ordering constants (e.g., `OrderByID`, `OrderByName`) and a `DefaultOrderBy`.

- **`event.go`** — Defines delegate actions (e.g., `ActionDeleted`) and their parameter types. Used for cross-domain communication via the delegate system.

### Storer Interface

Each domain defines a `Storer` interface that declares data access behavior:

```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, usr User) error
    Update(ctx context.Context, usr User) error
    Delete(ctx context.Context, usr User) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]User, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
    QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}
```

Storage implementations live in `stores/` subdirectories:
- **`userdb/`** — PostgreSQL implementation using `sqlx`. Has its own internal model types for database row mapping and conversion functions between database and domain models.
- **`usercache/`** — Decorator store that wraps another `Storer` with in-memory caching (using `sturdyc`). Caches individual entity lookups, invalidates on mutations.

Both stores implement `NewWithTx` to support database transactions.

### Extension Pattern

The extension pattern provides a way to layer cross-cutting concerns around business logic without modifying the core. It uses the decorator/wrapper pattern.

```go
type ExtBusiness interface {
    // Same methods as Business (Create, Update, Delete, Query, etc.)
    // but as an interface for wrapping
}

type Extension func(ExtBusiness) ExtBusiness
```

During construction, extensions wrap the business in reverse order so the first extension listed is the outermost wrapper:

```go
func NewBusiness(log, delegate, storer, extensions ...Extension) ExtBusiness {
    b := ExtBusiness(&Business{...})
    for i := len(extensions) - 1; i >= 0; i-- {
        b = extensions[i](b)
    }
    return b
}
```

Built-in extensions:
- **`userotel/`** — Adds OpenTelemetry span tracking to each operation
- **`useraudit/`** — Writes audit log entries for Create/Update/Delete operations

### Delegate System

The delegate (`business/sdk/delegate/`) provides cross-domain communication without circular imports. A domain can fire an action (e.g., "user deleted") and other domains can register handlers for that action.

```go
delegate.Register(domainType, actionType, handlerFunc)
delegate.Call(ctx, data)
```

This is a synchronous, in-process call mechanism. It is not an event bus.

### Value Types (`business/types/`)

The business layer avoids using raw primitive types like `string`, `int`, and `float64` in its domain models. These are considered "weak" types because they carry no semantic meaning and cannot be validated by the compiler. A `string` can hold any value—an empty string, a SQL injection, a 10,000-character blob—and the compiler will happily accept it. This means every function that receives a `string` must defensively validate it, and there is no guarantee that validation has already happened.

Instead, the business layer uses custom "strong" types that represent real data shapes. These types encapsulate their underlying value in an unexported field and can only be constructed through a `Parse` function that enforces validation rules. Once a value of a strong type exists, it is guaranteed to be valid. The compiler enforces this—you cannot pass a raw `string` where a `name.Name` is expected.

This is why the business models look like this:

```go
// Business model — strong types, compiler-enforced validity
type User struct {
    ID         uuid.UUID
    Name       name.Name           // NOT string
    Email      mail.Address        // NOT string
    Roles      []role.Role         // NOT []string
    Department name.Null           // NOT *string
    Password   password.Password   // NOT string
}
```

And NOT like this:

```go
// Weak model — primitives, no compiler protection
type User struct {
    ID         string
    Name       string
    Email      string
    Roles      []string
    Department string
    Password   string
}
```

#### Type Construction Pattern

Every value type follows the same structural pattern:

1. **Unexported value field** — The underlying data is stored in an unexported field, preventing direct construction or mutation:
   ```go
   type Name struct {
       value string
   }
   ```

2. **`Parse` function** — The only way to create a value. Validates the input and returns an error if invalid:
   ```go
   func Parse(value string) (Name, error) {
       if !nameRegEx.MatchString(value) {
           return Name{}, fmt.Errorf("invalid name %q", value)
       }
       return Name{value}, nil
   }
   ```

3. **`MustParse` function** — Convenience constructor that panics on error. Used in tests and for known-good constant values:
   ```go
   func MustParse(value string) Name {
       name, err := Parse(value)
       if err != nil {
           panic(err)
       }
       return name
   }
   ```

4. **`String` method** — Returns the underlying value for display and output.

5. **`Equal` method** — Provides comparison support for the `go-cmp` package used in testing.

6. **`MarshalText` method** — Provides support for logging and any marshaling needs.

#### Two Flavors of Strong Types

The value types fall into two categories based on their validation approach:

**Validated types** use rules (regex, range checks) to constrain what values are allowed:

- **`name.Name`** — Must match `^[a-zA-Z][a-zA-Z0-9' -]{2,19}$` (3-20 chars, starts with letter). Also provides `name.Null` for optional name fields with `sql.NullString` support.
- **`password.Password`** — Must match `^[a-zA-Z0-9#@!-]{3,19}$`. Provides `ParseConfirm` to validate password + confirmation match.
- **`money.Money`** — Must be > 0 and ≤ 1,000,000. Wraps `float64`. Provides `Value()` to access the underlying numeric.
- **`quantity.Quantity`** — Must be > 0 and ≤ 1,000,000. Wraps `int`. Provides `Value()` to access the underlying numeric.

**Enum types** use a closed set of known values registered at package init time:

- **`role.Role`** — Known values: `Admin`, `User`. Uses a `map[string]Role` registry. Provides `ParseMany` and `ParseToString` for slice conversions.
- **`home.Home`** — Known values: `Single`, `Condo`. Same registry pattern as Role.
- **`domain.Domain`** — Known values: `User`, `Product`, `Home`. Identifies which business domain an entity belongs to (used by the audit system).

The enum pattern uses a package-level map and unexported constructor:
```go
var roles = make(map[string]Role)

var (
    Admin = newRole("ADMIN")
    User  = newRole("USER")
)

func newRole(role string) Role {
    r := Role{role}
    roles[role] = r
    return r
}
```

This prevents any code outside the package from creating new role values. The `Parse` function is the only way to convert a string to a `Role`, and it only accepts values that exist in the registry.

#### Where Conversion Happens

Strong types are only used in the business layer and below. The app layer uses primitives (`string`, `float64`) with JSON tags for API serialization. Conversion between weak and strong types happens at the app-to-business boundary in the `toBusNewUser` / `toBusUpdateUser` functions inside each app domain's `model.go`:

```go
func toBusNewUser(app NewUser) (userbus.NewUser, error) {
    var errors errs.FieldErrors

    nme, err := name.Parse(app.Name)       // string → name.Name
    if err != nil {
        errors.Add("name", err)
    }

    roles, err := role.ParseMany(app.Roles) // []string → []role.Role
    if err != nil {
        errors.Add("roles", err)
    }

    // ... all fields validated ...

    if len(errors) > 0 {
        return userbus.NewUser{}, errors.ToError()
    }

    return userbus.NewUser{Name: nme, Roles: roles, ...}, nil
}
```

This means validation happens exactly once, at the edge. Once data enters the business layer, every function can trust that the data is valid because the type system guarantees it.

---

## Layer 3: app/

The app layer is the translation layer between HTTP requests/responses and business logic. It handles input decoding, output encoding, validation, routing, authentication, and middleware.

```
app/
├── domain/         # Domain-specific app handlers (one per business domain)
│   ├── auditapp/       # Audit endpoints
│   ├── authapp/        # Authentication endpoints
│   ├── checkapp/       # Health check endpoints (liveness, readiness)
│   ├── grpcauthapp/    # gRPC auth service implementation
│   ├── homeapp/        # Home CRUD endpoints
│   ├── oauthapp/       # OAuth endpoints (using goth)
│   ├── productapp/     # Product CRUD endpoints
│   ├── rawapp/         # Raw HTTP handler examples
│   ├── tranapp/        # Transaction example endpoints
│   ├── userapp/        # User CRUD endpoints
│   └── vproductapp/    # View Product (read-only) endpoints
└── sdk/            # Shared app support packages
    ├── apitest/        # API integration test helpers
    ├── auth/           # JWT auth + OPA authorization logic
    ├── authclient/     # Auth service client (HTTP and gRPC)
    ├── debug/          # Debug/metrics HTTP mux (expvar, pprof, statsviz)
    ├── errs/           # Error types implementing web.Encoder
    ├── metrics/        # Request metrics tracking
    ├── mid/            # HTTP middleware (auth, logging, errors, otel, panics, tx)
    ├── mux/            # Service mux builder (wires routes + middleware)
    └── query/          # Query result wrapper with pagination metadata
```

### App Domain Package Pattern (e.g., `app/domain/userapp/`)

Each app domain package follows a consistent structure:

```
userapp/
├── filter.go       # Query parameter parsing → business QueryFilter conversion
├── model.go        # App-layer models (JSON-tagged) + to/from business model converters
├── order.go        # App-layer order field mapping → business order constants
├── route.go        # Route registration function
└── userapp.go      # Handler methods (create, update, delete, query, queryByID)
```

**Key files explained:**

- **`route.go`** — Exports a `Routes(app *web.App, cfg Config)` function. This function:
  1. Creates middleware instances (authen, authorize, etc.)
  2. Constructs the app handler struct
  3. Registers each route with method, version group, path, handler, and per-route middleware

- **`model.go`** — Defines the app-layer data models with JSON tags. These are separate from business models. Contains:
  - Response models implementing `web.Encoder` (with `Encode()` method)
  - Input models implementing a `Decode([]byte) error` method
  - Conversion functions: `toAppUser(bus) app`, `toBusNewUser(app) bus`
  - All validation happens in the conversion functions using `errs.FieldErrors`

- **`userapp.go`** — The handler struct (unexported `app` type) with methods matching `web.HandlerFunc` signature. Each handler:
  1. Decodes input (for mutations)
  2. Converts app model to business model
  3. Calls business layer
  4. Converts business model to app model and returns it

### Middleware (`app/sdk/mid/`)

Middleware is applied in two ways:
1. **Global middleware** — Applied to all routes via `mux.WebAPI()`: Otel → Logger → Errors → Metrics → Panics
2. **Per-route middleware** — Applied to specific routes in `route.go`: Authenticate → Authorize

Available middleware:
- `mid.Otel` — Starts trace spans for each request
- `mid.Logger` — Logs request start/completion with timing
- `mid.Errors` — Catches and logs errors from handlers
- `mid.Metrics` — Tracks request counts, errors, goroutines
- `mid.Panics` — Recovers from panics and returns error responses
- `mid.Authenticate` — Validates JWT tokens via the auth client
- `mid.Authorize` — Checks role-based access using OPA rules
- `mid.AuthorizeUser` — Loads user by ID from path, checks ownership/admin
- `mid.BeginCommitRollback` — Wraps handler in a database transaction

Context values are set and retrieved through typed functions (e.g., `mid.GetUser(ctx)`, `mid.GetClaims(ctx)`).

### Mux Configuration (`app/sdk/mux/`)

The `mux.WebAPI()` function is the central wiring point. It:
1. Creates a `web.App` with global middleware
2. Applies CORS configuration
3. Calls `routeAdder.Add(app, cfg)` to bind routes
4. Adds file servers for static content

The `mux.Config` struct carries all dependencies through the system:
```go
type Config struct {
    Build       string
    Log         *logger.Logger
    DB          *sqlx.DB
    Tracer      trace.Tracer
    BusConfig   BusConfig       // All business domain instances
    SalesConfig SalesConfig     // Sales-service specific config (auth client)
    AuthConfig  AuthConfig      // Auth-service specific config (auth instance)
}
```

### Error Handling (`app/sdk/errs/`)

Errors implement `web.Encoder` so they flow through the same response path:
- `errs.Error` — Structured error with error code and message. Implements `HTTPStatus()` to map error codes to HTTP status codes.
- `errs.FieldErrors` — Collection of field-level validation errors.
- Error codes: `OK`, `Canceled`, `Unknown`, `InvalidArgument`, `NotFound`, `AlreadyExists`, `PermissionDenied`, `Internal`, `Aborted`, `Unauthenticated`, etc. (gRPC-style codes mapped to HTTP status codes).

### Authentication & Authorization (`app/sdk/auth/`)

- **Authentication**: JWT tokens (RS256) validated via the auth service. The auth service is a separate microservice.
- **Authorization**: OPA (Open Policy Agent) rules evaluated locally. Rules are defined in `.rego` files embedded in the binary.
- **Auth Client**: The sales service communicates with the auth service via HTTP or gRPC to validate tokens and check authorization.

---

## Layer 4: api/

The api layer contains the entry points—main packages that start the services, tooling binaries, and frontend applications.

```
api/
├── frontends/      # Frontend applications
│   └── admin/          # Admin UI (web frontend)
├── services/       # Microservice binaries
│   ├── auth/           # Auth service (JWT issuance/validation, gRPC + HTTP)
│   │   ├── build/          # Route composition (which app domains to include)
│   │   └── main.go         # Service entry point
│   ├── metrics/        # Metrics sidecar service
│   │   ├── collector/
│   │   ├── publisher/
│   │   └── main.go
│   └── sales/          # Primary API service
│       ├── build/          # Route composition using build tags
│       │   ├── all.go          # Default build: all routes
│       │   ├── crud.go         # `crud` build tag: CRUD routes only
│       │   └── reporting.go    # `reporting` build tag: reporting routes only
│       ├── static/         # Embedded static files
│       ├── tests/          # API-level integration tests
│       │   ├── userapi/
│       │   ├── homeapi/
│       │   ├── productapi/
│       │   └── ...
│       └── main.go         # Service entry point
└── tooling/        # CLI tools
    ├── admin/          # Admin CLI tool (key generation, migrations, etc.)
    └── logfmt/         # Log formatting tool (makes JSON logs readable)
```

### Service Entry Point Pattern (`main.go`)

Each service `main.go` follows the same pattern:

1. **Logger setup** — Create structured logger with event hooks
2. **Configuration** — Parse config from environment variables using `ardanlabs/conf`
3. **Database** — Open connection pool using `business/sdk/sqldb`
4. **Business layer construction** — Instantiate stores, extensions, and business types:
   ```go
   storage := usercache.NewStore(log, userdb.NewStore(log, db), time.Minute)
   bus := userbus.NewBusiness(log, delegate, storage, otelExt, auditExt)
   ```
5. **Auth initialization** — Set up auth client or auth server depending on service
6. **Tracing** — Initialize OpenTelemetry with Tempo exporter
7. **Debug server** — Start debug HTTP server (expvar, pprof) on separate port
8. **API server** — Build mux, create `http.Server`, start listening
9. **Graceful shutdown** — Wait for SIGINT/SIGTERM, drain with timeout

### Build Tags

The sales service uses Go build tags to control which routes are included in a build:
- Default (no tags): All routes (`all.go`)
- `-tags crud`: Only CRUD routes (`crud.go`)
- `-tags reporting`: Only reporting routes (`reporting.go`)

Each build file exports a `Routes()` function returning a type that implements `mux.RouteAdder`.

---

## zarf/ (Deployment Configuration)

```
zarf/
├── compose/        # Docker Compose files
├── docker/         # Dockerfiles
│   ├── dockerfile.auth
│   ├── dockerfile.metrics
│   └── dockerfile.sales
├── helm/           # Helm charts for Kubernetes deployment
├── k8s/            # Kubernetes manifests
│   ├── base/           # Base kustomize manifests
│   └── dev/            # Dev overlay (Kind cluster config, kustomizations)
├── keys/           # RSA key pairs for JWT signing
└── prompts/        # AI prompts and project documentation
```

---

## Key Design Patterns

### 1. Strict Model Separation

There are three distinct model layers that never mix:
- **Business models** (`business/domain/userbus/model.go`) — Use strong value types (`name.Name`, `role.Role`). No JSON tags. These are what the business logic works with.
- **App models** (`app/domain/userapp/model.go`) — Use primitive types (`string`, `bool`). Have JSON tags. These are the API contract.
- **Store models** (`business/domain/userbus/stores/userdb/model.go`) — Use database-compatible types. Have `db` tags. These map to database rows.

Conversion functions translate between layers. Validation happens during conversion from app models to business models.

### 2. Handler Return Pattern

Handlers return `web.Encoder` instead of writing responses directly:
```go
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
    // ... business logic ...
    return toAppUser(usr)   // success: returns model implementing Encoder
    return errs.New(...)    // error: returns error implementing Encoder
}
```

The framework's `Respond` function handles writing the response. This pattern ensures consistent response formatting and enables middleware to inspect/modify responses.

### 3. Extension Pattern (Decorator)

Business logic is extended without modification using the decorator pattern. Extensions wrap `ExtBusiness` and delegate to the inner implementation while adding behavior (tracing, auditing, etc.):

```go
userOtelExt := userotel.NewExtension()
userAuditExt := useraudit.NewExtension(auditBus)
userBus := userbus.NewBusiness(log, delegate, storage, userOtelExt, userAuditExt)
```

Call chain: OtelExtension → AuditExtension → Business → Store

### 4. Store Composition

Stores are composed using the same decorator pattern:
```go
storage := usercache.NewStore(log, userdb.NewStore(log, db), time.Minute)
```

The cache store wraps the database store. Both implement `Storer`.

### 5. Transaction Support

Every `Storer` and `ExtBusiness` implements `NewWithTx()`. When a handler needs a transaction:
1. The `BeginCommitRollback` middleware starts a transaction and puts it in context
2. Inside the handler, the business layer is reconstructed with the transaction
3. The middleware commits on success or rolls back on error

### 6. Route Registration

Routes are registered via a `Routes()` function in each app domain package. Build files in the service's `build/` directory compose which domains are included. This provides flexibility to build different service configurations from the same codebase.

### 7. Configuration via Environment

All configuration is driven by environment variables using the `ardanlabs/conf` package. Config is defined as a struct with `conf` tags providing defaults. Each service uses a unique prefix (e.g., `SALES`, `AUTH`).

---

## Testing Strategy

### Unit Tests
- Located alongside business logic: `business/domain/userbus/userbus_test.go`
- Use helpers in `business/sdk/unittest/`
- Test business logic against a real database using Docker

### Integration/API Tests
- Located in: `api/services/sales/tests/`
- Organized by domain: `userapi/`, `homeapi/`, `productapi/`, etc.
- Use helpers in `app/sdk/apitest/`
- Test the full HTTP request/response cycle
- Require a running database (managed via `business/sdk/dbtest/`)

---

## Infrastructure Dependencies

- **PostgreSQL** — Primary database (accessed via `pgx` driver + `sqlx`)
- **Kubernetes/Kind** — Container orchestration (development uses Kind clusters)
- **Docker** — Container builds
- **Kustomize** — Kubernetes manifest management
- **Grafana/Prometheus/Tempo/Loki/Promtail** — Observability stack
- **OPA** — Policy-based authorization (embedded, not external)

---

## Naming Conventions

| Layer | Package Suffix | Example |
|-------|---------------|---------|
| Business domain | `bus` | `userbus`, `productbus`, `homebus` |
| App domain | `app` | `userapp`, `productapp`, `homeapp` |
| Database store | `db` | `userdb`, `productdb`, `homedb` |
| Cache store | `cache` | `usercache` |
| OTel extension | `otel` | `userotel`, `productotel` |
| Audit extension | `audit` | `useraudit` |
| API tests | `api` | `userapi`, `productapi` |

---

## Creating a New Domain (Checklist)

To add a new domain (e.g., `order`):

1. **Business layer** (`business/domain/orderbus/`):
   - `model.go` — Define `Order`, `NewOrder`, `UpdateOrder` structs using value types
   - `filter.go` — Define `QueryFilter` struct
   - `order.go` — Define ordering constants and `DefaultOrderBy`
   - `orderbus.go` — Define `Storer` interface, `ExtBusiness` interface, `Extension` type, `Business` struct with CRUD methods
   - `event.go` — Define delegate actions if needed
   - `stores/orderdb/` — PostgreSQL implementation with its own `model.go` for row mapping
   - `stores/ordercache/` — Optional cache layer
   - `extensions/orderotel/` — OpenTelemetry extension
   - `extensions/orderaudit/` — Audit extension (if needed)
   - `testutil.go` — Test data generation helpers
   - `orderbus_test.go` — Business-level tests

2. **App layer** (`app/domain/orderapp/`):
   - `model.go` — JSON-tagged app models + conversion functions (toAppOrder, toBusNewOrder)
   - `filter.go` — Query parameter parsing
   - `order.go` — App-to-business order field mapping
   - `route.go` — Route registration with middleware
   - `orderapp.go` — Handler methods

3. **Mux wiring** (`app/sdk/mux/mux.go`):
   - Add `OrderBus` field to `BusConfig` struct

4. **Service wiring** (`api/services/sales/main.go`):
   - Instantiate store, extensions, and business
   - Add to `mux.BusConfig`

5. **Build registration** (`api/services/sales/build/all.go`):
   - Add `orderapp.Routes(app, ...)` call

6. **Tests** (`api/services/sales/tests/orderapi/`):
   - API integration tests

7. **Database migration** (`business/sdk/migrate/sql/`):
   - Add migration SQL file for the new table
