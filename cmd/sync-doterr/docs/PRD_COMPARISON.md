# PRD vs. Implementation Comparison

This document compares the original vision in [PRD.md](PRD.md) with the current `sync-doterr` implementation, and provides recommendations for what to preserve when building the generic _"drop-ins"_ system in [Gomion](https://github.com/mikeschinkel/gomion).

> **_NOTE_**: This comparison written by AI but not fully reviewed by humans.<br>
> — **Do not take anything written here as authoritative**.

## Executive Summary

The original PRD outlined an ambitious CLI tool with template management, multiple integration modes, interactive TUI, and comprehensive configuration. The implemented `sync-doterr` is a **minimal viable solution** that solves one specific problem: synchronizing embedded `doterr.go` files across packages.

**Current state:** `sync-doterr` does ~15% of the PRD vision, but covers the most critical use case.

## Feature Comparison

### ✅ Implemented in sync-doterr

| Feature | PRD Section | Implementation Notes |
|---------|-------------|---------------------|
| File synchronization | 6.2 (partial) | Basic sync from source to targets |
| Package name detection | 5.1 (partial) | Detects from existing .go files |
| Confirmation prompt | 6.2 (partial) | Simple y/N, not TUI |
| Skip unchanged files | 9 | Byte-level comparison |
| Preserve file permissions | 9 | Maintains existing mode |
| Recursive scanning | 8 | Walks directory tree |
| Source file exclusion | 8 | Never updates source itself |

### ❌ Not Implemented

| Feature | PRD Section | Impact | Priority for Gomion |
|---------|-------------|--------|-------------------|
| **Template system** | 5.1 | No variables (SPDX, version, timestamp) | **HIGH** - Essential for generic drop-ins |
| **Configuration file** (`doterr.json`) | 5.2 | No persistent settings | **HIGH** - Need package-level config |
| **Multiple modes** (dot/copy/ignore) | 10 | Only does "copy" | **MEDIUM** - Copy mode is most useful |
| **Interactive TUI** | 11 | No bubbletea interface | **HIGH** - Better UX than CLI flags |
| **`doterr init` command** | 6.1 | No project setup | **HIGH** - First-time setup is critical |
| **`doterr check` command** | 6.3 | No CI validation | **MEDIUM** - Useful for CI/CD |
| **`doterr upgrade` command** | 6.4 | No version migration | **LOW** - Premature for v1 |
| **Dry-run mode** | 6.2 | No preview | **MEDIUM** - Safety feature |
| **Deterministic builds** | 14 | Always includes timestamps (none currently) | **LOW** - Nice-to-have |
| **go:generate integration** | 7 | No directive management | **MEDIUM** - Automation value |
| **Dot-import management** | 10 | Doesn't handle import lines | **LOW** - Community prefers embedding |
| **Path glob patterns** | 5.2 | No pattern matching | **MEDIUM** - Useful for exclusions |
| **Auto-mode for new packages** | 6.2 | Always prompts | **LOW** - Interactive is fine |
| **Format with gofmt** | 9 | Assumes source is formatted | **LOW** - Tooling handles this |

## What Was Actually Built

`sync-doterr` is essentially:

```bash
# Pseudocode representation
find <root-dir> -name "doterr.go" |
  exclude <source-file> |
  for each target:
    detect package name from target dir
    replace "package X" line in source
    if content differs:
      confirm with user
      write to target
```

**Strengths:**
- Solves the immediate problem
- Simple to understand and use
- No external dependencies beyond go-cliutil and go-dt
- Fast and reliable

**Limitations:**
- No configuration persistence
- No project-level management
- No template variables
- doterr-specific (not generic)
- Manual workflow (no go:generate)

## Recommendations for [Gomion](https://github.com/mikeschinkel/gomion) _"Drop-ins"_

### Must-Have Features (from PRD)

1. **Template System** (PRD §5.1)
   - Support variable substitution: `{{ .PackageName }}`, `{{ .License }}`, etc.
   - Generic enough for any embeddable file type
   - Example use cases: error handlers, loggers, config loaders, test helpers

2. **Configuration File** (PRD §5.2)
   - Format: `.gomion.json` or `gomion.yaml`
   - Per-package overrides with glob patterns
   - Multiple drop-in definitions in one config
   - Example:
     ```json
     {
       "drop-ins": {
         "doterr": {
           "source": "templates/doterr.go.tmpl",
           "packages": {
             "./internal/**": true,
             "./cmd/**": true,
             "./vendor/**": false
           }
         },
         "logger": {
           "source": "templates/logger.go.tmpl",
           "packages": "./services/**"
         }
       }
     }
     ```

3. **Interactive Setup** (PRD §6.1, §11)
   - `gomion init` - Interactive TUI to select packages
   - Show package list with toggles
   - Preview changes before applying
   - Use bubbletea for consistent UX

4. **Apply Command** (PRD §6.2)
   - `gomion apply [drop-in-name]` - Apply configuration
   - Support `--dry-run` for safety
   - Smart package detection
   - Only update when content differs

5. **Check Command** (PRD §6.3)
   - `gomion check` - Verify no drift (for CI)
   - Exit non-zero if files out of sync
   - Report which files need updates

### Nice-to-Have Features

6. **Deterministic Builds** (PRD §14)
   - `--deterministic` flag to omit timestamps
   - Useful for reproducible builds
   - Lower priority - most projects don't need this

7. **go:generate Integration** (PRD §7)
   - Auto-add `//go:generate gomion apply` directives
   - Integrate with standard Go workflows
   - Medium priority - helpful but not critical

8. **Path Globs** (PRD §5.2)
   - Support patterns like `./internal/**`, `./cmd/*`
   - Exclude patterns: `!./vendor/**`
   - Medium priority - improves flexibility

### Features to Sunset (Not Worth Implementing)

1. **Dot-import Mode** (PRD §10)
   - **Reason:** User decided not to support dot-import mode and instead only support _"drop-in"_ mode
   - **Evidence:** The PRD itself recommends "copy" mode for libraries
   - **Decision:** Focus on embedded files only

2. **Template Versioning** (PRD §6.4, §16)
   - **Reason:** Over-engineered for typical use cases
   - **Evidence:** doterr API is stable; version churn unlikely
   - **Decision:** Keep templates simple; handle breaking changes manually
   - **HUMAN NOTE**: doterr is just one potential drop-in, versioning might be needed for others.

3. **Auto-mode for New Packages** (PRD §6.2)
   - **Reason:** Interactive prompts are better UX
   - **Evidence:** Developers want control over what gets embedded
   - **Decision:** Always prompt; let users decide
   - **HUMAN NOTE**: _Maybe_ sunset...

4. **Network Template Fetching** (PRD §3, §16)
   - **Reason:** Adds complexity and security concerns
   - **Evidence:** Local templates are simpler and more secure
   - **Decision:** Templates live in the repo

## Migration Path

### Phase 1: Current State (sync-doterr)
- ✅ Basic sync functionality
- ✅ Solves immediate need
- ✅ Zero external dependencies (except go-cliutil/go-dt)

### Phase 2: Gomion "Drop-ins" v1
- Implement must-have features (1-5 above)
- Generic template system
- Configuration file
- Interactive TUI setup
- Apply and check commands

### Phase 3: Gomion "Drop-ins" v2
- Add nice-to-have features (6-8 above)
- Deterministic builds
- go:generate integration
- Advanced glob patterns

### Deprecation
- Once Gomion v1 is stable, deprecate sync-doterr
- Update doterr README to reference Gomion
- Archive sync-doterr code for reference

## Key Insights

1. **The PRD got the core concept right** - Template-based drop-ins with package-level configuration is the right model.

2. **sync-doterr validated the need** - The minimal implementation proved the use case is real and valuable.

3. **Interactive > Declarative** - TUI setup is more important than comprehensive CLI flags.

4. **Generic > Specific** - Gomion should handle any embeddable file, not just doterr.go.

5. **Simple > Feature-rich** - Start with core workflow (init → apply → check), add polish later.

## Decision Matrix

For each PRD feature, should it be included in Gomion?

| Feature | Include? | Reasoning |
|---------|----------|-----------|
| Template system | ✅ YES | Core requirement |
| Config file | ✅ YES | Essential for multi-package projects |
| Interactive TUI | ✅ YES | Better UX than CLI-only |
| Apply command | ✅ YES | Core workflow |
| Check command | ✅ YES | CI integration value |
| Dry-run mode | ✅ YES | Safety and preview |
| Init command | ✅ YES | First-time setup |
| Path globs | ✅ YES | Flexibility |
| go:generate | ⚠️ MAYBE | Nice automation, medium priority |
| Deterministic | ⚠️ MAYBE | Edge case, low priority |
| Dot-import mode | ❌ NO | Community doesn't use it |
| Template versioning | ❌ NO | Over-engineered |
| Auto-mode | ❌ NO | Interactive is better |
| Network fetching | ❌ NO | Security and complexity |
| Upgrade command | ❌ NO | Manual migration is fine |

## Conclusion

The PRD outlined a **comprehensive vision** that's still valid for Gomion. The current `sync-doterr` is a **pragmatic MVP** that validated the need. When building Gomion's drop-ins system:

**Preserve:**
- Template system with variables
- Configuration file with package overrides
- Interactive TUI for setup
- Apply/check commands for workflows
- Dry-run and validation

**Sunset:**
- Dot-import management
- Template versioning
- Auto-mode flags
- Network fetching

**Result:** A focused, powerful tool for managing embedded source files across Go projects, applicable to any "drop-in" pattern, not just doterr.

---

||                                          |
|--|------------------------------------------|
|**Author:**| Claude Code                              |
|**Date:**| 2025-12-20                               |
|**Status:**| Analysis for planning development |
