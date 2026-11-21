# Verification: Radio Station Search and Browse

## Baseline (2026-11-21, branch `014-radio-station-search`, before any change)

| Command | Result |
|---|---|
| `gofmt -l .` | no output, exit 0 |
| `go vet ./...` | exit 0 |
| `go test -count=1 ./...` | exit 0; 1074 top-level tests passed, 0 failed, 0 skipped (`go test -v`, counting `--- PASS` lines) |
| `git ls-files --eol` | 0 tracked files `w/crlf` |

## api/openapi.json check (T002, 2026-11-21)

`info.version` is 0.1.35. Read with a script over the spec:

- Operations present: `searchRadioStations` (query `name`, `tag`, `countryCode`, `language`, `codec`,
  `bitrateMin`, `order`, `direction`, `limit`, `offset`), `getRadioStation` (`uuid`), `listRadioCountries`,
  `listRadioLanguages`, `listRadioTags` (`name`, `limit`, `offset`).
- `RadioStation` required: `clickCount`, `hls`, `languages`, `lastCheckOk`, `tags`, `uri`, `uuid`, `votes`;
  nullable: `name`, `favicon`, `homepage`, `countryCode`, `codec`, `bitrate` (int32). `votes`,
  `clickCount` int64. Matches `data-model.md`.
- `RadioStationPage` required `stations`, `nextOffset` nullable int32. `RadioTagPage` required `tags`,
  `nextOffset` nullable. `RadioCountry` (`code`, `name`, `stationCount` all required),
  `RadioLanguage` (`name`, `stationCount` required, `code` nullable), `RadioTag` (`name`,
  `stationCount` required). `RadioProblem` required `type`, `title`, `detail`, `status`; optional
  `parameter`. `Extension.version` is a nullable string.
- No change to `data-model.md` was needed.

## Public-package checks (T032, T033; 2026-11-21)

- `go test ./tests/unit -run 'Godoc|APISpec'`: pass. `TestHubGodoc` walks every exported function, type,
  method, const and var in `hub/` through `go/doc`, so it covers every new identifier without an explicit
  list; it also rejects spec-artifact references in any comment. Added
  `TestHubGodoc_TTSDiagnosisAliasIsDeprecated`, which checks the alias's `Deprecated: use
  ExtensionDiagnosis.` paragraph.
- `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./hub` lists only
  `github.com/Sonora-Multiroom/sonora-cli/hub`; `grep -rn "internal/" hub/*.go` finds nothing.
- sonora-mcp compile check, without touching that repository: a `go.work` in the session scratchpad
  (`use` for `D:\projects-sonora\sonora-mcp` and this working tree), selected with `GOWORK=`. In
  `sonora-mcp`: `go build ./...` exit 0, `go vet ./...` exit 0. `go test ./...` exit 1: five failures in
  `internal/tools` (`TestConformance`, `TestGetGroup`, `TestListGroups`, `TestListGroupsIncludeDisabled`,
  `TestSetGroupEnabled`; output has extra `origin`, `playable`, `unplayableReason`, `system` fields).
  The same five fail, with the same pass/fail list, when the workspace uses a `git archive main` copy of this
  repository instead of the branch. They come from `main` being ahead of the v0.3.4 that sonora-mcp pins
  and are not caused by this feature. `git status --short` in sonora-mcp was identical before and after
  the check (it already held an uncommitted `docs/backlog/INDEX.md` change and `adopt-route-chime.md`).

## Self-review (T035)

- **Principle I**: `search` is a new `case` in `run` and `dispatchSearch` does a string compare, then calls
  `radio.RunSearch`. The radio nouns are checked in `dispatchRadio` by string compare before `respath.Parse`.
  Neither does I/O or parsing before the command's own flag parsing, and no request is made before the one
  radio request (an inventory request follows only a "not offered" 404). Read in `cmd/sonora/main.go`
  (`run`, `dispatchSearch`, `dispatchRadio`). No benchmark is needed.
- **Principle III**: `git diff main -- go.mod go.sum` is empty; `go list -deps ./hub` shows no third-party
  package.
- **Principle IV**: every radio call uses `hub.NewClientWithTimeout(hub.RadioTimeout)` (10 s) and the
  diagnosis uses `hub.NewClient()` (5 s); each client is built once per command. Every failure path has a
  message and exit code: `tests/integration/radio_test.go` `TestSearchRadio_Failures`,
  `TestSearchRadio_RadioNotServed_*`, `TestSearchRadio_HubUnreachable_ExitFour`,
  `TestGetRadioStation_*`, `TestRadioLists_Failures`. `--verbose` prints the cause
  (`TestSearchRadio_VerboseKeepsTheCause`, `TestRadioReportError_Verbose*`).
- **Principle VI**: for each phase the tests were written first and `go vet ./tests/...` was run before
  any implementation: it failed with `undefined: hub.SearchStations` / `hub.RadioStation` /
  `radio.ReportError` / `hub.Station` / `hub.Countries` (the tests could not compile, which is the red
  state for new API). The tests that exercise changed existing code (`SafeText` escaping of `\a`, the
  ordering of exit code 15) were run against the implementation and two of my own test expectations were
  corrected (the `\a` escape form, and a source file where an editor had turned `\u202e` into a literal
  character: it now uses escapes). No test was changed to fit an implementation otherwise. Commits were
  not made during the run, so the order is not in git history.

## Final checks (T039, 2026-11-21)

| Command | Result |
|---|---|
| `gofmt -l .` | no output |
| `go vet ./...` | exit 0 |
| `go test -count=1 ./...` | exit 0; 1151 top-level tests passed, 0 failed, 0 skipped (was 1074 at baseline) |
| `git ls-files --eol` | 0 tracked files `w/crlf`; untracked new files have no CR either |
| `git diff main -- go.mod go.sum` | empty |

## Live run on `multiroom.lan` (T040, 2026-11-21, read-only)

Binary built from this branch; hub radio extension answered. Each step was one request.

1. `sonora search radio-stations kiss --country UA --limit 3`: three Kiss stations with `uuid`, `uri` and
   all 14 fields, `nextOffset: 3`.
2. `--limit 5 --json`: 5 stations, `nextOffset` 5; `--offset 5`: 5 stations, `nextOffset` 10.
3. `get radio-stations/bd9c1301-562d-451f-94bf-6a76e3bb898e`: the station, exit 0. The all-zero UUID: exit 5,
   `error: radio station not found: 00000000-0000-0000-0000-000000000000 (hub URL: …)`. `not-a-uuid`: exit 6,
   `error: uuid must be a station UUID (hub URL: …)`.
4. `list radio-countries` (United States 7060 first), `list radio-languages` (english 11400 first),
   `list radio-tags --name jazz --limit 3` (jazz 1072, `nextOffset: 3`).
5. `search radio-stations --limit 101`: exit 2, `--limit must be between 1 and 100`, no request.
7. `sonora get routes` before and after: identical.

Not run: the "not available" reasons (step 6: needs the extension disabled and a hub restart) and
SC-001 (playing a found station: makes sound; needs the user's permission under the production rules).
