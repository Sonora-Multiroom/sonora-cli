// Package hubinfo implements `sonora get info`, which shows which hub is
// running and since when.
package hubinfo

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

const getUsage = "usage: sonora get info [flags]"

// RunGet implements `sonora get info`: it parses this command's flags,
// resolves the hub URL, fetches the hub's identity, and renders it to stdout.
// Any failure is reported on stderr, never stdout.
func RunGet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("get info", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, getUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, getUsage)
		return 0
	}

	if err := fs.Parse(args); err != nil {
		return exitcode.Usage
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, getUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	info, err := hub.GetInfo(context.Background(), hub.NewClient(), baseURL)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return exitcode.For(class)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderInfoJSON(*info))
	} else {
		fmt.Fprint(stdout, render.RenderInfoYAML(*info))
	}
	return 0
}
