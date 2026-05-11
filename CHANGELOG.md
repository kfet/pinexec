# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.0.1] - 2026-05-10

Initial release. Extracted from `github.com/kfet/fir/pkg/exec`.

### Added

- `Execute(ctx, command, opts)` — run `$SHELL -c command` with combined
  stdout+stderr capture, ANSI/binary sanitization, line+byte tail
  truncation, optional temp-file spill, and cross-platform
  process-group cancellation.
- `Options.OnChunk` — live raw-output callback (ANSI preserved) for UIs.
  Triggers `CLICOLOR`/`CLICOLOR_FORCE`/`FORCE_COLOR` env injection so
  CLIs that gate colors on TTY detection still emit them.
- `StripAnsi`, `AppendColorEnv` — ANSI helpers reused by `Execute`,
  exported for callers shaping their own output.
- `TruncateHead`, `TruncateTail`, `TruncationOptions`,
  `TruncationResult`, `DefaultMaxLines`, `DefaultMaxBytes`,
  `FormatSize` — line + byte truncation primitives.

### Changed (from `fir/pkg/exec`)

- Package renamed `exec` → `pinexec` to avoid stutter and stdlib clash.
- `ExecuteBash` → `Execute`, `BashExecutorOptions` → `Options`,
  `BashResult` → `Result`.
- Dropped `ExecuteBashSimple` and `ExecuteBashCapture` — thin wrappers
  with no value-add over `Execute`.
- Temp-file prefix changed from `fir-bash-` to `pinexec-`.
