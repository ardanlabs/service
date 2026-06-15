---
name: review-pr
description: "Service Diffguard: read-only PR review lenses for correctness, error visibility, comment truthfulness, test coverage, type/contract boundaries, and simplification across every language in this repo (Go, Vue/JS/SCSS, Rego, protobuf, SQL, Helm/YAML, shell). Use for PR review, pre-commit review, unstaged-diff review, or a targeted review of one of those aspects."
---

# Service Diffguard

Service Diffguard reviews a diff with six focused **lenses**. Each lens runs as
a read-only Amp Task subagent, inspects the requested change, and returns
**findings** graded on one shared scale. Lenses advise only — they never edit,
stage, or commit code, never disable tests, and never bypass a failing check to
make it pass.

Default scope is the **unstaged `git diff`** unless the caller names specific
files, a branch, a commit range, or an existing PR.

## Gate scale

Every finding carries a Gate. The run's overall Gate is the highest any lens
returns.

| Gate   | Name    | Meaning                                                                                  | Action                                               |
|--------|---------|------------------------------------------------------------------------------------------|------------------------------------------------------|
| **G4** | Stop    | Likely correctness, security, data-integrity, compile, boundary, or reliability failure. | Block merge until fixed.                             |
| **G3** | Repair  | Likely user-visible bug, operational blind spot, real test gap, or misleading contract.  | Fix before the PR is ready.                          |
| **G2** | Tighten | Locally safe but leaves avoidable risk, ambiguity, weak design, or brittleness.          | Address if reasonably scoped; otherwise acknowledge. |
| **G1** | Polish  | Low-risk clarity, naming, comment, or small test refinement.                             | Optional; batch, don't spam.                         |
| **G0** | Clear   | No issue, or a positive observation.                                                     | None.                                                |

Every **G2 or higher** finding must include: `file:line`, **Proof** (evidence from
the diff), why it matters, a **Patch direction** (smallest safe fix), and how to
verify. Weak or speculative observations go under **Open checks**, not as a
finding. Cap **G1** items at five per lens unless exhaustive polish is requested.

## Lens catalog

| Lens                   | Focus                                  | Reference prompt                  |
|------------------------|----------------------------------------|-----------------------------------|
| **Service Steward**    | general code correctness & conventions | `reference/service-steward.md`    |
| **Error Tripwire**     | failure-path visibility                | `reference/error-tripwire.md`     |
| **Doc Drift Check**    | comment truthfulness                   | `reference/doc-drift-check.md`    |
| **Harness Map**        | behavior-to-test mapping               | `reference/harness-map.md`        |
| **Boundary Keeper**    | type & contract design across layers   | `reference/boundary-keeper.md`    |
| **Straight-Line Pass** | behavior-preserving simplification     | `reference/straight-line-pass.md` |

## Languages in this repo

Diffguard reviews a changed file in whatever language it is written in. Each lens
applies its concept (correctness, failure visibility, comment truth, test
coverage, contract design, simplification) through the idioms of that language —
Go is the primary codebase, not the only one.

| Area            | Where                                                      | Lens focus for this language                                                                                                                |
|-----------------|------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| Go services     | `app/**`, `business/**`, `foundation/**`, `api/**` (`.go`) | Full Go review; the layered type rules and Go skills apply here.                                                                            |
| Admin frontend  | `api/frontends/admin/src` (`.vue`, `.js`, `.scss`)         | Vue SFC structure, component `props`/`emits` contracts, reactivity, async/`fetch` error handling, accessibility, leaked watchers/listeners. |
| Auth policies   | `app/sdk/auth/rego` (`.rego`)                              | OPA/Rego rules: default-deny posture, `input` shape assumptions, unintended allows.                                                         |
| gRPC contracts  | `app/domain/grpcauthapp` (`.proto`)                        | Field-number stability, backward/forward compatibility, breaking changes.                                                                   |
| Database        | `business/sdk/migrate/sql` (`.sql`)                        | Migration reversibility/idempotency, indexing, locking, destructive changes, seed correctness.                                              |
| Deploy / config | `zarf/**` (`.yaml`, `.tpl`), `*.conf`                      | Helm/k8s manifests: resource limits, secrets handling, probes, value templating.                                                            |
| Scripts         | `*.sh`                                                     | Shell safety: `set -euo pipefail`, quoting, exit-code handling.                                                                             |

## Repo-local authority

All lenses treat **AGENTS.md** (and any nested AGENTS.md) as authoritative.

