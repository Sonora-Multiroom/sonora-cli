package integration

import (
	"strings"
	"testing"
)

// TestGlobalFlags_ConsistentAcrossRefactoredCommands closes an FR-009 gap:
// --json, --hub-url, and --verbose were each only exercised incidentally (a
// couple of commands here, --hub-url via runCLI everywhere), so nothing
// caught one of them being dropped from a rewritten flag set. This asserts
// all three are still accepted together on every refactored get/list form
// and on play. --hub-url pointing at an unreachable host proves --hub-url
// took effect (exit 4, a network error, rather than 2, a usage error for an
// unrecognized flag) and --json/--verbose were parsed without erroring;
// --verbose additionally asserts the raw error detail is appended.
func TestGlobalFlags_ConsistentAcrossRefactoredCommands(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"get inputs", []string{"get", "inputs"}},
		{"get inputs/id", []string{"get", "inputs/some-id"}},
		{"list inputs", []string{"list", "inputs"}},
		{"get outputs", []string{"get", "outputs"}},
		{"get outputs/id", []string{"get", "outputs/some-id"}},
		{"list outputs", []string{"list", "outputs"}},
		{"get groups", []string{"get", "groups"}},
		{"get groups/id", []string{"get", "groups/some-id"}},
		{"list groups", []string{"list", "groups"}},
		{"get routes", []string{"get", "routes"}},
		{"get routes/id", []string{"get", "routes/some-id"}},
		{"list routes", []string{"list", "routes"}},
		{"play", []string{"play", "https://stream.example.com/live.mp3", "outputs/some-id"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			args := append(append([]string{}, c.args...), "--hub-url", "http://127.0.0.1:1", "--json")
			res := runCLI(t, args...)
			if res.exitCode != 4 {
				t.Fatalf("%v exit code = %d, want 4 (proves --hub-url/--json were accepted); stderr: %s", args, res.exitCode, res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("%v stdout = %q, want empty on failure", args, res.stdout)
			}

			verboseArgs := append(append([]string{}, args...), "--verbose")
			verboseRes := runCLI(t, verboseArgs...)
			if verboseRes.exitCode != 4 {
				t.Fatalf("%v exit code = %d, want 4", verboseArgs, verboseRes.exitCode)
			}
			if !strings.Contains(verboseRes.stderr, "detail:") {
				t.Errorf("%v expected --verbose to append raw error detail, got stderr:\n%s", verboseArgs, verboseRes.stderr)
			}
			if strings.Contains(res.stderr, "detail:") {
				t.Errorf("%v expected no raw error detail without --verbose, got stderr:\n%s", args, res.stderr)
			}
		})
	}
}
