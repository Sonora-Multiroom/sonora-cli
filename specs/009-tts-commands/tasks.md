---

description: "Task list for `speak`, `get tts-cache`, `clear tts-cache`"
---

# Tasks: Text-to-Speech Announcements (`speak`, TTS cache)

**Input**: Design documents from `/specs/009-tts-commands/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-tts.md](contracts/cli-tts.md),
[quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory** in this project. Constitution Principle VI (Test-First,
NON-NEGOTIABLE) requires every command and API client method to have a test written and
watched to fail *before* its implementation. Every implementation task below is preceded by
the test task(s) that must land first and fail.

**Organization**: Tasks are grouped by user story from spec.md, in priority order
(P1 → P4). Each story is independently implementable and testable once Phase 2 is done.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: User story the task belongs to (US1–US5)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project, unchanged: `cmd/sonora/`, `internal/{hub,render,cli/tts}/`,
`tests/{unit,contract,integration}/`. Test packages are external (`package unit`,
`package contract`, `package integration`), so tests use exported API only.

---

## Phase 1: Setup

**Purpose**: Confirm the codebase is green before changing shared files.

- [ ] T001 Run `make check` (gofmt, go vet, `go test ./...`) from the repo root and confirm it passes on branch `009-tts-commands` before any change. Record any pre-existing failure in this file under Notes rather than fixing it as part of this feature.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the TTS error model, the new exit code, the configurable client timeout,
the extension inventory call and the shared failure path. Every story needs these, because
FR-010 to FR-012 apply to all three commands.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Phase 2 (write first, must fail) ⚠️

- [ ] T002 [P] Extend tests/unit/hub_client_test.go with:
  (a) `hub.ClassTTSUnavailable.ExitCode() == 13`, and all existing classes keep their codes (2, 3, 4, 5, 6, 8, 9, 10, 11, 12);
  (b) a table test of `hub.ClassifyError` for `*hub.TTSError`:
    - `TARGET_NOT_FOUND` → 12, `INVALID_REQUEST` → 6, `PROVIDER_NOT_FOUND` → 5;
    - `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR` and `FORMAT_NORMALIZATION_FAILED` → 10;
    - an unknown code or empty code gets 6 with status 400, 10 with 503 and 3 with 500;
    - the friendly message is `<code>: <message>` when there is a code, and `hub rejected the request (HTTP <status>): <message>` or `hub reported an error (HTTP <status>)` when there isn't (contracts/cli-tts.md, Failure messages);
  (c) `*hub.TTSUnavailableError` → 13 for every `hub.TTSDiagnosis` except `hub.DiagnosisHubAddress` → 4, with the exact message texts from contracts/cli-tts.md;
  (d) `hub.NewClientWithTimeout(50*time.Millisecond)` times out against a slow `httptest.Server`, and `hub.NewClient().Timeout == 5*time.Second`.
- [ ] T003 [P] Create tests/contract/extensions_test.go for `hub.ListExtensions(ctx, client, baseURL)`. It must cover:
  - `GET /api/v2/extensions`, decoding `loadingEnabled`, `extensions[].id`, `status` and `rejectionReason` (including a null reason);
  - a missing `extensions` array decodes as empty;
  - 404 returns an error that `errors.As` matches as `*hub.StatusError` with `StatusCode == 404`;
  - a non-JSON 200 body returns `*hub.DecodeError`.
  Shapes mirror `#/components/schemas/ExtensionInventory` and `#/components/schemas/Extension` in api/openapi.json.
- [ ] T004 [P] Create tests/unit/tts_report_test.go for `tts.ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int`. Point it at an `httptest.Server` serving `/api/v2/extensions`, and count requests.
  - Given `&hub.TTSNotOfferedError{}`, assert the exact stderr message and exit code for every row of the research.md §5 table: no `tts` entry; no entry with `loadingEnabled:false`; `DISABLED`; `REJECTED` with a reason; `REJECTED` with a null reason; `INERT`; `ACTIVE`; an unknown status; inventory 404 (exit 4, message names the base URL, `--hub-url`, `MULTIROOM_URL` and the config file); inventory 500; malformed inventory body; inventory unreachable (closed server). Every case except unreachable makes exactly one inventory request.
  - Matching uses `id == "tts"` only: an entry with `name: "tts"` and another id counts as absent.
  - Given any error other than `*hub.TTSNotOfferedError` (for example a `*hub.TTSError`), assert zero inventory requests.
  - With `verbose=true`, a `detail:` line is printed. Without it, the lookup's own error never appears.
  - Output format is `error: <message> (hub URL: <url>)`.

