package clihelp

import (
	"flag"
	"strings"
)

// ParsePositional collects args's positional arguments, letting flags and
// positionals interleave in any order — flag.Parse alone stops at the first
// non-flag token, so it can't do this by itself. It re-parses in a loop,
// peeling off one positional argument at a time.
//
// A "--" token ends flag parsing exactly once: everything after it is
// positional, including tokens shaped like flags (e.g. "--json"). Locating
// that terminator requires knowing which of fs's flags consume a following
// value: flag.Parse accepts any string, including "--", as such a flag's
// value unconditionally, so a "--" right after e.g. "--provider" is that
// flag's value, not a terminator, and must not be treated as one.
func ParsePositional(fs *flag.FlagSet, args []string) ([]string, error) {
	head, tail := args, []string(nil)
	if idx := terminatorIndex(fs, args); idx >= 0 {
		head, tail = args[:idx], args[idx+1:]
	}

	var positional []string
	remaining := head
	for {
		if err := fs.Parse(remaining); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		remaining = rest[1:]
	}
	return append(positional, tail...), nil
}

// terminatorIndex finds the index of the "--" token flag.Parse would treat
// as an end-of-flags terminator, or -1 if args contains none. It mirrors
// flag.Parse's own token-consumption rules — a "--flag=value" token is
// self-contained, a bool flag takes no following token, and any other known
// flag consumes the next token unconditionally as its value, even one
// spelled "--" — so a real terminator is never confused with a flag's value.
func terminatorIndex(fs *flag.FlagSet, args []string) int {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return i
		}
		if len(a) < 2 || a[0] != '-' {
			continue
		}
		name := strings.TrimLeft(a, "-")
		if strings.ContainsRune(name, '=') {
			continue
		}
		fl := fs.Lookup(name)
		if fl == nil {
			continue
		}
		if bf, ok := fl.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			continue
		}
		i++ // this flag's value is the next token, even if it is "--"
	}
	return -1
}
