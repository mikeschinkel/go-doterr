# ADR: Error Contracts, Sentinels, and Compatibility

## Status
Proposed - 2025-12-21

## North Star

Provide explicit, tool-enforceable stability and error contracts so small Go teams can evolve APIs without surprising downstream users.

## Context

Errors are part of the API when callers branch on them. In Go, error identity and semantics are often implied rather than documented, which makes compatibility hard to reason about and harder to enforce.

Go's error wrapping solved one problem (keeping a cause chain) but created another: a culture where new semantics are encoded in strings instead of stable classifiers. In a wrap-only culture there are effectively no structured semantics, only strings and a cause chain.

We need a small, explicit contract for exported error sentinels and metadata keys that:

- tells downstream code what is stable,
- makes deprecations actionable,
- and is enforceable by tooling.

OpenTelemetry provides prior art for stability levels and deprecation floors, but its maturity system targets component- and model-level contracts, not Go identifier contracts. This ADR adapts vocabulary and minimum deprecation floors while retaining symbol-level enforcement goals.

## Reference: General Stability Contracts

This ADR builds on the general stability contract system defined in the [go-stability](https://github.com/mikeschinkel/go-stability) repository.

**For general stability information, see go-stability:**
- [Stability Levels](https://github.com/mikeschinkel/go-stability#stability-levels) - Definitions and guarantees for all stability levels
- [Time-Based Guarantees](https://github.com/mikeschinkel/go-stability#time-based-guarantees) - Time-based (not version-based) deprecation commitments
- [Contract Comment Format](https://github.com/mikeschinkel/go-stability#contract-comment-format) - Machine-readable contract annotation syntax

**This ADR provides DotErr-specific guidance:**
- Error contracts as API
- Wrapping vs semantics philosophy
- DotErr adoption modes
- Metadata traces and structured context
- Error-specific contract format with Fields

## Decision

### 1) Error Contracts are API

Exported sentinels and documented metadata keys are part of the public API. Their meaning and stability must be explicit.

For each exported sentinel or condition-identifier, we attach a machine-readable contract block in its doc comment. The contract defines:

- `Stability`: Stable | Provisional | Experimental | Deprecated | Obsolete | Internal
- `Since`: version or date
- `Meaning`: one-line semantic description
- `Fields`: documented metadata keys and their semantics

Only metadata keys listed in `Fields` are considered part of the compatibility contract.

### 2) Wrapping is Narrative, Not Semantics

Wrapping preserves upstream identity, but it does not create new stable classifiers. If all errors are "wrapped all the way down," new semantics end up in strings and are not matchable by downstream code.

> Wrapping is fine for narrative context, but it is a bad foundation for stable semantics vs. providing explicit classifiers (sentinels) and metadata at the point of failure.

DotErr provides explicit classifiers plus metadata so downstream code can branch on stable semantics instead of parsing strings.

### 3) Adoption Modes are Incremental

DotErr is opt-in per package and does not replace Go errors. Minimal adoption is encouraged:

1. **Sentinels only**: stable identity + `errors.Is`
2. **MsgErr only**: structured message without strong matching contracts
3. **Full contract**: sentinel + metadata + optional metadata trace

### 4) Metadata Traces are Optional

DotErr records a **metadata trace** (not a stack trace): structured key/value context from each layer. This is optional and can be disabled or kept shallow. The intent is to capture causality context, not necessarily stack details.

## Compatibility and Deprecation Policy

**See [go-stability](https://github.com/mikeschinkel/go-stability) for:**
- Detailed stability level definitions and guarantees
- Time-based deprecation commitments (18 months for stable, 6 months for provisional)
- General deprecation and migration policies

**DotErr-specific notes:**
- All exported error sentinels and documented metadata keys follow the go-stability contract system
- Metadata keys listed in contract `Fields` are part of the compatibility guarantee
- Unlisted metadata keys may change without notice

## Contract Format (Errors)

**See [go-stability Contract Format](https://github.com/mikeschinkel/go-stability#contract-comment-format) for general contract syntax.**

DotErr extends the general contract format with error-specific fields:

**Error-specific contract block:**

```go
// ErrNotFound indicates the requested resource does not exist.
//
// Contract:
// - Stability: Stable
// - Since: v0.6.0
// - Meaning: Resource does not exist or is not visible to the caller.
// - Fields:
//   - resource: logical resource name (string)
//   - id: resource identifier (string|int)
var ErrNotFound = errors.New("not found")
```

**Additional fields for deprecated items** (see [go-stability examples](https://github.com/mikeschinkel/go-stability#deprecated) for details):

```go
// - Stability: deprecated
// - Since: v0.6.0 (2024-01-15)
// - RemoveAfter: 2026-01-15
// - RemoveVersion: v1.0.0
// - UseInstead: pkg.ErrNotFound
```

## Tooling Intent (Planned)

Tooling will extract `Contract:` blocks into a generated JSON index used for documentation, CI validation, and future CLI tooling. CI should detect:

- breaking changes to Stable contracts,
- missing deprecation metadata,
- and violations of removal floors.

## Consequences

### Positive

- Downstream code can rely on documented, stable error semantics.
- Deprecations become actionable and enforceable.
- Small teams get compatibility discipline without a platform org.

### Negative

- Requires additional documentation effort for exported sentinels.
- Enforcement tooling must be built to realize full benefits.
