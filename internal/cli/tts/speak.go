package tts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"sonora-cli/internal/cli/clihelp"
	"sonora-cli/internal/cli/respath"
	"sonora-cli/internal/config"
	"sonora-cli/internal/hub"
	"sonora-cli/internal/render"
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
		return hub.ClassUsage.ExitCode()
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
		return hub.ClassUsage.ExitCode()
	}
	textArg, targetArg := positional[0], positional[1]

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, name := range []string{"provider", "voice", "language"} {
		if !provided[name] {
			continue
		}
		val := map[string]string{"provider": *providerFlag, "voice": *voiceFlag, "language": *languageFlag}[name]
		if strings.TrimSpace(val) == "" {
			fmt.Fprintf(stderr, "error: --%s must not be empty\n", name)
			return hub.ClassUsage.ExitCode()
		}
	}

	speakTimeout := hub.SpeakTimeout
	if provided["timeout"] {
		d, err := time.ParseDuration(*timeoutFlag)
		if err != nil || d <= 0 {
			fmt.Fprintln(stderr, "error: --timeout must be a positive duration")
			return hub.ClassUsage.ExitCode()
		}
		speakTimeout = d
	}

	if strings.HasSuffix(targetArg, "/") {
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintln(stderr, "error: missing required argument: <target-path> must include an id")
		return hub.ClassUsage.ExitCode()
	}
	targetPath, err := respath.Parse(targetArg)
	if err != nil {
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintf(stderr, "sonora: %v\n", err)
		return hub.ClassUsage.ExitCode()
	}
	if targetPath.ID == "" {
		fmt.Fprintln(stderr, speakUsage)
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
		fmt.Fprintln(stderr, speakUsage)
		fmt.Fprintf(stderr, "error: speak target must be outputs/<id> or groups/<id>, got %q\n", targetArg)
		return hub.ClassUsage.ExitCode()
	}

	text := textArg
	if text == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "error: could not read text from standard input: %v\n", err)
			return hub.ClassUsage.ExitCode()
		}
		text = strings.TrimRight(string(data), "\r\n")
	}
	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(stderr, "error: text must not be empty")
		return hub.ClassUsage.ExitCode()
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return hub.ClassUsage.ExitCode()
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