### Implementation for Phase 2

- [ ] T005 Add the TTS error types. In internal/hub/tts.go (new file, `package hub`):
  - `TTSError{StatusCode int; Code, Message string}`;
  - `TTSNotOfferedError{}`;
  - `TTSDiagnosis` (`DiagnosisUnknown`, `DiagnosisNotInstalled`, `DiagnosisDisabled`, `DiagnosisRejected`, `DiagnosisInert`, `DiagnosisVersionMismatch`, `DiagnosisHubAddress`);
  - `TTSUnavailableError{Diagnosis TTSDiagnosis; Reason string; LoadingDisabled bool; BaseURL string; Cause error}`, each with `Error()`, and `Unwrap()` returning `Cause`.
  In internal/hub/errors.go:
  - add `ClassTTSUnavailable` to the `ErrorClass` enum with `ExitCode() == 13`, and update the ExitCode doc comment;
  - add `ClassifyError` branches for `*TTSUnavailableError` and `*TTSError`, **before** the existing branches, producing the classes and messages in data-model.md (Exit codes) and contracts/cli-tts.md (Failure messages). Hub messages are collapsed to a single line.
  - Leave every existing branch unchanged. Makes T002 (a)–(c) pass.
- [ ] T006 [P] In internal/hub/client.go, add `NewClientWithTimeout(d time.Duration) *http.Client` (same `http.DefaultTransport`) and make `NewClient()` return `NewClientWithTimeout(requestTimeout)`. Makes T002 (d) pass.
- [ ] T007 [P] Create internal/hub/extensions.go:
  - `ExtensionInventory{LoadingEnabled *bool; Extensions []Extension}` and `Extension{ID, Status string; RejectionReason *string}`, with JSON tags per data-model.md;
  - `ListExtensions(ctx, client, baseURL) (*ExtensionInventory, error)`: `GET {baseURL}/api/v2/extensions`. Any non-2xx is a `*StatusError`, and an undecodable body is a `*DecodeError`.
  Makes T003 pass.
- [ ] T008 Create internal/cli/tts/report.go with `ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int`:
  - On `*hub.TTSNotOfferedError`, call `hub.ListExtensions` exactly once with a fresh `hub.NewClient()` and build a `*hub.TTSUnavailableError` from the result per research.md §5, setting `BaseURL` for `DiagnosisHubAddress` and `Cause` for lookup failures.
  - Then, for any error, classify with `hub.ClassifyError`, print `error: <msg> (hub URL: <baseURL>)`, add `detail: <err>` when verbose, and return `class.ExitCode()`.
  Depends on T005 and T007. Makes T004 pass.

**Checkpoint**: `go test ./tests/unit/ ./tests/contract/` passes for T002–T004, and existing tests are still green.

---

## Phase 3: User Story 1 - Announce text on a single output (Priority: P1) 🎯 MVP

**Goal**: `sonora speak <text|-> outputs/<id>` requests an announcement and prints
`announcementId`, `cacheHit` and `queueDepth` (YAML, or JSON with `--json`), returning on 202.

**Independent Test**: Against a fake hub serving `POST /api/tts/speak`,
`sonora speak "Dinner is ready" outputs/kitchen` exits 0 and prints the three fields. The fake
hub saw exactly one request with `targetType: SINGLE_OUTPUT`, `targetName: kitchen`. Empty text
exits 2 with no request.

### Tests for User Story 1 (write first, must fail) ⚠️

