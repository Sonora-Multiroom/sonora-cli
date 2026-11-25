# Research: Radio Station Search and Browse

All facts below were read from `api/openapi.json` (hub 0.1.35, `info.version` 0.1.35), the extension's
contract `sonora-radio/specs/002-station-search-api/contracts/radio-rest-api.yaml` (0.2.0) and this
repository's code on 2026-11-21. No item was left open.

## R1. Is `api/openapi.json` current enough?

- **Decision**: use it as it is; no re-fetch.
- **Rationale**: it already carries the five `/api/radio/*` operations (`searchRadioStations`,
  `getRadioStation`, `listRadioCountries`, `listRadioLanguages`, `listRadioTags`) and the schemas
  `RadioStation`, `RadioStationPage`, `RadioCountry`, `RadioLanguage`, `RadioTag`, `RadioTagPage`,
  `RadioProblem`, matching the extension's 0.2.0 contract field for field. The 503's `Retry-After` is
  named in the response description only, not declared as a header; reading it is still described
  behaviour (Principle II), and the extension contract declares it. When the hub's own `/api-docs` next
  declares it as a header, the refreshed copy needs no code change.
- **Alternatives**: re-fetch from `multiroom.lan`: no change expected, and FR-014 asks only for a refresh
  when a shape is missing.

## R2. Request timeout

- **Decision**: a new `hub.RadioTimeout` of 10 s, used by every radio call; the inventory diagnosis keeps
  the standard 5 s client.
- **Rationale**: the hub bounds one radio request by its own budget, 5 s by default (radio FR-025, a
  setting), and answers 503 `catalogue-unavailable` when it runs out. With the CLI's standard 5 s bound
  the two race and the user would usually see "hub did not respond in time" instead of the hub's answer.
  This is the reason `hub.SpeakTimeout` exists. No `--timeout` flag: the hub's budget, not the provider,
  decides the wait, and a read gains nothing from waiting longer.
- **Alternatives**: the standard 5 s (races the budget); a `--timeout` flag (no use case).

## R3. Command layout and dispatch

- **Decision**: `search` is a new verb in `cmd/sonora/main.go`, taking `radio-stations` as its only
  resource. `get`/`list` gain `radio-stations/<uuid>` (get only), `radio-countries`, `radio-languages` and
  `radio-tags` (both verbs, as `chimes`), routed by name before `respath.Parse`, as `chimes`,
  `master-mute` and `tts-cache` are. The radio nouns are not added to `respath.ResourceKind`, because they
  are never a play or route target.
- **Bare `get radio-stations` / `list radio-stations`**: a usage error pointing to
  `sonora search radio-stations` (exit 2). It would otherwise be a search with no terms, which the spec
  rules out as surprising.
- **UUID argument**: the id segment of `radio-stations/<uuid>` is sent as given, path-escaped. It is not
  checked against `respath`'s id pattern or a UUID pattern: the hub validates it (400 `invalid-parameter`)
  and its message is shown (spec US2 scenario 3). It must be non-empty and contain no further `/`
  (usage error otherwise).
- **Alternatives**: a `radio` command group (rejected in clarification); adding the nouns to `respath`
  (they would become valid play targets).

## R4. Flags

| Flag | Command | Sent as | Local check |
|---|---|---|---|
| positional `[name]` | search | `name` | at most one positional |
| `--name` | `get radio-tags` | `name` | none |
| `--tag` | search | `tag` | none |
| `--country` | search | `countryCode` | none |
| `--language` | search | `language` | none |
| `--codec` | search | `codec` | none |
| `--min-bitrate N` | search | `bitrateMin` | none |
| `--order` | search | `order` | none |
| `--asc`, `--desc` | search | `direction` | not both |
| `--limit N` | search, tags | `limit` | 1–100 |
| `--offset N` | search, tags | `offset` | 0–10000 |

- **Decision**: only `--limit`, `--offset` and `--asc`/`--desc` are checked before the request; every
  other value goes to the hub unchanged and a refusal comes back as the hub's 400 message, which names the
  parameter. A flag left out is not sent, so the hub applies its default. An empty string flag
  (`--tag ""`) counts as not given.
- **Rationale**: FR-010 asks for page-size and offset checks before any request; checking the enums or
  the country pattern here too would copy hub rules that may change (Principle II forbids inventing
  behaviour). Paging values are checked because a wrong one is the commonest scripting mistake and the
  check costs nothing. `--offset` also gets the hub's 10000 cap, so the CLI never sends a value it knows
  the hub refuses.
- **Flag names**: `--country` rather than `--country-code` (users type a code either way; the help text
  says "ISO 3166-1 alpha-2 code"); `--min-bitrate` reads as English. `--asc`/`--desc` are booleans
  because `direction` has two values. `--limit`/`--offset` are new names in this CLI; no existing flag
  means the same thing, so no synonym rule is broken.
- **Alternatives**: `--direction asc|desc` (one more word to type, no gain); validating every enum
  locally (duplicates the hub).

## R5. Errors and their classes

The hub answers radio errors as `RadioProblem` (`type`, `title`, `detail`, `status`, `parameter`), with
`Retry-After` on 503.

- **Decision**: decode every non-2xx radio body into the existing `*hub.APIError`, which gains three
  fields: `Type`, `Parameter` and `RetryAfter` (a `time.Duration`, zero when absent). Two cases become
  other types:
  - 404 with `type` `urn:multiroom:radio:error:station-not-found` → `*hub.NotFoundError{Resource:
    "radio station", ID: <uuid>}`, which `ClassifyError` already maps to `ClassNotFound` (exit 5).
  - any other 404 (a core problem body from a hub without the extension, the extension's own `not-found`,
    or no body) → `*hub.RadioNotOfferedError`.
  A body that is not JSON keeps today's fallback, `*hub.StatusError`.
