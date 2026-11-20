// Package chimes implements `sonora list chimes` and `sonora get chimes[/<id>]`,
// which show the chimes the hub has loaded: the ids that `--chime` on `play`
// and `route` accepts.
package chimes

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
	listUsage = "usage: sonora list chimes [flags]"
	getUsage  = "usage: sonora get chimes[/<id>] [flags]"
)

// RunList implements `sonora list chimes` and `sonora get chimes`: it lists
// every chime the hub has loaded. verb names the command for usage messages.
func RunList(verb string, args []string, stdout, stderr io.Writer) int {
	usage := getUsage
	if verb == "list" {
		usage = listUsage
	}
	return run(usage, "", args, stdout, stderr)
}

// RunGet implements `sonora get chimes/<id>`: it prints the one chime with
// that id. The hub serves no single-chime endpoint, so this lists the chimes
// and picks the entry; an unknown id is a *hub.NotFoundError, as for other
// resources. The caller has already checked id against the identifier rule.
func RunGet(id string, args []string, stdout, stderr io.Writer) int {
	return run(getUsage, id, args, stdout, stderr)
}

func run(usage, id string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("chimes", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, usage)
		return 0
	}

	if err := fs.Parse(args); err != nil {
		return exitcode.Usage
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	fail := func(err error) int {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return exitcode.For(class)
	}

	list, err := hub.ListChimes(context.Background(), hub.NewClient(), baseURL)
	if err != nil {
		return fail(err)
	}

	if id == "" {
		if *jsonOut {
			fmt.Fprint(stdout, render.RenderChimesJSON(list))
		} else {
			fmt.Fprint(stdout, render.RenderChimesYAML(list))
		}
		return 0
	}

	for _, c := range list {
		if c.ID == id {
			if *jsonOut {
				fmt.Fprint(stdout, render.RenderChimeJSON(c))
			} else {
				fmt.Fprint(stdout, render.RenderChimeYAML(c))
			}
			return 0
		}
	}
	return fail(&hub.NotFoundError{Resource: "chime", ID: id})
}
