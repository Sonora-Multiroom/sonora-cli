// Command sonora is the Sonora Multiroom CLI entrypoint.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"sonora-cli/internal/cli/groups"
	"sonora-cli/internal/cli/inputs"
	"sonora-cli/internal/cli/outputs"
	"sonora-cli/internal/cli/play"
	"sonora-cli/internal/cli/respath"
	"sonora-cli/internal/cli/routes"
	"sonora-cli/internal/version"
)

const helpText = `Usage: sonora <verb> <resource>[/<id>] [flags]

Commands:
  get <resource>[/<id>]    Fetch a collection, or a single item by id
  list <resource>          Fetch a collection (synonym of 'get <resource>')
  play <uri> <outputs|groups>/<id>
                            Instant playback of an audio URI to an output or group
  help                     Show this help

Resources: inputs (in), outputs (out), groups (gr), routes (rt)

Global flags:
  --json               Output strict JSON instead of the default YAML
  --hub-url URL        Override the hub base URL
  --verbose            Print underlying error detail on failure
  --version, -v        Print the CLI version

Examples:
  sonora get outputs --include-disabled
  sonora get routes --status active
  sonora get groups/<id> --json
  sonora list outputs
  sonora play "https://stream.example.com/live.mp3" outputs/office-speaker --volume 40

Run 'sonora get <resource> --help' or 'sonora list <resource> --help' for the full flag
reference of any command.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches `sonora <verb> <resource>[/<id>] [flags]` to the matching
// command handler. An unrecognized verb (including any pre-refactor
// noun-first invocation) is a usage error (exit 2).
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

	switch args[0] {
	case "get", "list":
		return dispatchGetList(args[0], args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "sonora: unknown command %q\n", args[0])
		return 2
	}
}

// dispatchGetList resolves the resource-path argument following `get`/`list`
// via respath and translates it into the matching resource package's
// existing RunList/RunGet call (research.md §2). A `list` given a resource
// path that includes an id is a usage error (FR-003); a `get`/`list` with no
// resource argument at all gets a distinct message enumerating the valid
// resource names rather than the generic unrecognized-resource error
// (FR-006a).
func dispatchGetList(verb string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		if verb == "list" {
			fmt.Fprintln(stderr, "usage: sonora list <resource> [flags]")
		} else {
			fmt.Fprintln(stderr, "usage: sonora get <resource>[/<id>] [flags]")
		}
		fmt.Fprintf(stderr, "error: missing resource argument; valid resources: %s\n", strings.Join(respath.Names(), ", "))
		return 2
	}

	path, err := respath.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return 2
	}
	if verb == "list" && path.ID != "" {
		fmt.Fprintln(stderr, "usage: sonora list <resource> [flags]")
		fmt.Fprintln(stderr, "error: list does not accept an id; use 'sonora get <resource>/<id>' instead")
		return 2
	}

	rest := args[1:]
	callArgs := rest
	if path.ID != "" {
		callArgs = append([]string{path.ID}, rest...)
	}

	switch path.Kind {
	case respath.Inputs:
		if path.ID != "" {
			return inputs.RunGet(callArgs, stdout, stderr)
		}
		return inputs.RunList(callArgs, stdout, stderr)
	case respath.Outputs:
		if path.ID != "" {
			return outputs.RunGet(callArgs, stdout, stderr)
		}
		return outputs.RunList(callArgs, stdout, stderr)
	case respath.Groups:
		if path.ID != "" {
			return groups.RunGet(callArgs, stdout, stderr)
		}
		return groups.RunList(callArgs, stdout, stderr)
	default: // respath.Routes
		if path.ID != "" {
			return routes.RunGet(callArgs, stdout, stderr)
		}
		return routes.RunList(callArgs, stdout, stderr)
	}
}
