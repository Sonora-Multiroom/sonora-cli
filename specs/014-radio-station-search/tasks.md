# Tasks: Radio Station Search and Browse

**Input**: Design documents from `/specs/014-radio-station-search/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md,
contracts/hub-package.md, quickstart.md

**Tests**: Required. Constitution Principle VI: every test task comes before the code it covers and must
be seen to fail first.

**Organization**: grouped by user story. Phase 2 builds what every radio command shares (error classes,
the "radio not available" diagnosis, exit code 15, the CLI error path, terminal-safe escaping). US1
(search) also builds the shared HTTP plumbing in `hub/radio.go`; US2 and US3 reuse it and are independent
of each other. US4 (Go callers) checks the public-package rules over what US1 to US3 added.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: different files, no dependency on an unfinished task
- Paths are from the repository root `D:\projects-sonora\sonora-cli`
- Go checks run directly, not through `make` (this machine has no `make`):
  `gofmt -l .`, `go vet ./...`, `go test ./...`
- No `FR-`, `SC-`, `US` or research numbers in code or godoc (`tests/unit/hub_godoc_test.go` rejects them
  in `hub/`); `RADIO-002` may be cited in code comments.

---

## Phase 1: Setup

- [ ] T001 Confirm the baseline: run `gofmt -l .`, `go vet ./...` and `go test ./...` on branch
  `014-radio-station-search` and record commands, date and counts in
  `specs/014-radio-station-search/verification.md` (create it with a dated "Baseline" section).
- [ ] T002 Confirm `api/openapi.json` (`info.version` 0.1.35) declares what the feature mirrors: the
  operations `searchRadioStations`, `getRadioStation`, `listRadioCountries`, `listRadioLanguages`,
  `listRadioTags`; the schemas `RadioStation`, `RadioStationPage`, `RadioCountry`, `RadioLanguage`,
  `RadioTag`, `RadioTagPage`, `RadioProblem`; and `Extension.version`. Compare each property, `required`
  list and nullability with `data-model.md`; record the result in `verification.md` and fix
  `data-model.md` where the spec disagrees.

---

## Phase 2: Foundational (shared errors, diagnosis, exit code, escaping)

**Purpose**: everything the five commands share. Blocks Phases 3 to 5.

### Tests first (write, run, watch fail)

- [ ] T003 [P] Write failing `ClassifyError` cases in `tests/unit/hub_client_test.go`, building errors
  directly:
  - `*hub.APIError{StatusCode: 400, Type: "urn:multiroom:radio:error:invalid-parameter", Detail: "limit
    must be between 1 and 100", Parameter: "limit"}` → `hub.ClassValidation`, message is the detail;
  - `Type` `…:catalogue-bad-answer`, 502 → `hub.ClassSourceUnreachable`, message
    `the radio catalogue gave an unusable answer: <detail>`;
  - `Type` `…:catalogue-unavailable`, 503, `RetryAfter: 30*time.Second` →
    `hub.ClassServiceUnavailable`, message `the radio catalogue is unavailable; retry in 30 s`; with
    `RetryAfter` 0 the message ends `…is unavailable; retry later`;
  - `Type` `…:internal-error`, 500 → `hub.ClassHub`, message is the detail;
  - `*hub.RadioNotOfferedError{}` → new `hub.ClassRadioUnavailable`, `radio is not available on this hub`;
  - `*hub.RadioUnavailableError` with `DiagnosisHubAddress` → `hub.ClassNetwork`; any other diagnosis →
    `hub.ClassRadioUnavailable` with the error's own message;
  - an `*hub.APIError` 502/503 with an empty or non-radio `Type` keeps today's messages ("the audio
    source could not be reached", "the hub's playback service is temporarily unavailable"); every
    existing case in the file is unchanged.
- [ ] T004 [P] Write failing diagnosis tests in `tests/contract/extensions_test.go` with a fake
  `GET /api/v2/extensions`:
  - `hub.ListExtensions` decodes `version` into `Extension.Version` (`*string`; `null` and absent give
    nil);
  - `hub.DiagnoseRadioUnavailable` matches the extension by id `radio` (never by name) and gives
    `DiagnosisNotInstalled` (no `radio` entry; `LoadingDisabled` true when `loadingEnabled` is false),
    `DiagnosisDisabled` (`DISABLED`), `DiagnosisRejected` with `Reason` passed through `hub.SingleLine`
    (`REJECTED`), `DiagnosisInert` (`INERT`), `DiagnosisVersionMismatch` with `Version` set (`ACTIVE`),
    `DiagnosisHubAddress` with `BaseURL` (inventory 404), `DiagnosisUnknown` with `Cause` (inventory 500
    or malformed body); it makes exactly one request;
  - each diagnosis's `Error()` text, all starting `radio is not available on this hub`:
    `: the radio extension is not installed on this hub` (+ ` (extension loading is switched off in the
    hub's configuration)`), `: the radio extension is installed but disabled in the hub's
    configuration`, `: the radio extension failed to load: <reason>` (no `: <reason>` when empty),
    `: the radio extension is loaded but inactive`, `: the radio extension on this hub (version X) cannot
    search stations; it needs 0.2.0 or later` (`(version unknown)` when nil), and the bare head for
    `DiagnosisUnknown`; the hub-address text is the same as the TTS one;
  - every existing `DiagnoseTTSUnavailable` test still passes unchanged, and a variable declared as
    `hub.TTSDiagnosis` accepts `hub.DiagnosisDisabled` (the alias compiles).
- [ ] T005 [P] Add failing cases to `tests/unit/exitcode_test.go`: `hub.ClassRadioUnavailable` → 15;
  every existing class keeps its code; `hub.ClassRadioUnavailable` is numerically after
  `hub.ClassConflict`.
- [ ] T006 [P] Write failing tests in `tests/unit/radio_report_test.go` for `radio.ReportError(stderr,
  err, baseURL, verbose) int` (package `internal/cli/radio`), modelled on `tests/unit/tts_report_test.go`:
  on `*hub.RadioNotOfferedError` it makes exactly one inventory request and prints
  `error: radio is not available on this hub: <reason> (hub URL: <url>)`, exit 15; on any other error it
  makes no inventory request; the hub-address diagnosis prints `error: <url> is not serving the
  Multiroom Audio Hub API: …; set the correct address with --hub-url, MULTIROOM_URL, or the config file`
  without the `(hub URL: …)` suffix, exit 4; `--verbose` adds `detail: <err>` (the diagnosis `Cause`
  when there is one); a hub detail containing `\x1b[2J`, `\u009b` and `\u202e` is printed escaped
  (no raw ESC, C1 or format rune on stderr).
- [ ] T007 [P] Write failing tests in `tests/unit/render_radio_test.go` for an exported
  `render.SafeText(s string) string` used for message text: printable text (including non-ASCII letters
  such as "Київ") is unchanged; C0 controls, DEL, C1 controls (U+0080–U+009F) and Unicode format
  characters (category Cf, e.g. U+202E, U+200B) come out as `\xNN` / `\uNNNN` escapes; line breaks are
  escaped, never printed.

### Implementation

- [ ] T008 Add to `hub/errors.go`: `APIError` fields `Type string`, `Parameter string`,
  `RetryAfter time.Duration` with self-contained godoc; `ClassRadioUnavailable` appended after
  `ClassConflict`; the radio branches of T003 in `ClassifyError`, placed before the generic `*APIError`
  branch and selected by the `Type` prefix `urn:multiroom:radio:error:`. Make T003 pass.
- [ ] T009 In `hub/extensions.go` and `hub/tts.go`: add `Extension.Version *string` (`json:"version"`);
  introduce `type ExtensionDiagnosis int`, move the `Diagnosis*` constants to it, and make
  `TTSDiagnosis` an alias with godoc `TTSDiagnosis is the former name of ExtensionDiagnosis.` and a
  separate `Deprecated: use ExtensionDiagnosis. This alias will be removed in a later release.`
  paragraph; move the inventory walk of `DiagnoseTTSUnavailable` into unexported
  `diagnoseExtension(ctx, client, baseURL, id string)` and call it from `DiagnoseTTSUnavailable`; add
  `RadioNotOfferedError` (empty struct), `RadioUnavailableError{Diagnosis ExtensionDiagnosis; Reason
  string; LoadingDisabled bool; Version string; BaseURL string; Cause error}` with `Error()` and
  `Unwrap()`, and `DiagnoseRadioUnavailable` (never returns nil). TTS messages must not change. Make T004
  pass. Depends on T008 (both feed `ClassifyError`).
- [ ] T010 [P] Map `hub.ClassRadioUnavailable` to 15 in `internal/cli/exitcode/exitcode.go`. Make T005
  pass. Depends on T008.
- [ ] T011 [P] Add `SafeText` and the JSON escaping pass to `internal/render/radio.go` (new file):
  `SafeText` escapes as T007 says; `safeJSON([]byte) []byte` rewrites any DEL, C1 or Cf rune that
  `encoding/json` leaves raw as `\uNNNN`, keeping the JSON valid. Make T007 pass.
- [ ] T012 Add `internal/cli/radio/report.go` with package doc and `ReportError` as T006 describes:
  diagnose through `hub.DiagnoseRadioUnavailable` with a fresh `hub.NewClient()` (5 s), classify with
  `hub.ClassifyError`, pass the message through `hub.SingleLine` and `render.SafeText`, append the
  hub-address hint, return `exitcode.For(class)`. Make T006 pass. Depends on T009, T010, T011.

**Checkpoint**: `go test ./...` passes, including every TTS test unchanged.

---

## Phase 3: User Story 1 - Find a station and play it (P1) MVP

**Goal**: `hub.SearchStations` and `sonora search radio-stations [name]`.

**Independent Test**: search a test server by name and filters and compare request and output; live, a
name search on `multiroom.lan` lists stations whose `uri` `sonora play` accepts.

### Tests first (write, run, watch fail)

- [ ] T013 [P] [US1] Contract tests in `tests/contract/radio_test.go` for `hub.SearchStations` against a
  test server:
  - with every `StationSearch` field nil the request is `GET /api/radio/stations` with no query string;
  - each set field is sent once under its query name (`name`, `tag`, `countryCode`, `language`,
    `codec`, `bitrateMin`, `order`, `direction`, `limit`, `offset`), and only set fields; values with a
    space, `&`, `+`, `=` and non-ASCII ("Київ") are percent-encoded as UTF-8; the order constants send
    `name`, `votes`, `clickCount`, `bitrate`, the direction constants `asc`, `desc`;
  - a page fixture decodes every `RadioStation` field; its keys equal
    `specSchemaProperties(t, "RadioStation")` and `"RadioStationPage"`; marshalling a decoded station
    gives the same key set, with `null` for nil nullable fields;
  - `nextOffset` `null` and absent both give nil; `{"stations":[]}` is an empty page, not an error;
  - a missing `stations`, or a station missing any of the required `"clickCount", "hls", "languages",
    "lastCheckOk", "tags", "uri", "uuid", "votes"`, or an empty `uuid` or `uri`, is a `*hub.DecodeError`;
    `languages`/`tags` `[]` decode to empty slices;
  - errors: 400 `RadioProblem` → `*hub.APIError` with `Type`, `Title`, `Detail`, `Parameter`; 503 with
    `Retry-After: 30` → `RetryAfter` 30 s, with `Retry-After` absent or not an integer → 0; 502 and 500
    radio problems → `*hub.APIError` with their `Type`; 404 with a core problem body, with the radio
    `…:not-found` body, and with an empty body → `*hub.RadioNotOfferedError`; 400 with a non-JSON body
    → `*hub.StatusError`; an error body over 64 KiB is read only up to 64 KiB.
- [ ] T014 [P] [US1] Renderer tests in `tests/unit/render_radio_test.go` for `RenderStationPageYAML` and
  `RenderStationPageJSON`: exact YAML for a two-station page in the field order of `data-model.md`
  (`uuid`, `uri`, `name`, `favicon`, `homepage`, `countryCode`, `languages`, `tags`, `codec`, `bitrate`,
  `hls`, `lastCheckOk`, `votes`, `clickCount`), strings via `%q`, nil as `null`, lists as
  `["a", "b"]` (`[]` when empty), then `nextOffset: N` or `nextOffset: null`; an empty page prints
  `# no stations found`, `stations: []`, `nextOffset: null`; JSON is one object
  `{"stations":[…],"nextOffset":…}` plus a newline, with the same keys; a station whose `name`, `tags`
  and `homepage` hold `\x1b[2J`, `\u009b`, `\u202e` and `\u0007` produces no raw ESC, BEL, C1 or Cf rune
  in either format (SC-004), and the JSON still parses back to the original strings.
- [ ] T015 [P] [US1] Command tests in `tests/integration/radio_test.go` for `sonora search
  radio-stations` against a test server counting requests:
  - `kiss --tag pop --country UA --language ukrainian --codec MP3 --min-bitrate 128 --order bitrate
    --desc --limit 5 --offset 10` sends exactly those query values; flags before and after the name both
    work; no flags sends no query string;
  - `--asc` sends `direction=asc`; `--tag ""` sends no `tag`;
  - exit 2 with no request for: two positionals, `--asc --desc`, `--limit 0`, `--limit 101`,
    `--offset -1`, `--offset 10001`, `--limit abc`, an unknown flag; `search` with no resource and
    `search radio-tags` print `usage: sonora search radio-stations [name] [flags]`, exit 2;
    `search --help` and `search radio-stations --help` exit 0;
  - output equals the renderer's for the served page, YAML by default and JSON with `--json`; an empty
    page exits 0;
  - every row of the exit-code table in `contracts/cli.md` that applies to search: 400 → 6 with the
    hub's detail, 502 → 9, 503 → 10 with `retry in 30 s`, radio not served → 15 with exactly one extra
    `GET /api/v2/extensions`, inventory also 404 → 4, 500 radio problem → 3, malformed 200 → 3,
    unreachable hub → 4; on success exactly one request is made.

### Implementation

- [ ] T016 [US1] Create `hub/radio.go`: file doc citing RADIO-002; `RadioTimeout = 10 * time.Second`;
  `RadioStation` (data-model.md: `uuid` string required, `uri` string required, `name` `*string`
  "maxLength 256", `favicon`/`homepage` `*string` "maxLength 2048", `countryCode` `*string`,
  `languages` `[]string` "maxItems 10", `tags` `[]string` "maxItems 30", `codec` `*string` "maxLength 32",
  `bitrate` `*int32`, `hls` bool, `lastCheckOk` bool, `votes` int64, `clickCount` int64; nullable fields
  without `omitempty`) decoded through a wire struct with pointer required fields; `RadioStationPage`;
  `StationSearch` and the order/direction constants; `SearchStations`; and the shared unexported helpers
  used by every radio call: build the URL with `url.Values` in a fixed parameter order, send GET, on 2xx
  decode into the target, on non-2xx read at most 64 KiB and map 404 (`…:station-not-found` →
  `*NotFoundError{Resource: "radio station", ID: uuid}`, anything else → `*RadioNotOfferedError`), a
  `RadioProblem` body → `*APIError` with `Type`, `Title`, `Detail` (through `SingleLine`), `Parameter`,
  `RetryAfter` from an integer `Retry-After`, a non-JSON body → `*StatusError`. Godoc self-contained. Make
  T013 pass.
- [ ] T017 [US1] Add `RenderStationPageYAML` and `RenderStationPageJSON` (and an unexported
  per-station YAML writer reused by US2) to `internal/render/radio.go`, using `safeJSON` for JSON. Make
  T014 pass.
- [ ] T018 [US1] Add `internal/cli/radio/search.go` with `RunSearch(args, stdout, stderr) int`: flags
  `--tag`, `--country`, `--language`, `--codec`, `--min-bitrate`, `--order`, `--asc`, `--desc`,
  `--limit`, `--offset`, `--json`, `--verbose`, `--hub-url`; positionals through
  `clihelp.ParsePositional` (at most one, the name); a string flag is sent only when non-empty, an int
  flag only when given (track with `fs.Visit`); local checks exactly as research R4 (`--limit` 1–100,
  `--offset` 0–10000, not both `--asc` and `--desc`), each a usage error naming the flag; client
  `hub.NewClientWithTimeout(hub.RadioTimeout)`; errors through `radio.ReportError`. Depends on T012,
  T016, T017.
- [ ] T019 [US1] In `cmd/sonora/main.go` add a `search` case to `run` and `dispatchSearch`, which accepts
  only `radio-stations` and calls `radio.RunSearch`; anything else prints
  `usage: sonora search radio-stations [name] [flags]` and `error: search supports only radio-stations`,
  exit 2; `search --help` prints the usage, exit 0. Add the help lines from `contracts/cli.md` for
  `search radio-stations [name]` to `helpText` and the example
  `sonora search radio-stations "kiss fm" --country UA`. Make T015 pass.

**Checkpoint**: `go test ./...` passes; `sonora search radio-stations kiss` works against the mock hub
and, read-only, against `multiroom.lan`.

---

## Phase 4: User Story 2 - Look up one station (P2)

**Goal**: `hub.Station` and `sonora get radio-stations/<uuid>`.

**Independent Test**: look up a UUID from a search result and compare with the hub's answer.

### Tests first (write, run, watch fail)

- [ ] T020 [P] [US2] Contract tests in `tests/contract/radio_test.go` for `hub.Station`: the path is
  `/api/radio/stations/<uuid>` with the UUID path-escaped (a value with a space or `?` stays in the path
  segment); a station with `lastCheckOk: false` decodes; 404 `…:station-not-found` →
  `*hub.NotFoundError{Resource: "radio station", ID: <uuid>}`; 404 with a core body →
  `*hub.RadioNotOfferedError`; 400 `invalid-parameter` → `*hub.APIError` with `Parameter` `uuid`; a body
  missing a required field → `*hub.DecodeError`.
- [ ] T021 [P] [US2] Renderer tests in `tests/unit/render_radio_test.go` for `RenderStationYAML` (a bare
  record with the 14 fields, no `stations:` wrapper) and `RenderStationJSON` (one object plus newline),
  with the same escaping checks as T014.
- [ ] T022 [P] [US2] Command tests in `tests/integration/radio_test.go`:
  `get radio-stations/<uuid>` prints the station (YAML, `--json`); unknown UUID → exit 5,
  `error: radio station not found: <uuid> (hub URL: …)`; hub 400 → exit 6 with the hub's detail; radio
  not served → exit 15; usage errors with exit 2 and no request: `get radio-stations/` (empty id),
  `get radio-stations/a/b`, `list radio-stations/<uuid>`, and bare `get radio-stations` and
  `list radio-stations`, the last two with `error: use 'sonora search radio-stations' to find stations`.

### Implementation

- [ ] T023 [US2] Add `Station` to `hub/radio.go` (`url.PathEscape` on the UUID, otherwise sent as
  given). Make T020 pass. Depends on T016.
- [ ] T024 [US2] Add `RenderStationYAML` and `RenderStationJSON` to `internal/render/radio.go`, reusing the
  per-station writer from T017. Make T021 pass.
- [ ] T025 [US2] Add `internal/cli/radio/station.go` with `RunGetStation(uuid, args, stdout, stderr)
  int` (flags `--json`, `--verbose`, `--hub-url`; no positionals; client with `hub.RadioTimeout`; errors
  through `ReportError`). In `dispatchGetList` in `cmd/sonora/main.go`, before `respath.Parse`, route
  `radio-stations` and `radio-stations/…` as T022 requires; add the `get radio-stations/<uuid>` help
  line. Make T022 pass. Depends on T023, T024.

**Checkpoint**: `go test ./...` passes.

---

## Phase 5: User Story 3 - Browse countries, languages and tags (P2)

**Goal**: `hub.Countries`, `hub.Languages`, `hub.Tags` and `sonora get|list radio-countries`,
`radio-languages`, `radio-tags`.

**Independent Test**: list tags filtered by a fragment and use one as `--tag` in a search.

### Tests first (write, run, watch fail)

- [ ] T026 [P] [US3] Contract tests in `tests/contract/radio_test.go`: `hub.Countries` →
  `GET /api/radio/countries`, no query, decodes an array of `RadioCountry` (`code`, `name`,
  `stationCount`, all required) in the hub's order; `hub.Languages` → `GET /api/radio/languages`,
  `RadioLanguage` (`name` and `stationCount` required, `code` nullable); `hub.Tags` →
  `GET /api/radio/tags` with `name`, `limit`, `offset` sent once and only when set, decoding
  `RadioTagPage` (`tags` required, `nextOffset` nullable); fixture keys equal
  `specSchemaProperties` for each schema; a missing required field → `*hub.DecodeError`; a non-array 200
  for countries → `*hub.DecodeError`; 404 core body → `*hub.RadioNotOfferedError`; 503 →
  `*hub.APIError` with `RetryAfter`.
- [ ] T027 [P] [US3] Renderer tests in `tests/unit/render_radio_test.go` for countries
  (`countries:` list of `code`, `name`, `stationCount`), languages (`languages:` list of `name`, `code`
  or `null`, `stationCount`) and tag pages (`tags:` list of `name`, `stationCount`, then `nextOffset`),
  YAML and JSON; empty lists print `# no countries found` / `# no languages found` / `# no tags found`
  and `[]`; escaping checks as T014.
- [ ] T028 [P] [US3] Command tests in `tests/integration/radio_test.go`: `get` and `list` both work for
  `radio-countries`, `radio-languages` and `radio-tags`; `radio-tags --name jazz --limit 10 --offset 20`
  sends those values; `--limit`/`--offset` range checks as for search (exit 2, no request); an id segment
  (`radio-countries/UA`, `radio-tags/jazz`) and an unexpected positional are usage errors, exit 2;
  `--name`, `--limit` and `--offset` on `radio-countries` or `radio-languages` are unknown flags, exit 2;
  radio not served → exit 15; 503 → exit 10.

### Implementation

- [ ] T029 [US3] Add `RadioCountry`, `RadioLanguage`, `RadioTag`, `RadioTagPage`, `TagSearch`,
  `Countries`, `Languages` and `Tags` to `hub/radio.go`, using the shared helpers. Make T026 pass.
  Depends on T016.
- [ ] T030 [US3] Add the country, language and tag-page renderers to `internal/render/radio.go`. Make
  T027 pass.
- [ ] T031 [US3] Add `internal/cli/radio/lists.go` with `RunCountries(verb, args, …)`,
  `RunLanguages(verb, args, …)` and `RunTags(verb, args, …)` (tags: `--name`, `--limit`, `--offset`,
  with the same paging checks as search, shared through one unexported helper in the `radio` package).
  Route the three nouns in `dispatchGetList` in `cmd/sonora/main.go` for both verbs, rejecting an id
  segment; add the `list radio-countries|radio-languages` and `list radio-tags [--name TEXT]` help lines.
  Make T028 pass. Depends on T029, T030.

**Checkpoint**: `go test ./...` passes.

---

## Phase 6: User Story 4 - Search from Go code (P2)

**Goal**: the new `hub` surface meets the public-package rules. Behaviour tests are in T003, T004, T013,
T020, T026.

- [ ] T032 [US4] Run `tests/unit/hub_godoc_test.go` and `tests/unit/api_spec_test.go`. Confirm the godoc
  test covers every new exported identifier (`RadioTimeout`, the `Radio*` types, `StationSearch`,
  `TagSearch`, the order/direction constants, `SearchStations`, `Station`, `Countries`, `Languages`,
  `Tags`, `RadioNotOfferedError`, `RadioUnavailableError`, `DiagnoseRadioUnavailable`,
  `ExtensionDiagnosis`, `ClassRadioUnavailable`, and the new `APIError` and `Extension` fields); extend it
  if it does not enumerate them itself. Confirm `TTSDiagnosis`'s godoc has a `Deprecated:` paragraph.
  Fix failures in `hub/`.
- [ ] T033 [US4] Check `hub/` stays protocol-only and stdlib-only: `go list -deps -f '{{if not
  .Standard}}{{.ImportPath}}{{end}}' ./hub` lists only this module's own packages, and nothing in `hub/`
  imports `internal/`. Confirm that sonora-mcp's use of the diagnosis (`hub.DiagnosisHubAddress` and the
  other constants, `TTSUnavailableError.Diagnosis`) still compiles: in `D:\projects-sonora\sonora-mcp`,
  with a temporary, uncommitted `go.work` (or `replace`) pointing at this branch, run `go build ./...`
  and `go test ./...`, then remove the temporary file. Record both results in `verification.md`.

