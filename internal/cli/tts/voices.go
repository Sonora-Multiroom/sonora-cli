package tts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

const listVoicesUsage = "usage: sonora list tts-voices --provider NAME [flags]"

// listVoicesNote follows the flag list in --help: `speak` has no --engine
// flag, so a google-cloud short name (which needs an engine and language
// alongside it) can't be spoken as-is, while its fullName can.
const listVoicesNote = `
Only google-cloud and google-gemini providers list voices. Pass a voice's
fullName to 'sonora speak --voice'.`

// RunListVoices implements `sonora list tts-voices --provider NAME
// [--language CODE] [--engine NAME]`: it defines and parses this command's
// flags, resolves the hub URL, fetches the provider's voices from the hub,
// and renders them to stdout. Filters are passed through verbatim — the hub
// matches them case-insensitively and knows the engine aliases. Any failure
// is reported on stderr, never stdout, so scripts piping stdout never see
// error text.
func RunListVoices(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("list tts-voices", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, listVoicesUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	providerFlag := fs.String("provider", "", "list the voices of this `NAME`d provider (required)")
	languageFlag := fs.String("language", "", "only voices for this language-region `CODE` (e.g. uk-UA)")
	engineFlag := fs.String("engine", "", "only voices of this `ENGINE` (google-cloud only)")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, listVoicesUsage)
		fmt.Fprintln(stdout, listVoicesNote)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		return exitcode.Usage
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, listVoicesUsage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return exitcode.Usage
	}

	supplied := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { supplied[f.Name] = true })

	if !supplied["provider"] {
		fmt.Fprintln(stderr, listVoicesUsage)
		fmt.Fprintln(stderr, "error: --provider is required")
		return exitcode.Usage
	}
	optional := func(name string, value *string) (*string, bool) {
		if !supplied[name] {
			return nil, true
		}
		if strings.TrimSpace(*value) == "" {
			fmt.Fprintf(stderr, "error: --%s must not be empty\n", name)
			return nil, false
		}
		return value, true
	}
	if _, ok := optional("provider", providerFlag); !ok {
		return exitcode.Usage
	}
	language, ok := optional("language", languageFlag)
	if !ok {
		return exitcode.Usage
	}
	engine, ok := optional("engine", engineFlag)
	if !ok {
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	client := hub.NewClient()
	list, err := hub.ListTTSVoices(context.Background(), client, baseURL, *providerFlag, language, engine)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderTTSVoicesJSON(*list))
	} else {
		fmt.Fprint(stdout, render.RenderTTSVoicesYAML(*list))
	}
	return 0
}
