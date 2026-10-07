// Package exitcode maps hub error classes to the process exit codes of the
// sonora command.
package exitcode

import "github.com/Sonora-Multiroom/sonora-cli/hub"

// Usage is the exit code for a command-line usage error (bad flags, unknown
// command, missing argument).
const Usage = 2

// For returns the process exit code for a hub error classification. Exit
// code 7 is retired: it used to mean "target matches both an output and a
// group", a case that path-style target addressing makes structurally
// unreachable; it is never reused by another class.
func For(c hub.ErrorClass) int {
	switch c {
	case hub.ClassHub:
		return 3
	case hub.ClassNetwork:
		return 4
	case hub.ClassNotFound:
		return 5
	case hub.ClassValidation:
		return 6
	case hub.ClassRouteFailed:
		return 8
	case hub.ClassSourceUnreachable:
		return 9
	case hub.ClassServiceUnavailable:
		return 10
	case hub.ClassInputNotFound:
		return 11
	case hub.ClassTargetNotFound:
		return 12
	case hub.ClassTTSUnavailable:
		return 13
	case hub.ClassConflict:
		return 14
	default:
		return 0
	}
}
