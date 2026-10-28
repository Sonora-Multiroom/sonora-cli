// Package tts implements `sonora speak`, `sonora get tts-cache`, `sonora
// clear tts-cache`, and `sonora list tts-voices`, the CLI commands that wrap
// the hub's optional TTS extension. All four share one failure path: the extension's own
// {error, message} shape, and a 404 meaning "not offered", diagnosed via
// one inventory lookup (research.md §1, §5).
package tts

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
)

// hubAddressHint tells the user how to fix a wrong hub address; hub's own
// message (hub.TTSUnavailableError, DiagnosisHubAddress) is CLI-neutral, so
// this CLI-specific hint is appended here instead.
const hubAddressHint = "; set the correct address with --hub-url, MULTIROOM_URL, or the config file"

// versionMismatchMsg replaces hub's CLI-neutral DiagnosisVersionMismatch
// message with this CLI's own wording, naming the CLI rather than a generic
// "client".
const versionMismatchMsg = "text-to-speech is not available on this hub: the hub's TTS API does not match this CLI version"

// ReportError is the shared failure path for speak/get tts-cache/clear
// tts-cache/list tts-voices (research.md §5). On a *hub.TTSNotOfferedError it makes exactly
// one hub.ListExtensions call to diagnose why TTS is unavailable, building a
// *hub.TTSUnavailableError from the result. Any error is then classified
// with hub.ClassifyError; for the hub-address and version-mismatch
// diagnoses, this CLI's own hint/wording is added back on top of hub's
// CLI-neutral message (FR-015). The result is printed to stderr as `error:
// <msg> (hub URL: <url>)` — or just `error: <msg>` for the hub-address
// diagnosis, whose message already starts with the URL — with `detail:
// <err>` added when verbose is set. It returns the process exit code.
func ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int {
	var notOffered *hub.TTSNotOfferedError
	if errors.As(err, &notOffered) {
		err = diagnoseUnavailable(baseURL)
	}

	class, msg := hub.ClassifyError(err)

	var unavail *hub.TTSUnavailableError
	hubAddress := errors.As(err, &unavail) && unavail.Diagnosis == hub.DiagnosisHubAddress
	if hubAddress {
		msg += hubAddressHint
	} else if unavail != nil && unavail.Diagnosis == hub.DiagnosisVersionMismatch {
		msg = versionMismatchMsg
	}

	if hubAddress {
		fmt.Fprintf(stderr, "error: %s\n", msg)
	} else {
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
	}
	if verbose {
		var detail error = err
		if unavail != nil && unavail.Cause != nil {
			detail = unavail.Cause
		}
		fmt.Fprintf(stderr, "detail: %v\n", detail)
	}
	return exitcode.For(class)
}

// diagnoseUnavailable runs the research.md §5 diagnosis through
// hub.DiagnoseTTSUnavailable, with a fresh 5s hub.NewClient() (even when the
// failing TTS call used the longer speak client).
func diagnoseUnavailable(baseURL string) error {
	return hub.DiagnoseTTSUnavailable(context.Background(), hub.NewClient(), baseURL)
}
