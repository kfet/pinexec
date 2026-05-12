# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.0.4] - 2026-05-12

### Changed

- Internal: simplified the `Execute` implementation. Manual
  stdout/stderr pipe plumbing, the reader goroutines, the
  `sync.WaitGroup`, and the serializing mutex are gone — a single
  `io.Writer` is now assigned to both `cmd.Stdout` and `cmd.Stderr` and
  `os/exec` serializes writes for us. The rolling output window is now
  a `bytes.Buffer` trimmed in halves instead of a per-chunk ring. No
  public API change.

## [0.0.3] - 2026-05-11

### Changed

- `Execute` now uses [`os.CreateTemp`] for the >50KB output spill
  instead of hand-rolled `crypto/rand` + `os.Create`. Atomic uniqueness,
  no concurrent-collision edge case, fewer imports.
- `Execute` no longer retries temp-file creation on every chunk after
  the first failure. With a misconfigured `TMPDIR`, previous versions
  would re-attempt `os.Create` on every read; v0.0.3 attempts exactly
  once.
- `Execute` `defer`s the temp-file `Close` immediately after creation,
  so the file is closed and not leaked even if a user-supplied
  `onChunk` callback panics and unwinds through the read loop.

### Fixed

- Doc on `Result.Output` referenced `TruncationOptions` as the knob for
  truncation limits, but `Execute` uses hardcoded `DefaultMaxBytes` /
  `DefaultMaxLines`. Doc now points at the actual constants.
- `Execute`'s doc now notes that stdout and stderr are merged in
  *arrival* order; their relative ordering does not reflect which
  stream produced each chunk.

## [0.0.2] - 2026-05-10

### Changed (breaking)

- **`Options` struct removed.** `Execute` now takes the `onChunk`
  callback as a direct third parameter:

      // before (v0.0.1)
      res, err := pinexec.Execute(ctx, cmd, &pinexec.Options{OnChunk: cb})
      // after  (v0.0.2)
      res, err := pinexec.Execute(ctx, cmd, cb)

  `Options` had a single field; the struct was overhead with no
  forward-compat benefit (a future option would warrant a breaking change
  anyway, since the zero value is the no-op default). Pass `nil` for no
  streaming. Mirrors the upstream `fir/pkg/exec` simplification.

## [0.0.1] - 2026-05-10

Initial release. Extracted from `github.com/kfet/fir/pkg/exec`.

### Added

- `Execute(ctx, command, onChunk)` — run `$SHELL -c command` with combined
  stdout+stderr capture, ANSI/binary sanitization, line+byte tail
  truncation, optional temp-file spill, and cross-platform
  process-group cancellation.
- `onChunk` parameter — live raw-output callback (ANSI preserved) for UIs.
  Triggers `CLICOLOR`/`CLICOLOR_FORCE`/`FORCE_COLOR` env injection so
  CLIs that gate colors on TTY detection still emit them.
- `StripAnsi`, `AppendColorEnv` — ANSI helpers reused by `Execute`,
  exported for callers shaping their own output.
- `TruncateHead`, `TruncateTail`, `TruncationOptions`,
  `TruncationResult`, `DefaultMaxLines`, `DefaultMaxBytes`,
  `FormatSize` — line + byte truncation primitives.

### Changed (from `fir/pkg/exec`)

- Package renamed `exec` → `pinexec` to avoid stutter and stdlib clash.
- `ExecuteBash` → `Execute`, `BashResult` → `Result`.
- Dropped `ExecuteBashSimple` and `ExecuteBashCapture` — thin wrappers
  with no value-add over `Execute`.
- Temp-file prefix changed from `fir-bash-` to `pinexec-`.