- [ ] T009 [P] [US1] Create tests/contract/tts_test.go with `hub.Speak(ctx, client, baseURL, hub.SpeakRequest)` contract tests:
  - Request: `POST /api/tts/speak`, `Content-Type: application/json`, and a body with exactly `text`, `targetName`, `targetType` when the optional pointers are nil (no `providerName`/`voice`/`language` keys).
  - 202: decodes `announcementId`, `cacheHit` and `queueDepth`. Each missing field (three cases), and a non-JSON body, return `*hub.DecodeError` (FR-005a).
  - 400 and 503: each documented code returns `*hub.TTSError` with the status, code and message.
  - An unknown code (`SOMETHING_NEW`) is kept verbatim.
  - A core-shaped body (`{"title","detail"}`), an HTML body and an empty body return a `*hub.TTSError` with an empty code, the message taken from `detail`, then `title`, else empty.
  - A 1 MiB error body is handled without error. At most 64 KiB is read, and the message is a single line.
  - 404 returns `*hub.TTSNotOfferedError`.
  - `hub.SpeakTimeout == 15*time.Second`.
- [ ] T010 [P] [US1] Create tests/unit/cli_tts_speak_test.go for `tts.RunSpeak(args, stdin io.Reader, stdout, stderr) int`, using an `httptest.Server` fake hub. Cover:
  - success with the YAML field order and values, and `--json` producing strict one-line JSON;
  - flags before, between and after positionals;
  - `-` reading text from an injected `strings.Reader`, trimming trailing `\n` and `\r\n` while keeping inner newlines;
  - `-` with empty or whitespace-only input (exit 2, no request);
  - a failing reader (exit 2, `could not read text from standard input`);
  - a literal `""` and `"   "` text (exit 2, `text must not be empty`, no request);
  - `--` followed by `"-5 degrees"` spoken as text, and a later `--json` after `--` treated as a positional (so exit 2, unexpected argument);
  - missing text, missing target and extra positional (exit 2, messages per contracts/cli-tts.md);
  - `--help` (stdout, exit 0);
  - a 404 from speak followed by a diagnosis, confirming `ReportError` is used (exit 13).
- [ ] T011 [P] [US1] Create tests/unit/render_tts_test.go for `render.RenderSpeakYAML(hub.SpeakAccepted)` and `render.RenderSpeakJSON`: exact YAML `announcementId: "<id>"\ncacheHit: <bool>\nqueueDepth: <n>\n`, and JSON with exactly those three keys, newline-terminated.
- [ ] T012 [P] [US1] Create tests/integration/tts_test.go. Add:
  - a `runCLIWithStdin(t, stdin string, args ...string) cliResult` helper, same env handling as `runCLI` in tests/integration/outputs_list_test.go, with `cmd.Stdin` set;
  - a `mockTTSHub(t)` fake that serves `/api/tts/speak` and `/api/v2/extensions`, lets each test script responses, and records per-path request counts plus the last speak body.
  Tests:
  - speak success: YAML, then `--json`, with exactly 1 request total and 0 to `/api/v2/extensions` (SC-005a);
  - `echo`-style piped text via `runCLIWithStdin` with `-`;
  - `TARGET_NOT_FOUND` → exit 12, stderr contains `TARGET_NOT_FOUND:` and the hub message;
  - `INVALID_REQUEST` (text too long) → 6;
  - `PROVIDER_TIMEOUT` → 10;
  - a 404 with the inventory listing `tts` as `DISABLED` → 13 and exactly 1 inventory request;
  - a 404 with the inventory also 404 → 4, and the message names the URL;
  - an unreachable hub (closed server URL) → 4;
  - a malformed 202 → 3;
  - `TestSpeak_SlowProviderStillReceived`: the handler sleeps 11 s and then answers 503 `PROVIDER_TIMEOUT`, expecting exit 10, not 4 (SC-003). Call `t.Skip` under `testing.Short()`.
- [ ] T013 [P] [US1] Extend cmd/sonora/main_test.go: `run([]string{"help"}, …)` output contains the `speak <text|-> <target>` line and a `sonora speak` example; `run([]string{"speak"}, …)` exits 2 with the speak usage line on stderr.

### Implementation for User Story 1

