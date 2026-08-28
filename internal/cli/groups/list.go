// Package groups implements the `sonora groups` commands.
package groups

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

const listUsage = "usage: sonora get|list groups [flags]"

// RunList implements `sonora get|list groups`: it defines and parses this
// command's flags, resolves the hub URL, fetches groups from the hub
// (enabled-only by default, or all groups with --include-disabled), and
// renders them to stdout. Any failure is reported on stderr, never stdout,
// so scripts piping stdout never see error text. It returns the process
// exit code per the exit code classes in data-model.md.
func RunList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("groups list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, listUsage)

	includeDisabled := fs.Bool("include-disabled", false, "include disabled groups in the results")
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

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	groupList, err := hub.ListGroups(context.Background(), client, baseURL, *includeDisabled)
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
		rendered = render.RenderGroupsJSON(groupList)
	} else {
		rendered = render.RenderGroupsYAML(groupList)
	}
	fmt.Fprint(stdout, rendered)
	return 0
}
