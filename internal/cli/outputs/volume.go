package outputs

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const setVolumeUsage = "usage: sonora set outputs/<output-id> volume <0-100> [--json] [--verbose] [--hub-url URL]"

// RunSetVolume implements `sonora set outputs/<output-id> volume <0-100>`:
// it defines and parses this command's flags, resolves the hub URL, sets
// the named output's volume via the hub, and renders the confirmation to
// stdout. Any failure is reported on stderr, never stdout, so scripts
// piping stdout never see error text. It returns the process exit code per
// the exit code classes in internal/hub/errors.go.
func RunSetVolume(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("set outputs volume", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, setVolumeUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base URL override")

	// flag.Parse stops at the first non-flag argument, so the positional
	// <output-id>/volume/<value> triple (per the documented invocation
	// shape) would otherwise be mistaken for the end of flags. Re-parse in
	// a loop, peeling off one positional argument at a time, so flags can
	// appear anywhere relative to the positionals.
	var positional []string
	remaining := args
	for {
		// The <0-100> value is the one positional that can legitimately
		// start with '-'. flag.Parse would read a negative value as an
		// undefined flag and fail with "flag provided but not defined: -5",
		// hiding the range error below, so peel it off before parsing.
		if len(positional) == 2 && len(remaining) > 0 && looksNumeric(remaining[0]) {
			positional = append(positional, remaining[0])
			remaining = remaining[1:]
			continue
		}
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
	if len(positional) != 3 {
		fmt.Fprintln(stderr, setVolumeUsage)
		switch {
		case len(positional) < 3:
			fmt.Fprintln(stderr, "error: missing required arguments: <output-id> volume <0-100>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[3:])
		}
		return hub.ClassUsage.ExitCode()
	}
	outputID, attr, valueArg := positional[0], positional[1], positional[2]

	if attr != "volume" {
		fmt.Fprintln(stderr, setVolumeUsage)
		fmt.Fprintf(stderr, "error: unsupported attribute %q for outputs; only volume is supported\n", attr)
		return hub.ClassUsage.ExitCode()
	}

	volume, err := strconv.Atoi(valueArg)
	if err != nil || volume < 0 || volume > 100 {
		fmt.Fprintln(stderr, setVolumeUsage)
		fmt.Fprintf(stderr, "error: volume must be an integer between 0 and 100, got %q\n", valueArg)
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	ov, err := hub.SetOutputVolume(context.Background(), client, baseURL, outputID, volume)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderOutputVolumeJSON(*ov))
	} else {
		fmt.Fprint(stdout, render.RenderOutputVolumeYAML(*ov))
	}
	return 0
}

// looksNumeric reports whether s is a decimal integer literal, including one
// too large to fit an int64 — such a value still belongs to the volume slot,
// where Atoi rejects it with the range error, rather than to flag parsing.
func looksNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil || errors.Is(err, strconv.ErrRange)
}
