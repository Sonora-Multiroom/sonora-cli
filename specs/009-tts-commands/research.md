# Research: Text-to-Speech Announcements (`speak`, TTS cache)

**Feature**: `009-tts-commands` | **Date**: 2026-09-21 | **Plan**: [plan.md](plan.md)

The Technical Context had no open `NEEDS CLARIFICATION` items: language, dependencies and test
layers are fixed by earlier features. The spec's clarification session settled exit codes,
required response fields, the `clear` output format and reading text from standard input.
This document records the design decisions that come from mapping that spec onto the existing
code.

---

## 1. Package layout: one new `internal/cli/tts` package, two new hub files

**Decision**: add:

- `internal/hub/tts.go`: `Speak`, `GetTTSCacheStats`, `ClearTTSCache`, their wire types, and
  the TTS error types (§3).
- `internal/hub/extensions.go`: `ListExtensions` plus the `Extension` and
  `ExtensionInventory` types, used only for the FR-010 diagnosis.
- `internal/cli/tts/`: `speak.go` (`RunSpeak`), `cache.go` (`RunGetCache`, `RunClearCache`),
  and `report.go` (the shared failure path: 404 → diagnosis → message → exit code).
- `internal/render/tts.go`: the YAML and JSON renderers for the three results.

**Rationale**: the three commands share one error model (the extension's `{error, message}`
shape, 404 meaning "not offered", diagnosis via the inventory), and nothing else in the CLI
shares it. Putting the commands in one package keeps that model in one place. The existing
per-resource packages (`outputs`, `groups`, …) each wrap one core-API resource, and TTS is
not one of those.

**Alternatives considered**: a `speak` package plus a cache package next to `mastermute`,
which would duplicate the failure path or push it into `hub`. Folding the commands into
`play`, which has no error-model overlap.

## 2. Dispatch: `speak` is a top-level special case; `tts-cache` is a singleton keyword

**Decision**:

- `cmd/sonora/main.go` routes `args[0] == "speak"` straight to
  `tts.RunSpeak(args[1:], os.Stdin, stdout, stderr)`, next to the existing `play`, `route`
  and `transfer` special cases.
- `dispatchGetList` gains a `tts-cache` keyword check before `respath.Parse`, exactly like
  `master-mute`. `list tts-cache` is a usage error pointing at `get tts-cache`.
- A new `case "clear":` calls `dispatchClear`. It accepts only `tts-cache` (FR-007); any
  other argument, a missing argument, or a respath resource is a usage error that names
  `tts-cache` as the only supported resource.
- `tts-cache` is **not** added to `respath`: it has no id, no alias, and no collection form,
  which is the same reason `master-mute` isn't there.

**Standard input plumbing**: `run(args, stdout, stderr)` keeps its signature. `main` passes
`os.Stdin` only into the `speak` branch, so the 34 existing `run(...)` call sites in
`cmd/sonora/main_test.go` stay unchanged. `RunSpeak` takes an `io.Reader`, so unit tests inject
standard input directly. Integration tests pipe it into the built binary (§9).

**Alternatives considered**: changing `run` to take `stdin`, which churns every existing test
for a single command's benefit. Having `RunSpeak` read `os.Stdin` itself, which can't be
unit-tested.

## 3. Error model: `TTSError`, `TTSNotOfferedError`, `TTSUnavailableError`

**Decision**: add three error types to `internal/hub`:

| Type | Produced by | Meaning |
| --- | --- | --- |
| `*TTSError{StatusCode, Code, Message}` | `Speak`/`ClearTTSCache`/`GetTTSCacheStats` on 400 or 503 | The extension rejected the request. `Code` is the machine-readable code; it may be empty (§4). |
| `*TTSNotOfferedError{}` | any TTS operation on 404 | The extension is not serving on this hub. Never shown to the user directly. |
| `*TTSUnavailableError{Diagnosis, Reason, LoadingDisabled, BaseURL, Cause}` | `tts.ReportError` after running the diagnosis (§5) | The user-facing "TTS not available" failure, carrying why. |

`hub.ClassifyError` gets two new branches, checked before the existing ones:

- `*TTSUnavailableError` → `ClassTTSUnavailable` (**13**), except that `DiagnosisHubAddress`
  maps to `ClassNetwork` (4) (FR-010a).
- `*TTSError` → by `Code` first: `TARGET_NOT_FOUND` → `ClassTargetNotFound` (12);
  `INVALID_REQUEST` → `ClassValidation` (6); `PROVIDER_NOT_FOUND` → `ClassNotFound` (5);
  `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR` and
  `FORMAT_NORMALIZATION_FAILED` → `ClassServiceUnavailable` (10). An unrecognised or empty
  code falls back to the status: 400 → 6, 503 → 10, anything else → `ClassHub` (3)
  (FR-012).

`ClassTTSUnavailable` is added to the `ErrorClass` enum with `ExitCode() == 13`. None of the
existing branches of `ClassifyError` change, so no existing command's exit codes move
(SC-006).