- [ ] T014 [US1] In internal/hub/tts.go, add:
  - `SpeakTimeout = 15 * time.Second`;
  - `SpeakRequest{Text, TargetName, TargetType string; ProviderName, Voice, Language *string}` with `omitempty` on the pointers;
  - `SpeakAccepted{AnnouncementID string; CacheHit bool; QueueDepth int32}`;
  - an unexported `decodeTTSError(resp *http.Response) error`, which reads at most 64 KiB, handles the `{error,message}` shape with a fallback to `detail`/`title`, and returns a `*TTSError`, per research.md §4;
  - `Speak(ctx, client, baseURL, req) (*SpeakAccepted, error)`. It posts JSON to `/api/tts/speak`. A 404 returns `*TTSNotOfferedError`. Any other non-2xx goes through `decodeTTSError`. A 2xx decodes into an unexported wire struct with `*string`/`*bool`/`*int32` fields, and a missing or empty field returns `*DecodeError`.
  Depends on T005. Makes T009 pass.
- [ ] T015 [P] [US1] Create internal/render/tts.go with `RenderSpeakYAML(hub.SpeakAccepted) string` and `RenderSpeakJSON(hub.SpeakAccepted) string`, in the style of internal/render/play.go. Makes T011 pass.
- [ ] T016 [US1] Create internal/cli/tts/speak.go with `RunSpeak(args []string, stdin io.Reader, stdout, stderr io.Writer) int`:
  - Define the flags `--json`, `--verbose` and `--hub-url`, and set up `clihelp.SetUsage`/`clihelp.Requested` with the usage line `usage: sonora speak <text|-> <outputs|groups>/<id> [flags]`.
  - Parse with `play`'s re-parse loop, except that once `--` has been consumed all remaining args are positional (research.md §7).
  - Require exactly 2 positionals. Parse the target with `respath.Parse`, require an id, and accept only `respath.Outputs` (→ `SINGLE_OUTPUT`) or `respath.Groups` (→ `OUTPUT_GROUP`); anything else is exit 2.
  - Validate the target before reading standard input. When the text is `-`, read with `io.ReadAll(stdin)` and trim trailing CR/LF.
  - Reject empty or whitespace-only text (exit 2).
  - Resolve the URL with `config.ResolveHubURL`, build the client with `hub.NewClientWithTimeout(hub.SpeakTimeout)` and call `hub.Speak`. On error, return `ReportError(...)`. On success, render YAML or JSON.
  Depends on T008, T014 and T015. Makes T010 pass.
- [ ] T017 [US1] In cmd/sonora/main.go, add `if args[0] == "speak" { return tts.RunSpeak(args[1:], os.Stdin, stdout, stderr) }` next to the `play`/`route`/`transfer` special cases. Add the helpText command line `speak <text|-> <target>` ("Speak text on an output or group via the hub's TTS extension") and the example `sonora speak "Dinner is ready" outputs/kitchen`. Depends on T016. Makes T012 and T013 pass.

**Checkpoint**: US1 is fully functional: the MVP. `go test ./...` is green, and `go test -short ./...` skips only the slow test.

---

## Phase 4: User Story 2 - Announce text on an output group (Priority: P2)

**Goal**: The same command works with `groups/<id>` and aliases (`gr/`, `out/`). Other kinds
and bare resources are usage errors before contacting the hub.

**Independent Test**: `sonora speak "Motion detected in garden" groups/all-rooms` exits 0 and
the fake hub receives `targetType: OUTPUT_GROUP`, `targetName: all-rooms`. `routes/x`,
`inputs/x` and `outputs` exit 2 with zero requests.

### Tests for User Story 2 (write first) ⚠️

- [ ] T018 [P] [US2] Extend tests/unit/cli_tts_speak_test.go with a table test of targets:
  - `groups/all-rooms` and `gr/all-rooms` → `OUTPUT_GROUP`/`all-rooms`;
  - `out/kitchen` → `SINGLE_OUTPUT`/`kitchen`;
  - `routes/x` and `inputs/x` → exit 2 with `speak target must be outputs/<id> or groups/<id>`;
  - `outputs` and `gr` (no id) → exit 2 with `must include an id`;
  - `bogus/x` → exit 2 with the respath error.
  Every usage case asserts zero hub requests.
