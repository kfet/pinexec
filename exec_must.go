package pinexec

import (
	"io"
	"os/exec"
)

// mustStdoutPipe wraps [exec.Cmd.StdoutPipe] and panics on error.
//
// StdoutPipe is documented to fail only when cmd.Stdout is already set
// or cmd.Start has already been called. pinexec never assigns
// cmd.Stdout and only calls this before cmd.Start, so a failure here
// is a programmer error in pinexec, not a runtime condition the
// caller can react to.
func mustStdoutPipe(cmd *exec.Cmd) io.ReadCloser {
	p, err := cmd.StdoutPipe()
	if err != nil {
		panic("pinexec: StdoutPipe failed; this is a bug: " + err.Error())
	}
	return p
}

// mustStderrPipe is the [Cmd.StderrPipe] counterpart of
// [mustStdoutPipe]; same panic-on-error rationale.
func mustStderrPipe(cmd *exec.Cmd) io.ReadCloser {
	p, err := cmd.StderrPipe()
	if err != nil {
		panic("pinexec: StderrPipe failed; this is a bug: " + err.Error())
	}
	return p
}
