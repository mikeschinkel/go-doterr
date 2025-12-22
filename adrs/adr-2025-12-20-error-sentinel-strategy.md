# ADR: Error Sentinel Strategy and Naming

## Status
Proposed - 2025-12-20

**Note:** Sentinel naming conventions are still under discussion (see Open Questions below).

## Context

### The Error Sentinel Problem

Error sentinels are a critical part of Go's error handling, but they present unique challenges:

1. **Sentinel proliferation** - Easy to create too many sentinels, making APIs cluttered
2. **Naming is hard** - Hard to know what good sentinel names are upfront
3. **Part of API contract** - Once exported, sentinels lock you into names
4. **By the time you understand the domain, the API is locked** - Similar to other symbols, but more acute for errors

### Why This Needs a Separate ADR

While error sentinels follow the general stability levels defined in `go-dt/adrs/adr-2025-12-20-stability-levels.md`, they have unique concerns:

- **Naming conventions** specific to error sentinels
- **MsgErr workflow** for rapid development
- **When to define sentinels** vs using ad-hoc errors
- **Storage location** (dt package vs local package)
- **Migration strategies** specific to errors

### Reference: General Stability Levels

This ADR builds on the general stability contract system defined in the [go-stability](https://github.com/mikeschinkel/go-stability) repository.

**For general stability information, see go-stability:**
- [Stability Levels](https://github.com/mikeschinkel/go-stability#stability-levels) - Definitions of stable, provisional, experimental, obsolete, deprecated, internal
- [Time-Based Guarantees](https://github.com/mikeschinkel/go-stability#time-based-guarantees) - 18 months for stable, 6 months for provisional
- [Contract Comment Format](https://github.com/mikeschinkel/go-stability#contract-comment-format) - General contract annotation syntax
- [Changelog Requirements](https://github.com/mikeschinkel/go-stability#changelog-requirements) - All stability levels require changelog entries
- [Migration Strategies](https://github.com/mikeschinkel/go-stability#migration-strategies) - General approach to deprecations

**This ADR provides error-specific guidance:**
- Error sentinel naming conventions and patterns
- MsgErr workflow for rapid development
- When to define sentinels vs using ad-hoc errors
- Storage location (dt package vs local package)
- Error-specific migration examples

## Decision

### Sentinel Naming Conventions

**⚠️ TODO:** This section is NOT finalized. We need more discussion before finality. Likely we need more guidelines and specificity about use-cases where adjectives are used with nouns.

**TODO:** Also, updating this section means we'll need to update all sections whose assumptions are based on this section's assertions.

#### Pattern: Err + State/Condition

**Format:** `Err<State>`

**Examples:**
```go
// ✅ GOOD - State nouns
ErrNotFound
ErrClosed
ErrTimeout
ErrInvalid
ErrConflict
ErrUnauthorized
ErrExhausted

// ❌ BAD - Verb phrases
ErrFailedToOpen
ErrCouldNotConnect
ErrUnableToProcess

// ❌ BAD - Inconsistent tense
ErrOpeningFailed  // past tense
ErrFailsValidation // present tense

// ❌ BAD - Inconsistent structure
ErrNoUserFound  // "No" prefix
ErrUserNotFound // "Not" suffix
```

#### Guidelines

1. **Prefix:** Always `Err`
2. **Body:** Noun or adjective describing the error state
3. **Avoid:**
   - Verbs ("Failed", "Unable", "Could not")
   - Past tense ("-ed" endings)
   - "No"/"Not" inconsistency
4. **Specificity:** Specific enough to be useful, general enough to be stable

#### Domain-Specific Examples

**Validation Errors:**
```go
// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
var (
    ErrInvalidInput  = errors.New("invalid input")  // not ErrInputInvalid or ErrValidationFailed
    ErrMissingField  = errors.New("missing field")  // not ErrFieldMissing or ErrNoField
    ErrOutOfRange    = errors.New("out of range")   // not ErrRangeExceeded
    ErrInvalidFormat = errors.New("invalid format") // not ErrFormatInvalid
)
```

**Resource Errors:**
```go
// Contract:
// - Stability: stable
// - Since: v1.0.0 (2023-01-10)
var (
    ErrNotFound      = errors.New("not found")         // general - stable
    ErrAlreadyExists = errors.New("already exists")    // not ErrExists or ErrDuplicate
    ErrConflict      = errors.New("conflict")          // not ErrVersionConflict (unless needed)
)

// Contract:
// - Stability: provisional
// - Since: v1.5.0 (2024-03-01)
// - Note: May consolidate into ErrNotFound
var ErrUserNotFound = errors.New("user not found")  // specific - if needed
```

**State Errors:**
```go
// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
var (
    ErrClosed      = errors.New("closed")       // connection/file closed
    ErrTimeout     = errors.New("timeout")      // operation timed out
    ErrExhausted   = errors.New("exhausted")    // resource exhausted
    ErrUnavailable = errors.New("unavailable")  // resource temporarily unavailable
)
```

**Authorization Errors:**
```go
// Contract:
// - Stability: stable
// - Since: v1.3.0 (2023-09-20)
var (
    ErrUnauthorized = errors.New("unauthorized") // not ErrNotAuthorized
    ErrForbidden    = errors.New("forbidden")    // not ErrAccessDenied
    ErrExpired      = errors.New("expired")      // token/session expired
)
```

#### Naming Decision Tree

```
1. Is it a missing resource?
   → ErrNotFound (or ErrXNotFound if specific needed)

2. Is it a validation issue?
   → ErrInvalid<X> or ErrMissing<X>

3. Is it a state problem?
   → Err<State> (Closed, Timeout, Exhausted)

4. Is it a conflict?
   → ErrConflict (or Err<X>Conflict if needed)

5. Is it authorization?
   → ErrUnauthorized or ErrForbidden

6. When in doubt?
   → Use MsgErr during development
   → Analyze usage patterns
   → Promote to sentinel with proper name
```

#### Consistency Within Packages

**Prefer consistent patterns:**

```go
// ✅ GOOD - Consistent pattern
// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrRoleNotFound  = errors.New("role not found")
    ErrGroupNotFound = errors.New("group not found")
)

// ❌ BAD - Inconsistent
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrNoRole        = errors.New("no role")
    ErrMissingGroup  = errors.New("missing group")
)
```

### Storage Location

#### dt Package (Domain Types)

**Location:** `github.com/mikeschinkel/go-dt`

**What goes here:**
- General-purpose sentinels used across multiple packages
- Domain-agnostic errors (ErrNotFound, ErrInvalid, etc.)
- Can be any stability level

**Examples:**
```go
// dt/errors.go

// Contract:
// - Stability: stable
// - Since: v1.0.0 (2023-01-10)
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
    ErrTimeout  = errors.New("timeout")
)
```

#### Local Package

**Location:** Each package's error definitions

**What goes here:**
- Package-specific sentinels
- Domain-specific errors
- Can be any stability level

**Examples:**
```go
// gitutils/errors.go

// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
var (
    ErrRepoNotFound  = errors.New("repository not found")
    ErrInvalidGitRef = errors.New("invalid git reference")
)
```

**Key insight:** Stability is orthogonal to location - both dt and local packages can have stable sentinels.

### MsgErr Workflow: From Ad-Hoc to Stable

**⚠️ TODO:** This workflow does not acknowledge that we should (maybe?) defer defining sentinels unless and until we have a use-case where we actually need them. (Although the concern that others might need them even if we don't implies that we need to consider providing sentinels even before we need them. We still need to develop a moderating strategy for this tension.)

MsgErr (implemented in go-doterr) allows creating ad-hoc errors during rapid development, then promoting them to sentinels when patterns emerge.

#### Phase 1: Rapid Development

Use MsgErr during active development:

```go
func loadUser(id string) (User, error) {
    if id == "" {
        err := doterr.NewErr(doterr.MsgErr("user not found"), "user_id", id)
        return User{}, err
    }
    // ...
}
```

**MsgErr creates an error wrapper that:**
- Can be used like a sentinel
- Doesn't require pre-defining sentinel variables
- Signals "this might become a sentinel later"

#### Phase 2: Observation

- Code is working
- Error messages stabilize
- Usage patterns emerge
- Team understands domain better

#### Phase 3: Analysis

Run tooling to analyze error usage:

```bash
# Scan for MsgErr usage (hypothetical squire command)
squire errors scan

# Output:
MsgErr instances found:
  "user not found": 15 occurrences across 5 files
  "config invalid": 8 occurrences across 3 files
  "processing failed": 2 occurrences in 1 file
```

#### Phase 4: Naming

Apply naming conventions:
- "user not found" → `ErrUserNotFound` (or `ErrNotFound` if general enough)
- "config invalid" → `ErrInvalidConfig` (or `ErrInvalid`)
- "processing failed" → (too vague, needs refinement or stays as MsgErr)

#### Phase 5: Promotion

Create sentinel with appropriate stability:

```go
// errors.go

// Contract:
// - Stability: experimental
// - Since: v1.5.0 (2024-03-01)
// - Note: Promoted from MsgErr, may be refined
var ErrUserNotFound = errors.New("user not found")
```

#### Phase 6: Stabilization

After 1-2 versions of usage, promote to stable:

```go
// Contract:
// - Stability: stable
// - Since: v1.5.0 (2024-03-01)
// - Note: Stabilized in v1.7.0 after confirming naming
var ErrUserNotFound = errors.New("user not found")
```

#### Phase 7: Refinement (if needed)

If name needs to change:

```go
// New sentinel (more general)
// Contract:
// - Stability: stable
// - Since: v1.7.0 (2024-07-15)
var ErrNotFound = errors.New("not found")

// Old sentinel (compatibility)
// Contract:
// - Stability: deprecated
// - Since: v1.5.0 (2024-03-01)
// - RemoveAfter: 2026-01-15  // 18 months from deprecation
// - RemoveVersion: v2.0.0
// - UseInstead: ErrNotFound
var ErrUserNotFound = ErrNotFound
```

### When to Define Sentinels

Three approaches, each with tradeoffs:

#### Approach 1: Defer Until Needed (Internal Use)

**When:** Package is mainly for internal use

**Strategy:**
- Use `MsgErr` during development
- Only create sentinels when `errors.Is()` checks become necessary
- Promote based on actual usage patterns

**Pros:**
- Avoids premature sentinel proliferation
- Names emerge from real usage
- Less API surface to maintain

**Cons:**
- Consumers can't use `errors.Is()` on ad-hoc errors
- May delay stabilization of error handling

#### Approach 2: Provide Early (Public API)

**When:** Package has external consumers who need error inspection

**Strategy:**
- Define sentinels early (mark as `experimental` or `evolving`)
- Provide consumers with `errors.Is()` capability from the start
- Stabilize names after observing usage

**Pros:**
- Consumers can write robust error handling immediately
- Clear error handling contract from the start
- Better developer experience

**Cons:**
- May create sentinels that aren't actually used
- Names might need refinement later (requires deprecation)

#### Approach 3: Hybrid (Analyze and Decide)

**When:** Package has both internal and public APIs

**Strategy:**
- Use `MsgErr` for internal errors
- Define sentinels for errors that cross API boundaries
- Use tooling to track MsgErr usage
- Promote high-frequency MsgErr instances to sentinels

**Pros:**
- Balances flexibility with consumer needs
- Data-driven decision making
- Minimizes unnecessary sentinels

**Cons:**
- Requires tooling to track usage
- More complex decision process

### Migration Strategies for Error Sentinels

#### Renaming a Stable Sentinel (18 Month Process)

**Step 1: Add new sentinel (v1.7.0, Jan 1, 2024)**
```go
// Contract:
// - Stability: stable
// - Since: v1.7.0 (2024-01-01)
var ErrNotFound = errors.New("not found")
```

**Step 2: Deprecate old sentinel (same release)**
```go
// Contract:
// - Stability: deprecated
// - Since: v1.2.0 (2023-06-15)
// - RemoveAfter: 2025-07-01  // 18 months from deprecation
// - RemoveVersion: v2.0.0
// - UseInstead: ErrNotFound
var ErrRecordNotFound = ErrNotFound  // Alias keeps it working
```

**Step 3: Update CHANGELOG**
```markdown
## v1.7.0 (2024-01-01)

### Added
- `ErrNotFound` - General not found error

### Deprecated
- `ErrRecordNotFound` - Use `ErrNotFound` instead
- **Removal date:** No earlier than July 1, 2025 (18 month notice)
- **Target version:** v2.0.0

### Migration Guide
Replace `errors.Is(err, ErrRecordNotFound)` with `errors.Is(err, ErrNotFound)`
```

**Step 4: Wait for notice period**
Cannot remove before Jul 1, 2025 (even if you want to release v2.0.0 sooner).

**Step 5: Remove after notice period**
Can remove in v2.0.0 or later, but not before Jul 1, 2025.

#### Renaming a Provisional Sentinel (6 Month Process)

**Step 1: Announce change (v1.5.0, Jan 1, 2024)**
```go
// Contract:
// - Stability: provisional
// - Since: v1.3.0 (2023-10-15)
// - Note: Will rename to ErrInvalidUser in v1.6.0 (no earlier than Jul 1, 2024)
var ErrUserValidation = errors.New("user validation failed")
```

**Step 2: Make change after notice (v1.6.0, Jul 1, 2024 or later)**
```go
// New name
// Contract:
// - Stability: provisional
// - Since: v1.6.0 (2024-07-15)
var ErrInvalidUser = errors.New("invalid user")

// Keep alias for transition
// Contract:
// - Stability: deprecated
// - Since: v1.3.0 (2023-10-15)
// - RemoveAfter: 2025-01-15  // 6 more months
// - RemoveVersion: v2.0.0
// - UseInstead: ErrInvalidUser
var ErrUserValidation = ErrInvalidUser
```

#### Consolidating Multiple Sentinels

**When:** You have too many specific sentinels that could be one general sentinel

**Example:**
```go
// Old specific sentinels (being consolidated)
// Contract:
// - Stability: deprecated
// - Since: v1.0.0 (2023-01-10)
// - RemoveAfter: 2025-07-10
// - RemoveVersion: v2.0.0
// - UseInstead: ErrNotFound
var (
    ErrUserNotFound  = ErrNotFound  // Alias
    ErrRoleNotFound  = ErrNotFound  // Alias
    ErrGroupNotFound = ErrNotFound  // Alias
)

// New general sentinel
// Contract:
// - Stability: stable
// - Since: v1.7.0 (2024-01-01)
var ErrNotFound = errors.New("not found")
```

**Migration guidance:**
```go
// Old code
if errors.Is(err, ErrUserNotFound) {
    // handle
}

// New code (use structured metadata for specifics)
if errors.Is(err, ErrNotFound) {
    // Extract resource type from metadata if needed
    resourceType := doterr.ErrValue[string](err, "resource")
    // handle
}
```

### Mixing MsgErr and Sentinels

MsgErr and sentinels work together:

```go
var ErrInvalid = errors.New("invalid")

// Use sentinel for category, MsgErr for specifics
err := doterr.NewErr(
    ErrInvalid,
    doterr.MsgErr("email format is incorrect"),
    "email", userEmail,
)

// Consumer can check category
if errors.Is(err, ErrInvalid) {
    // Handle validation error
}
```

This allows:
- **Sentinels** for stable, checkable categories
- **MsgErr** for specific, evolving error details
- **Best of both:** stability + specificity

## Consequences

### Positive

- **Gradual error API evolution** without version churn
- **MsgErr enables rapid development** without premature sentinel definition
- **Clear migration paths** for renaming/consolidating sentinels
- **Consumers can plan** error handling changes around dates
- **Tooling can track** MsgErr usage patterns to guide promotion

### Negative

- **Must honor time-based commitments** for error sentinels
- **Learning curve** for MsgErr workflow
- **Temptation to over-stabilize** (marking everything stable too early)
- **Need tooling** to track MsgErr usage effectively
- **Sentinel proliferation** still possible if not careful

## Examples

### Complete Example: All Stability Levels for Errors

```go
package mypackage

import "errors"

// Stable errors - guaranteed for 18 months minimum
// =================================================

// ErrNotFound indicates the requested resource does not exist.
//
// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
var ErrNotFound = errors.New("not found")

// ErrInvalid indicates the input is invalid.
//
// Contract:
// - Stability: stable
// - Since: v1.0.0 (2023-01-10)
var ErrInvalid = errors.New("invalid input")

// Provisional errors - may change with 6 months notice
// ======================================================

// ErrUserValidation validates user-specific data.
//
// Contract:
// - Stability: provisional
// - Since: v1.5.0 (2024-03-01)
// - Note: May rename to ErrInvalidUser for consistency
var ErrUserValidation = errors.New("user validation failed")

// ErrConfigFormat validates configuration file format.
//
// Contract:
// - Stability: provisional
// - Since: v1.6.0 (2024-05-15)
// - Note: Considering merge with ErrInvalid in future
var ErrConfigFormat = errors.New("config format error")

// Experimental errors - may change or disappear anytime
// ======================================================

// ErrProcessingTemp is used during active refactoring.
//
// Contract:
// - Stability: experimental
// - Since: v1.7.0 (2024-11-20)
// - Note: May be replaced with more specific errors
var ErrProcessingTemp = errors.New("processing failed")

// Obsolete errors - don't use, but won't remove
// ==============================================

// ErrRecordMissing is inconsistent with current naming.
//
// Contract:
// - Stability: obsolete
// - Since: v1.0.0 (2023-01-10)
// - UseInstead: ErrNotFound
// - Note: Use ErrNotFound for new code, kept for compatibility
var ErrRecordMissing = ErrNotFound

// Deprecated errors - scheduled for removal
// ==========================================

// ErrRecordNotFound is deprecated.
//
// Contract:
// - Stability: deprecated
// - Since: v1.0.0 (2023-01-10)
// - RemoveAfter: 2025-07-10
// - RemoveVersion: v2.0.0
// - UseInstead: ErrNotFound
// - Note: 18 month notice (stable item deprecated)
var ErrRecordNotFound = ErrNotFound

// ErrBadInput is deprecated.
//
// Contract:
// - Stability: deprecated
// - Since: v1.1.0 (2023-03-15)
// - RemoveAfter: 2025-09-15
// - RemoveVersion: v2.0.0
// - UseInstead: ErrInvalid
// - Note: 18 month notice (stable item deprecated)
var ErrBadInput = ErrInvalid

// Internal errors - not for external use
// =======================================

// ErrInternalState is used for internal state tracking.
//
// Contract:
// - Stability: internal
// - Note: Exported for testing only, do not use
var ErrInternalState = errors.New("internal state error")
```

### Real-World Pattern: MsgErr to Sentinel Promotion

```go
// Initial development - use MsgErr
func processFile(path string) error {
    // Rapid development phase
    err := doterr.MsgErr("file processing failed")
    err = doterr.WithErr(err, "path", path)
    err = doterr.WithErr(err, "operation", "validate")
    return err
}

// After usage analysis - promote to sentinel
// Contract:
// - Stability: experimental
// - Since: v1.8.0 (2025-01-15)
// - Note: Promoted from MsgErr, observing usage
var ErrProcessingFailed = errors.New("processing failed")

func processFile(path string) error {
    err := doterr.NewErr(ErrProcessingFailed, "path", path)
    err = doterr.WithErr(err, "operation", "validate")
    return err
}

// After stabilization - mark as stable
// Contract:
// - Stability: stable
// - Since: v1.8.0 (2025-01-15)
// - Note: Stabilized in v2.0.0
var ErrProcessingFailed = errors.New("processing failed")
```

## Best Practices

### Start with MsgErr for New Errors

Don't rush to create sentinels. Use MsgErr and let patterns emerge:

```go
// Early development
err := doterr.MsgErr("unexpected response format")

// After patterns emerge
var ErrInvalidFormat = errors.New("invalid format")
```

### Prefer General Sentinels Over Specific

```go
// ✅ GOOD - General, stable
var ErrNotFound = errors.New("not found")

// ❌ RISKY - Too specific, might need to change
var ErrUserWithIDNotFoundInDatabase = errors.New("user with id not found in database")
```

Use structured metadata for specifics:
```go
err := doterr.NewErr(ErrNotFound, "resource", "user", "user_id", id)
```

### Use Hierarchical Error Wrapping

```go
// Broad category (stable)
var ErrValidation = errors.New("validation error")

// Specific case (can evolve)
err := fmt.Errorf("invalid email: %w", ErrValidation)
```

Consumers checking `ErrValidation` are unaffected by changes to specific error messages.

### Document Error Contract Explicitly

```go
// Contract:
// - Stability: stable
// - Since: v1.2.0 (2023-06-15)
// - Note: Check with errors.Is() for programmatic handling
var ErrNotFound = errors.New("not found")
```

## Open Questions

### 1. Sentinel Naming Finalization

**TODO (line 287 from original doc):** Needs more discussion - adjectives + nouns

**Current state:** Pattern defined, but edge cases need clarification:
- When to use adjectives vs nouns?
- Specificity guidelines (ErrNotFound vs ErrUserNotFound)?
- Domain-specific naming patterns?

**Action:** Continue discussion, gather real-world examples, refine guidelines

### 2. When to Define Sentinels

**TODO (line 438 from original doc):** Defer until needed vs provide for consumers

**Current state:** Three approaches documented (Defer, Provide Early, Hybrid) but no clear decision

**Tension:**
- Internal preference: Defer with MsgErr (avoid proliferation)
- Consumer need: Early sentinels (enable errors.Is() checking)

**Action:** Gather more information and experience before deciding. May be context-dependent (internal vs public packages).

## Related ADRs and Documentation

- **General Stability Levels:** See [go-stability](https://github.com/mikeschinkel/go-stability) for stability level definitions and time-based guarantees
- **go-doterr Package:** Implements MsgErr and structured error handling
- **Breaking Change Detection:** Future tooling will validate error sentinel stability contracts
- **Changelog Generation:** Future tooling will support automated changelog generation from Contract: annotations

## References

- **go-stability:** https://github.com/mikeschinkel/go-stability
- **Go Error Handling:** https://go.dev/blog/error-handling-and-go
- **Error Wrapping:** https://go.dev/blog/go1.13-errors
