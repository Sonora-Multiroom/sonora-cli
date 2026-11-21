// Package radio implements `sonora search radio-stations`, `sonora get
// radio-stations/<uuid>` and `sonora get|list radio-countries|radio-languages|
// radio-tags`, the CLI commands that wrap the hub's optional radio extension
// (RADIO-002). All share one failure path: the extension's own problem
// details, and a 404 meaning "not served", diagnosed with one inventory
// lookup.
package radio

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

// hubAddressHint tells the user how to fix a wrong hub address; the hub
// package's own message is CLI-neutral, so this CLI-specific hint is added
// here.
const hubAddressHint = "; set the correct address with --hub-url, MULTIROOM_URL, or the config file"

// ReportError is the shared failure path of every radio command. On a
// *hub.RadioNotOfferedError it makes exactly one hub.ListExtensions call to
// diagnose why radio is unavailable. The error is then classified with
// hub.ClassifyError and printed to stderr as `error: <msg> (hub URL: <url>)`,
// or `error: <msg>` for the hub-address diagnosis, whose message already
// starts with the URL. Hub-supplied text is put on one line and escaped, so
// it cannot drive the terminal. With verbose, `detail: <err>` follows. It
// returns the process exit code.
func ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int {
	var notOffered *hub.RadioNotOfferedError
	if errors.As(err, &notOffered) {
		err = hub.DiagnoseRadioUnavailable(context.Background(), hub.NewClient(), baseURL)
	}

	class, msg := hub.ClassifyError(err)
	msg = render.SafeText(hub.SingleLine(msg))

	var unavail *hub.RadioUnavailableError
	hubAddress := errors.As(err, &unavail) && unavail.Diagnosis == hub.DiagnosisHubAddress
	if hubAddress {
		fmt.Fprintf(stderr, "error: %s%s\n", msg, hubAddressHint)
	} else {
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
	}
	if verbose {
		var detail error = err
		if unavail != nil && unavail.Cause != nil {
			detail = unavail.Cause
		}
		fmt.Fprintf(stderr, "detail: %s\n", render.SafeText(hub.SingleLine(detail.Error())))
	}
	return exitcode.For(class)
}
