// Package route implements `sonora route inputs/<id> <outputs|groups>/<id>`.
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

const usage = "usage: sonora route inputs/<input-id> <outputs|groups>/<target-id> [--json] [--verbose] [--hub-url URL]"

// Run implements `sonora route inputs/<id> <outputs|groups>/<id>`: it
// defines and parses this command's flags, validates both resource paths
// (the input path must be inputs/<id>, the target path outputs/<id> or
// groups/<id> — no auto-detect, per FR-002a/FR-002b), verifies the input
// and target each already exist, calls the hub to create a route between
// them, and renders the result to stdout. Any failure is reported on
// stderr, never stdout, so scripts piping stdout never see error text. It
// returns the process exit code per data-model.md's exit code table.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("route", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")

	// flag.Parse stops at the first non-flag argument, so <input-path>/
	// <target-path> preceding a flag (per the documented invocation shape)
	// would otherwise be mistaken for the end of flags. Re-parse in a loop,
	// peeling off one positional argument at a time, so flags can appear
	// before, between, or after the two identifiers (same pattern play.Run
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
		fmt.Fprintln(stderr, usage)
		switch {
		case len(positional) == 0:
			fmt.Fprintln(stderr, "error: missing required argument: <input-path>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <target-path>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return hub.ClassUsage.ExitCode()
	}
	inputArg, targetArg := positional[0], positional[1]

	inputPath, err := respath.Parse(inputArg)
	if err != nil {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}
	if inputPath.Kind != respath.Inputs {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: input path must start with inputs/ or in/, got %q\n", inputArg)
		return hub.ClassUsage.ExitCode()
	}
	if inputPath.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: input path must include an id")
		return hub.ClassUsage.ExitCode()
	}
	inputID := inputPath.ID

	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, usage)
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
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: route target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return hub.ClassUsage.ExitCode()
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, usage)
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

	if _, err := hub.GetInput(ctx, client, baseURL, inputID); err != nil {
		var notFoundErr *hub.NotFoundError
		if errors.As(err, &notFoundErr) {
			return reportNotFound(stderr, notFoundErr, baseURL, *verbose, hub.ClassInputNotFound)
		}
		return reportError(stderr, err, baseURL, *verbose)
	}

	if err := hub.ResolveTarget(ctx, client, baseURL, targetID, targetType); err != nil {
		var notFoundErr *hub.NotFoundError
		if errors.As(err, &notFoundErr) {
			return reportNotFound(stderr, notFoundErr, baseURL, *verbose, hub.ClassTargetNotFound)
		}
		return reportError(stderr, err, baseURL, *verbose)
	}

	req := hub.CreateRouteRequest{InputID: inputID, TargetID: targetID, TargetType: targetType}
	created, err := hub.CreateRoute(ctx, client, baseURL, req)
	if err != nil {
		var notFoundErr *hub.NotFoundError
		if errors.As(err, &notFoundErr) {
			return reportNotFound(stderr, notFoundErr, baseURL, *verbose, hub.ClassTargetNotFound)
		}
		return reportError(stderr, err, baseURL, *verbose)
	}

	message := fmt.Sprintf("Routed %s to %s.", inputArg, targetArg)
	if *jsonOut {
		fmt.Fprint(stdout, render.RenderRouteCreatedJSON(*created, message))
	} else {
		fmt.Fprint(stdout, render.RenderRouteCreatedYAML(*created, message))
	}
	return 0
}

func reportNotFound(stderr io.Writer, notFoundErr *hub.NotFoundError, baseURL string, verbose bool, class hub.ErrorClass) int {
	fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", notFoundErr.Error(), baseURL)
	if verbose {
		fmt.Fprintf(stderr, "detail: %v\n", notFoundErr)
	}
	return class.ExitCode()
}

func reportError(stderr io.Writer, err error, baseURL string, verbose bool) int {
	class, msg := hub.ClassifyError(err)
	fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
	if verbose {
		fmt.Fprintf(stderr, "detail: %v\n", err)
	}
	return class.ExitCode()
}
