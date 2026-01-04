# ksrc Stack

## Primary Language
- Go (1.22+). Single‑binary CLI, fast startup, strong stdlib for process control and zip handling.

## CLI Framework
- `spf13/cobra` for subcommands and flags. Minimal dependency surface, common patterns, easy to extend.

## Build & Release
- `go build` for local builds.
- `goreleaser` for reproducible releases and artifact signing.

## Gradle Integration
- Execute `./gradlew` (preferred) or `gradle` via `os/exec`.
- Per‑invocation Gradle init script (Kotlin DSL) generated at runtime and passed with `-I`.
- Parse init script stdout for `group:artifact:version|/path/to/sources.jar`.

## Search Engine
- External `rg` (ripgrep) invocation for `ksrc search`.
- Use `rg --search-zip` to scan source JARs without extraction.

## File Read (`ksrc cat`)
- Implement file extraction in Go using `archive/zip` to avoid external tools.
- Line ranges handled in‑process (1‑based, inclusive).

## Runtime Dependencies (External)
- Gradle wrapper or `gradle` on PATH.
- `rg` on PATH.

## Internal Structure (Modules)
- `cmd/`: CLI entry points and command wiring.
- `gradle/`: init script generation, Gradle execution, output parsing.
- `resolve/`: version selection and module filtering logic.
- `search/`: rg invocation + result parsing/formatting.
- `cat/`: zip file read and line slicing.

## Testing
- Table‑driven unit tests for parsing and resolution.
- Golden tests for CLI output.
- Integration tests with a minimal Gradle fixture (no repo mutation).
