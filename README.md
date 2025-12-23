# doterr
## An Embeddable "Drop-In" Error Package for Go

**`doterr`** is a single-file error handling package designed to be **embedded directly into your Go packages** as a _**"drop-in"**_ source file. Unlike traditional dependencies,  you copy `doterr.go` into your package—no external dependencies, no version conflicts, just a single file that becomes part of your package namespace.

This approach is inspired by [ShadCN](https://ui.shadcn.com/docs) for React components: just like you might copy a utility function into your codebase, `doterr.go` becomes part of your package.

## Status

This is **beta** _(aka experimental)_: in active use during development of a large Go project and feels stable and close to v1.0. As of December 2025 I am actively working on it and using it in current projects.

If you find value in this project and want to use it, please start a discussion to let me know. If you discover any issues with it, please open an issue or submit a pull request.

## What it is

`doterr` introduces **two small structural concepts** built on top of Go's `errors.Join`, plus **convenience helpers** for common patterns:

### Core Concepts

1. **Entries** — lightweight layers that attach sentinel errors and key/value metadata for a single call frame.
2. **Combined errors** — minimal composite wrappers for bundling *independent* failures _(like other multi-error packages)._

### Convenience Helpers

- **MsgErr()** — Create ad-hoc error messages during rapid development
- **Typed KV functions** — Type-safe metadata (StringKV, IntKV, BoolKV, etc.)
- **AppendKV()** — Accumulate metadata with lazy evaluation

Every error value returned by any function of `doterr` returns a Go standard library `error`. The only exported type besides `error` is the `ErrKV` interface for metadata key/value pairs. There are no exported concrete types, no reflection, and no dependency lock-in. You can use `doterr` with any Go app that uses standard Go error handling, and you can adopt it incrementally over time.

Use `doterr` to:

* Preserve **typed sentinel categories** like `ErrRepo`, `ErrConstraint`, or `ErrTemplate`.
* Attach **contextual metadata** like `"param=item_id"`, `"location=query"` or `"status=active"`.
* Compose **cause chains** naturally with `NewErr()` passing the cause as the trailing argument, with one entry per function by convention.
* Combine **independent failures** safely using `CombineErrs()`.
* Remain **100% interoperable** with the standard library and app-specific types —
  such as, for example, an RFC 9457 error, or a domain-specific type.
* Optionally **extract** custom errors with `FindErr[T](err)` when needed.

`doterr` does not try to replace Go's error handling. `doterr` makes error handling in Go  **layered, inspectable, and ergonomic**, without ever leaving `error` and `errors.Join`.

## How to Use doterr

### Embed the Source File

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
      err,  // err is the trailing cause
    )
  }
  return nil
}
```

**Why this is the intended use case:**
- No external dependencies or version conflicts
- Functions like `NewErr()` and `WithErr()` appear consistently across all packages
- Enables seamless cross-package error composition
- Each package owns its copy—modify if needed

### Synchronizing Multiple Copies

When using doterr in multiple packages, use `make sync` to keep all embedded copies synchronized:

```bash
cd /path/to/go-doterr
make sync DIR=~/Projects
```

This tool:
- Finds all `doterr.go` files under the specified directory
- Updates each with the correct package name
- Ensures bug fixes and updates propagate to all copies
- Prompts for confirmation before making changes

**Example:**
```bash
cd /Users/you/Projects/go-pkgs/go-doterr
make sync DIR=~/Projects
```

See [cmd/sync-doterr/README.md](cmd/sync-doterr/README.md) for full documentation.

> **Note:** This is a temporary solution while waiting for generalized drop-in management in [Squire](https://github.com/mikeschinkel/squire). Future tooling will support managing any embeddable source files, not just doterr.go.

## Core principles

| Principle                                                                                                                                                                                         | Description                                                                                                                                                                    |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Pure standard Go library**                                                                                                                                                                      | Only depends on Go's built-in `errors` package.                                                                                                                                |
| **Embeddable by design**                                                                                                                                                                          | Copy doterr.go into your package—it becomes part of your package namespace.                                              |
| **Fully composable**                                                                                                                                                                              | Always returns the native `error` type. No runtime type assertions required. No need to learn multiple custom error type APIs. Natural composition without type compatibility issues.                                                                                                                        |
| **Explicit layering**                                                                                                                                                                             | By convention, each `func` builds one entry and passes its cause as the trailing argument.                                                                                     |
| **Sentinel-driven**                                                                                                                                                                               | By convention, every layer identifies itself using sentinel errors.                                                                                            |
| **Structured metadata**                                                                                                                                                                           | Provide key/value pairs (`"key", value`) alongside sentinels at each level.                                                                                                    |
| **One way, not many**| `doterr` exposes a single canonical `func` for each behavior. No alternative function names or multiple ways to do the same thing. |
| **Consistent naming**                 | Embedding ensures `NewErr()` and `WithErr()` appear consistently across all packages for seamless interoperability.                                             |

> **Note:** `NewErr()` accepts an optional trailing cause parameter for wrapping errors. `WithErr()` enriches an existing error (which may itself contain causes) by adding metadata, but does not accept a trailing cause parameter—it only adds to what already exists.

### Why no exported custom error types?

Less experienced Go developers often think _"I will just define my own error type with fields."_
At first blush that seems harmless, until developers try to use it with existing code and especially when trying to write reusable packages that export those types.

Here is what happens:

1. **Developers must type-assert errors** — To access their custom properties and methods, or to pass to a `func` or assign to a `var` or method properties typed for the custom error, developers are forced to type assert, and _then_ write _more_ error handling code to deal with errors that don't type assert as expected. If you always uses errors of type `error`, this problem effectively disappears.

2. **Developers cannot mix types cleanly** — If multiple packages define their own custom error types you end up in situations were you can use one or the other, _but not both_. If you always use `error` instead, you can always unify errors with one mechanism: `errors.Join()`.

3. **Custom errors are often not composable** — Standard helpers (`errors.Is`, `errors.As`, `errors.Join`) only work if you expose or wrap correctly. Many libraries forget to implement `Unwrap()` or do not do it properly — causing lost context.

Go's own `os.PathError` is a classic example. It wraps valuable info (`Op`, `Path`, `Err`)
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
  ErrDriver   = errors.New("driver error")      // lowest level
  ErrRepo     = errors.New("repository error")  // middle layer
  ErrService  = errors.New("service error")     // top layer
  ErrTemplate = errors.New("template error")    // domain-specific category
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

The first five are used for `NewErr()` argument validation. `ErrCrossPackageError` is automatically added when `WithErr()` detects you're mixing errors from different `doterr` copies (see [Cross-package error detection](#cross-package-error-detection) below).

Include one or two sentinels when constructing an entry — **always first:**

```go
err := someOperation()
return NewErr(ErrDriver,
  "sql", query,
  "param", id,
  err,  // err is the trailing cause
)
```

Callers can then reason about context:

```go
if errors.Is(err, ErrDriver) {
  log.Println("driver-level failure")
}
```

Why provide **two (2)** sentinels? It can often be useful to provide both a general purpose error — e.g. `ErrNotFound` — and a more-specific error — e.g. `ErrWidgetNotFound` — when characterizing errors.

See [adrs/adr-2025-12-20-error-sentinel-strategy.md](adrs/adr-2025-12-20-error-sentinel-strategy.md) for sentinel naming conventions and guidance.

## MsgErr() for Rapid Development

`MsgErr()` creates ad-hoc error messages without requiring a sentinel error. This is a convenience for rapid development—use sentinels for production code when patterns stabilize.

**Two forms:**

```go
// Create error with message
err := MsgErr("config validation failed")

