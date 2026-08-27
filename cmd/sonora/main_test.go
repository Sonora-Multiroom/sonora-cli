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
