// Package clihelp gives every `sonora <verb> <resource>[/<id>]` command the
// same formatted --help output: its usage line, then an aligned list of
// flags, in place of the flag package's default "Usage of <name>:" dump.
package clihelp

import (
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
)

// SetUsage installs a formatted fs.Usage function that writes usageLine
// followed by an aligned "Flags:" section (omitted if fs defines no flags)
// to w. w is the failure stream: fs.Usage runs when a parse error is being
// reported, so an explicit help request is served by PrintUsage instead.
func SetUsage(fs *flag.FlagSet, w io.Writer, usageLine string) {
	fs.Usage = func() { PrintUsage(fs, w, usageLine) }
}

// PrintUsage writes usageLine followed by an aligned "Flags:" section
// (omitted if fs defines no flags) to w. Call it directly, with stdout, to
// serve an explicit --help; SetUsage wires it to stderr for parse failures.
func PrintUsage(fs *flag.FlagSet, w io.Writer, usageLine string) {
	fmt.Fprintln(w, usageLine)

	var any bool
	fs.VisitAll(func(*flag.Flag) { any = true })
	if !any {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fs.VisitAll(func(f *flag.Flag) {
		// UnquoteUsage names the flag's argument: whatever a `back-quoted`
		// span in the usage text says, else a type word ("string", "int"),
		// else "" for a boolean, which takes no argument at all.
		arg, usage := flag.UnquoteUsage(f)
		label := "--" + f.Name
		if arg != "" {
			label += " " + arg
		}
		fmt.Fprintf(tw, "  %s\t%s\n", label, usage)
	})
	tw.Flush()
}

// Requested reports whether args contains an explicit help flag, stopping at
// a "--" terminator exactly as flag.Parse does. The flag package turns help
// into a parse error (flag.ErrHelp), which would make asking for help exit
// non-zero with its output on stderr; commands check this before parsing so
// help is a success printed to stdout.
//
// Only the two spellings AGENTS.md sanctions are recognized: the single-dash
// multi-letter "-help" the flag package would also accept is not valid in
// this CLI.
func Requested(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}
