# doterr
## Embeddable, zero-dependency error composition for Go

Three (3) usage patterns: 

1. **Embed** the source file _(recommended)_, 
2. **Dot-import** from `go get` _(alternative),_ or 
3. Normal import _(if you must)_.


## Status

This is **pre-alpha** and in development thus **subject to change**, although I am trying to bring to v1.0 as soon as I feel confident its architecture will not need to change. As of Novemeber 2025 I am actively working on it and using it in current projects.

If you find value in this project and want to use it, please start a discuss to let me know. If you discuver any issues with it, please open an issue or submit a pull request.


## What it is

`doterr` is an **embeddable source file for building rich, composable errors in Go**. Unlike traditional packages, you embed `doterr.go` directly into your code—no external dependencies, no version conflicts, just a single file that becomes part of your package.

It introduces two small concepts built on top of Go's `errors.Join`:

1. **Entries** — lightweight layers that attach sentinel errors and key/value metadata for a single call frame.
2. **Combined errors** — minimal composite wrappers for bundling *independent* failures _(like other multi-error packages)._

Every error value returned by any function of `doterr` returns a Go standard library `error`. The only exported type besides `error` is the `KV` interface for metadata key/value pairs. There are no exported concrete types, no reflection, and no dependency lock-in. You can use `doterr` with any Go app that uses standard Go error handling, and you can adopt it incrementally over time.

Use `doterr` to:

* Preserve **typed sentinel categories** like `ErrRepo`, `ErrConstraint`, or `ErrTemplate`.
* Attach **contextual metadata** like `"param=item_id"`, `"location=query"` or `"status=active"`.
* Compose **cause chains** naturally with `NewErr()` passing the cause as the trailing argument, with one entry per function by convention.
* Combine **independent failures** safely using `CombineErrs()`.
* Remain **100% interoperable** with the standard library and app-specific types —
  such as, for example, an RFC 9457 error, or a domain-specific type.
* Optionally **extract** custom errors with `FindErr[T](err)` when needed.

`doterr` does not try to replace Go’s error handling. `doterr` makes error handling in Go  **layered, inspectable, and ergonomic**, without ever leaving `error` and `errors.Join`.

## Design intent

Go developers already understand embedding from `//go:embed` directives and struct embedding—`doterr` extends this concept to **source file embedding**. There are three mutually exclusive ways to use `doterr` in each package:

### Option 1: Embed the source file (STRONGLY RECOMMENDED)

Copy `doterr.go` directly into your package. The functions become part of your package namespace—no import statement needed.

```go
package myapp

// No import needed - doterr.go is embedded in this package

func processUser(id int) error {
  err := doSomething()
  if err != nil {
    return NewErr(
      ErrNotFound,
      "user_id", id,
      err,  // trailing cause
    )
  }
  return nil
}
```

**Why this is recommended:**
- No external dependencies or version conflicts
- Functions like `NewErr()` and `WithErr()` appear consistently across all packages
- Enables seamless cross-package error composition
- Each package owns its copy—modify if needed