- [ ] T019 [P] [US2] Extend tests/integration/tts_test.go: a `gr/all-rooms` target succeeds with the same YAML fields as US1 and the recorded body has `targetType: OUTPUT_GROUP`; `routes/x` exits 2 with 0 requests recorded.

### Implementation for User Story 2

- [ ] T020 [US2] Run T018 and T019 against `RunSpeak` from T016. If any case fails, fix the target-kind handling in internal/cli/tts/speak.go only. No new flags or branches are expected, because T016 already maps both kinds generically.

**Checkpoint**: US1 and US2 both pass independently.

---

## Phase 5: User Story 3 - Choose provider, voice, or language per announcement (Priority: P3)

**Goal**: `--provider`, `--voice` and `--language` are sent only when supplied. Empty values are
usage errors. An unknown provider gives `PROVIDER_NOT_FOUND` with exit 5, distinct from target
not found.

**Independent Test**: `sonora speak "Hi" out/kitchen --provider piper-local --voice
en_US-ryan-medium --language en-US` sends all three fields. Omitting `--voice` sends no `voice`
key. `--provider no-such` against a hub answering `PROVIDER_NOT_FOUND` exits 5.

### Tests for User Story 3 (write first, must fail) ⚠️

- [ ] T021 [P] [US3] Extend tests/contract/tts_test.go: a `hub.SpeakRequest` with `ProviderName`, `Voice` and `Language` set produces exactly those three extra keys with the given values, and a request with only `Voice` set produces only `voice`.
- [ ] T022 [P] [US3] Extend tests/unit/cli_tts_speak_test.go:
  - each of `--provider`, `--voice` and `--language`, given alone, is forwarded, and the other keys are absent from the recorded request body;
  - `--provider ""`, `--voice ""`, `--language "  "` → exit 2 with `--<flag> must not be empty` and zero requests;
  - `--help` lists all six flags.
- [ ] T023 [P] [US3] Extend tests/integration/tts_test.go: `--provider no-such` with the hub answering 400 `PROVIDER_NOT_FOUND` → exit 5, stderr contains `PROVIDER_NOT_FOUND:`; the same target with `TARGET_NOT_FOUND` → exit 12, confirming the two are distinct (US3 scenario 3).

### Implementation for User Story 3

- [ ] T024 [US3] In internal/cli/tts/speak.go, add the `--provider NAME`, `--voice ID` and `--language TAG` string flags, with usage text per contracts/cli-tts.md. Detect whether each was supplied with `fs.Visit`. A supplied value that is empty after `strings.TrimSpace` is exit 2 (`error: --<flag> must not be empty`), checked before any I/O. Set the matching `*string` field on `hub.SpeakRequest` only when supplied. Makes T021–T023 pass.

**Checkpoint**: US1–US3 pass. `speak` is feature-complete.

---

## Phase 6: User Story 4 - Inspect the TTS cache (Priority: P3)

**Goal**: `sonora get tts-cache` prints total entries, total size, max size and the
per-provider breakdown, in YAML (default) or JSON.

**Independent Test**: Against a fake hub serving `GET /api/tts/cache/stats`,
`sonora get tts-cache` exits 0 with the four fields and providers sorted. An empty cache (no
`entriesByProvider`) prints `entriesByProvider: {}`.

### Tests for User Story 4 (write first, must fail) ⚠️

- [ ] T025 [P] [US4] Extend tests/contract/tts_test.go with `hub.GetTTSCacheStats(ctx, client, baseURL)`:
  - `GET /api/tts/cache/stats`, decoding all four fields;
  - `entriesByProvider` absent and `null` both yield a non-nil empty map;
  - each missing required field (`totalEntries`, `totalSizeBytes`, `maxSizeBytes`) returns `*hub.DecodeError`;
  - 404 returns `*hub.TTSNotOfferedError`;
  - 500 with a `{error,message}` body returns `*hub.TTSError`.
- [ ] T026 [P] [US4] Extend tests/unit/render_tts_test.go for `render.RenderTTSCacheStatsYAML`/`JSON`:
  - exact YAML for a populated map, with providers sorted and names quoted, and for an empty map (`entriesByProvider: {}`);
  - JSON always contains an `entriesByProvider` object (`{}` when empty, never `null`);
  - `totalSizeBytes` above 2^31 renders exactly.
