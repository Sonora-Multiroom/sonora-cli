// Command sonora is the Sonora Multiroom CLI entrypoint.
package main

import (
	"fmt"
	"io"
	"os"

	"sonora-cli/internal/cli/groups"
	"sonora-cli/internal/cli/inputs"
	"sonora-cli/internal/cli/outputs"
	"sonora-cli/internal/cli/routes"
	"sonora-cli/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches `sonora <noun> <verb> [flags]` to the matching command
// handler. Unrecognized noun/verb is a usage error (exit 2).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) >= 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Fprintln(stdout, version.Version)
		return 0
	}

	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: sonora <noun> <verb> [flags]")
		return 2
	}

	noun, verb, rest := args[0], args[1], args[2:]
	switch noun {
	case "inputs":
		switch verb {
		case "list":
			return inputs.RunList(rest, stdout, stderr)
		case "get":
			return inputs.RunGet(rest, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "sonora: unknown verb %q for %q\n", verb, noun)
			return 2
		}
	case "outputs":
		switch verb {
		case "list":
			return outputs.RunList(rest, stdout, stderr)
		case "get":
			return outputs.RunGet(rest, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "sonora: unknown verb %q for %q\n", verb, noun)
			return 2
		}
	case "routes":
		switch verb {
		case "list":
			return routes.RunList(rest, stdout, stderr)
		case "get":
			return routes.RunGet(rest, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "sonora: unknown verb %q for %q\n", verb, noun)
			return 2
		}
	case "groups":
		switch verb {
		case "list":
			return groups.RunList(rest, stdout, stderr)
		case "get":
			return groups.RunGet(rest, stdout, stderr)
		default:
			fmt.Fprintf(stderr, "sonora: unknown verb %q for %q\n", verb, noun)
			return 2
		}
	default:
		fmt.Fprintf(stderr, "sonora: unknown command %q\n", noun)
		return 2
	}
}
