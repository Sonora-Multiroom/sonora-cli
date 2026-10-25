// Package play implements `sonora play <uri> <outputs|groups>/<id>`.
package play

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/joinmode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/respath"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

const usage = "usage: sonora play <uri> <outputs|groups>/<id> [flags]"

// Run implements `sonora play <uri> <outputs|groups>/<id>`: it defines and
// parses this command's flags, validates a client-side volume range and
// join mode, resolves the target path (its type is already known from the
// path prefix, per internal/cli/respath), calls the hub's playback endpoint, and renders
// the result to stdout. Any failure is reported on stderr, never stdout, so
// scripts piping stdout never see error text. It returns the process exit
// code per the exit code classes in data-model.md's exit code table.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	outputVolumeFlag := fs.Int("output-volume", -1, "set the target's volume to `N` (0-100) before playback starts; it stays after playback ends")
	routeVolumeFlag := fs.Int("route-volume", -1, "play this playback only at volume `N` (0-100), whatever the target's volume; the room's volume control does not reach it")
	duckingLevelFlag := fs.Float64("ducking-level", 0, "with --join-mode DUCK_OTHERS, lower what else plays by `DB` decibels (-60 to 0); without it, the hub's configured level")
	displayNameFlag := fs.String("display-name", "", "display `NAME` for the input, kept as given; without it the hub names a SoundCloud or YouTube link after its track and anything else with a generated name")
	joinModeFlag := fs.String("join-mode", "", "how the playback joins what already plays on the target: `MODE` is REPLACE, MIX or DUCK_OTHERS (any case); without it, REPLACE")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, usage)
		return 0
	}

	// flag.Parse stops at the first non-flag argument, so <uri>/<target-path>
	// preceding a flag (per the documented invocation shape) would otherwise
	// be mistaken for the end of flags. clihelp.ParsePositional re-parses in a
	// loop, peeling off one positional argument at a time, so flags can
	// appear before, between, or after the two identifiers.
	positional, err := clihelp.ParsePositional(fs, args)
	if err != nil {
		return exitcode.Usage
	}
	if len(positional) != 2 {
		fmt.Fprintln(stderr, usage)
		switch {
		case len(positional) == 0:
			fmt.Fprintln(stderr, "error: missing required argument: <uri>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <target-path>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return exitcode.Usage
	}
	uri, targetArg := positional[0], positional[1]

	if strings.HasSuffix(targetArg, "/") {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <target-path> must include an id")
		return exitcode.Usage
	}
	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return exitcode.Usage
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <target-path> must include an id")
		return exitcode.Usage
	}
	var targetType string
	switch targetPath.Kind {
	case respath.Outputs:
		targetType = "SINGLE_OUTPUT"
	case respath.Groups:
		targetType = "OUTPUT_GROUP"
	default:
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: play target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return exitcode.Usage
	}
	targetID := targetPath.ID

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	joinMode, ok := joinmode.Parse(*joinModeFlag)
	if provided["join-mode"] && !ok {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, joinmode.ErrorLine(*joinModeFlag))
		return exitcode.Usage
	}
	// The hub refuses a ducking level for any other join mode, and play's
	// join mode is REPLACE unless named, so only an explicit DUCK_OTHERS
	// can take one.
	if provided["ducking-level"] && (!provided["join-mode"] || joinMode != hub.JoinModeDuckOthers) {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: --ducking-level needs --join-mode DUCK_OTHERS")
		return exitcode.Usage
	}
	for _, v := range []struct {
		name  string
		value int
	}{{"output-volume", *outputVolumeFlag}, {"route-volume", *routeVolumeFlag}} {
		if provided[v.name] && (v.value < 0 || v.value > 100) {
			fmt.Fprintf(stderr, "error: --%s must be between 0 and 100\n", v.name)
			return exitcode.For(hub.ClassValidation)
		}
	}
	if provided["ducking-level"] && (*duckingLevelFlag < -60 || *duckingLevelFlag > 0) {
		fmt.Fprintln(stderr, "error: --ducking-level must be between -60 and 0")
		return exitcode.For(hub.ClassValidation)
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	ctx := context.Background()
	client := hub.NewClient()

	if err := hub.ResolveTarget(ctx, client, baseURL, targetID, targetType); err != nil {
		return reportError(stderr, err, baseURL, *verbose)
	}

	req := hub.PlaybackRequest{URI: uri, TargetID: targetID, TargetType: targetType}
	if provided["output-volume"] {
		req.OutputVolume = outputVolumeFlag
	}
	if provided["route-volume"] {
		req.RouteVolume = routeVolumeFlag
	}
	if provided["ducking-level"] {
		req.DuckingLevelDb = duckingLevelFlag
	}
	if *displayNameFlag != "" {
		req.DisplayName = displayNameFlag
	}
	if provided["join-mode"] {
		req.JoinMode = &joinMode
	}

	resp, err := hub.Playback(ctx, client, baseURL, req)
	if err != nil {
		return reportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderPlaybackJSON(*resp))
	} else {
		fmt.Fprint(stdout, render.RenderPlaybackYAML(*resp))
	}
	return 0
}

func reportError(stderr io.Writer, err error, baseURL string, verbose bool) int {
	class, msg := hub.ClassifyError(err)
	fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
	if verbose {
		fmt.Fprintf(stderr, "detail: %v\n", err)
	}
	return exitcode.For(class)
}