- [ ] T027 [P] [US4] Create tests/unit/cli_tts_cache_test.go for `tts.RunGetCache(args, stdout, stderr) int`: success YAML and `--json`; an extra positional → exit 2; `--help` → stdout, exit 0; a 404 → diagnosis via `ReportError` (exit 13).
- [ ] T028 [P] [US4] Extend tests/integration/tts_test.go and `mockTTSHub` with `/api/tts/cache/stats`:
  - `get tts-cache` success with exactly 1 request;
  - `--json`;
  - an empty cache;
  - a 404 with the inventory showing `REJECTED` and a reason → 13, with stderr containing the reason;
  - `list tts-cache` → exit 2 with `use 'sonora get tts-cache'`.
- [ ] T029 [P] [US4] Extend cmd/sonora/main_test.go: the help text contains `get tts-cache`, and `run([]string{"list", "tts-cache"}, …)` exits 2 with the contract's message.

### Implementation for User Story 4

- [ ] T030 [US4] In internal/hub/tts.go, add `TTSCacheStats{TotalEntries int32; TotalSizeBytes, MaxSizeBytes int64; EntriesByProvider map[string]int32}` and `GetTTSCacheStats(ctx, client, baseURL) (*TTSCacheStats, error)`.
  - A 404 returns `*TTSNotOfferedError`, and any other non-2xx goes through `decodeTTSError`.
  - Decode into a pointer-field wire struct, enforce the three required fields, and normalise a nil map to an empty map.
  Makes T025 pass.
- [ ] T031 [P] [US4] In internal/render/tts.go, add `RenderTTSCacheStatsYAML` and `RenderTTSCacheStatsJSON` per data-model.md (Rendered views): sorted provider keys via `sort.Strings`, and `{}` when empty. Makes T026 pass.
- [ ] T032 [US4] Create internal/cli/tts/cache.go with `RunGetCache(args, stdout, stderr) int`:
  - flags `--json`, `--verbose` and `--hub-url`, usage line `usage: sonora get tts-cache [flags]`;
  - reject extra positionals;
  - `hub.NewClient()` (5 s), `hub.GetTTSCacheStats`, `ReportError` on failure, then render.
  Depends on T008, T030 and T031. Makes T027 pass.
- [ ] T033 [US4] In cmd/sonora/main.go `dispatchGetList`, add a `tts-cache` keyword check next to `master-mute`. `get` calls `tts.RunGetCache(args[1:], …)`. `list` prints the usage line and `error: list does not support tts-cache; use 'sonora get tts-cache' instead`, then exits 2. Add the helpText line `get tts-cache` ("Fetch TTS audio-cache statistics") and the example `sonora get tts-cache`. Depends on T032. Makes T028 and T029 pass.

**Checkpoint**: US4 works independently of US1–US3 (it needs only Phase 2).

---

## Phase 7: User Story 5 - Clear the TTS cache (Priority: P4)

**Goal**: `sonora clear tts-cache [--provider NAME]` clears all entries, or one provider's, and
prints `cleared: all` or `cleared: provider` + `provider: "<name>"`. It is idempotent.

**Independent Test**: Against a fake hub serving `DELETE /api/tts/cache`,
`sonora clear tts-cache` exits 0 with `cleared: all` and no query string.
`--provider openai` sends `?providerName=openai` and prints both lines. A 400
`PROVIDER_NOT_FOUND` exits 5.

### Tests for User Story 5 (write first, must fail) ⚠️

- [ ] T034 [P] [US5] Extend tests/contract/tts_test.go with `hub.ClearTTSCache(ctx, client, baseURL, provider *string)`:
  - `DELETE /api/tts/cache`, with no query string when `provider` is nil;
  - `providerName=openai` when set, with a value containing a space or `&` URL-encoded;
  - 204 returns nil;
  - 400 `PROVIDER_NOT_FOUND` returns `*hub.TTSError`;
  - 404 returns `*hub.TTSNotOfferedError`.
