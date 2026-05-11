# AGENTS.md

Guidance for AI agents working on `pinexec`.

## Scope

`pinexec` is a **small, focused** package that runs `$SHELL -c` commands
with cancellation, output sanitization, and live streaming. Building
blocks:

- `Execute` / `Options` / `Result` — the headline API (`exec.go`)
- Process-group setup — `setProcGroup` (`procgroup_unix.go`,
  `procgroup_windows.go`)
- ANSI handling — `StripAnsi`, `AppendColorEnv` (`ansi.go`)
- Truncation — `TruncateHead`, `TruncateTail`, `TruncationOptions`,
  `TruncationResult`, `DefaultMaxLines`, `DefaultMaxBytes` (`truncate.go`)

`doc.go` is the source of truth for the public API surface; keep it and
this list in sync.

**Do not** add provider-specific or domain-specific helpers (path
resolution, recipe DSLs, retry loops, …). They belong in the consumer's
codebase.

## Constraints

- **Stdlib only.** No third-party deps. Ever. If you reach for one, stop
  and ask first.
- **Go 1.21+.** Don't use language features newer than that without a
  real need; bumping the minimum cuts users.
- **No global state.** No `init()` registries. No package-level
  mutables.
- **Tests run real commands** (`echo`, `sleep`, `printf`, `seq`, …) on
  the host shell. No mocking the runner.

## Workflow

- `make all` runs gofmt + go vet + staticcheck (if installed) + race
  tests with a **100% coverage gate** (excluding paths in `.covignore`).
  Must pass before any commit.
- Add a `## [Unreleased]` entry in `CHANGELOG.md` for any user-visible
  change.
- Update `doc.go`, `README.md`, and `AGENTS.md` when the public API
  changes.

## Public API

Anything exported from the package is API. Treat it as semver-stable
from v0.1.0 onward. Renames or signature changes need a major bump.

## Coverage exemptions

`procgroup_windows.go` is excluded from the coverage gate via
`.covignore` — it can't be exercised on the macOS/Linux CI runners. The
same coverage line applies to the `_` discard error paths in `exec.go`:
if a CSPRNG read, temp-file create, or temp-file write fails, pinexec
proceeds with degraded behaviour (no temp-file spill, no buffered
flush) rather than aborting the run. These are structurally hard to
exercise in tests; the file-level guard is in `.covignore`.