- **Classification** (new branch in `ClassifyError`, before the generic `*APIError` branch, chosen by
  `Type` prefix `urn:multiroom:radio:error:`):

  | Type | Status | Class (exit) | Message |
  |---|---|---|---|
  | `invalid-parameter` | 400 | `ClassValidation` (6) | the hub's detail |
  | `catalogue-bad-answer` | 502 | `ClassSourceUnreachable` (9) | "the radio catalogue gave an unusable answer" + detail |
  | `catalogue-unavailable` | 503 | `ClassServiceUnavailable` (10) | "the radio catalogue is unavailable; retry in N s" |
  | any other radio type | any | `ClassHub` (3) | the hub's detail |

  Without the prefix the existing `*APIError` mapping applies unchanged, so no existing message moves.
- **Rationale**: the spec (FR-003) and the backlog ask for the existing error type; adding fields to it
  is additive (Principle VII). Branching on `Type` keeps the core API's "audio source" and "playback
  service" messages away from radio failures, where they would be wrong. Exit codes 9 and 10 already mean
  "upstream failed" and "temporarily unavailable, retry".
- **Alternatives**: a separate `RadioError` type (contradicts FR-003, and every consumer would need a
  third error type to check); reusing `ClassHub` for 502/503 (loses the retry signal scripts need).

## R6. "Radio is not available" diagnosis, shared with TTS

`hub.DiagnoseTTSUnavailable` already maps the extension inventory onto a reason. Radio needs the same
mapping for the extension id `radio`.

- **Decision**:
  - Rename the diagnosis type to `ExtensionDiagnosis` and keep `TTSDiagnosis` as a type alias
    (`type TTSDiagnosis = ExtensionDiagnosis`) marked `Deprecated:`, so the constants
    (`DiagnosisNotInstalled`, …) and every existing caller compile unchanged and tools flag the old
    name. The alias is removed once sonora-mcp has moved (plan, "Follow-up records").
  - Move the inventory walk into an unexported `diagnoseExtension(ctx, client, baseURL, id)`, used by
    `DiagnoseTTSUnavailable` and a new `DiagnoseRadioUnavailable`. Behaviour for TTS stays byte-for-byte
    the same (its tests are the guard).
  - New `*hub.RadioUnavailableError{Diagnosis, Reason, LoadingDisabled, BaseURL, Version, Cause}` with
    messages starting "radio is not available on this hub". `DiagnosisVersionMismatch` (extension `ACTIVE`
    yet 404) reads "the radio extension on this hub (version X) cannot search stations; it needs 0.2.0 or
    later", using the inventory's `version`, which `hub.Extension` starts decoding (additive field).
  - New `ClassRadioUnavailable`, exit code **15**; `DiagnosisHubAddress` maps to `ClassNetwork`, as for
    TTS.
  - CLI: `internal/cli/radio/report.go` mirrors `tts.ReportError`: on `*hub.RadioNotOfferedError` it makes
    the one inventory call, appends the same `--hub-url` hint for the hub-address diagnosis, prints
    `error: … (hub URL: …)` and `detail:` with `--verbose`.
- **Rationale**: one inventory walk for two extensions; the alias keeps the public API backward
  compatible; a separate exit code lets scripts tell "radio missing" from "TTS missing" and from a real
  not-found, as SC-003 requires.
- **Alternatives**: copy the TTS diagnosis (two copies to keep in step); a generic
  `ExtensionUnavailableError` replacing `TTSUnavailableError` (breaking for sonora-mcp); reuse exit 13
  (its name and docs say TTS).

## R7. Output and terminal safety

- **YAML** (default): hand-written fixed-shape renderers like `render/inputs.go`: every field always
  emitted, strings with `%q`, null as `null`, lists as YAML flow sequences of quoted strings. Go's `%q`
  escapes every non-printable rune (C0, DEL, C1, and format characters such as U+202E), so catalogue text
  cannot drive the terminal.
- **JSON**: `json.Marshal` of the `hub` types, which already have the hub's property names, then one pass
  that rewrites any rune `%q` would escape and `encoding/json` does not (DEL, C1 controls, Unicode format
  characters) as `\uXXXX`. The result is still valid JSON with the same values.
- **Error messages**: hub details are passed through `hub.SingleLine` (as today) and the same rune
  escaping before printing.
- **Page output**: `{stations: [...], nextOffset: N|null}` and `{tags: [...], nextOffset: N|null}` in both
  formats, so a script reads the next offset from the output. An empty page prints `stations: []` with a
  `# no stations found` comment line in YAML, as `inputs` does.
- **Alternatives**: a table view (rejected in clarification: same fields as JSON); relying on the hub's
  stripping alone (FR-011 makes the CLI responsible).

## R8. Test fixtures

- **Decision**: hand-written JSON in the tests, each fixture's keys checked against `api.Spec` with the
  existing `specSchemaProperties` helper; one fixture holds escape sequences and C1/format characters for
  SC-004.
- **Rationale**: the spec's assumption; the helper already exists in `tests/contract`.

## R9. Release

`hub/` changes are additive (new types, functions, fields, constants, an alias): a minor tag after merge,
then the `sonora-mcp` bump. The exit code 15 is new CLI behaviour and goes into the README's exit-code
table.