// Wrap existing error (preserves error chain for errors.Is)
err := MsgErr(existingErr)
```

**Usage with doterr functions:**

```go
// In NewErr() as a sentinel replacement
err := NewErr(MsgErr("file not found"), "path", configPath)

// With metadata
err := MsgErr("config invalid")
err = WithErr(err, "path", "/etc/app.conf")

// As trailing cause
err := NewErr(ErrProcessing, "step", "validate", MsgErr("format error"))
```

**Migration path:**

1. **Rapid development** — Use MsgErr during active coding
2. **Observation** — Let error patterns emerge
3. **Analysis** — Use tooling to find common MsgErr messages
4. **Promotion** — Convert frequent patterns to sentinels
5. **Stabilization** — Mark sentinels as stable after validation

Future tooling can detect MsgErr usage and suggest/generate appropriate sentinel errors.

**Mixing MsgErr and sentinels:**

```go
// Use sentinel for category, MsgErr for specifics
err := NewErr(
    ErrInvalid,
    MsgErr("email format is incorrect"),
    "email", userEmail,
)

// Consumer can check category
if errors.Is(err, ErrInvalid) {
    // Handle validation error
}
```

## Parameter types for NewErr()

`NewErr(parts ...any)` accepts a variadic parameter, but **not all types are valid**:

| Valid | Type | Position | Example |
|-------|------|----------|---------|
| ✅ | Sentinel error | First positions | `ErrRepo`, `ErrNotFound` |
| ✅ | Metadata key (string) | Followed by value | `"table"` |
| ✅ | Metadata value (any) | After key | `"users"`, `42`, `obj` |
| ✅ | Cause error | Last position only | `err` |
| ❌ | Descriptive string | Any | `"failed to connect"` |
| ❌ | fmt.Sprintf result | Any | `fmt.Sprintf("error: %s", msg)` |

**Anti-pattern to avoid:**

```go
// ❌ WRONG - descriptive strings are not sentinels
return NewErr(ErrCmd, "failed to resolve config directory", err)
return NewErr(ErrCmd, fmt.Sprintf("no branch found (tried: %s)", list))

