---
name: ksrc
description: CLI for searching/reading Kotlin dependency sources from a Gradle project. Use when you need to locate a class/function in a library and inspect dependency source files.
---

## Quick start
- Search a module:
  `ksrc search org.jetbrains.kotlinx:kotlinx-datetime -q "class LocalDate"`
- Read a file by id:
  `ksrc cat org.jetbrains.kotlinx:kotlinx-datetime:0.6.1!/kotlinx/datetime/LocalDate.kt --lines 1,200`

## Commands
### `ksrc search [<module>] -q <pattern> [-- <rg-args>]`
Search dependency sources.

Common flags:
- `--all` search across all resolved deps (required if `<module>` omitted)
- `<module>` supports glob patterns (same as `--module`)
- `--project <path>` project root (default `.`)
- `--subproject <name>` limit to a subproject (repeatable)
- `--targets <list>` limit KMP targets (comma‑separated: `jvm,android,iosX64`)
- `--config <name>` resolve specific configuration(s) (comma‑separated)
- `--scope <compile|runtime|test|all>`
- `--module <glob>` module filter (`group:artifact[:version]`)
- `--group <glob>` / `--artifact <glob>` / `--version <glob>`
- `--offline` only use cached sources
- `--refresh` force dependency refresh
- `--context <n>` shortcut for `rg -C <n>` (context lines emit column `0`)
- `--rg-args <args>` extra rg args (comma‑separated)
- `-- <rg-args>` pass through raw rg args
- `--show-extracted-path` include temp extracted paths in output (off by default)

Output format:
`<file-id> <line>:<col>:<match>` (use `--show-extracted-path` for temp paths)

### `ksrc cat <file-id|path>`
Print file contents.

Common flags:
- `--lines <start,end>` 1‑based inclusive range
- `--module <glob>` / `--group` / `--artifact` / `--version` to disambiguate when using a path

### `ksrc open <file-id|path>`
Open in `$PAGER` (defaults to `less -R`). Same flags as `cat`.

### `ksrc deps`
List resolved dependencies and source availability.

### `ksrc resolve`
Resolve and print source JARs: `group:artifact:version|/path/to/sources.jar`.

### `ksrc fetch <coord>`
Ensure sources for a coordinate exist: `group:artifact:version`.

### `ksrc where <path|coord>`
Locate cached source JAR or file.

### `ksrc doctor`
Basic diagnostics for environment issues.

## File-id format
`group:artifact:version!/path/inside/jar.kt`

## Common issues
- `E_NO_MODULE`: provide `<module>` or pass `--all`.
- `E_NO_SOURCES`: dependency sources not available; try `ksrc deps`, `ksrc fetch <coord>`, or rerun without `--offline`.
- Gradle not found: run in a Gradle project or set `--project` to the root.
- `rg` not found: install ripgrep and ensure it’s on PATH.
- Ambiguous modules: use `--module`, `--group`, or `--artifact` to narrow.
