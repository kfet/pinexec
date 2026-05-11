# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

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