**Rationale**: `*NotFoundError` and `*APIError` both encode core-API semantics ("resource X
not found", problem-details `title`/`detail`), and TTS has neither. Reusing them would force
a TTS 404 through the "target not found" message that FR-010 forbids. A separate type also
keeps the machine-readable code available to the renderer (FR-011).

**Alternatives considered**: mapping `*TTSError` into `*APIError`, which loses the code and
routes 503 to the playback-service message "the hub's playback service is temporarily
unavailable", wrong for a provider failure. A single "TTS failure" class for everything,
which violates FR-012.

## 4. Error body decoding and fallbacks

**Decision**: on 400 or 503 (and on any other non-2xx that isn't 404), each TTS function
tries to decode the body as `{error, message}`:

- If `error` is non-empty, the result is `*TTSError{Code: error, Message: message}`, with the
  code kept verbatim even when unrecognised (FR-011, FR-016).
- If the body isn't that shape (empty, HTML from a proxy, or the core problem-details
  shape), the result is `*TTSError` with an empty code and a message taken, in order, from
  the core shape's `detail`, then its `title`. If neither is present there is no message,
  and the command prints the generic status line from
  [contracts/cli-tts.md](contracts/cli-tts.md#failure-messages-all-three-commands). The exit
  class then follows the status per §3.
- The body is read through `io.LimitReader` (64 KiB), so an oversized proxy error page
  can't flood memory or the terminal. Any message printed is trimmed to a single line.

**Rationale**: spec edge cases "error code the CLI does not recognise" and "error body in the
core API's shape, or no parseable body" both require a correct exit class plus a clean
message. Falling back to the status class (rather than blanket `ClassHub`) matches FR-012's
"unrecognised code falls back to the class implied by the HTTP status" and keeps 503 → 10
for scripts.

## 5. "TTS not available" diagnosis (FR-010, FR-010a)

**Decision**: `tts.ReportError(stderr, err, baseURL, verbose)` checks
`errors.As(err, *TTSNotOfferedError)`. If it matches,
it makes exactly one `hub.ListExtensions` call and builds a `*TTSUnavailableError`:

| Inventory outcome | `Diagnosis` | Message tail |
| --- | --- | --- |
| 200, no entry with `id == "tts"`, `loadingEnabled` true or absent | `NotInstalled` | "the TTS extension is not installed on this hub" |
| 200, no `tts` entry, `loadingEnabled == false` | `NotInstalled` | "the TTS extension is not installed on this hub (extension loading is switched off in the hub's configuration)" |
| `status: DISABLED` | `Disabled` | "the TTS extension is installed but disabled in the hub's configuration" |
| `status: REJECTED` | `Rejected` | "the TTS extension failed to load: <rejectionReason>". The reason is omitted when null or empty. |
| `status: INERT` | `Inert` | "the TTS extension is loaded but inactive" |
| `status: ACTIVE` | `VersionMismatch` | "the hub's TTS API does not match this CLI version" |
| any other `status` value | `Unknown` | none (generic) |
| inventory answers 404 | `HubAddress` | see below |
| network error, timeout, non-2xx other than 404, malformed body | `Unknown` | none; underlying error in `--verbose` only |

The generic head of the message is always "text-to-speech is not available on this hub". The
tail is appended after a colon when there is one.

`HubAddress` produces a different message, and exit code 4: "<URL> is not serving the
Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is
not installed or not loaded; set the correct address with --hub-url, MULTIROOM_URL, or the
config file".

Matching uses the machine id `tts` only, never `name` (spec Assumptions).

**Where it runs**: in `internal/cli/tts/report.go`, not inside the `hub` functions. `hub.Speak`
stays a faithful single-operation wrapper (Principle II), and the success path can't
accidentally trigger the lookup (FR-010a, SC-005a).

**Client for the lookup**: a fresh `hub.NewClient()` (standard 5 s bound), even for `speak`,
whose own client has the longer bound (§6). Both share `http.DefaultTransport`, so the
keep-alive connection from the failed TTS call is reused (Performance Standards, "connection
reuse").

**`loadingEnabled == false` refinement**: a plain "no `tts` entry" doesn't separate
"nobody installed it" from "the hub has extension loading switched off". The inventory
answers that at no extra cost, so the message says so (FR-010 has a row for it). Diagnosis
and exit class are the same as for "not installed".

**Output format**: every diagnosis is printed as `error: <message> (hub URL: <url>)`, except
`HubAddress`, whose message already starts with the URL and so is printed without the
suffix (contracts/cli-tts.md).

## 6. Request timeout for `speak`

**Decision**: add `hub.NewClientWithTimeout(d time.Duration)`. `NewClient()` becomes
`NewClientWithTimeout(requestTimeout)`, and its behaviour doesn't change. `internal/hub/tts.go`
exports `SpeakTimeout = 15 * time.Second`, and `RunSpeak` builds its client with it. The cache
commands and the inventory lookup keep the standard 5 s client (FR-013).

**Rationale**: 15 s = the hub's 10 s default provider timeout + 5 s margin (spec Assumptions),
so SC-003's "hub's timeout error, not a CLI timeout" holds. A single constructor parameter
keeps one timeout mechanism (`http.Client.Timeout`) rather than mixing in per-call contexts.

**Alternatives considered**: a `--timeout` flag, which the spec puts out of scope
(Assumptions). A context deadline around one call, which works but would be the only place
in the CLI that bounds a request differently.

## 7. Positional parsing, `--`, and standard input

**Decision**: `RunSpeak` uses `play`'s re-parse loop (flags before, between or after the two
positionals), with one fix: once a `--` terminator has been consumed, every remaining argument
is positional and the loop stops calling `fs.Parse`. The loop detects this by checking whether
the slice passed to `fs.Parse` contained `"--"` before the first returned positional.

- Go's `flag` package already treats a lone `-` as a positional argument, not a flag, so
  `sonora speak - outputs/kitchen` needs no special parsing.
- A text of exactly `-` means standard input (FR-003a). It is read with `io.ReadAll`, then
  `strings.TrimRight(s, "\r\n")` is applied. Inner line breaks are kept. A read error is a
  usage error (2).
- There is **no** size cap on standard input. FR-003 puts the length limit on the hub; the
  hub's own `INVALID_REQUEST` covers oversized text.
- Empty or whitespace-only text (after the trim, for standard input) is a usage error (2).
  The hub is not contacted.
- `--provider`, `--voice` and `--language` are `fs.String` flags. "Supplied" is detected with
  `fs.Visit`, and a supplied value that is empty or whitespace-only is a usage error (FR-004).
  `clear tts-cache --provider` applies the same rule (FR-007).
  Supplied values go into the request as pointers, so omitted ones are left out of the JSON
  entirely (`omitempty` on `*string`).

**Why fix `--` for `speak` only**: `play` has the same latent quirk, but changing it is out of
scope. For `speak` it matters, because free text beginning with a dash is a listed edge case.

## 8. Rendering

**Decision** (`internal/render/tts.go`), following `render/play.go`'s flat-record style:

- **Speak**: `announcementId` (quoted string), `cacheHit` (bool), `queueDepth` (int), in that
  order. JSON is the same three keys.
- **Cache stats**: `totalEntries`, `totalSizeBytes`, `maxSizeBytes`, then `entriesByProvider`
  as a nested mapping with provider names quoted and **sorted** for stable output. An empty
  breakdown renders as `entriesByProvider: {}` in YAML and `{}` in JSON, never `null` or an
  omitted key (User Story 4, scenario 2).
- **Clear**: `cleared: all`, or `cleared: provider` followed by `provider: "<name>"`. JSON:
  `{"cleared":"all"}` or `{"cleared":"provider","provider":"<name>"}`.

Sizes are shown as raw byte counts, the same numbers the hub reports (FR-006). No
human-readable rounding is added, since that would make the YAML output unparseable for
scripts.

## 9. Testing strategy (Principle VI)

Tests are written first and must fail before implementation:

- **Contract** (`tests/contract/tts_test.go`, `extensions_test.go`): `httptest.Server`s
  shaped from `speak`, `getCacheStats`, `clearCache` and `listExtensions` in
  `api/openapi.json`. They check the method, path, query (`providerName` present only when
  given), the request body (optional fields omitted), and decoding of every documented 202,
  200, 204, 400 and 503 body. They also cover each required-field omission (FR-005a),
  absent/null `entriesByProvider`, a 404 returning `*TTSNotOfferedError`, unknown codes,
  non-TTS-shaped error bodies, and `ListExtensions` decode/404.
- **Unit**: `cli_tts_speak_test.go` covers parsing (order, extra args, `--`, `-` with an
  injected reader, empty/whitespace/CRLF-trimmed text, empty flag values, target kinds and
  aliases). `cli_tts_cache_test.go` covers flags and extra args. `tts_report_test.go` is a
  table test of every diagnosis row in §5 and every code-to-class row in §3.
  `render_tts_test.go` checks field order, sorted providers, the `{}` empty map and strict
  JSON. `hub_errors_test.go` (or an extension of the existing one) asserts that exit code 13
  exists and that the existing classes are unchanged.
- **Integration** (`tests/integration/tts_test.go`): runs the built binary against a fake hub
  that serves `/api/tts/**` and `/api/v2/extensions` and **counts requests per path**. It
  checks one request on success (SC-005a); 404 followed by one inventory call; every exit
  code in the contract's table; stdin piping via a new `runCLIWithStdin` helper in the same
  file; and a slow-handler test showing a response after about 11 s is still received
  (SC-003). The slow test is skipped under `-short`, so it runs at most once per full
  run.
- `cmd/sonora/main_test.go`: help text lists `speak`, `get tts-cache` and `clear tts-cache`;
  the `clear` dispatcher's usage errors.

## 10. Startup cost (Principle I)

No new dependency. Nothing runs before argument parsing: clients are constructed inside the
`Run*` functions after flags, positionals and standard input are validated. Standard input is
read only when the text argument is `-`, and only after the target path has been validated,
so a typo in the target fails immediately instead of waiting on a pipe. No benchmark is
required: the startup path is unchanged apart from three string comparisons in `run`.
