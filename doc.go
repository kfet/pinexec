// Package pinexec runs shell commands with cancellation, output sanitization,
// and live streaming.
//
// pinexec is a sandboxed "$SHELL -c" runner shaped for AI coding agents:
//
//   - Combined stdout+stderr capture.
//   - Cross-platform cancellation that kills the entire process group
//     (so go run's compiled binary, npm spawn, make recipes, etc. all die
//     when ctx is cancelled — not just the leader).
//   - Dual output: live raw chunks (ANSI preserved) via an optional
//     callback for UIs, plus a final ANSI-stripped, binary-sanitized
//     output for LLM context.
//   - Line + byte truncation with tail-keep; full output spills to a
//     temp file when it exceeds the in-memory threshold.
//   - Color env injection (CLICOLOR/CLICOLOR_FORCE/FORCE_COLOR) when a
//     live callback is provided, so CLIs that gate ANSI on TTY detection
//     still emit colors.
//
// The headline API is [Execute]. The truncation and ANSI helpers
// ([TruncateHead], [TruncateTail], [StripAnsi], [AppendColorEnv]) are
// exported for callers that want to apply the same shape to output they
// produce by other means.
//
// pinexec is dependency-free and Go 1.21+.
package pinexec
