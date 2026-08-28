package main

import (
	"bytes"
	"strings"
	"testing"

	"sonora-cli/internal/version"
)

func TestRunHelp(t *testing.T) {
	cases := [][]string{
		{"help"},
		{"-h"},
		{"--help"},
		{},
	}

	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := run(args, &stdout, &stderr)

		if code != 0 {
			t.Errorf("run(%v) exit code = %d, want 0", args, code)
		}
		if stderr.Len() != 0 {
			t.Errorf("run(%v) stderr = %q, want empty", args, stderr.String())
		}

		out := stdout.String()
		for _, want := range []string{"sonora", "get", "list", "outputs", "inputs", "routes", "groups", "play", "-json", "-hub-url", "-verbose"} {
			if !strings.Contains(out, want) {
				t.Errorf("run(%v) stdout missing %q, got:\n%s", args, want, out)
			}
		}
	}
}

func TestRunVersionPrecedence(t *testing.T) {
	cases := [][]string{
		{"--version", "--help"},
		{"-v", "--help"},
	}

	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := run(args, &stdout, &stderr)

		if code != 0 {
			t.Errorf("run(%v) exit code = %d, want 0", args, code)
		}
		if stderr.Len() != 0 {
			t.Errorf("run(%v) stderr = %q, want empty", args, stderr.String())
		}

		out := stdout.String()
		if out != version.Version+"\n" {
			t.Errorf("run(%v) stdout = %q, want version output %q (not help)", args, out, version.Version+"\n")
		}
		if strings.Contains(out, "Commands:") || strings.Contains(out, "Flags:") {
			t.Errorf("run(%v) printed help content instead of version:\n%s", args, out)
		}
	}
}

// unreachableHubURL is a fast, deterministic connection failure: nothing
// listens on port 1. Used to prove dispatch reached the resource's
// RunList/RunGet (exit 4, a network error) rather than stopping short with
// a usage error.
const unreachableHubURL = "http://127.0.0.1:1"

func TestRunGet_CollectionForm_DispatchesToRunList(t *testing.T) {
	for _, resource := range []string{"inputs", "outputs", "groups", "routes"} {
		t.Run(resource, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{"get", resource, "--hub-url", unreachableHubURL}, &stdout, &stderr)

			if code != 4 {
				t.Fatalf("exit code = %d, want 4 (dispatch reached RunList); stderr: %s", code, stderr.String())
			}
		})
	}
}

func TestRunGet_SingleItemForm_DispatchesToRunGet(t *testing.T) {
	for _, resource := range []string{"inputs", "outputs", "groups", "routes"} {
		t.Run(resource, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{"get", resource + "/some-id", "--hub-url", unreachableHubURL}, &stdout, &stderr)

			if code != 4 {
				t.Fatalf("exit code = %d, want 4 (dispatch reached RunGet); stderr: %s", code, stderr.String())
			}
		})
	}
}

func TestRunList_CollectionForm_IsSynonymOfGet(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"list", "outputs", "--hub-url", unreachableHubURL}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4 (dispatch reached RunList); stderr: %s", code, stderr.String())
	}
}

func TestRunList_WithID_UsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"list", "outputs/some-id"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Error("stderr = empty, want usage error explaining list takes no id")
	}
}

func TestRunGetList_UnrecognizedResource_UsageError(t *testing.T) {
	for _, verb := range []string{"get", "list"} {
		t.Run(verb, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{verb, "bogus"}, &stdout, &stderr)

			if code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
			if !strings.Contains(stderr.String(), "bogus") {
				t.Errorf("stderr = %q, want mention of unrecognized resource", stderr.String())
			}
		})
	}
}

func TestRunGetList_NoResourceArgument_EnumeratesValidResources(t *testing.T) {
	for _, verb := range []string{"get", "list"} {
		t.Run(verb, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{verb}, &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2", code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			for _, want := range []string{"inputs", "outputs", "groups", "routes"} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("stderr = %q, want it to enumerate %q (FR-006a)", stderr.String(), want)
				}
			}
		})
	}
}

