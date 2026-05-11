package pinexec

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Options configures a single [Execute] call.
//
// The zero value is valid: no streaming callback, default truncation,
// no color env injection.
type Options struct {
	// OnChunk, when non-nil, is invoked with each raw output chunk as it
	// arrives. ANSI escape sequences are preserved so the chunks are
	// suitable for live display. The chunk is the combined stdout+stderr
	// stream of the running command; ordering reflects arrival order, not
	// stream of origin.
	//
	// When OnChunk is non-nil, color-forcing environment variables
	// (see [AppendColorEnv]) are injected so tools that gate ANSI output
	// on TTY detection still emit colors.
	//
	// OnChunk is called serially from a single goroutine and may block
	// briefly without affecting correctness, but a slow callback will
	// back-pressure the underlying read loop and may cause the child
	// process to block on output. Keep it fast; copy to a channel if
	// you need to do real work.
	OnChunk func(chunk string)
}

// Result holds the outcome of an [Execute] call.
type Result struct {
	// Output is the combined stdout+stderr of the command, with ANSI
	// escape sequences stripped, binary bytes replaced with '?', and
	// '\r' removed. If the raw output exceeded the truncation limits
	// (see [TruncationOptions]) only the tail is retained and
	// [Result.Truncated] is true.
	Output string

	// ExitCode is the process exit code, or -1 if the command was
	// cancelled via context, killed, or did not produce an exit
	// status for any other reason.
	ExitCode int

	// Cancelled is true if the call's context was cancelled before
	// the command finished.
	Cancelled bool

	// Truncated is true if [Result.Output] was truncated.
	Truncated bool

	// FullOutputPath, when non-empty, is the path to a temp file
	// containing the (sanitized, ANSI-stripped) full output of the
	// command. The file is created lazily once total output exceeds
	// [DefaultMaxBytes] and is not removed by pinexec; the caller
	// owns its lifecycle.
	FullOutputPath string
}

// sanitizeBinaryOutput replaces non-printable bytes with '?'. Common
// whitespace ('\n', '\t', '\r') and ESC (used in ANSI escape sequences)
// are preserved, as is anything outside ASCII (assumed to be UTF-8
// continuation bytes or printable Unicode).
func sanitizeBinaryOutput(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\t' || r == '\r' || r == '\x1b' || (r >= 32 && r < 127) || r >= 128 {
			b.WriteRune(r)
		} else {
			b.WriteRune('?')
		}
	}
	return b.String()
}

// Execute runs command via $SHELL -c (falling back to /bin/sh), capturing
// combined stdout+stderr.
//
// The command runs in its own process group on Unix so cancelling ctx
// kills the entire group, not just the shell — this is important for
// commands like `go run` that spawn a compiled binary the shell does
// not directly track. On Windows the standard process-tree termination
// from [exec.CommandContext] is used.
//
// Output is sanitized (binary bytes replaced, '\r' stripped) and
// ANSI-stripped before being stored in [Result.Output]. Live raw output
// (ANSI preserved) is streamed via [Options.OnChunk] when set.
//
// If total output exceeds [DefaultMaxBytes], the full sanitized output
// is also streamed to a temp file whose path is returned in
// [Result.FullOutputPath]. The in-memory output is kept to roughly
// 2×[DefaultMaxBytes] via a rolling window, then further trimmed by
// [TruncateTail] to [DefaultMaxBytes] / [DefaultMaxLines] for the
// final [Result.Output].
//
// Execute returns a non-nil error only if the child process could not
// be started or its pipes could not be opened. A non-zero exit status
// is reported via [Result.ExitCode], not as an error.
//
// Execute is safe to call concurrently.
func Execute(ctx context.Context, command string, opts *Options) (Result, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Env = os.Environ()

	// Run in its own process group so cancellation kills all children.
	setProcGroup(cmd)

	// When streaming for display, inject env vars that tell CLI tools to
	// emit ANSI colors even when stdout is not a TTY.
	if opts != nil && opts.OnChunk != nil {
		cmd.Env = AppendColorEnv(cmd.Env)
	}

	stdoutPipe := mustStdoutPipe(cmd)
	stderrPipe := mustStderrPipe(cmd)

	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("start command: %w", err)
	}

	maxOutputBytes := DefaultMaxBytes * 2

	var (
		mu           sync.Mutex
		outputChunks []string
		outputBytes  int
		totalBytes   int
		tempFilePath string
		tempFile     *os.File
		onChunk      func(string)
	)
	if opts != nil {
		onChunk = opts.OnChunk
	}

	handleData := func(data []byte) {
		mu.Lock()
		defer mu.Unlock()

		totalBytes += len(data)

		// Sanitize binary but preserve ANSI for display.
		rawText := sanitizeBinaryOutput(string(data))
		rawText = strings.ReplaceAll(rawText, "\r", "")

		// Stream raw (ANSI-preserved) output to display callback.
		if onChunk != nil {
			onChunk(rawText)
		}

		// Strip ANSI for stored output.
		text := StripAnsi(rawText)

		// Start temp file once total output crosses the threshold.
		if totalBytes > DefaultMaxBytes && tempFile == nil {
			id := make([]byte, 8)
			// crypto/rand.Read never fails in practice on supported
			// platforms; on the off chance it does we accept an
			// all-zero id rather than aborting the run.
			_, _ = rand.Read(id)
			tempFilePath = filepath.Join(os.TempDir(), "pinexec-"+hex.EncodeToString(id)+".log")
			if f, err := os.Create(tempFilePath); err == nil {
				tempFile = f
				// Flush already-buffered chunks to the temp file.
				for _, chunk := range outputChunks {
					_, _ = tempFile.WriteString(chunk)
				}
			} else {
				// Couldn't create the file; surface nothing rather
				// than crashing the run.
				tempFilePath = ""
			}
		}

		if tempFile != nil {
			_, _ = tempFile.WriteString(text)
		}

		// Keep a rolling window of recent chunks for in-memory output.
		outputChunks = append(outputChunks, text)
		outputBytes += len(text)
		for outputBytes > maxOutputBytes && len(outputChunks) > 1 {
			removed := outputChunks[0]
			outputChunks = outputChunks[1:]
			outputBytes -= len(removed)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	readStream := func(r io.Reader) {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				handleData(buf[:n])
			}
			if readErr != nil {
				return
			}
		}
	}
	go readStream(stdoutPipe)
	go readStream(stderrPipe)
	wg.Wait()

	waitErr := cmd.Wait()
	if tempFile != nil {
		_ = tempFile.Close()
	}

	mu.Lock()
	fullOutput := strings.Join(outputChunks, "")
	mu.Unlock()

	truncationResult := TruncateTail(fullOutput, TruncationOptions{})
	cancelled := ctx.Err() != nil

	exitCode := -1
	if !cancelled {
		if waitErr == nil {
			exitCode = 0
		} else {
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				exitCode = exitErr.ExitCode()
			}
		}
	}

	output := fullOutput
	if truncationResult.Truncated {
		output = truncationResult.Content
	}

	return Result{
		Output:         output,
		ExitCode:       exitCode,
		Cancelled:      cancelled,
		Truncated:      truncationResult.Truncated,
		FullOutputPath: tempFilePath,
	}, nil
}
