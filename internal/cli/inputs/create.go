package inputs

import (
	"context"
	"flag"
	"fmt"
	"io"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const createUsage = "usage: sonora create inputs/<input-id> <uri> --display-name <name> [flags]"

// RunCreate implements `sonora create inputs/<input-id> <uri> --display-name
// <name> [--auto-remove] [--disabled]`: it defines and parses this command's
// flags, resolves the hub URL, registers a new ephemeral input with the hub,
// and renders the result to stdout. Any failure is reported on stderr, never
// stdout, so scripts piping stdout never see error text. It returns the
// process exit code per the exit code classes in data-model.md's exit code
// table.
func RunCreate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("inputs create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, createUsage)

	displayName := fs.String("display-name", "", "human-readable display `name` (required)")
	autoRemove := fs.Bool("auto-remove", false, "auto-remove the input when its route is stopped")
	disabled := fs.Bool("disabled", false, "create the input in a disabled state")
	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, createUsage)
		return 0
	}

	// flag.Parse stops at the first non-flag argument, so the positional
	// <input-id>/<uri> preceding a flag (per the documented invocation
	// shape) would otherwise be mistaken for the end of flags. Re-parse in a
	// loop, peeling off one positional argument at a time, so flags can
	// appear before, between, or after the two identifiers (same pattern
	// route.Run uses).
	var positional []string
	remaining := args
	for {
		if err := fs.Parse(remaining); err != nil {
			return hub.ClassUsage.ExitCode()
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		remaining = rest[1:]
	}
	if len(positional) != 2 {
		fmt.Fprintln(stderr, createUsage)
		switch {
		case len(positional) == 0:
			fmt.Fprintln(stderr, "error: missing required argument: <input-id>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <uri>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return hub.ClassUsage.ExitCode()
	}
	inputID, uri := positional[0], positional[1]

	if *displayName == "" {
		fmt.Fprintln(stderr, createUsage)
		fmt.Fprintln(stderr, "error: missing required flag: --display-name")
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	req := hub.CreateInputRequest{
		InputID:     inputID,
		DisplayName: *displayName,
		URI:         uri,
		Enabled:     !*disabled,
		AutoRemove:  *autoRemove,
	}
	created, err := hub.CreateInput(context.Background(), client, baseURL, req)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderInputJSON(*created))
	} else {
		fmt.Fprint(stdout, render.RenderInputYAML(*created))
	}
	return 0
}
