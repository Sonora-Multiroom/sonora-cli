package radio

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

// StationUsage is the usage line of `sonora get radio-stations/<uuid>`.
const StationUsage = "usage: sonora get radio-stations/<uuid> [flags]"

// RunGetStation implements `sonora get radio-stations/<uuid>`: it fetches one
// station from the hub's radio catalogue with one request and prints it as a
// bare record. The caller has checked that uuid is one non-empty path segment;
// the hub checks the UUID itself and its refusal is shown as it words it.
func RunGetStation(uuid string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("get radio-stations", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, StationUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, StationUsage)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		return exitcode.Usage
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, StationUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	st, err := hub.Station(context.Background(), hub.NewClientWithTimeout(hub.RadioTimeout), baseURL, uuid)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderStationJSON(*st))
	} else {
		fmt.Fprint(stdout, render.RenderStationYAML(*st))
	}
	return 0
}
