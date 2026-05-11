package pinexec_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/kfet/pinexec"
)

// Basic usage: capture combined stdout+stderr and the exit code.
func ExampleExecute() {
	ctx := context.Background()
	res, err := pinexec.Execute(ctx, "echo hello; echo world", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("exit:", res.ExitCode)
	fmt.Println(strings.TrimSpace(res.Output))
	// Output:
	// exit: 0
	// hello
	// world
}

// Stream live chunks (ANSI preserved) while the command runs. The final
// [Result.Output] is still ANSI-stripped for downstream consumers.
func ExampleExecute_streaming() {
	ctx := context.Background()
	var live strings.Builder
	res, _ := pinexec.Execute(ctx, "echo hi", &pinexec.Options{
		OnChunk: func(chunk string) {
			live.WriteString(chunk)
		},
	})
	fmt.Println("live:", strings.TrimSpace(live.String()))
	fmt.Println("stored:", strings.TrimSpace(res.Output))
	// Output:
	// live: hi
	// stored: hi
}
