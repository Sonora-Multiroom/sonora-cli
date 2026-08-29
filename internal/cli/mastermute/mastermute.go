// Package mastermute implements `sonora get master-mute`, `sonora mute all`,
// and `sonora unmute all` against the system-wide master-mute singleton.
package mastermute

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

const getUsage = "usage: sonora get master-mute [flags]"

// RunGet implements `sonora get master-mute`: it defines and parses this
// command's flags, resolves the hub URL, fetches the current master-mute
// state from the hub, and renders it to stdout. Any failure is reported on
// stderr, never stdout, so scripts piping stdout never see error text.
func RunGet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("get master-mute", flag.ContinueOnError)
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
		return hub.ClassUsage.ExitCode()
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, getUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	mm, err := hub.GetMasterMute(context.Background(), client, baseURL)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderMasterMuteJSON(*mm))
	} else {
		fmt.Fprint(stdout, render.RenderMasterMuteYAML(*mm))
	}
	return 0
}

// RunMute implements `sonora mute all`.
func RunMute(args []string, stdout, stderr io.Writer) int {
	return runSetMuted("mute", true, args, stdout, stderr)
}

// RunUnmute implements `sonora unmute all`.
func RunUnmute(args []string, stdout, stderr io.Writer) int {
	return runSetMuted("unmute", false, args, stdout, stderr)
}

// runSetMuted implements the shared body of RunMute/RunUnmute: it defines
// and parses this command's flags, resolves the hub URL, sets the
// system-wide master-mute state via the hub, and renders the updated state
// to stdout. Any failure is reported on stderr, never stdout, so scripts
// piping stdout never see error text.
func runSetMuted(verb string, muted bool, args []string, stdout, stderr io.Writer) int {
	usage := fmt.Sprintf("usage: sonora %s all [flags]", verb)

	fs := flag.NewFlagSet(verb+" all", flag.ContinueOnError)
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
		return hub.ClassUsage.ExitCode()
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	mm, err := hub.SetMasterMute(context.Background(), client, baseURL, muted)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return class.ExitCode()
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderMasterMuteJSON(*mm))
	} else {
		fmt.Fprint(stdout, render.RenderMasterMuteYAML(*mm))
	}
	return 0
}
