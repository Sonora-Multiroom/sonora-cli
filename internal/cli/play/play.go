// Package play implements `sonora play <uri> <target-id>`.
package play

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

const usage = "usage: sonora play <uri> <target-id> [--group | --output] [--volume N] [--name NAME] [--json] [--verbose] [--hub-url URL]"

// Run implements `sonora play <uri> <target-id>`: it defines and parses this
// command's flags, validates a client-side volume range, resolves the
// target's type (single output or group), calls the hub's playback
// endpoint, and renders the result to stdout. Any failure is reported on
// stderr, never stdout, so scripts piping stdout never see error text. It
// returns the process exit code per the exit code classes in
// data-model.md's exit code table.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")
	groupFlag := fs.Bool("group", false, "force the target to be resolved as an output group")
	outputFlag := fs.Bool("output", false, "force the target to be resolved as a single output")
	volumeFlag := fs.Int("volume", -1, "starting volume (0-100) to set before playback starts")
	nameFlag := fs.String("name", "", "display name for the created ephemeral input")

	// flag.Parse stops at the first non-flag argument, so <uri>/<target-id>
	// preceding a flag (per the documented invocation shape) would otherwise
	// be mistaken for the end of flags. Re-parse in a loop, peeling off one
	// positional argument at a time, so flags can appear before, between, or
	// after the two identifiers.
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
			fmt.Fprintln(stderr, "error: missing required argument: <uri>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <target-id>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return hub.ClassUsage.ExitCode()
	}
	uri, targetID := positional[0], positional[1]

	if *groupFlag && *outputFlag {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: --group and --output are mutually exclusive")
		return hub.ClassUsage.ExitCode()
	}

	volumeProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "volume" {
			volumeProvided = true
		}
	})
	if volumeProvided && (*volumeFlag < 0 || *volumeFlag > 100) {
		fmt.Fprintln(stderr, "error: volume must be between 0 and 100")
		return hub.ClassValidation.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	ctx := context.Background()
	client := hub.NewClient()

	targetType, err := hub.ResolveTarget(ctx, client, baseURL, targetID, *groupFlag, *outputFlag)
	if err != nil {
		return reportError(stderr, err, baseURL, *verbose)
	}

	req := hub.PlaybackRequest{URI: uri, TargetID: targetID, TargetType: targetType}
	if volumeProvided {
		req.Volume = volumeFlag
	}
	if *nameFlag != "" {
		req.DisplayName = nameFlag
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
	return class.ExitCode()
}