- [ ] T035 [P] [US5] Extend tests/unit/render_tts_test.go for `render.RenderTTSCacheClearedYAML(provider *string)`/`JSON`: `cleared: all\n` and `{"cleared":"all"}`; `cleared: provider\nprovider: "openai"\n` and `{"cleared":"provider","provider":"openai"}`.
- [ ] T036 [P] [US5] Extend tests/unit/cli_tts_cache_test.go for `tts.RunClearCache(args, stdout, stderr) int`: with and without `--provider`, the recorded query and output match; `--provider ""` → exit 2 with zero requests; an extra positional → 2; `--json`; `--help` lists `--provider`.
- [ ] T037 [P] [US5] Extend tests/integration/tts_test.go and `mockTTSHub` with `DELETE /api/tts/cache`:
  - clear all;
  - clear with a provider;
  - clear twice, both exit 0 (FR-008);
  - an unknown provider → 5;
  - a 404 with the inventory showing `INERT` → 13;
  - `clear` with no argument, `clear outputs/x` and `clear routes` → exit 2 with `clear supports only tts-cache` and 0 requests.
- [ ] T038 [P] [US5] Extend cmd/sonora/main_test.go: the help text contains `clear tts-cache`; `run([]string{"clear"}, …)` and `run([]string{"clear", "master-mute"}, …)` exit 2 with the usage line `usage: sonora clear tts-cache [flags]`; `run([]string{"clear", "--help"}, …)` prints that usage on stdout and exits 0.

### Implementation for User Story 5

- [ ] T039 [US5] In internal/hub/tts.go, add `ClearTTSCache(ctx, client, baseURL string, provider *string) error`. It sends `DELETE /api/tts/cache`, adding `providerName` via `url.Values` only when `provider` is non-nil. A 404 returns `*TTSNotOfferedError`, any other non-2xx goes through `decodeTTSError`, and 2xx returns nil. Makes T034 pass.
- [ ] T040 [P] [US5] In internal/render/tts.go, add `RenderTTSCacheClearedYAML(provider *string)` and `RenderTTSCacheClearedJSON(provider *string)`, using a payload struct `{Cleared string "json:cleared"; Provider *string "json:provider,omitempty"}`. Makes T035 pass.
- [ ] T041 [US5] In internal/cli/tts/cache.go, add `RunClearCache(args, stdout, stderr) int`:
  - flags `--json`, `--verbose`, `--hub-url` and `--provider NAME`, usage line `usage: sonora clear tts-cache [flags]`;
  - an empty `--provider` is exit 2; extra positionals are exit 2;
  - `hub.NewClient()`, `hub.ClearTTSCache`, `ReportError` on failure, then render.
  Depends on T008, T039 and T040. Makes T036 pass.
- [ ] T042 [US5] In cmd/sonora/main.go, add `case "clear": return dispatchClear(args[1:], stdout, stderr)` and write `dispatchClear`:
  - `--help` with no resource prints usage to stdout and exits 0;
  - a first argument of `tts-cache` calls `tts.RunClearCache(args[1:], …)`;
  - no argument, or anything else, prints the usage line plus `error: clear supports only tts-cache` and exits 2.
  Add the helpText line `clear tts-cache` ("Clear the TTS audio cache (all, or one provider's)") and the example `sonora clear tts-cache --provider openai`. Depends on T041. Makes T037 and T038 pass.

**Checkpoint**: All five stories pass independently.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T043 [P] Update README.md: add `speak`, `get tts-cache` and `clear tts-cache` to the command table and examples, note that they need the hub's TTS extension, and document exit code 13 wherever the README lists exit codes (FR-015).
- [ ] T044 [P] Update docs/cli-command-landscape.md: add a "tts (extension)" section with ✅ rows mapping `sonora speak <text|-> <outputs|groups>/<id> [--provider] [--voice] [--language]` → `POST /api/tts/speak`, `sonora get tts-cache` → `GET /api/tts/cache/stats` and `sonora clear tts-cache [--provider]` → `DELETE /api/tts/cache`. Note that `listExtensions` is used internally only, and update the operationId count in the intro.
- [ ] T045 Review cmd/sonora/main.go `helpText` as a whole: the new lines are aligned with existing entries, and the `Commands:` column widths are unchanged (FR-015). Re-run `go test ./cmd/...`.
- [ ] T046 Run `make check` (gofmt, go vet, the full `go test ./...` including the slow SC-003 test) and fix any findings in files touched by this feature.
- [ ] T047 Regression (SC-006): diff the `sonora help` output and the exit codes of existing commands against `main`, using `go test ./tests/... ./cmd/...` on both branches. Confirm no existing test changed expectation apart from the additive help-text assertions.
- [ ] T048 Walk through specs/009-tts-commands/quickstart.md sections 1–2. Run section 3 against a real hub if one is available; otherwise record in this file under Notes which manual rows were not run.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: depends on Setup. **Blocks all stories.** Every command uses
  `ReportError`, the TTS error types and exit 13.
