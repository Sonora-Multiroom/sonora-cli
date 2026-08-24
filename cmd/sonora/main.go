// Command sonora is the Sonora Multiroom CLI entrypoint.
package main

import (
	"fmt"
	"io"
	"os"

	"sonora-cli/internal/cli/outputs"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches `sonora <noun> <verb> [flags]` to the matching command
// handler. Unrecognized noun/verb is a usage error (exit 2).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: sonora <noun> <verb> [flags]")
		return 2
	}

	noun, verb, rest := args[0], args[1], args[2:]
	switch noun {
	case "outputs":
		switch verb {
		case "list":
			return outputs.Run(rest, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "sonora: unknown verb %q for %q\n", verb, noun)
			return 2
		}
	default:
		fmt.Fprintf(stderr, "sonora: unknown command %q\n", noun)
		return 2
	}
}
