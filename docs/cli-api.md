# ksrc CLI API Spec

## Purpose
Provide single‑command, `rg`‑style search and file read for Kotlin dependency sources, with zero project mutation and minimal context pollution. The UX must work in a freshly opened project with no caches or dependencies downloaded, and must rely on Gradle’s own caches (no ksrc cache/index).

## Goals
- One‑liner search: `ksrc search <module> -q "fun foo"` with automatic dependency resolution and source download.
- One‑liner file read: `ksrc cat <file-id>` or `ksrc open <file-id>`.
- No changes to the project working tree.
- Use Gradle caches only; no additional ksrc cache or index.
- Deterministic, repeatable dependency resolution.

## Non‑Goals
- IDE integration.
- Build, test, or runtime execution.
- Cross‑language search beyond Kotlin/JVM dependency sources.

---

## End‑to‑End (0 → hero) Path
**Scenario:** brand‑new project, no Gradle caches, no sources downloaded.

1) User runs a single command:
```
ksrc search kotlinx.datetime -q "public class LocalDate" --project .
```

2) CLI behavior (no prompts unless `--interactive`):
- Detects a supported Kotlin project (Gradle) or fails fast with actionable error.
- Resolves dependency graph for the selected configuration(s).
- Ensures sources are present in Gradle caches (downloads if missing).
- Runs search scoped to the resolved source JARs and prints `rg`‑style results plus a file identifier.

3) Follow‑up file read, one‑liner:
```
ksrc cat org.jetbrains.kotlinx:kotlinx-datetime:0.6.1!/kotlinx/datetime/LocalDate.kt --lines 1,200
```
The CLI resolves the file identifier and prints contents.

**No files in the repo are created, modified, or deleted.**

---

## Command Overview

### `ksrc search <module>`
Search dependency sources for a target module or group.

**Usage**
```
ksrc search [<module>] -q <pattern> [flags]
```
`<module>` is required unless `--all` is provided.

**Key Flags**
- `--project <path>`: Project root (default: `.`)
- `--all`: Search across all resolved dependencies (required if `<module>` is omitted)
- `--subproject <name>`: Limit resolution to a subproject (repeatable)
- `--targets <list>`: Limit KMP targets (comma‑separated, e.g. `jvm,android,iosX64`)
- `--config <name>`: Dependency configuration(s) to resolve (default: inferred)
- `--module <glob>`: Filter by `group:artifact` glob (alias of positional `<module>`)
- `--group <glob>`: Filter by group
- `--artifact <glob>`: Filter by artifact
- `--version <glob>`: Filter by version
- `--scope <compile|runtime|test|all>`: Dependency scope (default: `compile`)
- `--refresh`: Re‑resolve and re‑download sources
- `--offline`: Only use cached sources, error if missing
- `--max-results <n>`: Limit output
- `--rg-args <args>`: Extra args passed to `rg`
- `--emit-id <always|auto|never>`: Include file identifiers (default: `always`)

**Output (default)**
`<file-id> <file>:<line>:<col>:<match>` (rg‑compatible with leading file id)

**Aliases**
- `ksrc rg` is an alias of `ksrc search`

---

### `ksrc cat <file-id|path>`
Print file contents to stdout. Resolves the file from dependency sources.

**Usage**
```
ksrc cat <file-id|path> [flags]
```

**Path Forms**
- Relative source path: `org/jetbrains/kotlinx/coroutines/flow/Flow.kt`
- Fully qualified path: `group:artifact:version!/org/.../Flow.kt`

**Flags**
- `--project <path>`
- `--module <glob>` (disambiguate)
- `--lines <start,end>`: Output a line range (1‑based, inclusive; sed‑style)

---

### `ksrc open <path>`
Open a file in `$PAGER` (defaults to `less -R`).

**Usage**
```
ksrc open <path> [flags]
```

---

### `ksrc deps`
List resolved dependencies and source availability.

**Usage**
```
ksrc deps [flags]
```

**Output (default)**
`group:artifact:version  [sources: yes|no]  [path: <gradle cache path>]`

---

### `ksrc fetch <coord>`
Ensure sources for a coordinate exist in Gradle caches.

**Usage**
```
ksrc fetch org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1
```

**Flags**
- `--project <path>` (optional, if resolving via project)
- `--refresh`

---

### `ksrc where <path|coord>`
Locate the Gradle cached source artifact or file.