**Checkpoint**: sonora-mcp can adopt the radio calls from the next minor tag.

---

## Phase 7: Polish and follow-up records

- [ ] T034 [P] Update `README.md`: a radio section with the five commands, an example each
  (`search radio-stations "kiss fm" --country UA`, then `play radiobrowser://<uuid> outputs/office`),
  the search and tag flags, paging with `--limit`/`--offset` and `nextOffset`, the YAML and JSON output,
  the failure cases from `contracts/cli.md`, and that the commands need the hub's radio extension 0.2.0
  or later; add exit code 15 ("radio is not available on this hub") to the "Exit codes" table.
- [ ] T035 [P] Add a "Startup" line to `verification.md`: the `search` case and the radio name checks
  add no work before argument parsing and no request before the radio request (Principle I, FR-013); no
  benchmark is needed. Confirm by reading `run` and `dispatchGetList` in `cmd/sonora/main.go`.
- [ ] T036 [P] Write the sonora-mcp backlog entry `D:\projects-sonora\sonora-mcp\docs\backlog\adopt-radio-search.md`
  following that repository's existing items (header block `Status`, `Origin`, `Depends on`, `Effort`;
  Problem, Proposal, Open questions) and add a row to its `docs/backlog/INDEX.md` (P3, "API adoption",
  Small–Medium, "Needs a sonora-cli release first"). Content: after this feature's tag, bump with
  `go get github.com/Sonora-Multiroom/sonora-cli@vX.Y.Z`; add radio search and browse tools over
  `hub.SearchStations`, `hub.Station`, `hub.Countries`, `hub.Languages`, `hub.Tags`; map
  `hub.ClassRadioUnavailable` and the `*hub.RadioUnavailableError` reasons onto its error categories (as
  `internal/tools/errors.go` does for TTS); use `hub.ExtensionDiagnosis` wherever a diagnosis type is
  named and never `hub.TTSDiagnosis`, which is deprecated and will be removed (on 2026-11-21 sonora-mcp
  names only the `Diagnosis*` constants and the `Diagnosis` field, so this is a check plus new code).
  Origin: CLI-014 (this feature). Leave it uncommitted in sonora-mcp and tell the user.
