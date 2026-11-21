package radio

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

// SearchUsage is the usage line of `sonora search radio-stations`.
const SearchUsage = "usage: sonora search radio-stations [name] [flags]"

// The hub refuses a page size outside 1-100 and an offset outside 0-10000;
// the CLI checks them first so a wrong value costs no request.
const (
	minLimit  = 1
	maxLimit  = 100
	maxOffset = 10000
)

// supplied reports which flags the command line set.
func supplied(fs *flag.FlagSet) map[string]bool {
	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	return set
}

// optionalString returns the flag's value for sending, or nil when the flag
// was left out or is empty.
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// optionalInt returns a pointer to the flag's value when the flag was given.
func optionalInt(set map[string]bool, name string, value int) *int {
	if !set[name] {
		return nil
	}
	return &value
}

// checkPaging reports a --limit or --offset the hub would refuse, as a usage
// error naming the flag. It returns false after printing one.
func checkPaging(stderr io.Writer, usage string, set map[string]bool, limit, offset int) bool {
	if set["limit"] && (limit < minLimit || limit > maxLimit) {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: --limit must be between %d and %d\n", minLimit, maxLimit)
		return false
	}
	if set["offset"] && (offset < 0 || offset > maxOffset) {
		fmt.Fprintln(stderr, usage)
		fmt.Fprintf(stderr, "error: --offset must be between 0 and %d\n", maxOffset)
		return false
	}
	return true
}

// RunSearch implements `sonora search radio-stations [name] [flags]`: it
// defines and parses the command's flags, searches the hub's radio catalogue
// with one request, and prints the page to stdout. A flag left out is not
// sent, so the hub applies its default. Failures go to stderr, never stdout.
func RunSearch(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("search radio-stations", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, SearchUsage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	tag := fs.String("tag", "", "only stations with this whole `TAG` (see 'sonora get radio-tags')")
	country := fs.String("country", "", "only stations of this ISO 3166-1 alpha-2 `CODE`")
	language := fs.String("language", "", "only stations in this whole `NAME`d language (see 'sonora get radio-languages')")
	codec := fs.String("codec", "", "only stations with this `CODEC` label, e.g. MP3 or AAC+")
	minBitrate := fs.Int("min-bitrate", 0, "only stations of at least this bitrate in kbit/s (`N`)")
	order := fs.String("order", "", "sort by `FIELD`: name, votes, clickCount or bitrate (hub default: votes)")
	asc := fs.Bool("asc", false, "sort from lowest to highest")
	desc := fs.Bool("desc", false, "sort from highest to lowest")
	limit := fs.Int("limit", 0, "page size, 1-100 (hub default: 20) (`N`)")
	offset := fs.Int("offset", 0, "where the page starts, 0-10000; use the previous page's nextOffset (`N`)")

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, SearchUsage)
		return 0
	}
	positional, err := clihelp.ParsePositional(fs, args)
	if err != nil {
		return exitcode.Usage
	}
	if len(positional) > 1 {
		fmt.Fprintln(stderr, SearchUsage)
		fmt.Fprintf(stderr, "error: expected at most one station name, got %d (quote a name with spaces)\n", len(positional))
		return exitcode.Usage
	}

	set := supplied(fs)
	if *asc && *desc {
		fmt.Fprintln(stderr, SearchUsage)
		fmt.Fprintln(stderr, "error: --asc and --desc cannot be used together")
		return exitcode.Usage
	}
	if !checkPaging(stderr, SearchUsage, set, *limit, *offset) {
		return exitcode.Usage
	}

	q := hub.StationSearch{
		Tag: optionalString(*tag), CountryCode: optionalString(*country), Language: optionalString(*language),
		Codec: optionalString(*codec), BitrateMin: optionalInt(set, "min-bitrate", *minBitrate),
		Order: optionalString(*order), Limit: optionalInt(set, "limit", *limit), Offset: optionalInt(set, "offset", *offset),
	}
	if len(positional) == 1 {
		q.Name = optionalString(positional[0])
	}
	switch {
	case *asc:
		d := hub.RadioDirectionAsc
		q.Direction = &d
	case *desc:
		d := hub.RadioDirectionDesc
		q.Direction = &d
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	page, err := hub.SearchStations(context.Background(), hub.NewClientWithTimeout(hub.RadioTimeout), baseURL, q)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}

	if *jsonOut {
		fmt.Fprint(stdout, render.RenderStationPageJSON(*page))
	} else {
		fmt.Fprint(stdout, render.RenderStationPageYAML(*page))
	}
	return 0
}
