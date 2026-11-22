package inputs

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

const (
	keepUsage       = "usage: sonora keep inputs/<input-id> [flags]"
	keepDescription = "Keeps a runtime input across hub restarts (origin STORED). It stops auto-removing, and a route playing it is not touched. An input that is already kept is returned unchanged; one from the hub's configuration cannot be kept, and the hub refuses it with 400."
)

// RunKeep implements `sonora keep inputs/<input-id>` (hub 0.1.35 and later):
// it promotes a runtime input to a kept one and renders the input as it now
// reads. Failures go to stderr; the exit code follows the exit code classes.
func RunKeep(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("keep inputs", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, keepUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, keepUsage+"\n\n"+keepDescription)
		return 0
	}

	// Peel off positional arguments one at a time so flags may follow the id.
	var positional []string
	remaining := args
	for {
		if err := fs.Parse(remaining); err != nil {
			return exitcode.Usage
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		remaining = rest[1:]
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, keepUsage)
		if len(positional) == 0 {
			fmt.Fprintln(stderr, "error: missing required argument: <input-id>")
		} else {
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[1:])
		}
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	input, err := hub.PersistInput(context.Background(), hub.NewClient(), baseURL, positional[0])
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return exitcode.For(class)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderInputJSON(*input))
	} else {
		fmt.Fprint(stdout, render.RenderInputYAML(*input))
	}
	return 0
}