// ✅ CORRECT - use sentinels + metadata
return NewErr(ErrCmd, ErrResolvingConfigDir, err)
return NewErr(ErrCmd, ErrNoBranchFound, "tried", list)

// ✅ ALTERNATIVE - use MsgErr during development
return NewErr(ErrCmd, MsgErr("failed to resolve config directory"), err)
```

Descriptive strings bypass `errors.Is()` and cannot be matched programmatically. If you need to describe a failure condition, create a sentinel error for it or use MsgErr() during development.

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
      err,  // err is the trailing cause (original database error)
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
      err,  // err is the trailing cause (from driver layer)
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
      err,  // err is the trailing cause (from repository layer)
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

### Adding function-level context at the end label

When using the Clear Path style with `goto end`, add function-level context once at the `end:` label instead of repeating it at every error creation point:

**Repeated metadata; not recommended:**
```go
func (c *Cmd) processFile() (err error) {
    var filepath dt.Filepath

    filepath = dt.FilepathJoin(c.Dir, "config.json")
    data, err := filepath.ReadFile()
    if err != nil {
        // Repetitive - filepath added here
        err = NewErr(ErrFileRead, "filepath", filepath, err)
        goto end
    }

    err = c.parseData(data)
    if err != nil {
        // And again here
        err = NewErr(ErrParsing, "filepath", filepath, err)
        goto end
    }

end:
    return err
}
```

**Preferred pattern:**
```go
func (c *Cmd) processFile() (err error) {
    var filepath dt.Filepath
    var data []byte

    filepath = dt.FilepathJoin(c.Dir, "config.json")
    data, err = filepath.ReadFile()
    if err != nil {
        err = NewErr(ErrFileRead, err)
        goto end
    }

    err = c.parseData(data)
    if err != nil {
        err = NewErr(ErrParsing, err)
        goto end
    }

end:
    if err != nil {
        err = WithErr(err, "filepath", filepath)
    }
    return err
}
```

This keeps error creation clean (just sentinels + cause) and adds shared function-level context in one place.

### Accumulating errors with AppendErr

`AppendErr` is a convenience function for accumulating errors in a slice, skipping nil errors automatically:

```go
var errs []error
err := doSomething()
errs = AppendErr(errs, err)  // Only appends if err != nil
```

This is equivalent to but more concise than:

```go
var errs []error
err := doSomething()
if err != nil {
    errs = append(errs, err)
}
```

**Note:** If you're adding context to an error with `NewErr()`, use regular `append()` since you know the error is non-nil:

```go
errs = append(errs, NewErr(ErrSomethingFailed, err))
```

## Type-Safe Metadata with KV Functions

doterr provides **typed ErrKV constructors** for creating metadata, inspired by the `log/slog.Attr` pattern:

```go
// Typed constructors
StringKV(key, value string) ErrKV
IntKV(key string, value int) ErrKV
Int64KV(key string, value int64) ErrKV
BoolKV(key string, value bool) ErrKV
Float64KV(key string, value float64) ErrKV
AnyKV(key string, value any) ErrKV
ErrorKV(key string, value error) ErrKV  // For errors as metadata (not causes)
```

**Benefits:**
- Type safety at creation time
- Explicit about value types
- Self-documenting code
- Consistent with log/slog patterns

**Usage:**

```go
err := NewErr(ErrDatabase,
    StringKV("table", "users"),
    IntKV("user_id", 42),
    BoolKV("retry_enabled", true),
    err, // err is the trailing cause
)
```

**Mixing old and new styles:**

```go
// Both styles work together
err := NewErr(ErrValidation,
    "field", "email",              // Old style: string pairs
    StringKV("pattern", emailRx),  // New style: typed ErrKV
    IntKV("length", len(email)),   // New style: typed ErrKV
    "required", true,              // Old style: string pairs
)
```

String pairs remain concise for simple cases. Use typed KV functions when you want explicit type safety or are building complex metadata.

## Accumulating Metadata with AppendKV()

`AppendKV()` allows you to **build up metadata throughout a function** before creating an error. It supports three forms and includes **lazy evaluation** for expensive operations:

**Three forms:**

```go
// 1. Individual ErrKV values
kvs = AppendKV(kvs, StringKV("name", "Alice"))

