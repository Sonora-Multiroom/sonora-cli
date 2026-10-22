package inputs

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

// renameDescription follows the usage line in --help, saying how long a
// rename lasts.
const renameDescription = "Renaming does not interrupt playback. A rename lasts until the hub restarts for a configured input, and until removal for a runtime input. --reset restores the original name."

// RunRename implements `sonora rename inputs/<input-id> <name>` and
// `sonora rename inputs/<input-id> --reset`: it parses this command's
// flags, resolves the hub URL, sets (or, with --reset, restores) the named
// input's display name via the hub, and renders the updated input to
// stdout. Any failure is reported on stderr, never stdout. It returns the
// process exit code per the exit code classes in data-model.md's exit code
// table.
func RunRename(args []string, stdout, stderr io.Writer) int {
	usage := "usage: sonora rename inputs/<input-id> <name>|--reset [flags]"

	fs := flag.NewFlagSet("rename inputs", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, usage)

	reset := fs.Bool("reset", false, "restore the input's original name instead of setting one")
	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")

	// An explicit --help is a request, not a failure: serve it on stdout
	// and exit 0. Left to flag.Parse it would surface as flag.ErrHelp,
	// printing to stderr and exiting 2.
	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, usage+"\n\n"+renameDescription)
		return 0
	}

	// flag.Parse stops at the first non-flag argument; re-parse in a loop,
	// peeling off one positional argument at a time, so flags can appear
	// before or after the identifier and the name.
	var positional []string
	remaining := args
	for {
		if err := fs.Parse(remaining); err != nil {
			return exitcode.Usage
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		remaining = rest[1:]
	}
	if len(positional) == 0 {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <input-id>")
		return exitcode.Usage
	}
	inputID := positional[0]

	var name *string
	switch {
	case *reset && len(positional) > 1:
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: give either a <name> or --reset, not both")
		return exitcode.Usage
	case *reset:
	case len(positional) == 1:
		fmt.Fprintln(stderr, usage)
		fmt.Fprintln(stderr, "error: missing required argument: <name> (or --reset to restore the original name)")
		return exitcode.Usage
	case len(positional) > 2:
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v (quote a name that contains spaces)\n", positional[2:])
		return exitcode.Usage
	default:
		name = &positional[1]
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	client := hub.NewClient()
	input, err := hub.SetInputDisplayName(context.Background(), client, baseURL, inputID, name)
	if err != nil {
		class, msg := hub.ClassifyError(err)
		fmt.Fprintf(stderr, "error: %s (hub URL: %s)\n", msg, baseURL)
		if *verbose {
			fmt.Fprintf(stderr, "detail: %v\n", err)
		}
		return exitcode.For(class)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderInputJSON(*input))
	} else {
		fmt.Fprint(stdout, render.RenderInputYAML(*input))
	}
	return 0
}
