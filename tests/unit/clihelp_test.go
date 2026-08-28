package unit

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"sonora-cli/internal/cli/clihelp"
)

func TestSetUsage(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("status", "", "only return routes with this status")
	fs.Bool("json", false, "emit strict JSON instead of the default YAML")

	var buf bytes.Buffer
	clihelp.SetUsage(fs, &buf, "usage: sonora routes list [--status STATUS] [--json]")
	fs.Usage()

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	if len(lines) == 0 || lines[0] != "usage: sonora routes list [--status STATUS] [--json]" {
		t.Errorf("first line = %q, want the usage line", out)
	}
	if !strings.Contains(out, "Flags:") {
		t.Errorf("output missing Flags: header, got:\n%s", out)
	}
	if !strings.Contains(out, "--status string") {
		t.Errorf("expected string flag to show its type as the argument, got:\n%s", out)
	}
	if strings.Contains(out, "--json string") || strings.Contains(out, "--json value") {
		t.Errorf("expected bool flag to take no argument, got:\n%s", out)
	}
	if !strings.Contains(out, "only return routes with this status") {
		t.Errorf("output missing flag description, got:\n%s", out)
	}
	if strings.Contains(out, "Usage of") {
		t.Errorf("expected the default flag.PrintDefaults() header to be replaced, got:\n%s", out)
	}
}

func TestSetUsage_NoFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	var buf bytes.Buffer
	clihelp.SetUsage(fs, &buf, "usage: sonora help")
	fs.Usage()

	if strings.Contains(buf.String(), "Flags:") {
		t.Errorf("expected no Flags: header when fs has no flags, got:\n%s", buf.String())
	}
}

// customBoolFlag is a hand-rolled flag.Value (not one of the flag package's
// own bool/string/int types) that reports itself as boolean via
// IsBoolFlag(), to prove SetUsage detects bool-ness through that interface
// generically rather than only for flag.Bool's built-in type.
type customBoolFlag struct{}

func (customBoolFlag) String() string   { return "" }
func (customBoolFlag) Set(string) error { return nil }
func (customBoolFlag) IsBoolFlag() bool { return true }

func TestSetUsage_CustomBoolFlagValue(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Var(customBoolFlag{}, "custom", "a custom bool-like flag")

	var buf bytes.Buffer
	clihelp.SetUsage(fs, &buf, "usage: test")
	fs.Usage()

	out := buf.String()
	if !strings.Contains(out, "Flags:") {
		t.Fatalf("output missing Flags: header, got:\n%s", out)
	}
	if strings.Contains(out, "--custom value") {
		t.Errorf("expected custom bool-like flag to omit the value placeholder, got:\n%s", out)
	}
}

// A back-quoted span in a flag's usage text names that flag's argument, so
// help can read "--hub-url URL" rather than the generic "--hub-url string".
func TestSetUsage_BackQuotedArgumentName(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("hub-url", "", "hub base `URL` override")

	var buf bytes.Buffer
	clihelp.SetUsage(fs, &buf, "usage: test")
	fs.Usage()

	out := buf.String()
	if !strings.Contains(out, "--hub-url URL") {
		t.Errorf("expected the back-quoted argument name, got:\n%s", out)
	}
	if strings.Contains(out, "`") {
		t.Errorf("expected the back-quotes to be stripped from the description, got:\n%s", out)
	}
	if !strings.Contains(out, "hub base URL override") {
		t.Errorf("output missing the unquoted description, got:\n%s", out)
	}
}

// PrintUsage is what serves an explicit --help, so it must render the same
// block SetUsage installs for failures, just to a caller-chosen writer.
func TestPrintUsage_MatchesSetUsage(t *testing.T) {
	newFlagSet := func() *flag.FlagSet {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("hub-url", "", "hub base `URL` override")
		fs.Bool("json", false, "emit strict JSON")
		return fs
	}

	var viaSet bytes.Buffer
	fs := newFlagSet()
	clihelp.SetUsage(fs, &viaSet, "usage: test")
	fs.Usage()

	var viaPrint bytes.Buffer
	clihelp.PrintUsage(newFlagSet(), &viaPrint, "usage: test")

	if viaSet.String() != viaPrint.String() {
		t.Errorf("PrintUsage and SetUsage disagree.\nSetUsage:\n%s\nPrintUsage:\n%s", viaSet.String(), viaPrint.String())
	}
}

func TestRequested(t *testing.T) {
	// AGENTS.md: long flags require "--", short flags are single-letter with
	// "-". The single-dash multi-letter "-help" the flag package would accept
	// is not valid in this CLI.
	yes := [][]string{
		{"--help"},
		{"-h"},
		{"outputs", "--help"},
		{"--json", "-h", "outputs"},
	}
	for _, args := range yes {
		if !clihelp.Requested(args) {
			t.Errorf("Requested(%v) = false, want true", args)
		}
	}

	no := [][]string{
		{},
		{"outputs"},
		{"-help"},
		{"--helpful"},
		{"--", "--help"},
		{"play", "--", "-h"},
	}
	for _, args := range no {
		if clihelp.Requested(args) {
			t.Errorf("Requested(%v) = true, want false", args)
		}
	}
}
