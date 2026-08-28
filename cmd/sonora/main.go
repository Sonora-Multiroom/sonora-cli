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
	"sonora-cli/internal/cli/route"
	"sonora-cli/internal/cli/routes"
	"sonora-cli/internal/version"
)

const helpText = `Usage: sonora <verb> <resource>[/<id>] [flags]

Commands:
  get <resource>[/<id>]    Fetch a collection, or a single item by id
  list <resource>          Fetch a collection (synonym of 'get <resource>')
  play <uri> <outputs|groups>/<id>
                            Instant playback of an audio URI to an output or group
  route inputs/<id> <outputs|groups>/<id>
                            Connect an existing input to an existing output or group
  transfer routes/<id> <outputs|groups>/<id>
                            Move an active route's playback to a new output or group
  delete routes/<id>       Stop and remove a route
  stop routes/<id>         Alias of 'delete routes/<id>'
  enable inputs/<id>       Enable a disabled input
  disable inputs/<id>      Disable an input
  set outputs/<id> volume <0-100>
                            Set an output's volume level
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
  sonora route inputs/spotify-1 outputs/office-speaker
  sonora transfer routes/<route-id> outputs/bedroom-speaker
  sonora delete routes/<route-id>
  sonora enable inputs/<input-id>
  sonora disable inputs/<input-id>
  sonora set outputs/<output-id> volume 40

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
	if args[0] == "route" {
		return route.Run(args[1:], stdout, stderr)
	}
	if args[0] == "transfer" {
		return route.RunTransfer(args[1:], stdout, stderr)
	}

	switch args[0] {
	case "get", "list":
		return dispatchGetList(args[0], args[1:], stdout, stderr)
	case "delete", "stop":
		return dispatchDelete(args[0], args[1:], stdout, stderr)
	case "enable", "disable":
		return dispatchEnabled(args[0], args[1:], stdout, stderr)
	case "set":
		return dispatchSet(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "sonora: unknown command %q\n", args[0])
		return 2
	}
}

// dispatchDelete resolves the resource-path argument following `delete`/
// `stop` via respath and calls routes.RunDelete — the only resource `delete`
// currently supports is `routes` (`stop` is an exact alias of `delete`, per
// docs/cli-command-landscape.md). Any other resource is a usage error.
func dispatchDelete(verb string, args []string, stdout, stderr io.Writer) int {
	usage := fmt.Sprintf("usage: sonora %s routes/<route-id> [flags]", verb)
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing resource argument")
		return 2
	}

	path, err := respath.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return 2
	}
	if path.Kind != respath.Routes {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: %s does not support %s; only routes/<route-id> is supported\n", verb, path.Kind)
		return 2
	}
	if path.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <route-id>")
		return 2
	}

	callArgs := append([]string{path.ID}, args[1:]...)
	return routes.RunDelete(callArgs, stdout, stderr)
}

// dispatchEnabled resolves the resource-path argument following `enable`/
// `disable` via respath and calls inputs.RunEnable/RunDisable — the only
// resource `enable`/`disable` currently support is `inputs`. Any other
// resource is a usage error.
func dispatchEnabled(verb string, args []string, stdout, stderr io.Writer) int {
	usage := fmt.Sprintf("usage: sonora %s inputs/<input-id> [flags]", verb)
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing resource argument")
		return 2
	}

	path, err := respath.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return 2
	}
	if path.Kind != respath.Inputs {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: %s does not support %s; only inputs/<input-id> is supported\n", verb, path.Kind)
		return 2
	}
	if path.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <input-id>")
		return 2
	}

	callArgs := append([]string{path.ID}, args[1:]...)
	if verb == "enable" {
		return inputs.RunEnable(callArgs, stdout, stderr)
	}
	return inputs.RunDisable(callArgs, stdout, stderr)
}

// dispatchSet resolves the resource-path argument following `set` via
// respath and calls outputs.RunSetVolume — the only resource/attribute
// combination `set` currently supports is `outputs`/`volume` (the literal
// `volume` word and its `<0-100>` value are validated by RunSetVolume
// itself). Any other resource is a usage error.
func dispatchSet(args []string, stdout, stderr io.Writer) int {
	usage := "usage: sonora set outputs/<output-id> volume <0-100> [flags]"
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing resource argument")
		return 2
	}

	path, err := respath.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return 2
	}
	if path.Kind != respath.Outputs {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: set does not support %s; only outputs/<output-id> volume <0-100> is supported\n", path.Kind)
		return 2
	}
	if path.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <output-id>")
		return 2
	}

	callArgs := append([]string{path.ID}, args[1:]...)
	return outputs.RunSetVolume(callArgs, stdout, stderr)
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