// 2. String key-value pairs
kvs = AppendKV(kvs, "age", 30)

// 3. Lazy evaluation with func()ErrKV
kvs = AppendKV(kvs, func()ErrKV{
    return IntKV("stats", generateExpensiveStats())
})
```

**Real-world example:**

```go
func processFile(path string, size int) (err error) {
    var kvs []ErrKV
    kvs = AppendKV(kvs, "path", path)
    kvs = AppendKV(kvs, "size", size)

    if size > 1000 {
        kvs = AppendKV(kvs, "truncated", true)
    }

    // Lazy evaluation - only computed if error occurs
    kvs = AppendKV(kvs, func()ErrKV{
        return IntKV("checksum", computeChecksum(path))
    })

    err = validate(path)
    if err != nil {
        return NewErr(ErrValidation, kvs, err) // kvs is passed as []ErrKV
    }
    return nil  // Happy path - lazy functions never evaluated
}
```

**Lazy evaluation benefits:**

- Expensive operations (checksums, stats, formatting) only run on error path
- Happy path stays fast
- Uses `sync.Once` internally - safe for concurrent access
- Evaluated exactly once when error is created

**Integration with NewErr() and WithErr():**

```go
// Accumulate metadata
var kvs []ErrKV
kvs = AppendKV(kvs, "user_id", userID)
kvs = AppendKV(kvs, "attempt", retryCount)

// Use in NewErr()
err := NewErr(ErrAuth, kvs, err)