func TestRunOldStyleInvocation_UnknownCommand(t *testing.T) {
	cases := [][]string{
		{"outputs", "list"},
		{"outputs", "get", "office-speaker"},
		{"inputs", "list"},
		{"routes", "get", "route-1"},
		{"groups", "list"},
	}

	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := run(args, &stdout, &stderr)

		if code != 2 {
			t.Errorf("run(%v) exit code = %d, want 2", args, code)
		}
		if stdout.Len() != 0 {
			t.Errorf("run(%v) stdout = %q, want empty", args, stdout.String())
		}
		if stderr.Len() == 0 {
			t.Errorf("run(%v) stderr = empty, want unknown-command usage error", args)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"bogus", "list"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "bogus") {
		t.Errorf("stderr = %q, want mention of unknown command", stderr.String())
	}
}

// A verb with no usable resource path still answers --help on stdout with a
// zero exit, rather than reporting "--help" as an unrecognized resource.
func TestRunVerbHelp(t *testing.T) {
	cases := map[string][]string{
		"get":     {"get", "--help"},
		"list":    {"list", "--help"},
		"delete":  {"delete", "--help"},
		"stop":    {"stop", "-h"},
		"enable":  {"enable", "--help"},
		"disable": {"disable", "--help"},
		"set":     {"set", "--help"},
	}

	for name, args := range cases {
		var stdout, stderr bytes.Buffer
		code := run(args, &stdout, &stderr)

		if code != 0 {
			t.Errorf("run(%v) exit code = %d, want 0; stderr: %s", args, code, stderr.String())
		}
		if stderr.Len() != 0 {
			t.Errorf("run(%v) stderr = %q, want empty", args, stderr.String())
		}
		if !strings.Contains(stdout.String(), "usage: sonora "+name) {
			t.Errorf("run(%v) stdout missing the %s usage line, got:\n%s", args, name, stdout.String())
		}
	}
}

// A resource path present alongside --help means the resource's own command
// answers, so the reply carries that command's flag reference.
func TestRunResourceHelpBeatsVerbHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"get", "outputs", "--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Flags:") || !strings.Contains(out, "--include-disabled") {
		t.Errorf("expected the outputs command's own flag reference, got stdout:\n%s", out)
	}
}

// helpSection returns the lines of helpText under the given header, up to the
// next blank line.
func helpSection(t *testing.T, header string) []string {
	t.Helper()

	lines := strings.Split(helpText, "\n")
	start := -1
	for i, l := range lines {
		if l == header {
			start = i + 1
			break
		}
	}
	if start == -1 {
		t.Fatalf("help text has no %q section:\n%s", header, helpText)
	}

	var section []string
	for _, l := range lines[start:] {
		if strings.TrimSpace(l) == "" {
			break
		}
		section = append(section, l)
	}
	if len(section) == 0 {
		t.Fatalf("section %q is empty", header)
	}
	return section
}

// Descriptions used to be misaligned: single-line commands landed in one
// column while the wrapped ones (play, route, transfer, set) landed one
// column further right. Every row must share a description column.
func TestHelpTextColumnsAlign(t *testing.T) {
	for _, header := range []string{"Commands:", "Flags:"} {
		var want int
		for i, line := range helpSection(t, header) {
			gap := strings.Index(line[2:], "  ")
			if gap == -1 {
				t.Errorf("%s row %q has no description", header, line)
				continue
			}
			col := 2 + gap
			for col < len(line) && line[col] == ' ' {
				col++
			}

			if i == 0 {
				want = col
				continue
			}
			if col != want {
				t.Errorf("%s: description of %q starts at column %d, want %d (aligned with the first row)", header, line, col, want)
			}
		}
	}
}

// Help is read in a terminal; nothing may rely on the reader's window being
// wider than the classic 80 columns.
func TestHelpTextFitsEightyColumns(t *testing.T) {
	for _, line := range strings.Split(helpText, "\n") {
		if len(line) > 80 {
			t.Errorf("help line is %d chars, want <= 80:\n%s", len(line), line)
		}
	}
}

// Every command the help advertises must actually dispatch, and every verb
// run() dispatches must be advertised.
func TestHelpTextListsEveryDispatchedVerb(t *testing.T) {
	listed := map[string]bool{}
	for _, line := range helpSection(t, "Commands:") {
		listed[strings.Fields(line)[0]] = true
	}

	for _, verb := range []string{"get", "list", "play", "route", "transfer", "delete", "stop", "enable", "disable", "set", "help"} {
		if !listed[verb] {
			t.Errorf("help text does not list the %q command", verb)
		}
	}
}