- [ ] T037 [P] Write `docs/backlog/remove-tts-diagnosis-alias.md` in this repository (same header
  block and sections as the other items) and add its row to `docs/backlog/INDEX.md` (P4, "Cleanup",
  Small, "Only after sonora-mcp has adopted the release with `ExtensionDiagnosis`"): delete the
  `TTSDiagnosis` alias from `hub/`; depends on the sonora-mcp item from T036; a breaking change in a
  public package, so it is named in the release notes.
- [ ] T038 Update the backlog item this feature came from: delete `docs/backlog/radio-search.md` and its
  `INDEX.md` row, since the item ships here (the sonora-mcp part now lives in T036's entry and the
  sonora-mobile part in that repository's own `radio-stations.md`).
- [ ] T039 Run `gofmt -l .` (no output), `go vet ./...`, `go test ./...`, and `git ls-files --eol` (no
  tracked file `w/crlf`). Record commands, date and counts in `verification.md`.
- [ ] T040 Run the `quickstart.md` live steps 1 to 5 and 7 against `multiroom.lan` (read-only, allowed
  without asking) and record the results in `verification.md`. The play step (SC-001) makes sound: only
  with the user's permission, following the production rules in the repository-parent `AGENTS.md`
  (record state, `office` output, low volume, stop the route first, then restore, then the ALSA and
  metrics checks).
- [ ] T041 Mark every finished task `[X]` in this file.
- [ ] T042 Release (only when the user asks): after merge, cut a minor tag for the additive `hub` API so
  sonora-mcp can follow T036's entry; never move or re-create a published tag.

---

## Dependencies & Execution Order

- Phase 1, then Phase 2 (blocks everything after it). Inside Phase 2: T003–T007 together; T008, then
  T009 and T010 and T011 (T009 after T008), then T012.
- Phase 3 (US1) builds the shared radio HTTP helpers in T016, so T023 (US2) and T029 (US3) depend on it.
  Their tests (T020–T022, T026–T028) can be written before T016 lands.
- Phases 4 and 5 are independent of each other apart from edits to `hub/radio.go`,
  `internal/render/radio.go`, `cmd/sonora/main.go` and the test files they share: do those edits one after
  the other.
- Phase 6 needs Phases 3 to 5. Phase 7 last; T036 and T037 can be written any time after the plan, and
  T038 only once the feature is merged or about to be.
- Inside each story: tests, then `hub`, then renderers, then the command, then `main.go`. Watch each test
  fail before its implementation task.

### Parallel opportunities

- Phase 2 tests: T003, T004, T005, T006, T007 together; T010 and T011 together.
- Phase 3 tests: T013, T014, T015 together.
- Phase 4 tests T020–T022 and Phase 5 tests T026–T028 together, once Phase 2 is done.
- Phase 7: T034, T035, T036, T037 together.

## Implementation Strategy

1. T001–T012: baseline, shared errors, diagnosis, exit code, escaping. Every TTS test must still pass.
2. Phase 3 (MVP): search ships alone if needed; validate with the mock hub, then read-only live.
3. Phases 4 and 5: station lookup and browse lists, in either order.
4. Phase 6: public-package rules and the sonora-mcp compile check.
5. Phase 7: README, follow-up backlog entries, final checks, live run; tag only when asked.
