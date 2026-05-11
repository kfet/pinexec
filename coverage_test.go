package pinexec

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf8"
)

// SHELL="" → defaults to /bin/sh
func TestExecute_DefaultsToBinSh(t *testing.T) {
	t.Setenv("SHELL", "")
	ctx := context.Background()
	res, err := Execute(ctx, "echo ok", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d", res.ExitCode)
	}
	if !strings.Contains(res.Output, "ok") {
		t.Errorf("output = %q", res.Output)
	}
}

// SHELL pointing at a non-existent binary causes cmd.Start to fail
func TestExecute_StartError(t *testing.T) {
	t.Setenv("SHELL", "/nonexistent/shell/binary/pinexec-test")
	_, err := Execute(context.Background(), "echo hi", nil)
	if err == nil {
		t.Fatal("expected start error, got nil")
	}
	if !strings.Contains(err.Error(), "start command") {
		t.Errorf("error = %v, want wrapped 'start command'", err)
	}
}

// TMPDIR pointing at a non-writable path → os.Create fails inside
// handleData. The run still completes; FullOutputPath is empty.
func TestExecute_TempFileCreateFailure(t *testing.T) {
	t.Setenv("TMPDIR", "/nonexistent/dir/for/pinexec-tempfile-test")
	// Generate output > DefaultMaxBytes (50KB) so the temp-file branch fires.
	cmd := "for i in $(seq 1 5000); do echo 'line '$i' with padding xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'; done"
	res, err := Execute(context.Background(), cmd, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d", res.ExitCode)
	}
	if res.FullOutputPath != "" {
		t.Errorf("FullOutputPath = %q, expected empty after create failure", res.FullOutputPath)
	}
}

// procgroup cancel-before-start branch (procgroup_unix.go: cmd.Process == nil).
func TestSetProcGroup_NilProcessCancel(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "true")
	setProcGroup(cmd)
	// Cancel before Start → cmd.Process is nil → must return nil.
	if err := cmd.Cancel(); err != nil {
		t.Errorf("Cancel with nil Process = %v, want nil", err)
	}
}

// (UTF-8 boundary skip is covered by truncate_test.go but the existing
// case starts on a valid rune boundary, so the in-loop advance is dead.
// This case lands inside a multi-byte rune to force the loop body.)
func TestTruncateStringToBytesFromEnd_InsideMultiByte(t *testing.T) {
	// "日" = 3 bytes (0xE6, 0x97, 0xA5). maxBytes=2 → start=1 which is
	// inside the rune; loop advances past 0x97 and 0xA5 to len(s)=3,
	// returning the empty string.
	s := "日"
	out := truncateStringToBytesFromEnd(s, 2)
	if !utf8.ValidString(out) {
		t.Errorf("invalid UTF-8: %q", out)
	}
	if out != "" {
		t.Errorf("got %q, want \"\"", out)
	}
}
