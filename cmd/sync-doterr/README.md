# sync-doterr

A CLI tool for synchronizing `doterr.go` files across multiple Go packages.

> **TEMPORARY SOLUTION:** This tool is a minimal implementation to solve the immediate need of syncing embedded `doterr.go` files. It will be replaced by generic "drop-ins" functionality in [Gomion](https://github.com/mikeschinkel/gomion) (or its successor), which will support managing any embeddable source files, not just doterr.go.
>
> This app and its documentation was almost entirely written by AI and as such may or may not be a good representation of best practices of use of `doterr.go`, any of the imports it uses, or Go coding in general.

## Purpose

`sync-doterr` automates the process of keeping `doterr.go` files consistent across multiple packages when used as an  **embedded source _"drop-in"_**, as described in the [doterr README](../../README.md).

Since `doterr` is designed to be embedded directly into each package, you may have multiple copies of `doterr.go` throughout your codebase. When you update the implementation or fix bugs in one copy, this tool ensures all other copies stay in sync.

This tool **currently** does not concern itself with versioning because the intent is that `doterr.go` **should never become backward incompatible**, meaning that a developer should always be able to update to the latest version.

## Build and Run


Use `make sync` to build and run sync-doterr _("installtion" is not generally needed.)_

```bash
cd /path/to/go-doterr
make sync DIR=<root_dir>
```

This automatically:
- Builds `sync-doterr` to `/path/to/go-doterr/bin/sync-doterr`
- Runs it from `/path/to/go-doterr`
- Syncs from from `/path/to/go-doterr/doterr.go` to all target paths
- Determines target paths based on finding files named `doterr.go` withing `<root_dir>`
- Presents a list of files to update and requests confirmation before updating files.

Note that `DIR` can accept tilde paths, e.g. `DIR=~/Projects`

### Example

Sync all `doterr.go` files under the `/Users/you/Projects/go-pkgs` directory:

```bash
cd /Users/you/Projects/go-pkgs/go-doterr    # Contains source doterr.go
make sync DIR=~/Projects
```

### What it does

1. **Scans** the root directory recursively for all `doterr.go` files
2. **Detects** the correct package name for each target directory by reading other `.go` files
3. **Shows** you a preview of:
   - Source file path
   - All target file paths
   - Count of files to update
4. **Prompts** for confirmation before making changes
5. **Updates** each target file by:
   - Replacing the `package` declaration with the target's package name
   - Preserving the rest of the source file exactly
   - Maintaining file permissions
6. **Reports** results:
   - Number of files updated
   - Number of files skipped (already up-to-date)

### Example output

```
Scanning ~/Projects/go-pkgs for doterr.go files...

Source file:
  ~/Projects/go-pkgs/go-doterr/doterr.go

Target files (8):
  ~/Projects/go-pkgs/go-cliutil/doterr.go
  ~/Projects/go-pkgs/go-dt/doterr.go
  ~/Projects/go-pkgs/go-fsfix/doterr.go
  ~/Projects/go-pkgs/go-jsontest/doterr.go
  ~/Projects/go-pkgs/go-jsonxtractr/doterr.go
  ~/Projects/go-pkgs/go-logutil/doterr.go
  ~/Projects/go-pkgs/go-rfc9457/doterr.go
  ~/Projects/go-pkgs/go-testutil/doterr.go

Update these files? (y/N): y

  ~/Projects/go-pkgs/go-cliutil/doterr.go updated
  ~/Projects/go-pkgs/go-dt/doterr.go skipped
  ~/Projects/go-pkgs/go-fsfix/doterr.go updated
  ~/Projects/go-pkgs/go-jsontest/doterr.go updated
  ~/Projects/go-pkgs/go-jsonxtractr/doterr.go updated
  ~/Projects/go-pkgs/go-logutil/doterr.go updated
  ~/Projects/go-pkgs/go-rfc9457/doterr.go updated
  ~/Projects/go-pkgs/go-testutil/doterr.go updated


Updated: 7
Skipped: 1 (skipped files were already up-to-date)
```

## Package name detection

The tool automatically detects the correct package name for each target directory by:

1. Reading all `.go` files in the directory (excluding `doterr.go` itself)
2. Finding the first valid `package` declaration
3. Skipping `*_test.go` package declarations (to avoid `package foo_test`)

If no Go files are found, it falls back to using the directory's base name and shows a warning.

## Safety features

- **Confirmation prompt** - Always asks before making changes
- **Skip unchanged files** - Detects when target files are already up-to-date
- **Preserve permissions** - Maintains existing file mode/permissions
- **Source file exclusion** - Never attempts to update the source file itself
- **Byte-level comparison** - Only updates files when content actually differs

## Global options

Inherited from `go-cliutil`:

- `--quiet` - Suppress non-essential output
- `--verbose` - Increase verbosity (can be used multiple times: `-v`, `-vv`, `-vvv`)

## Use cases

Cases #2 and #3 are hopefully very rare:

1. **Consistency:** When you want to ensure all embedded copies stay identical _(except for the declared `package`  name, of course)._
1. **Multi-package projects:** When you have multiple packages in a monorepo or organized directory structure, each with its own embedded `doterr.go`.
1. **Bug fixes or Feature additions:** When we fix a bug in `doterr.go` and need to propagate to all your copies 


## Error handling

The tool uses the `doterr` pattern itself for structured error handling. Common errors include:

- **Missing source file** - Current directory must contain `doterr.go`
- **Invalid root directory** - Target directory must exist
- **Permission errors** - Must have read/write access to target files
- **Invalid package names** - Target directories must contain valid Go packages

## Implementation notes

- This `sync-doterr` CLI tool Built using `go-cliutil` for CLI framework
- Uses `go-dt` domain types for type-safe path handling
- Implements the `doterr` error pattern (self-hosting!)
- Single-pass file transformation with minimal memory footprint
- Preserves all content except, appropriately, the original `package` name.

## See also

- [doterr README](../../README.md) - Main documentation for the doterr pattern
- [ADR: Error Sentinel Strategy](../../adrs/adr-2025-12-20-error-sentinel-strategy.md) - Design decisions

## License

MIT - © 2025 Mike Schinkel
