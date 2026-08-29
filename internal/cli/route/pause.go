package route

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

// RunPause implements `sonora pause routes/<route-id>`.
func RunPause(args []string, stdout, stderr io.Writer) int {
	return runSetPause("pause", true, args, stdout, stderr)
}

// RunResume implements `sonora resume routes/<route-id>`.
func RunResume(args []string, stdout, stderr io.Writer) int {
	return runSetPause("resume", false, args, stdout, stderr)
}

// runSetPause implements the shared body of RunPause/RunResume: it defines
// and parses this command's flags, resolves the hub URL, sets the named
// route's paused state via the hub (idempotent — pausing an already-paused
// route, or resuming an already-active one, still succeeds), and renders
// the updated route to stdout. Any failure is reported on stderr, never
// stdout, so scripts piping stdout never see error text. It returns the
// process exit code per the exit code classes in data-model.md's exit code
// table.
func runSetPause(verb string, paused bool, args []string, stdout, stderr io.Writer) int {
	usage := fmt.Sprintf("usage: sonora %s routes/<route-id> [flags]", verb)

	fs := flag.NewFlagSet(verb+" routes", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, usage)
		return 0
	}

	// flag.Parse stops at the first non-flag argument, so a positional
	// <route-id> preceding a flag (per the documented invocation shape)
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
		fmt.Fprintln(stderr, usage)
		if len(positional) == 0 {
			fmt.Fprintf(stderr, "error: missing required argument: <route-id>\n")
		} else {
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[1:])
		}
		return hub.ClassUsage.ExitCode()
	}
	routeID := positional[0]

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	updated, err := hub.SetPauseState(context.Background(), client, baseURL, routeID, paused)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	verbPast := "paused"
	if !paused {
		verbPast = "resumed"
	}
	message := fmt.Sprintf("Route %s %s.", routeID, verbPast)
	if *jsonOut {
		fmt.Fprint(stdout, render.RenderRoutePauseJSON(*updated, message))
	} else {
		fmt.Fprint(stdout, render.RenderRoutePauseYAML(*updated, message))
	}
	return 0
}
