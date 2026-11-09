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
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

const warmUsage = "usage: sonora warm tts-cache <text|-> [flags]"

// nothingStoredNote is appended to a 503 from the warm call: the hub stored
// nothing, so the user can simply repeat the command. It is CLI wording, so
// it lives here and not in package hub.
const nothingStoredNote = "; nothing was stored, repeating the request is safe"

// RunWarm implements `sonora warm tts-cache <text|->`: it has the hub
// synthesize the phrase into its audio cache without playing it, so a later
// `speak` with the same text and voice options starts at once. The text and
// "-" (standard input) handling and the --provider, --voice, --language and
// --timeout checks are those of `speak`. Failures go to stderr; it returns
// the process exit code.
func RunWarm(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("warm tts-cache", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, warmUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	providerFlag := fs.String("provider", "", "TTS provider `NAME` override")
	voiceFlag := fs.String("voice", "", "provider-specific voice `ID` override")
	languageFlag := fs.String("language", "", "BCP 47 language `TAG` override")
	timeoutFlag := fs.String("timeout", "", "override the default 15s response-wait bound")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, warmUsage)
		return 0
	}

	positional, err := clihelp.ParsePositional(fs, args)
	if err != nil {
		return exitcode.Usage
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, warmUsage)
		if len(positional) == 0 {
			fmt.Fprintln(stderr, "error: missing required argument: <text>")
		} else {
			fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", positional[1:])
		}
		return exitcode.Usage
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, name := range []string{"provider", "voice", "language"} {
		if !provided[name] {
			continue
		}
		val := map[string]string{"provider": *providerFlag, "voice": *voiceFlag, "language": *languageFlag}[name]
		if strings.TrimSpace(val) == "" {
			fmt.Fprintf(stderr, "error: --%s must not be empty\n", name)
			return exitcode.Usage
		}
	}

	warmTimeout := hub.SpeakTimeout
	if provided["timeout"] {
		d, err := time.ParseDuration(*timeoutFlag)
		if err != nil || d <= 0 {
			fmt.Fprintln(stderr, "error: --timeout must be a positive duration")
			return exitcode.Usage
		}
		warmTimeout = d
	}

	text := positional[0]
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

	req := hub.WarmRequest{Text: text}
	if provided["provider"] {
		req.ProviderName = providerFlag
	}
	if provided["voice"] {
		req.Voice = voiceFlag
	}
	if provided["language"] {
		req.Language = languageFlag
	}

	client := hub.NewClientWithTimeout(warmTimeout)
	res, err := hub.WarmTTSCache(context.Background(), client, baseURL, req)
	if err != nil {
		if class, _ := hub.ClassifyError(err); class == hub.ClassServiceUnavailable {
			return reportWarmUnavailable(stderr, err, baseURL, *verbose)
		}
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderWarmJSON(*res))
	} else {
		fmt.Fprint(stdout, render.RenderWarmYAML(*res))
	}
	return 0
}

// reportWarmUnavailable reports a 503 from the warm call like ReportError
// does, with the note that nothing was stored placed after the hub's message.
func reportWarmUnavailable(stderr io.Writer, err error, baseURL string, verbose bool) int {
	class, msg := hub.ClassifyError(err)
	fmt.Fprintf(stderr, "error: %s%s (hub URL: %s)\n", msg, nothingStoredNote, baseURL)
	if verbose {
		fmt.Fprintf(stderr, "detail: %v\n", err)
	}
	return exitcode.For(class)
}
