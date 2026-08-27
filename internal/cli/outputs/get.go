package outputs

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

const getUsage = "usage: sonora get outputs/<output-id> [--json] [--verbose] [--hub-url URL]"

// RunGet implements `sonora get outputs/<output-id>`: it defines and parses
// this command's flags, resolves the hub URL, fetches the single named
// output from the hub, and renders it to stdout. Any failure is reported on
// stderr, never stdout, so scripts piping stdout never see error text. It
// returns the process exit code per the exit code classes in
// specs/002-outputs-get/data-model.md's exit code table.
func RunGet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("outputs get", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, getUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")

	// flag.Parse stops at the first non-flag argument, so a positional
	// <output-id> preceding a flag (per the documented invocation shape)
	// would otherwise be mistaken for the end of flags. Re-parse in a loop,
	// peeling off one positional argument at a time, so flags can appear
	// before or after the identifier.
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
	if len(positional) != 1 {
		fmt.Fprintln(stderr, getUsage)
		if len(positional) == 0 {
			fmt.Fprintf(stderr, "error: missing required argument: <output-id>\n")
		} else {
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[1:])
		}
		return hub.ClassUsage.ExitCode()
	}
	outputID := positional[0]

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	output, err := hub.GetOutput(context.Background(), client, baseURL, outputID)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderOutputJSON(*output))
	} else {
		fmt.Fprint(stdout, render.RenderOutputYAML(*output))
	}
	return 0
}
