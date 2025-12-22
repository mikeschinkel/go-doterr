# Stability Contracts and DotErr: Small-Team Compatibility Without a Platform Org

> **IMPORTANT**: This was written by ChatGPT and have not yet been reviewed by a human for accuracy nor emphasis although a cursory scan indicates there are many aspects that need to be changed before this is "published."

Go's compatibility story works well for the language and standard library, but most teams are not the Go team. They do not have decades-long horizons, large review staffs, or a massive user base to absorb churn. Small teams still need to evolve quickly without surprising downstream users, and that requires explicit, lightweight compatibility discipline.

This article argues for symbol-level stability contracts and structured error semantics, using `dt` and `doterr` as concrete examples.

## Why "stability by location" does not scale

The Go ecosystem often signals stability by location: `std` vs `x/` vs `internal`. That works when you can split experimental APIs into new packages. But Go has a hard constraint: you cannot add methods to types in another package. For small repos, "make a new package" becomes impractical when you need to evolve a shared core type without breaking imports or duplicating types. Symbol-level stability is the missing tool for that world.

## OTel is prior art, not a replacement

OpenTelemetry provides a maturity ladder and a minimum deprecation floor, and it supports tooling through machine-readable models and code generation. That is useful prior art. It is not a replacement for symbol-level contracts on Go identifiers.

We borrow vocabulary and the minimum deprecation floor from OTel, but we explicitly target Go API identifiers and contract enforcement.

## A compressed stability ladder

OTel uses a richer ladder. We intentionally compress it:

- **Experimental**: maximum volatility (OTel Development + Alpha)
- **Provisional**: likely safe for production but not yet committed (OTel Beta + RC)
- **Stable**: committed (OTel Stable)
- **Deprecated**: removal scheduled (OTel Deprecated + minimum floor)
- **Obsolete**: kept for compatibility
- **Internal**: exported for technical reasons, not a public contract

Provisional is not "beta" and not "release candidate." It means "this is probably OK for production, but if you require strict compatibility, wait."

## Wrapping is narrative, not semantics

Go's `%w` solved a real problem, but it encouraged a culture where new semantics are encoded in strings. Wrapping preserves upstream classifiers. It does not create new stable classifiers at the point of failure.

If it is all "wrapped errors, all the way down," who watches the watchmen?

Wrapping is fine for narrative context, but it is a bad foundation for stable semantics vs. providing explicit classifiers (sentinels) and metadata at the point of failure.

### The join illustration

This example shows why multiple classifiers matter:

```go
err := causeSentinelError()
switch {
case errors.Is(err, ErrDamnItsEven):
    println("It was even!")
case errors.Is(err, ErrDamnItsOdd):
    println("It was odd!")
}

err = causeWrapperError()
switch {
case strings.Contains(err.Error(), "its even"):
    println("It was even!")
case strings.Contains(err.Error(), "its odd"):
    println("It was odd!")
}
```

`errors.Join(ErrEven, ErrSentinel)` allows multiple independent classifiers. A single `%w` wrap preserves only one classifier and pushes new semantics into the error string. This is an illustration, not a recommendation to "wrap + join" as a general workaround. DotErr exists to make multiple classifiers and metadata ergonomic without string parsing.

## Typed errors: obvious downstream pain

Typed errors are not wrong, but they do not compose well at scale:

1. **Type-switch matrix**: Package A returns `*AError`, package B returns `*BError`. Downstream code grows a brittle type-switch matrix and misses cases unless it lives inside those types every day.
2. **Interface evolution is breaking**: Adding a method to a shared error interface breaks every implementation downstream.
3. **Multiple taxonomies**: Each dependency invents its own hierarchy, so retryable vs. fatal vs. temporary becomes inconsistent across a codebase.

Sentinel identity plus documented metadata keys provides a stable, cross-package classifier without forcing downstream code to understand every concrete type.

## DotErr in one sentence

DotErr provides stable classifiers (sentinels), optional metadata, and an optional metadata trace, while remaining 100% compatible with Go's `error` type and `errors.Is/As/Join`.

### Metadata trace (not a stack trace)

DotErr records a metadata trace: key/value context captured per layer. It is optional and can be disabled or kept shallow. The purpose is causality context, not necessarily a full call stack.

## Adoption path: start small

DotErr is opt-in per package and does not replace Go errors. You can adopt it incrementally:

1. **Sentinels only**: stable identity + `errors.Is`
2. **MsgErr only**: structured message without strong matching contracts
3. **Full contract**: sentinel + metadata + optional metadata trace

Linters can teach at the edges: "exported errors must use a sentinel" is a small, learnable rule. The rest can remain unchanged.

## Deprecation floor and stronger guarantees

For Deprecated symbols, the minimum removal floor is **2 minor releases or 6 months, whichever is later**. We meet the OTel minimum and may exceed it; each symbol's contract is authoritative and enforceable.

## Why this is small-team friendly

This approach does not require a platform org or a giant process. It is lightweight documentation plus simple tooling:

- contract blocks in doc comments,
- a JSON index for browsing and CI validation,
- `apidiff` to detect breaks,
- and a small policy layer that applies to **Stable** vs **Provisional** symbols.

Compatibility discipline is respect for stability as well as downstream developer's and user's time.

## Open items

Comparison: go-semver-audit vs go-nextver

This will be added as an appendix when findings are available.