This is the intended use case, inspired by [ShadCN's approach](https://www.shadcn.io/ui/installation-guide#what-makes-shadcnui-different) for React.

### Option 2: go get + dot-import (alternative to avoid duplication)

If you want to avoid file duplication, use `go get` with dot-import:

```go
package myapp

import . "github.com/mikeschinkel/go-doterr"

func processUser(id int) error {
  err := doSomething()
  if err != nil {
    return NewErr(
      ErrNotFound,
      "user_id", id,
      err,  // trailing cause
    )
  }
  return nil
}
```

**Benefits:**
- Same unqualified function names as embedding (`NewErr()`, `WithErr()`)
- Maintains cross-package compatibility
- Single dependency source instead of multiple copies

**Trade-offs:**
- Traditional dependency (version management required)
- Cannot customize the implementation per-package

### Option 3: go get + normal import (if you must)

If dot-imports are forbidden in your project:

```go
package myapp

import "github.com/mikeschinkel/go-doterr"

func processUser(id int) error {
  err := doSomething()
  if err != nil {
    return doterr.NewErr(
      ErrNotFound,
      "user_id", id,
      err,  // trailing cause
    )
  }
  return nil
}
```

**Note:** This works but loses the primary benefit. When different packages use different qualifiers (`doterr.NewErr()` vs `pkgname.NewErr()`), cross-package error composition becomes inconsistent. Use this only if organizational constraints require it.

### Why function names matter

With Option 1 or 2, every package uses `NewErr()` and `WithErr()`—creating predictable, interoperable error handling across your entire codebase. With Option 3, each package might use different qualifiers, breaking that consistency.

### Linters & imports

If your linter or style guide restricts dot-imports:
1. **Best:** Use Option 1 (embedding)—no import means no linter issue
2. **Alternative:** Configure linter to allow dot-import for `go-doterr`
3. **Last resort:** Use Option 3, understanding the compatibility trade-offs

## Core principles

| Principle                                                                                                                                                                                         | Description                                                                                                                                                                    |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Pure standard Go library**                                                                                                                                                                      | Only depends on Go's built-in `errors` package.                                                                                                                                |
| **Embeddable by design**                                                                                                                                                                          | Three usage patterns: embed the source file (recommended), dot-import via `go get` (alternative), or normal import (if you must).                                             |
| **Fully composable**                                                                                                                                                                              | Always returns the native `error` type.                                                                                                                                        |
| **Explicit layering**                                                                                                                                                                             | By convention, each `func` builds one entry and passes its cause as the trailing argument.                                                                                     |
| **Sentinel-driven**                                                                                                                                                                               | By convention, every layer identifies itself using sentinel errors.                                                                                                            |
| **Structured metadata**                                                                                                                                                                           | Provide key/value pairs (`"key", value`) alongside sentinels at each level.                                                                                                    |
| **One way, not many**| `doterr` exposes a single canonical `func` for each behavior. No alternative function names or multiple ways to do the same thing. |
| **Consistent naming**                 | Embedding or dot-import ensures `NewErr()` and `WithErr()` appear consistently across all packages for seamless interoperability.                                             |

> **Note:** `NewErr()` accepts an optional trailing cause parameter for wrapping errors. `WithErr()` enriches an existing error (which may itself contain causes) by adding metadata, but does not accept a trailing cause parameter—it only adds to what already exists.

### Why no exported custom error types?

Less experienced Go developers often think _"I will just define my own error type with fields."_
At first blush that seems harmless, until developers try to use it with existing code and especially when trying to write reusable packages that export those types.

Here is what happens:

1. **Developers must type-assert errors** — To access their custom properties and methods, or to pass to a `func` or assign to a `var` or method properties typed for the custom error, developers are forced to type assert, and _then_ write _more_ error handling code to deal with errors that don't type assert as expected. If you always uses errors of type `error`, this problem effectively disappears.  

2. **Developers cannot mix types cleanly** — If multiple packages define their own custom error types you end up in situations were you can use one or the other, _but not both_. If you always use `error` instead, you can always unify errors with one mechanism: `errors.Join()`.

3. **Custom errors are often not composable** — Standard helpers (`errors.Is`, `errors.As`, `errors.Join`) only work if you expose or wrap correctly. Many libraries forget to implement `Unwrap()` or do not do it properly — causing lost context.


Go’s own `os.PathError` is a classic example. It wraps valuable info (`Op`, `Path`, `Err`)
but forces you to `errors.As(err, &os.PathError{})` instead of using consistent metadata patterns.

By contrast, `doterr` keeps **every layer** a plain `error` — enriched with sentinel and key-value metadata:

* No forced type assertions.
* No impossible combination of multiple custom errors.
* Full compatibility with the standard error ecosystem.

`doterr` deliberately avoids defining or exporting any new concrete error type, and generally, you should to. `doterr` lets you **standardize semantics** without breaking composability.

## Sentinel errors

A **sentinel error** is a package-level constant identifying a specific class or layer of failure.
They make your error tree 1.) **type-safe**, 2.) **searchable**, and 3.) **idiomatic** via `errors.Is()` and `errors.As()` instead of brittle string matching.

