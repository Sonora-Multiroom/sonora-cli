// Package tts implements `sonora speak`, `sonora get tts-cache`, and
// `sonora clear tts-cache`, the CLI commands that wrap the hub's optional
// TTS extension. All three share one failure path: the extension's own
// {error, message} shape, and a 404 meaning "not offered", diagnosed via
// one inventory lookup (research.md §1, §5).
package tts

import (
	"context"
	"errors"
	"fmt"
	"io"

	"sonora-cli/internal/hub"
)

// ReportError is the shared failure path for speak/get tts-cache/clear
// tts-cache (research.md §5). On a *hub.TTSNotOfferedError it makes exactly
// one hub.ListExtensions call to diagnose why TTS is unavailable, building a
// *hub.TTSUnavailableError from the result. Any error is then classified
// with hub.ClassifyError and printed to stderr as `error: <msg> (hub URL:
// <url>)` — or just `error: <msg>` for the hub-address diagnosis, whose
// message already starts with the URL — with `detail: <err>` added when
// verbose is set. It returns the process exit code.
func ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int {
	var notOffered *hub.TTSNotOfferedError
	if errors.As(err, &notOffered) {
		err = diagnoseUnavailable(baseURL)
	}

	class, msg := hub.ClassifyError(err)

	var unavail *hub.TTSUnavailableError
	hubAddress := errors.As(err, &unavail) && unavail.Diagnosis == hub.DiagnosisHubAddress

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
	return class.ExitCode()
}

// diagnoseUnavailable runs the research.md §5 diagnosis: it calls
// hub.ListExtensions once, with a fresh 5s hub.NewClient() (even when the
// failing TTS call used the longer speak client), and maps the result onto
// a TTSDiagnosis. Matching uses the extension's machine id "tts" only, never
// its display name.
func diagnoseUnavailable(baseURL string) error {
	client := hub.NewClient()
	inv, err := hub.ListExtensions(context.Background(), client, baseURL)
	if err != nil {
		var statusErr *hub.StatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == 404 {
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisHubAddress, BaseURL: baseURL}
		}
		return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisUnknown, Cause: err}
	}

	for _, ext := range inv.Extensions {
		if ext.ID != "tts" {
			continue
		}
		switch ext.Status {
		case "DISABLED":
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisDisabled}
		case "REJECTED":
			reason := ""
			if ext.RejectionReason != nil {
				reason = hub.SingleLine(*ext.RejectionReason)
			}
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisRejected, Reason: reason}
		case "INERT":
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisInert}
		case "ACTIVE":
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisVersionMismatch}
		default:
			return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisUnknown}
		}
	}

	loadingDisabled := inv.LoadingEnabled != nil && !*inv.LoadingEnabled
	return &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisNotInstalled, LoadingDisabled: loadingDisabled}
}
