package tts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/respath"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

const speakUsage = "usage: sonora speak <text|-> <outputs|groups>/<id> [flags]"

// RunSpeak implements `sonora speak <text|-> <outputs|groups>/<id>`: it
// defines and parses this command's flags, resolves the target and text
// (reading standard input when the text argument is "-"), calls the hub's
// TTS speak endpoint, and renders the result to stdout. Any failure is
// reported on stderr, never stdout, so scripts piping stdout never see
// error text. It returns the process exit code per the exit code classes in
// data-model.md's exit code table.
func RunSpeak(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("speak", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, speakUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	providerFlag := fs.String("provider", "", "TTS provider `NAME` override")
	voiceFlag := fs.String("voice", "", "provider-specific voice `ID` override")
	languageFlag := fs.String("language", "", "BCP 47 language `TAG` override")
	timeoutFlag := fs.String("timeout", "", "override the default 15s response-wait bound")
	playbackModeFlag := fs.String("playback-mode", "", "play over other audio as `MODE` (duck-others or mix); without it, the TTS extension's configured mode")
	routeVolumeFlag := fs.Int("route-volume", -1, "play this announcement only at volume `N` (0-100), whatever the target's volume; the room's volume control does not reach it")
	duckingLevelFlag := fs.Float64("ducking-level", 0, "with playback mode duck-others, lower what else plays by `DB` decibels (-60 to 0); without it, the configured level")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, speakUsage)
		return 0
	}

	// flag.Parse stops at the first non-flag argument, so <text>/<target-path>
	// preceding a flag would otherwise be mistaken for the end of flags.
	// clihelp.ParsePositional re-parses in a loop, peeling off one positional
	// argument at a time, so flags can appear before, between, or after the
	// two identifiers, and handles a "--" terminator (research.md §7)
	// without confusing it with a preceding flag's own value.
	positional, err := clihelp.ParsePositional(fs, args)
	if err != nil {
		return exitcode.Usage
	}
	if len(positional) != 2 {
		fmt.Fprintln(stderr, speakUsage)
		switch {
		case len(positional) == 0:
			fmt.Fprintln(stderr, "error: missing required argument: <text>")
		case len(positional) == 1:
			fmt.Fprintln(stderr, "error: missing required argument: <target-path>")
		default:
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[2:])
		}
		return exitcode.Usage
	}
	textArg, targetArg := positional[0], positional[1]

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, name := range []string{"provider", "voice", "language", "playback-mode"} {
		if !provided[name] {
			continue
		}
		val := map[string]string{"provider": *providerFlag, "voice": *voiceFlag, "language": *languageFlag, "playback-mode": *playbackModeFlag}[name]
		if strings.TrimSpace(val) == "" {
			fmt.Fprintf(stderr, "error: --%s must not be empty\n", name)
			return exitcode.Usage
		}
	}
	// The playback mode and whether it takes a ducking level are left to the
	// hub: the TTS extension's configured mode applies when none is named.
	if provided["route-volume"] && (*routeVolumeFlag < 0 || *routeVolumeFlag > 100) {
		fmt.Fprintln(stderr, "error: --route-volume must be between 0 and 100")
		return exitcode.For(hub.ClassValidation)
	}
	if provided["ducking-level"] && (*duckingLevelFlag < -60 || *duckingLevelFlag > 0) {
		fmt.Fprintln(stderr, "error: --ducking-level must be between -60 and 0")
		return exitcode.For(hub.ClassValidation)
	}

	speakTimeout := hub.SpeakTimeout
	if provided["timeout"] {
		d, err := time.ParseDuration(*timeoutFlag)
		if err != nil || d <= 0 {
			fmt.Fprintln(stderr, "error: --timeout must be a positive duration")
			return exitcode.Usage
		}
		speakTimeout = d
	}

	if strings.HasSuffix(targetArg, "/") {
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintln(stderr, "error: missing required argument: <target-path> must include an id")
		return exitcode.Usage
	}
	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return exitcode.Usage
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, speakUsage)
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
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintf(stderr, "error: speak target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return exitcode.Usage
	}

	text := textArg
	if text == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "error: could not read text from standard input: %v\n", err)
			return exitcode.Usage
		}
		text = strings.TrimRight(string(data), "\r\n")
	}
	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(stderr, "error: text must not be empty")
		return exitcode.Usage
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	req := hub.SpeakRequest{Text: text, TargetName: targetPath.ID, TargetType: targetType}
	if provided["provider"] {
		req.ProviderName = providerFlag
	}
	if provided["voice"] {
		req.Voice = voiceFlag
	}
	if provided["language"] {
		req.Language = languageFlag
	}
	if provided["playback-mode"] {
		req.PlaybackMode = playbackModeFlag
	}
	if provided["route-volume"] {
		req.RouteVolume = routeVolumeFlag
	}
	if provided["ducking-level"] {
		req.DuckingLevelDb = duckingLevelFlag
	}

	client := hub.NewClientWithTimeout(speakTimeout)
	resp, err := hub.Speak(context.Background(), client, baseURL, req)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderSpeakJSON(*resp))
	} else {
		fmt.Fprint(stdout, render.RenderSpeakYAML(*resp))
	}
	return 0
}
