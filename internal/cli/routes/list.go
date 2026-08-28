// Package routes implements the `sonora routes` commands.
package routes

import (
	"context"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const listUsage = "usage: sonora get|list routes [flags]"

// routeStatuses is the RouteResponse.status enum from api/openapi.json
// (constitution Principle II). The hub rejects anything else, so --status is
// checked against it here and upper-cased before it goes on the wire —
// `--status active` is the natural thing to type and is accepted.
var routeStatuses = []string{"STARTING", "ACTIVE", "STOPPING", "STOPPED", "FAILED"}

// RunList implements `sonora get|list routes`: it defines and parses this
// command's flags, resolves the hub URL, fetches routes from the hub
// (optionally narrowed by --status/--input-id/--target-id), and renders
// them to stdout. Any failure is reported on stderr, never stdout, so
// scripts piping stdout never see error text. It returns the process exit
// code per the exit code classes in data-model.md.
func RunList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("routes list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, listUsage)

	status := fs.String("status", "", "filter by `STATUS`: "+strings.Join(routeStatuses, ", "))
	inputID := fs.String("input-id", "", "only return routes sourced from this input `ID`")
	targetID := fs.String("target-id", "", "only return routes pointed at this target `ID`")
	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, listUsage)
		return 0
	}

	if err := fs.Parse(args); err != nil {
		return hub.ClassUsage.ExitCode()
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, listUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", fs.Args())
		return hub.ClassUsage.ExitCode()
	}

	statusFilter := strings.ToUpper(*status)
	if statusFilter != "" && !slices.Contains(routeStatuses, statusFilter) {
		fmt.Fprintln(stderr, listUsage)
		fmt.Fprintf(stderr, "error: unknown status %q; valid statuses: %s\n", *status, strings.Join(routeStatuses, ", "))
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	routeList, err := hub.ListRoutes(context.Background(), client, baseURL, statusFilter, *inputID, *targetID)
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
		rendered = render.RenderRoutesJSON(routeList)
	} else {
		rendered = render.RenderRoutesYAML(routeList)
	}
	fmt.Fprint(stdout, rendered)
	return 0
}
