// Package clihelp gives every `sonora <noun> <verb>` command the same
// formatted --help output: its usage line, then an aligned list of flags,
// in place of the flag package's default "Usage of <name>:" dump.
package clihelp

import (
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
)

// boolFlag mirrors the unexported interface the flag package's own
// PrintDefaults uses to tell a value-less flag (e.g. --json) apart from one
// that takes a value (e.g. --hub-url).
type boolFlag interface {
	IsBoolFlag() bool
}

// SetUsage installs a formatted fs.Usage function that writes usageLine
// followed by an aligned "Flags:" section (omitted if fs defines no flags)
// to w.
func SetUsage(fs *flag.FlagSet, w io.Writer, usageLine string) {
	fs.Usage = func() {
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
			name := "--" + f.Name
			if bf, ok := f.Value.(boolFlag); !ok || !bf.IsBoolFlag() {
				name += " value"
			}
			fmt.Fprintf(tw, "  %s\t%s\n", name, f.Usage)
		})
		tw.Flush()
	}
}