```go
var (
  ErrDriver   = errors.New("driver")   // lowest level
  ErrRepo     = errors.New("repo")     // middle layer
  ErrService  = errors.New("service")  // top layer
  ErrTemplate = errors.New("template") // domain-specific category
)
```

**Built-in validation sentinels:**

`doterr` includes several built-in sentinel errors for validation and safety:

```go
var (
  ErrMissingSentinel     = errors.New("missing required sentinel error")
  ErrTrailingKey         = errors.New("trailing key without value")
  ErrMisplacedError      = errors.New("error in wrong position")
  ErrInvalidArgumentType = errors.New("invalid argument type")
  ErrOddKeyValueCount    = errors.New("odd number of key-value arguments")
  ErrCrossPackageError   = errors.New("error from different doterr package")
)
```

The first five are used for `NewErr()` argument validation. `ErrCrossPackageError` is automatically added when `WithErr()` detects you're mixing errors from different `doterr` copies (see [Cross-package error detection](#-cross-package-error-detection) below).

Include one or two sentinels when constructing an entry — **always first:**

```go
cause := someOperation()
return NewErr(ErrDriver,
  "sql", query,
  "param", id,
  cause,  // trailing cause
)
```

Callers can then reason about context:

```go
if errors.Is(err, ErrDriver) {
  log.Println("driver-level failure")
}
```

Why provide **two (2)** sentinels? It can often be useful to provide both a general purpose error — e.g. `ErrNotFound` — and a more-specific error — e.g. `ErrNoWidgetMatchedSearchTerm` — when characterizing errors.

## Layered composition example

Each function layer defines **its own sentinel** and passes the error from the inner function as the trailing cause.
This produces a clean, typed, layered tree that mirrors your call stack.

```go
// innermost: driver layer
var db *sql.DB
func readDriver() (Result, error) {
  query := "SELECT * FROM users WHERE id=?"
  id := 42
  result, err := db.Query(query, id)
  if err != nil {
    return nil, NewErr(ErrDriver,
      "sql", query,
      "param", id,
      err,  // trailing cause
    )
  }
  return result, nil
}

// middle: repository layer
func readRepo() error {
  _, err := readDriver()
  if err != nil {
    return NewErr(ErrRepo,
      "table", "users",
      err,  // trailing cause
    )
  }
  return nil
}

// outer: service layer
func readService() error {
  err := readRepo()
  if err != nil {
    return NewErr(ErrService,
      "op", "GetUser",
      err,  // trailing cause
    )
  }
  return nil
}
```

**Inspection:**

```go
err := readService()

fmt.Println(err)

if errors.Is(err, ErrDriver)  { fmt.Println("driver error") }
if errors.Is(err, ErrRepo)    { fmt.Println("repository layer failed") }
if errors.Is(err, ErrService) { fmt.Println("service layer failed") }
```

Each function contributes one entry and one sentinel — composable, testable, and human-readable.

### Enrichment in the same function

If you want to add fields or tags within the same function, use `WithErr`:

```go
err = WithErr(err, "attempt", retryCount)
```

If the rightmost entry is already a `doterr` entry, it merges into it.
If not, it creates a new one and joins it automatically. `WithErr()` never accepts a cause — it's for enrichment only.

## API summary

