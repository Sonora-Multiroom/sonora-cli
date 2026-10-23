// Package joinmode checks the join mode a user gives on the command line
// (`--join-mode`, `--default-join-mode`) before it goes to the hub.
package joinmode

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Values is the join mode enum from api/openapi.json (constitution Principle
// II), in the order the help and error text name them. The hub rejects
// anything else, so a flag value is checked against it here and upper-cased
// before it goes on the wire — `--join-mode mix` is accepted.
var Values = []string{hub.JoinModeReplace, hub.JoinModeMix, hub.JoinModeDuckOthers}

// Parse upper-cases s and reports whether it is one of Values.
func Parse(s string) (string, bool) {
	mode := strings.ToUpper(s)
	return mode, slices.Contains(Values, mode)
}

// ErrorLine is the stderr line for a join mode Parse rejected, so every
// command that takes one reports it the same way.
func ErrorLine(v string) string {
	return fmt.Sprintf("error: unknown join mode %q; valid modes: %s", v, strings.Join(Values, ", "))
}
