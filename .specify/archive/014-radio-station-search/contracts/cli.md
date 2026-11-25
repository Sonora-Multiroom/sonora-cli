# CLI contract: radio commands

Every command takes `--json`, `--verbose`, `--hub-url URL` and `--help`/`-h`; flags and positionals may
interleave. Each makes exactly one radio request (10 s bound); after a "not offered" 404 it makes one
extension-inventory request (5 s bound). Output goes to standard output, errors to standard error.

## `sonora search radio-stations [name] [flags]`

| Flag | Meaning |
|---|---|
| `[name]` | part of the station name, any letter case (at most one positional; quote names with spaces) |
| `--tag TAG` | one whole tag (see `get radio-tags`) |
| `--country CODE` | ISO 3166-1 alpha-2 code |
| `--language NAME` | one whole language name (see `get radio-languages`) |
| `--codec CODEC` | the catalogue's codec label, e.g. MP3, AAC+ |
| `--min-bitrate N` | lowest bitrate in kbit/s |
| `--order FIELD` | `name`, `votes`, `clickCount` or `bitrate` (hub default `votes`) |
| `--asc` / `--desc` | sort direction (hub default: `desc` for votes and clickCount, `asc` otherwise) |
| `--limit N` | page size, 1–100 (hub default 20) |
| `--offset N` | where the page starts, 0–10000; pass the previous output's `nextOffset` |

YAML output (JSON has the same keys):

```yaml
stations:
  - uuid: "9617a958-0601-11e8-ae97-52543be04c81"
    uri: "radiobrowser://9617a958-0601-11e8-ae97-52543be04c81"
    name: "Kiss FM"
    favicon: null
    homepage: "https://www.kissfm.ua/"
    countryCode: "UA"
    languages: ["ukrainian"]
    tags: ["dance", "electronic"]
    codec: "MP3"
    bitrate: 128
    hls: false
    lastCheckOk: true
    votes: 1520
    clickCount: 37
nextOffset: 20
```

An empty page prints `# no stations found`, `stations: []` and `nextOffset: null`.

## `sonora get radio-stations/<uuid> [flags]`

Prints one station as a bare record (the fields above, no `stations:` wrapper). `list radio-stations/…`
and a bare `get|list radio-stations` are usage errors that name `sonora search radio-stations`.

## `sonora get|list radio-countries`, `sonora get|list radio-languages`

No paging flags. Output `countries:` / `languages:` lists of `{code, name, stationCount}` /
`{name, code, stationCount}`, largest first, in the hub's order. An id segment (`radio-countries/UA`) is a
usage error.

## `sonora get|list radio-tags [--name TEXT] [--limit N] [--offset N]`

Output `tags:` list of `{name, stationCount}` and `nextOffset`, as for search.

## Exit codes and messages

| Situation | Exit | Standard error |
|---|---|---|
| success, including an empty result | 0 | — |
| bad flag, two positionals, `--asc` with `--desc`, `--limit` outside 1–100, `--offset` outside 0–10000, bare `get radio-stations` | 2 | usage line and `error: …` |
| hub 400 `invalid-parameter` | 6 | `error: <hub detail> (hub URL: …)` |
| hub 404 `station-not-found` | 5 | `error: radio station not found: <uuid> (hub URL: …)` |
| hub 502 `catalogue-bad-answer` | 9 | `error: the radio catalogue gave an unusable answer: <detail> (hub URL: …)` |
| hub 503 `catalogue-unavailable` | 10 | `error: the radio catalogue is unavailable; retry in 30 s (hub URL: …)` (`retry later` without a readable `Retry-After`) |
| radio not served (any other 404) | 15 | `error: radio is not available on this hub: <reason> (hub URL: …)` |
| … and the inventory also 404 | 4 | `error: <url> is not serving the Multiroom Audio Hub API: …; set the correct address with --hub-url, MULTIROOM_URL, or the config file` |
| other radio problem (500, 405) | 3 | `error: <hub detail> (hub URL: …)` |
| malformed 2xx body | 3 | `error: hub returned an unexpected or malformed response (hub URL: …)` |
| hub unreachable or too slow | 4 | as every other command |

Reasons for exit 15: "the radio extension is not installed on this hub" (+ "(extension loading is
switched off in the hub's configuration)"), "… is installed but disabled in the hub's configuration",
"… failed to load: <reason>", "… is loaded but inactive", "the radio extension on this hub (version X)
cannot search stations; it needs 0.2.0 or later". An unreadable inventory leaves the bare "radio is not
available on this hub".

Every catalogue string in output and in error messages is escaped (research R7).

## Help

`helpText` in `cmd/sonora/main.go` gains:

```text
  search radio-stations [name]        Search the hub's radio catalogue
  get radio-stations/<uuid>           Fetch one radio station
  list radio-countries                List radio catalogue countries
  list radio-languages                List radio catalogue languages
  list radio-tags [--name TEXT]       List radio catalogue tags
```