| Function                                                                                                                  | Purpose                                                                      |
|---------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| `NewErr(parts ...any)`                                                                                                    | Create a new entry with sentinels first, metadata, and optional trailing cause. |
| `WithErr(err error, parts ...any)`                                                                                        | Enrich existing error by merging into rightmost entry (enrichment only).    |
| `CombineErrs(errs []error)`                                                                                               | Join multiple independent errors (skips `nil`s, preserves order).           |
| `ErrMeta(err error) []KV`                                                                                                 | Return metadata key/value pairs from first entry (unwraps one level).       |
| `Errors(err error) []error`                                                                                               | Return sentinel/typed errors from first entry (unwraps one level).          |
| `FindErr[T](err error) (T, bool)`                                                                                         | Extract first typed error of type T using `errors.As`.                      |

### Implementation notes

* Each entry is a minimal struct implementing `Error()` and `Unwrap() []error`.
* `WithErr()` scans one join level right-to-left for an entry to enrich.
* No recursion deeper than one join level.
* No reflection or third-party dependencies.
* Every exported function returns the **built-in `error` type**.

## Cross-package error detection

Since `doterr` is designed to be **embedded** into independent packages (like `pathvars`, `common`, `dbqvars`), each embedded copy has its own unique package identity. To prevent subtle bugs from accidentally mixing errors between different `doterr` instances, the package includes **automatic cross-package detection**.

**How it works:**

1. Each embedded `doterr` instance generates a unique `uniqueId` at init time
2. Every `entry` created by that instance stores this `id`
3. When `WithErr()` receives an error to enrich or join:
   - It checks if the error is an `entry` from a different `doterr` instance
   - If the IDs don't match, it wraps the error with `ErrCrossPackageError`
   - The wrapped error includes diagnostic metadata: `package_id` and `expected_id`

**Why this matters:**

```go
// In package pathvars (has embedded doterr.go):
func ValidateParam() error {
  return NewErr(ErrValidation, "param", "id")
}

// In package common (has its own embedded doterr.go):
func ProcessWithValidation() error {
  err := pathvars.ValidateParam()
  // Trying to enrich an error from a different doterr instance
  return WithErr(err, "extra", "data")  // ⚠️ Cross-package detected!
}
```

Without this check, mixing entries from different packages could cause:
- Lost metadata when enrichment fails silently
- Type assertion failures in internal code
- Confusing error chains that are hard to debug

**The detection wraps automatically:**

```go
if errors.Is(err, doterr.ErrCrossPackageError) {
    // Developer is warned they're mixing errors across package boundaries
    // Metadata tells them which packages are involved
}
```

**Best practice:** Each independent package should create and manage its own `doterr` errors. When passing errors between packages, use them as **trailing causes** in `NewErr()` rather than trying to enrich them with `WithErr()`.

### Why no type introspection?

The API intentionally **does not** provide functions like `IsEntry(err) bool` or `IsCombined(err) bool` to detect internal types.

**Rationale:**

1. **Violates encapsulation** — The whole point of "everything returns `error`" is that consumers shouldn't care about concrete types. Exposing type checks undermines this principle.

2. **Use stdlib interfaces instead** — To detect multi-unwrappers (entries, combined, or stdlib joins), use the standard pattern:
   ```go
   if u, ok := err.(interface{ Unwrap() []error }); ok {
       // This is a multi-unwrapper (entry, combined, or errors.Join)
       for _, child := range u.Unwrap() {
           // traverse
       }
   }
   ```

3. **Existing API handles common cases** — `ErrMeta()` and `Errors()` already unwrap one level automatically, covering most needs. For deeper traversal, use `Unwrap()` directly.

4. **Slippery slope** — Today it's "is this an entry?", tomorrow it's "give me the raw entry", then we've lost all abstraction benefits.

5. **Unclear use case** — Most operations (`errors.Is()`, `errors.As()`, `ErrMeta()`, `Errors()`) work uniformly regardless of concrete type. If you need to distinguish types, you're probably doing something the API should handle for you.

**Historical note:** Prior art (like Go's stdlib hiding `joinError`) shows that keeping error structure opaque encourages robust, interface-based code. Type introspection leads to fragile coupling to implementation details.

## License

MIT — © 2025 Mike Schinkel [mike@newclarity.net](mailto:mike@newclarity.net)