// Use in WithErr()
err = WithErr(err, kvs)
```

## API summary

| Function                                                                                                                  | Purpose                                                                      |
|---------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| `NewErr(parts ...any)`                                                                                                    | Create a new entry with sentinels first, metadata, and optional trailing cause. |
| `MsgErr(msg any)`                                                                                                         | Create ad-hoc error message (string) or wrap error for rapid development.   |
| `WithErr(err error, parts ...any)`                                                                                        | Enrich existing error by merging into rightmost entry (enrichment only).    |
| `CombineErrs(errs []error)`                                                                                               | Join multiple independent errors (skips `nil`s, preserves order).           |
| `AppendErr(errs []error, err error) []error`                                                                              | Append error to slice only if non-nil (convenience to avoid if checks).     |
| `StringKV(key, value string) ErrKV`                                                                                          | Create string ErrKV pair.                                                       |
| `IntKV(key string, value int) ErrKV`                                                                                         | Create int ErrKV pair.                                                          |
| `Int64KV(key string, value int64) ErrKV`                                                                                     | Create int64 ErrKV pair.                                                        |
| `BoolKV(key string, value bool) ErrKV`                                                                                       | Create bool ErrKV pair.                                                         |
| `Float64KV(key string, value float64) ErrKV`                                                                                 | Create float64 ErrKV pair.                                                      |
| `AnyKV(key string, value any) ErrKV`                                                                                         | Create ErrKV pair with any value type.                                         |
| `ErrorKV(key string, value error) ErrKV`                                                                                     | Create ErrKV pair with error value (for errors as metadata, not causes).       |
| `AppendKV(kvs []ErrKV, parts ...any) []ErrKV`                                                                                   | Accumulate ErrKV pairs with support for lazy evaluation via func()ErrKV.          |
| `ErrMeta(err error) []ErrKV`                                                                                                 | Return metadata key/value pairs from all doterr entries (scans recursively).|
| `ErrValue[T](err error, key string) (T, bool)`                                                                            | Extract single metadata value by key with type safety.                      |
| `Errors(err error) []error`                                                                                               | Return sentinel/typed errors from first entry (unwraps one level).          |
| `FindErr[T](err error) (T, bool)`                                                                                         | Extract first typed error of type T using `errors.As`.                      |

### Implementation notes

* Each entry is a minimal struct implementing `Error()` and `Unwrap() []error`.
* `WithErr()` scans one join level right-to-left for an entry to enrich.
* No recursion deeper than one join level.
* No reflection or third-party dependencies.
* Every exported function returns the **built-in `error` type**.
* Lazy ErrKV evaluation uses `sync.Once` for deferred computation.

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

**What is a "trailing cause"?**

A **trailing cause** is the error you pass as the last argument to `NewErr()`. It represents the underlying error from a lower layer that caused this layer's error:

```go
err := lowerLayerFunc()
if err != nil {
    return NewErr(ErrRepo,
        "table", "users",
        err, // ← This is the "trailing cause"
    )
}
```

The trailing cause:
- Is always the **last argument** to NewErr()
- Represents the **underlying error** being wrapped
- Creates an **error chain** for errors.Is() and errors.As()
- Should be passed **between layers** but not enriched with WithErr() across package boundaries

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

## Security Analysis

`doterr` has an **extremely minimal security surface** and is very unlikely to introduce security vulnerabilities into your application. This section analyzes the security characteristics of the package.

### What doterr Does

From a security perspective, `doterr` is a pure data structure library that:

- Creates error values with structured metadata (key/value pairs)
- Wraps and unwraps errors using Go's standard `errors` package
- Stores metadata in Go slices and structs
- Provides helper functions to extract and format metadata
- Uses `errors.Join()` to compose error chains
- Performs string formatting for error messages using Go's standard `fmt` package

### Security Surface Analysis

**✓ No external I/O**
- **No network operations:** no HTTP, TCP, UDP, or any network protocols
- **No file system operations:** no reading or writing files
- **No external process execution:** no os/exec
- **No environment variable** manipulation
- **Completely isolated** from external systems

**✓ No parsing of untrusted data**
- **No complex data format** parsing: no JSON, XML, YAML, etc.
- **No user input or untrusted strings** parsing
- Simply **stores key/value pairs** provided by the caller
- **No interpretation or evaluation** of stored data

**✓ No code execution**
- **No code execution**
- **No string evaluation**
- **No use of `unsafe` package** for memory manipulation
- **No dynamic dispatch** via reflection or otherwise
- **Lazy evaluation via `func()ErrKV` **only execute user's own code**, not `doterr`'s

**✓ No cryptography**
- **No cryptographic** operations
- **No hashing, encryption, or signing**
- Uses `crypto/rand` only for generating unique package instance IDs at init time

**✓ Memory safety**
- Relies on **Go's built-in memory safety guarantees**
- **No pointer arithmetic** 
- **No unsafe memory access**
- **No unbounded memory growth** from doterr operations
- Memory usage is bounded by caller's metadata size

**✓ Concurrency safety**
- Uses `sync.Once` for lazy evaluation _(safe for concurrent access)_
- Lazy KV functions **evaluate exactly once**, thread-safe
- **No shared mutable state** across goroutines
- **Error structures are immutable** after creation

**✓ Limited reflection usage**
- Type assertions **only used for Go generics**, e.g., `ErrValue[T]`
- **No dynamic type** creation or modification
- **No reflection-based security bypasses**

### Potential Concerns (All Low Risk)

**1. Memory exhaustion**
- **Risk:** Caller could create errors with massive amounts of metadata
- **Mitigation:** Metadata size is bounded by what the caller provides (user-controlled)
- **Impact:** Same risk as any Go data structure - not specific to `doterr`
- **Assessment:** Not a security concern _(caller controls their own resource usage)_

**2. Panic conditions**
- **Risk:** Package includes panics for development-time validation errors
- **Examples:** `AppendKV()` panics on trailing keys without values
- **Mitigation:** These catch programmer errors during development, not runtime security issues
- **Impact:** Panics are documented and expected for API misuse
- **Assessment:** Design feature for fail-fast behavior on incorrect usage

**3. Error message injection**
- **Risk:** Could someone inject misleading error messages via metadata?
- **Mitigation:**
  - Metadata is structured (not parsed), preventing injection attacks
  - Error messages are formatted using `fmt.Sprintf` which is safe
  - No interpretation of metadata content as code or commands
- **Impact:** No more risk than Go's standard `error` interface
- **Assessment:** Not a security concern

**4. Cross-package ID collision**
- **Risk:** Two doterr instances could theoretically generate the same `uniqueId`
- **Mitigation:** Uses `math/rand.Int()` which has sufficient entropy for process-local uniqueness
- **Impact:** Would only affect cross-package error detection, not security
- **Assessment:** Not a security concern (worst case: false negative in development warning)

### Risk Profile Comparison

`doterr`'s security profile is **similar to Go's standard library `errors` package**:

| Package | External I/O | Parsing | Code Execution | Attack Surface |
|---------|--------------|---------|----------------|----------------|
| `errors` (stdlib) | None | None | None | Essentially zero |
| `doterr` | None | None | None | Essentially zero |
| `fmt` (stdlib) | None (just formatting) | Format strings | None | Minimal |
| `encoding/json` | None | JSON parsing | None | Low (DoS via large inputs) |

### Conclusion

**`doterr` is extremely unlikely to introduce security vulnerabilities** because:

1. **Zero attack surface** - No external inputs, outputs, or untrusted data handling
2. **Pure data structures** - Only stores and retrieves metadata provided by the caller
3. **No code evaluation** - Does not interpret or execute stored data
4. **Memory safe** - Relies on Go's memory safety with no unsafe operations
5. **Dependency-free** - Only uses Go standard library (`errors`, `fmt`, `strings`, `sync`)

The security risk of using `doterr` is comparable to using any other standard Go data structure like a `map[string]any` or `[]error`. The package does not process untrusted input, perform privileged operations, or interact with external systems.

**Recommendation:** Treat `doterr` like a utility data structure (similar to `container/list` or `sync.Map`) - it manipulates data you provide but does not introduce security concerns beyond standard Go safety guarantees.

## Objections

### Yet another error package?

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td><code>doterr</code> is <strong>complementary</strong> to existing error packages, not an alternative.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li>Keeps everything as the built-in <code>error</code> type.</li>
<li>Works with <code>errors.Is()</code>, <code>errors.As()</code>, and <code>errors.Join()</code>.</li>
<li>Enhances most existing error packages instead of competing with them.</li>
<li>Supports metadata like <code>slog.Logger</code> supports log attributes.</li>
<li>Adds error sentinels so downstream code can check errors with <code>errors.Is()</code> instead of parsing strings.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use only sentinels with <code>errors.Is()</code> and keep everything else as-is.</td>
</tr>
</table>

### Metadata keys become schema, and I do not want that

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>Metadata keys are optional. Use as few or as many as you need.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li><code>doterr</code> treats metadata as structured context, like <code>slog</code> treats log attributes.</li>
<li>You control which keys you use and how you use them.</li>
<li>Use metadata to provide context like <code>"user_id"</code> or <code>"path"</code> without encoding it in the error message string.</li>
<li>Keys are just strings - there's no required schema or validation.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use sentinels only with <code>errors.Is()</code> and no metadata at all.</td>
</tr>
</table>

### I do not want context traces

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>Metadata traces are optional and not required.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li><code>doterr</code> can capture key/value context at each error-handling layer (not a stack trace).</li>
<li>This is entirely optional - you can keep errors as shallow as you like.</li>
<li>You still benefit from sentinel identity with <code>errors.Is()</code> even without any metadata.</li>
<li>Use metadata only where it adds value for debugging or logging.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use only sentinels with <code>errors.Is()</code> and skip all metadata.</td>
</tr>
</table>

### I do not want to teach new developers a new way

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td><code>doterr</code> works like standard Go errors - just with added metadata support.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li><code>doterr</code> errors are still just <code>error</code> types that work with <code>errors.Is()</code> and <code>errors.As()</code>.</li>
<li>Developers already familiar with error sentinels will recognize the pattern.</li>
<li>The metadata is optional - developers can use as much or as little as makes sense.</li>
<li>Adopt incrementally: start with sentinels at public boundaries, keep internal code unchanged.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use <code>doterr</code> only for defining error sentinels, nothing else.</td>
</tr>
</table>

### I do not want lints policing errors

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td><code>doterr</code> is a library, not a linter. Use it however you want.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li><code>doterr</code> provides functions for creating and enriching errors - no linting required.</li>
<li>Future tooling may help validate error patterns, but it's completely optional.</li>
<li>Use <code>doterr</code> with or without any tooling or enforcement.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use <code>doterr</code> as a simple library with no tooling required.</td>
</tr>
</table>

### String matching is the caller's fault

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>It's still your API - give callers a better option.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li>When you return errors, developers <a href="https://www.hyrumslaw.com/"><strong>will</strong></a> branch on whatever affordance you expose.
<li>Can you expose sentinel errors, or just the text of your error messages, but you stil expose something.</li>
<li>If the only affordance is <code>.Error()</code>, callers will parse your message message text as strings.</li>
<li>Sentinels let callers use <code>errors.Is()</code> instead of string matching.</li>
<li>This makes error handling more reliable and less brittle.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Export a few sentinels and let error message strings change freely.</td>
</tr>
</table>

### Typed errors are enough

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>Typed errors don't compose well across dependencies.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li>Downstream code often must type-switch across many concrete error types from different packages.</li>
<li>Adding a method to an error interface is a breaking change for implementers.</li>
<li>Different libraries invent different type hierarchies, making cross-cutting error handling inconsistent.</li>
<li>Sentinels work with <code>errors.Is()</code> and don't require type assertions.</li>
<li>Sentinels enable more reliable error handling with just one quetstion: <em>"What sentinel do I check for?"</em></li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Use sentinels for public error categories; keep typed errors for internal details.</td>
</tr>
</table>

### I am shipping software, not taxonomy

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>Keep it simple - a few sentinels go a long way.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li>You don't need a complex error taxonomy.</li>
<li>Define sentinels for the error cases you actually have (not theoretical ones).</li>
<li>Maybe <code>3</code> to <code>5</code> well-chosen sentinels cover most libraries' needs.</li>
<li>Focus on errors that callers need to handle differently.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Define only the sentinels you already return in practice.</td>
</tr>
</table>

### This is overkill for small libraries

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td>Start small - even one sentinel is useful.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li>Small libraries often export just a few error types.</li>
<li>Using sentinels lets callers check errors with <code>errors.Is()</code> instead of parsing strings.</li>
<li>You can start with one or two sentinels and no metadata.</li>
<li>The overhead is minimal - just define error variables.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Define 1-3 sentinel errors for your library's main error cases.</td>
</tr>
</table>

### I don't want another dependency to manage

<table>
<tr>
<td valign="top"><strong>Short answer:</strong></td>
<td><code>doterr</code> is embedded vs. imported, and essentially finished; <strong>no need to ever update</strong>.</td>
</tr>
<tr>
<td valign="top"><strong>Longer answer:</strong></td>
<td>
<ul>
<li><code>doterr</code> is feature-complete for its intended purpose - no new features are planned.</li>
<li>Drop in <code>doterr</code> and <strong><em>you'll likely never need to update it</em></strong></li>
<li>You won't need to update unless you actively want to use new features.</li>
<li>It works like your own code, once you add it, it just works.</li>
<li>The security surface is minimal, see <a href="#security-analysis">Security Analysis</a>.</li>
<li>However, any future changes <strong>will</strong> be backward-compatible<sup>*</sup>.
</li>
<li>Treat it like part of the Go standard library with the same compatibility guarantees.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Security profile:</strong></td>
<td>
<ul>
<li><code>doterr</code> has essentially <strong>zero security surface area</strong>.</li>
<li>It performs <strong>no I/O</strong> <em>(no network, files, or external communication)</em>.</li>
<li>It <strong>does not parse untrusted data</strong>.</li>
<li>It <strong>does not execute untrusted code</strong>.</li>
<li>It is a <strong>pure data structure library</strong> like the standard <code>errors</code> package.</li>
<li>The risk profile is similar to using Go's built-in error handling - <em>virtually none</em>.</li>
</ul>
</td>
</tr>
<tr>
<td valign="top"><strong>Minimal adoption:</strong></td>
<td>Add it as a dependency once:
<ul>
<li><em>Then</em> <strong>forget about it</strong>.</li>
<li><code>doterr</code> will continue working without updates.</li>
</ul>
</td>
</tr>
</table>

`*` Backward compatibility is guaranteed, _with_ **one** small caveat:
1. Any new public functions could conflict with existing functions in packages that embed `doterr`.
    - However developers should notice immediately since the updated package will no longer compile.
    - If there is breakage it will likely be easy for a developer to resolve; just rename the function.
    - On the other hand there is _(almost?)_ no reason developers will be **required** to update.
    - By the very nature of `doterr`, different versions of `doterr` can coexists across different packages.
    - _Still,_ we have **NO PLANS** to add new functions, and
    - We will not decide to add a new function without serious consideration of potential breakage.

## Stability and Deprecation

doterr follows the stability levels and deprecation policies defined in the go-dt package:

- **Stability Levels**: See [go-dt/adrs/adr-2025-12-20-stability-levels.md](/Users/mikeschinkel/Projects/go-dt/adrs/adr-2025-12-20-stability-levels.md) for definitions of stable, evolving, experimental, deprecated, obsolete, and internal stability levels.
- **Error Sentinels**: See [adrs/adr-2025-12-20-error-sentinel-strategy.md](adrs/adr-2025-12-20-error-sentinel-strategy.md) for error-specific naming conventions and migration strategies.

**Tooling support:** Automated stability validation and deprecation tracking will be provided by [Squire](https://github.com/mikeschinkel/squire) as part of its API Stability Management features.

## License

MIT — © 2025 Mike Schinkel [mike@newclarity.net](mailto:mike@newclarity.net)
