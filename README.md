# pinexec

A small, dependency-free Go package for running shell commands with
cancellation, output sanitization, and live streaming.

[![pkg.go.dev](https://pkg.go.dev/badge/github.com/kfet/pinexec.svg)](https://pkg.go.dev/github.com/kfet/pinexec)

## Why

`os/exec` is the right primitive when you need fine-grained control, but
plugging it into an AI coding agent (or any UI that wants to display
commands live while also persisting clean output) involves the same
half-dozen chores every time:

- cancel via context **and** make the kill recursive across the child's
  whole process group (so `go run`'s compiled binary actually dies);
- strip ANSI escapes from stored output without losing them for live
  display;
- replace stray binary bytes (UTF-8 islands of `\x01`/`\x02` when a
  program misdetects a pipe as a terminal) so the output is safe to feed
  into an LLM;
- truncate runaway output, but keep the tail and spill the full version
  to a temp file;
- force colors via `CLICOLOR_FORCE`/`FORCE_COLOR` when a UI is watching,
  even though stdout is a pipe.

`pinexec` does all of that behind one call:

```go
res, err := pinexec.Execute(ctx, "make test", func(s string) {
    ui.Write(s) // live, ANSI preserved
})
// res.Output     — ANSI-stripped, sanitized, tail-truncated
// res.ExitCode   — 0 on success, -1 on cancel, …
// res.Cancelled  — true if ctx fired
// res.Truncated  — true if Output was tail-truncated
// res.FullOutputPath — non-empty if a temp-file spill was created
```

## Install

```bash
go get github.com/kfet/pinexec
```

Requires Go 1.21+. Zero external dependencies.

## Scope

Intentionally small. `pinexec` is a runner shaped for AI coding agents,
not a general process-management library. It does **not**:

- parse, lint, or rewrite shell input (use [mvdan.cc/sh](https://mvdan.cc/sh) for that);
- offer a fluent pipeline DSL (see [bitfield/script](https://github.com/bitfield/script));
- expose every `os/exec.Cmd` knob (use `os/exec` directly).

What it does cover, and what makes it different from the alternatives,
is the combination above: dual-output sandboxed `$SHELL -c` with
cross-platform pgroup-kill, ANSI/binary sanitization, line+byte tail
truncation, and live raw-chunk callbacks.

## License

MIT — see [LICENSE](LICENSE).
