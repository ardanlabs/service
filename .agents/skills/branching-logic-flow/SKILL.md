---
name: branching-logic-flow
description: Prefer defaulting and naked switches over if/else chains for shallow branching logic in Go. Use when writing or refactoring conditional logic with 1-3 branches.
---

# Branching Logic Flow

For shallow branching (1-3 cases), prefer **defaulting** or **naked switches** over if/else chains. This reduces
nesting, avoids duplication, and lowers the cognitive load.

## When to Apply

- Writing new conditional logic with 1-3 branches
- Refactoring nested if/else blocks
- Type assertions on `any` / `map[string]any`
- Any branching where each branch assigns to the same variable

## Rule 1: Default first, override later

Initialize the variable with its default value, then override only in the special case.

❌ Avoid:

```go
var result map[string]any
if outer, ok := input["outer"].(map[string]any); ok {
    if inner, ok := outer["inner"].(map[string]any); ok {
        result = make(map[string]any, len(inner))
        maps.Copy(result, inner)
    } else {
        result = make(map[string]any)
    }
} else {
    result = make(map[string]any)
}
```

✅ Prefer:

```go
result := make(map[string]any)
if outer, ok := input["outer"].(map[string]any); ok {
    if inner, ok := outer["inner"].(map[string]any); ok {
        result = make(map[string]any, len(inner))
        maps.Copy(result, inner)
    }
}
```

## Rule 2: Naked switch over if/else if/else ladders

❌ Avoid:

```go
if a > b {
    result = a
} else if a == b {
    result = a + b
} else {
    result = b
}
```

✅ Prefer:

```go
switch {
case a > b:
    result = a

case a == b:
    result = a + b

default:
    result = b
}
```

For simple two-way numeric comparisons, prefer the `min`/`max` built-ins (Go 1.21+):

```go
result := max(a, b)
```

## When NOT to Apply

- Default initialization is heavyweight (DB call, large allocation, external API)
- The "default" case is exceptional and should fail fast with an error
- Conditions are deeply nested (>3 levels) — extract to a separate function instead
- Defaulting could mask a bug that should be surfaced explicitly

## Related Modern Go Patterns

- Use `maps.Copy(dst, src)` instead of manual copy loops (destination must be non-nil)
- Use `cmp.Or(a, b, "default")` to pick the first non-zero value across fallbacks
- Always use the two-value form `v, ok := x.(T)` for type assertions
