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

const deleteUsage = "usage: sonora delete routes/<route-id> [--json] [--verbose] [--hub-url URL]"

// RunDelete implements `sonora delete routes/<route-id>` (and its `sonora
// stop routes/<route-id>` alias): it defines and parses this command's
// flags, resolves the hub URL, deletes the named route via the hub, and
// renders the result to stdout. Any failure is reported on stderr, never
// stdout, so scripts piping stdout never see error text. It returns the
// process exit code per the exit code classes in data-model.md's exit code
// table.
func RunDelete(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("delete routes", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, deleteUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")

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
		fmt.Fprintln(stderr, deleteUsage)
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
	if err := hub.DeleteRoute(context.Background(), client, baseURL, routeID); err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	message := fmt.Sprintf("Stopped and removed routes/%s.", routeID)
	if *jsonOut {
		fmt.Fprint(stdout, render.RenderRouteDeletedJSON(routeID, message))
	} else {
		fmt.Fprint(stdout, render.RenderRouteDeletedYAML(routeID, message))
	}
	return 0
}
