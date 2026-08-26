// Command sonora is the Sonora Multiroom CLI entrypoint.
package main

import (
	"fmt"
	"io"
	"os"

	"sonora-cli/internal/cli/groups"
	"sonora-cli/internal/cli/inputs"
	"sonora-cli/internal/cli/outputs"
	"sonora-cli/internal/cli/play"
	"sonora-cli/internal/cli/routes"
	"sonora-cli/internal/version"
)

const helpText = `Usage: sonora <noun> <verb> [flags]

Commands:
  outputs list|get   Manage multiroom outputs
  inputs  list|get   Manage multiroom inputs
  routes  list|get   Manage multiroom routes
  groups  list|get   Manage multiroom output groups
  play <uri> <target-id>
                      Instant playback of an audio URI to an output or group
  help                Show this help

Global flags:
  --json               Output strict JSON instead of the default YAML
  --hub-url URL        Override the hub base URL
  --verbose            Print underlying error detail on failure
  --version, -v        Print the CLI version

Examples:
  sonora outputs list --include-disabled
  sonora routes list --status active
  sonora groups get <id> --json
  sonora play "https://stream.example.com/live.mp3" office-speaker --volume 40

Run 'sonora <noun> <verb> --help' for the full flag reference of any command.
`

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

	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	if args[0] == "play" {
		return play.Run(args[1:], stdout, stderr)
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
