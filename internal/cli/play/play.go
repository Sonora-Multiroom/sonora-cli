// Package play implements `sonora play <uri> <outputs|groups>/<id>`.
package play

import (
	"context"
	"flag"
	"fmt"
	"io"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/cli/respath"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const usage = "usage: sonora play <uri> <outputs|groups>/<id> [--volume N] [--display-name NAME] [--json] [--verbose] [--hub-url URL]"

// Run implements `sonora play <uri> <outputs|groups>/<id>`: it defines and
// parses this command's flags, validates a client-side volume range,
// resolves the target path (its type is already known from the path prefix,
// per internal/cli/respath), calls the hub's playback endpoint, and renders
// the result to stdout. Any failure is reported on stderr, never stdout, so
// scripts piping stdout never see error text. It returns the process exit
// code per the exit code classes in data-model.md's exit code table.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")
	volumeFlag := fs.Int("volume", -1, "starting volume (0-100) to set before playback starts")
	displayNameFlag := fs.String("display-name", "", "display name for the created ephemeral input")

	// flag.Parse stops at the first non-flag argument, so <uri>/<target-path>
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
			fmt.Fprintln(stderr, "error: missing required argument: <target-path>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return hub.ClassUsage.ExitCode()
	}
	uri, targetArg := positional[0], positional[1]

	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <target-path> must include an id")
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
		fmt.Fprintf(stderr, "error: play target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return hub.ClassUsage.ExitCode()
	}
	targetID := targetPath.ID

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

	if err := hub.ResolveTarget(ctx, client, baseURL, targetID, targetType); err != nil {
		return reportError(stderr, err, baseURL, *verbose)
	}

	req := hub.PlaybackRequest{URI: uri, TargetID: targetID, TargetType: targetType}
	if volumeProvided {
		req.Volume = volumeFlag
	}
	if *displayNameFlag != "" {
		req.DisplayName = displayNameFlag
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
