package routes

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

const stopAllUsage = "usage: sonora stop routes [flags]"

// RunStopAll implements `sonora stop routes` (no id): it defines and parses
// this command's flags, resolves the hub URL, stops every active route
// system-wide via the hub, and renders the result to stdout. Any failure is
// reported on stderr, never stdout, so scripts piping stdout never see error
// text.
func RunStopAll(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("stop routes", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, stopAllUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, stopAllUsage)
		return 0
	}

	if err := fs.Parse(args); err != nil {
		return hub.ClassUsage.ExitCode()
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, stopAllUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	result, err := hub.StopAllRoutes(context.Background(), client, baseURL)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderBulkStopJSON(*result))
	} else {
		fmt.Fprint(stdout, render.RenderBulkStopYAML(*result))
	}
	return 0
}
