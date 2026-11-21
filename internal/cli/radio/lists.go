package radio

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/clihelp"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
	"github.com/Sonora-Multiroom/sonora-cli/internal/config"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

// listCommand is what `radio-countries`, `radio-languages` and `radio-tags`
// share: flags, hub URL, one request, and the failure path. paging says
// whether --name, --limit and --offset exist.
type listCommand struct {
	usage  string
	paging bool
	// fetch makes the one radio request and returns the YAML and JSON texts.
	fetch func(ctx context.Context, client *http.Client, baseURL string, q hub.TagSearch) (yaml, json string, err error)
}

func (c listCommand) run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("radio list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	clihelp.SetUsage(fs, stderr, c.usage)

	jsonOut := fs.Bool("json", false, "emit strict JSON instead of the default YAML")
	verbose := fs.Bool("verbose", false, "print the underlying error detail on failure")
	hubURLFlag := fs.String("hub-url", "", "hub base `URL` override")
	var name *string
	var limit, offset *int
	if c.paging {
		name = fs.String("name", "", "only tags containing this `TEXT`, in any letter case")
		limit = fs.Int("limit", 0, "page size, 1-100 (hub default: 20) (`N`)")
		offset = fs.Int("offset", 0, "where the page starts, 0-10000; use the previous page's nextOffset (`N`)")
	}

	if clihelp.Requested(args) {
		clihelp.PrintUsage(fs, stdout, c.usage)
		return 0
	}
	if err := fs.Parse(args); err != nil {
		return exitcode.Usage
	}
	if rest := fs.Args(); len(rest) > 0 {
		fmt.Fprintln(stderr, c.usage)
		fmt.Fprintf(stderr, "error: unexpected argument(s): %v\n", rest)
		return exitcode.Usage
	}

	var q hub.TagSearch
	if c.paging {
		set := supplied(fs)
		if !checkPaging(stderr, c.usage, set, *limit, *offset) {
			return exitcode.Usage
		}
		q = hub.TagSearch{Name: optionalString(*name), Limit: optionalInt(set, "limit", *limit), Offset: optionalInt(set, "offset", *offset)}
	}

	baseURL, err := config.ResolveHubURL(*hubURLFlag)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitcode.Usage
	}

	yaml, js, err := c.fetch(context.Background(), hub.NewClientWithTimeout(hub.RadioTimeout), baseURL, q)
	if err != nil {
		return ReportError(stderr, err, baseURL, *verbose)
	}
	if *jsonOut {
		fmt.Fprint(stdout, js)
	} else {
		fmt.Fprint(stdout, yaml)
	}
	return 0
}

// verbUsage names the verb the user typed in a usage line.
func verbUsage(verb, rest string) string { return "usage: sonora " + verb + " " + rest }

// RunCountries implements `sonora get|list radio-countries`: it prints every
// country of the hub's radio catalogue with its station count.
func RunCountries(verb string, args []string, stdout, stderr io.Writer) int {
	return listCommand{
		usage: verbUsage(verb, "radio-countries [flags]"),
		fetch: func(ctx context.Context, client *http.Client, baseURL string, _ hub.TagSearch) (string, string, error) {
			list, err := hub.Countries(ctx, client, baseURL)
			if err != nil {
				return "", "", err
			}
			return render.RenderCountriesYAML(list), render.RenderCountriesJSON(list), nil
		},
	}.run(args, stdout, stderr)
}

// RunLanguages implements `sonora get|list radio-languages`: it prints every
// language of the hub's radio catalogue with its station count.
func RunLanguages(verb string, args []string, stdout, stderr io.Writer) int {
	return listCommand{
		usage: verbUsage(verb, "radio-languages [flags]"),
		fetch: func(ctx context.Context, client *http.Client, baseURL string, _ hub.TagSearch) (string, string, error) {
			list, err := hub.Languages(ctx, client, baseURL)
			if err != nil {
				return "", "", err
			}
			return render.RenderLanguagesYAML(list), render.RenderLanguagesJSON(list), nil
		},
	}.run(args, stdout, stderr)
}

// RunTags implements `sonora get|list radio-tags [--name TEXT] [--limit N]
// [--offset N]`: it prints one page of the hub's radio catalogue tags.
func RunTags(verb string, args []string, stdout, stderr io.Writer) int {
	return listCommand{
		usage:  verbUsage(verb, "radio-tags [--name TEXT] [--limit N] [--offset N] [flags]"),
		paging: true,
		fetch: func(ctx context.Context, client *http.Client, baseURL string, q hub.TagSearch) (string, string, error) {
			page, err := hub.Tags(ctx, client, baseURL, q)
			if err != nil {
				return "", "", err
			}
			return render.RenderTagPageYAML(*page), render.RenderTagPageJSON(*page), nil
		},
	}.run(args, stdout, stderr)
}
