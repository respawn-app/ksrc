# AGENTS.md

## Project Purpose
`ksrc` is a CLI for one‑liner search and file read of Kotlin dependency sources. It resolves dependency versions from the project, ensures source JARs are present in **Gradle caches**, and runs `rg --search-zip` over those JARs. Output includes a `<file-id>` so `ksrc cat` can read files without extra resolution steps.

## Stack
- Language: Go 1.22+
- CLI: cobra
- Search: external `rg` (`--search-zip`)
- Gradle integration: per‑invocation init script (`-I <temp>`), prefer `./gradlew`, fallback to `gradle`
- File read: Go `archive/zip` + line slicing

## Philosophy / Rules
- Zero project mutation: no files written to the repo.
- No custom cache/index: use Gradle caches only.
- One‑liner UX: `ksrc search <module> -q <pattern>` and `ksrc cat <file-id>`.
- Deterministic resolution: prefer project‑resolved versions; only fall back to cache‑latest if absent.
- Keep output stable and parseable.

## Directory / Module Structure (planned)
- `cmd/` — CLI entry points and wiring
- `gradle/` — init script generation, Gradle execution, output parsing
- `resolve/` — version selection, module filtering, cache scanning
- `search/` — rg invocation, result formatting, file‑id emission
- `cat/` — zip file read + `--lines` slicing
- `internal/` — shared helpers (logging, error codes)
- `testdata/` — minimal Gradle fixtures
- `docs/` — CLI spec and stack

## Common Tasks
- Add/adjust commands: update cobra wiring in `cmd/`, keep flags consistent with `docs/cli-api.md`.
- Resolution changes: keep init script minimal and compatible with multiple Gradle versions.
- Search changes: must keep `rg` call scoped to resolved JARs only.
- Cat changes: must support `<file-id>` and `--lines start:end`.
- After code changes, rebuild the binary to `./bin/ksrc` so the symlinked CLI updates.

## Tests
- Unit: parsing, version selection, file‑id handling.
- Integration: run against `testdata/` Gradle fixture; no repo mutation; asserts on output format.

## Clean Merge Expectations
- Keep changes focused;
- Update docs when CLI flags or output formats change.
- Update AGENTS.md (this file) with learnings/rules and project info.
