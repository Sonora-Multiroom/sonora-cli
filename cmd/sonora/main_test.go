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
		for _, want := range []string{"sonora", "outputs", "inputs", "routes", "groups", "play", "-json", "-hub-url", "-verbose"} {
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

func TestRunUnknownNoun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"bogus", "list"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "bogus") {
		t.Errorf("stderr = %q, want mention of unknown noun", stderr.String())
	}
}

func TestRunMissingVerb(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"outputs"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Error("stderr = empty, want usage error")
	}
}

func TestRunUnknownVerb(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"outputs", "bogus"}, &stdout, &stderr)

	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "bogus") {
		t.Errorf("stderr = %q, want mention of unknown verb", stderr.String())
	}
}
