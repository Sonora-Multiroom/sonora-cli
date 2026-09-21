package tts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
)

const getCacheUsage = "usage: sonora get tts-cache [flags]"

// RunGetCache implements `sonora get tts-cache`: it defines and parses this
// command's flags, resolves the hub URL, fetches TTS audio-cache statistics
// from the hub, and renders them to stdout. Any failure is reported on
// stderr, never stdout, so scripts piping stdout never see error text.
func RunGetCache(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("get tts-cache", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, getCacheUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, getCacheUsage)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		return hub.ClassUsage.ExitCode()
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, getCacheUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	stats, err := hub.GetTTSCacheStats(context.Background(), client, baseURL)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderTTSCacheStatsJSON(*stats))
	} else {
		fmt.Fprint(stdout, render.RenderTTSCacheStatsYAML(*stats))
	}
	return 0
}

const clearCacheUsage = "usage: sonora clear tts-cache [flags]"

// RunClearCache implements `sonora clear tts-cache [--provider NAME]`: it
// defines and parses this command's flags, resolves the hub URL, clears the
// TTS audio cache (all entries, or one provider's), and renders the result
// to stdout. Any failure is reported on stderr, never stdout, so scripts
// piping stdout never see error text. It does not prompt for confirmation
// and is idempotent (FR-008): clearing an already-empty cache still exits 0.
func RunClearCache(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("clear tts-cache", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, clearCacheUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	providerFlag := fs.String("provider", "", "clear only this `NAME`d provider's cache entries")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, clearCacheUsage)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		return hub.ClassUsage.ExitCode()
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, clearCacheUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return hub.ClassUsage.ExitCode()
	}

	var providerSupplied bool
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "provider" {
			providerSupplied = true
		}
	})
	var provider *string
	if providerSupplied {
		if strings.TrimSpace(*providerFlag) == "" {
			fmt.Fprintln(stderr, "error: --provider must not be empty")
			return hub.ClassUsage.ExitCode()
		}
		provider = providerFlag
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}

	client := hub.NewClient()
	if err := hub.ClearTTSCache(context.Background(), client, baseURL, provider); err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderTTSCacheClearedJSON(provider))
	} else {
		fmt.Fprint(stdout, render.RenderTTSCacheClearedYAML(provider))
	}
	return 0
}