Apply these repo skills **only to `.go` files**, where relevant:

- `use-modern-go` — Go implementation style.
- `layered-architecture-types` — App ↔ Business ↔ Storage: primitives at the
  edges, strong types in Business, named converters only.
- `business-layer-extensions` — the `ExtBusiness` / `Extension` decorator seam.
- `branching-logic-flow` — defaulting and shallow control flow.

`validate-pr-title` applies whenever the PR title is in scope, regardless of
language.

## Diff intake

1. Resolve scope from the caller.
2. With no explicit scope, inspect unstaged changes:
   - `git status --short`
   - `git diff --name-only`
   - `git diff`
   - For a branch/PR: `git diff --name-only <base>...HEAD`, or `gh pr diff`.
3. Build a short change inventory: which languages/areas changed (see *Languages
   in this repo*); production vs test files; App/Business/Storage files; comments
   or doc strings; error/failure handling; new or modified types, interfaces,
   converters, component contracts, proto messages, or SQL schema; PR title if a
   PR is in scope.

## Lens routing

Route by what changed unless the caller asks for "full", "all", or "PR-ready".

- **Always** (any production code changed, any language): Service Steward.
- **Production code or tests changed, or coverage asked:** Harness Map.
- **Diff touches any failure path** — Go `err`/`errors.Is/As`/wrapping/`defer`/
  `recover`/context, JS `try/catch`/promises/`await`/`fetch`, SQL transactions,
  Rego allow/deny logic, or shell exit codes; plus logging and fallback/default
  behavior: Error Tripwire.
- **Any contract or shape changed** — Go structs/interfaces/constructors/
  converters/Business strong types/DB rows/stores, Vue `props`/`emits`, proto
  messages, or SQL schema: Boundary Keeper.
- **Comments, doc strings (Go doc, JSDoc), examples, README snippets, or
  TODO/FIXME changed:** Doc Drift Check.
- **Polish pass:** Straight-Line Pass — run in a second wave after G4/G3 findings are
  resolved, or in the first wave only when simplification is explicitly requested.

## Execution

Run lenses as Task subagents. Because they are read-only, **parallel is safe** and
preferred for a broad PR sweep; prefer **sequential** when scope is ambiguous, the
diff is very large, or only one concern was requested.

Default is a two-wave plan:

- **Wave 1 (risk):** Service Steward, Harness Map, plus Error Tripwire, Boundary
  Keeper, and Doc Drift Check where applicable.
- **Wave 2 (polish, optional):** Straight-Line Pass, after high-risk findings are
  addressed or when requested.

Launch each lens with a prompt of this shape:

```text
Use `.agents/skills/review-pr/reference/<lens>.md` as your review instructions.

Scope: <exact scope>
Diff source: <unstaged git diff | PR diff | branch range | file list>
Languages in scope: <e.g. Go; Vue/JS; Rego; proto; SQL; Helm/YAML; shell>
Repo context: polyglot monorepo (Go primary). Follow AGENTS.md; apply Go skills to .go files only.
Mode: advisory/read-only — do not modify files.

Return a Gate report only, using the G0-G4 scale.
```

## Aggregation

Merge lens reports into one Diffguard report. Deduplicate overlapping findings by
failure mode and location; on disagreement keep the higher Gate.

```markdown
# Service Diffguard Report

Scope: <scope reviewed>
Overall Gate: G[0-4] — <Clear/Polish/Tighten/Repair/Stop>

## Run matrix
| Lens               | Result        | Notes |
|--------------------|---------------|-------|
| Service Steward    | Gx            | ...   |
| Error Tripwire     | Gx / skipped  | ...   |
| Doc Drift Check    | Gx / skipped  | ...   |
| Harness Map        | Gx / skipped  | ...   |
| Boundary Keeper    | Gx / skipped  | ...   |
| Straight-Line Pass | Gx / deferred | ...   |

## Fix queue
### G4 Stop
- [lens] <title> — `file.go:line`
  - Proof:
  - Patch direction:
  - Verify:
### G3 Repair
### G2 Tighten
### G1 Polish

## Clear signals
- <What looks solid and why.>

## Open checks
- <Anything not reviewed, unavailable, or needing human confirmation.>

## Suggested verification
- Go: `make test-only`, or `go test ./path/...` for a single package.
- Frontend: the project's JS test/lint/build scripts in `api/frontends/admin`.
- SQL/Rego/proto: the relevant migration, policy, or codegen check for that area.
- Re-run affected lenses after fixes.
```
