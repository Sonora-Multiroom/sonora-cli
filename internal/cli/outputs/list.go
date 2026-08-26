// Package outputs implements the `sonora outputs` commands.
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

const listUsage = "usage: sonora outputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]"

// RunList implements `sonora outputs list`: it defines and parses this
// command's flags, resolves the hub URL, fetches outputs from the hub, and
// renders them to stdout. Any failure is reported on stderr, never stdout,
// so scripts piping stdout never see error text. It returns the process
// exit code per the exit code classes in research.md §6.
func RunList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("outputs list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, listUsage)

	includeDisabled := fs.Bool("include-disabled", false, "include disabled outputs in the results")
	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")

	if err := fs.Parse(args); err != nil {
		return hub.ClassUsage.ExitCode()
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, listUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", fs.Args())
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	outputs, err := hub.ListOutputs(context.Background(), client, baseURL, *includeDisabled)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	var rendered string
	if *jsonOut {
		rendered = render.RenderJSON(outputs)
	} else {
		rendered = render.RenderYAML(outputs)
	}
	fmt.Fprint(stdout, rendered)
	return 0
}
