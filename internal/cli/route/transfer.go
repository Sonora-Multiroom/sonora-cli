package route

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/cli/respath"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const transferUsage = "usage: sonora transfer routes/<route-id> <outputs|groups>/<target-id> [flags]"

// RunTransfer implements `sonora transfer routes/<route-id>
// <outputs|groups>/<target-id>`: it defines and parses this command's
// flags, validates both resource paths (the first must be routes/<id>, the
// second outputs/<id> or groups/<id> — no auto-detect, mirroring Run's
// input/target validation), verifies the target already exists, calls the
// hub to transfer the route, and renders the result to stdout. The hub
// replaces the old route with a new one, so the rendered routeId is the
// *new* route's id. Any failure is reported on stderr, never stdout, so
// scripts piping stdout never see error text. It returns the process exit
// code per data-model.md's exit code table.
func RunTransfer(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("transfer", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, transferUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, transferUsage)
		return 0
	}

	// flag.Parse stops at the first non-flag argument, so <route-path>/
	// <target-path> preceding a flag (per the documented invocation shape)
	// would otherwise be mistaken for the end of flags. Re-parse in a loop,
	// peeling off one positional argument at a time, so flags can appear
	// before, between, or after the two identifiers (same pattern Run
	// uses).
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
		fmt.Fprintln(stderr, transferUsage)
		switch {
		case len(positional) == 0:
			fmt.Fprintln(stderr, "error: missing required argument: <route-path>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <target-path>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return hub.ClassUsage.ExitCode()
	}
	routeArg, targetArg := positional[0], positional[1]

	routePath, err := respath.Parse(routeArg)
	if err != nil {
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}
	if routePath.Kind != respath.Routes {
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintf(stderr, "error: route path must start with routes/ or rt/, got %q\n", routeArg)
		return hub.ClassUsage.ExitCode()
	}
	if routePath.ID == "" {
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintln(stderr, "error: route path must include an id")
		return hub.ClassUsage.ExitCode()
	}
	routeID := routePath.ID

	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}
	var targetType string
	switch targetPath.Kind {
	case respath.Outputs:
		targetType = "SINGLE_OUTPUT"
	case respath.Groups:
		targetType = "OUTPUT_GROUP"
	default:
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintf(stderr, "error: transfer target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return hub.ClassUsage.ExitCode()
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, transferUsage)
		fmt.Fprintln(stderr, "error: target path must include an id")
		return hub.ClassUsage.ExitCode()
	}
	targetID := targetPath.ID

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	ctx := context.Background()
	client := hub.NewClient()

	if err := hub.ResolveTarget(ctx, client, baseURL, targetID, targetType); err != nil {
		var notFoundErr *hub.NotFoundError
		if errors.As(err, &notFoundErr) {
			return reportNotFound(stderr, notFoundErr, baseURL, *verbose, hub.ClassTargetNotFound)
		}
		return reportError(stderr, err, baseURL, *verbose)
	}

	req := hub.TransferRequest{TargetID: targetID, TargetType: targetType}
	transferred, err := hub.TransferRoute(ctx, client, baseURL, routeID, req)
	if err != nil {
		var notFoundErr *hub.NotFoundError
		if errors.As(err, &notFoundErr) {
			return reportNotFound(stderr, notFoundErr, baseURL, *verbose, hub.ClassNotFound)
		}
		return reportError(stderr, err, baseURL, *verbose)
	}

	message := fmt.Sprintf("Transferred %s to %s.", routeArg, targetArg)
	if *jsonOut {
		fmt.Fprint(stdout, render.RenderRouteCreatedJSON(*transferred, message))
	} else {
		fmt.Fprint(stdout, render.RenderRouteCreatedYAML(*transferred, message))
	}
	return 0
}