**Usage**
```
ksrc where org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1
ksrc where org/jetbrains/kotlinx/coroutines/flow/Flow.kt
```

---

### `ksrc resolve`
Resolve the dependency graph without search. No project files are modified.

**Usage**
```
ksrc resolve [flags]
```

---

### `ksrc doctor`
Diagnostics for project detection, Gradle cache accessibility, and source availability.

---

## File Identifier
`<file-id>` is a fully qualified path to a file inside a source JAR:
`group:artifact:version!/path/inside/jar.kt`

`ksrc search` emits `<file-id>` in every result line so clients can call `ksrc cat <file-id>` with no extra resolution steps.

---

## Resolution Behavior

### Project Detection
- Default project root is `.`
- Supported: Gradle (Groovy or Kotlin DSL)
- If no supported project is found, error and suggest `ksrc fetch <coord>` or `ksrc search <module> -q <pattern>`.

### Dependency Resolution
- Uses project dependency configuration(s) based on `--scope`/`--config`.
- Must be deterministic and reproducible.
- Does **not** modify project files.
- Uses Gradle to resolve and download source artifacts.

### Resolution Mechanism
- Default runner: project Gradle wrapper (`./gradlew`); fallback: `gradle` on PATH.
- Uses a per‑invocation Gradle init script passed via `-I <temp-file>`.
- The init script resolves source artifacts and prints `group:artifact:version|/path/to/sources.jar` to stdout.
- The init script lives in system temp (e.g., `/tmp/ksrc-init-<pid>.gradle.kts`) and is deleted after the command completes.
- No files are written to the repo.

### Default Resolution Scope (KMP)
When `--config` is not provided, resolve **all main compile classpaths** across all subprojects:
- `commonMainCompileClasspath` (if present)
- `<target>MainCompileClasspath` for all detected KMP targets

`--subproject`, `--targets`, and `--config` let callers narrow or expand the scope.

### Version Selection
When the module is specified without an explicit version:
1) Use the version resolved by the project’s dependency graph (authoritative).
2) If not present in the project graph, pick the highest version already in Gradle caches (Maven version ordering).
3) If still ambiguous or missing, fail and require `--version`.

### Source Acquisition
- For each dependency: attempts to locate source artifacts in Gradle caches.
- If missing and not `--offline`: Gradle downloads sources before search.
- If `--offline`: missing sources are reported and search excludes them.

### Temporary Files
- `/tmp` (or system temp) may be used for transient extraction to support `cat --lines`.
- `/tmp` (or system temp) may be used for transient extraction to enable `search` over source JARs.
- Temp files are deleted after command completion.

---

## Caveats
- Gradle version variance: keep the init script minimal and on stable APIs.
- Performance: each run starts Gradle (first run may be slow).
- Policy restrictions: some environments disallow `-I` init scripts; fail with a clear error and suggest plugin use.
- Multi‑project builds: default resolves all subprojects; use `--subproject`/`--targets`/`--config` to narrow.
- Offline mode: if sources aren’t cached, `--offline` fails by design.

---

## Error Handling
- Non‑zero exit code on failures
- Suggested remediation in trailing text line
- Common cases:
  - `E_NO_PROJECT`: No Gradle project detected
  - `E_NO_DEPS`: No dependencies to resolve
  - `E_NO_MODULE`: `<module>` omitted without `--all`
  - `E_NO_SOURCES`: Sources not found and `--offline` set
  - `E_AMBIGUOUS`: Multiple modules match; requires `--module`

---

## Examples

### One‑liner search in a fresh project
```
ksrc search kotlinx.coroutines -q "CoroutineScope" --project .
```

### One‑liner file read
```
ksrc cat org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1!/kotlinx/coroutines/CoroutineScope.kt --lines 1,200
```

### Filter search to a module
```
ksrc search org.jetbrains.kotlinx:kotlinx-coroutines-core -q "internal"
```

### Resolve only, no search
```
ksrc resolve --project .
```

### Offline search using cache only
```
ksrc search kotlinx.coroutines -q "StateFlow" --offline
```

---

## Compatibility Requirements
- Works from a clean checkout with no existing Gradle caches.
- Provides a one‑command path to search and read sources.
- Does not mutate the project directory.
- No interactive prompts unless `--interactive`.

---

## Versioning
- CLI output and flags are semver‑governed.
- Breaking changes require a major version bump.