- **US1 (Phase 3)**: depends on Phase 2.
- **US2 (Phase 4)**: depends on US1's `RunSpeak` (T016) and dispatch (T017). It adds tests and
  at most a fix.
- **US3 (Phase 5)**: depends on US1 (T016). It is independent of US2 and can run in parallel
  with it.
- **US4 (Phase 6)**: depends on Phase 2 only. It is independent of US1–US3, apart from sharing
  files: `internal/hub/tts.go` and `internal/render/tts.go` need T014/T015 to have created
  them, or US4 creates them if run first.
- **US5 (Phase 7)**: depends on Phase 2. It shares `internal/cli/tts/cache.go` with US4
  (T041 edits the file T032 creates), so run it after US4 or create the file in whichever
  lands first.
- **Polish (Phase 8)**: after all stories.

### Within Each Story

- Test tasks first. They must compile-fail or fail before implementation starts
  (Principle VI).
- hub function → render → `Run*` → main.go dispatch.
- T005 → T008. T006 and T007 are independent of T005. T008 needs T005 and T007.

### Parallel Opportunities

- Phase 2 tests: T002, T003 and T004 are different files.
- Phase 2 implementation: T006 and T007 in parallel with T005. T008 comes last.
- US1 tests: T009–T013 are five different files. US1 implementation: T015 in parallel with T014.
- US2 (T018, T019) and US3 (T021–T023) tests in parallel with each other, after US1.
- US4 tests T025–T029 and implementation T031 in parallel with T030.
- US5 tests T034–T038 and implementation T040 in parallel with T039.
- Polish: T043 and T044 in parallel.

---

## Parallel Example: User Story 1

```bash
# All US1 tests together (different files):
Task: "Speak contract tests in tests/contract/tts_test.go"
Task: "RunSpeak unit tests in tests/unit/cli_tts_speak_test.go"
Task: "Speak render tests in tests/unit/render_tts_test.go"
Task: "Speak integration tests + runCLIWithStdin in tests/integration/tts_test.go"
Task: "Help-text/speak usage tests in cmd/sonora/main_test.go"

# Then, in parallel:
Task: "hub.Speak + decodeTTSError in internal/hub/tts.go"
Task: "RenderSpeakYAML/JSON in internal/render/tts.go"

# Then sequentially:
Task: "RunSpeak in internal/cli/tts/speak.go"
Task: "speak dispatch + helpText in cmd/sonora/main.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1: Setup (T001).
2. Phase 2: Foundational (T002–T008). This gives the error model, exit 13, the inventory
   diagnosis and the 15 s client.
3. Phase 3: US1 (T009–T017).
4. **Stop and validate**: quickstart rows 1, 2, 4, 5, 7, 8, 13 and 14. `speak` to one output
   is shippable.

### Incremental Delivery

1. MVP (US1), then US2 (group targets, mostly tests), then US3 (per-announcement overrides).
   At that point `speak` is complete.
2. US4 (`get tts-cache`), then US5 (`clear tts-cache`). The cache commands can be delivered
   independently of `speak`, after Phase 2.
3. Polish: docs, help text, full `make check`, regression, quickstart.

## Notes

- [P] = different files, no dependency on incomplete tasks.
- Verify each test fails before implementing. Commit after each task or logical group, using
  Conventional Commits (`feat:`, `test:`, `docs:`), with no attribution trailers (AGENTS.md).
- No new dependency (Principle III). No change to existing `ClassifyError` branches or
  existing exit codes (SC-006).
