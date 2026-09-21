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
- **[Story]**: User story the task belongs to (US1–US6)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project, unchanged: `cmd/sonora/`, `internal/{hub,render,cli/tts}/`,
`tests/{unit,contract,integration}/`. Test packages are external (`package unit`,
`package contract`, `package integration`), so tests use exported API only.

---

## Phase 1: Setup

**Purpose**: Confirm the codebase is green before changing shared files.

- [X] T001 Run `make check` (gofmt, go vet, `go test ./...`) from the repo root and confirm it passes on branch `009-tts-commands` before any change. Record any pre-existing failure in this file under Notes rather than fixing it as part of this feature.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the TTS error model, the new exit code, the configurable client timeout,
the extension inventory call and the shared failure path. Every story needs these, because
FR-010 to FR-012 apply to all three commands.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Phase 2 (write first, must fail) ⚠️

- [X] T002 [P] Extend tests/unit/hub_client_test.go with:
  (a) `hub.ClassTTSUnavailable.ExitCode() == 13`, and all existing classes keep their codes (2, 3, 4, 5, 6, 8, 9, 10, 11, 12);
  (b) a table test of `hub.ClassifyError` for `*hub.TTSError`:
    - `TARGET_NOT_FOUND` → 12, `INVALID_REQUEST` → 6, `PROVIDER_NOT_FOUND` → 5;
    - `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR` and `FORMAT_NORMALIZATION_FAILED` → 10;
    - an unknown code or empty code gets 6 with status 400, 10 with 503 and 3 with 500;
    - the friendly message is `<code>: <message>` when there is a code, and `hub rejected the request (HTTP <status>): <message>` or `hub reported an error (HTTP <status>)` when there isn't (contracts/cli-tts.md, Failure messages);
  (c) `*hub.TTSUnavailableError` → 13 for every `hub.TTSDiagnosis` except `hub.DiagnosisHubAddress` → 4, with the exact message texts from contracts/cli-tts.md;
  (d) `hub.NewClientWithTimeout(50*time.Millisecond)` times out against a slow `httptest.Server`, and `hub.NewClient().Timeout == 5*time.Second`.
- [X] T003 [P] Create tests/contract/extensions_test.go for `hub.ListExtensions(ctx, client, baseURL)`. It must cover:
  - `GET /api/v2/extensions`, decoding `loadingEnabled`, `extensions[].id`, `status` and `rejectionReason` (including a null reason);
  - a missing `extensions` array decodes as empty;
  - 404 returns an error that `errors.As` matches as `*hub.StatusError` with `StatusCode == 404`;
  - a non-JSON 200 body returns `*hub.DecodeError`.
  Shapes mirror `#/components/schemas/ExtensionInventory` and `#/components/schemas/Extension` in api/openapi.json.
- [X] T004 [P] Create tests/unit/tts_report_test.go for `tts.ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int`. Point it at an `httptest.Server` serving `/api/v2/extensions`, and count requests.
  - Given `&hub.TTSNotOfferedError{}`, assert the exact stderr message and exit code for every row of the research.md §5 table: no `tts` entry; no entry with `loadingEnabled:false`; `DISABLED`; `REJECTED` with a reason; `REJECTED` with a null reason; `INERT`; `ACTIVE`; an unknown status; inventory 404 (exit 4, message names the base URL, `--hub-url`, `MULTIROOM_URL` and the config file); inventory 500; malformed inventory body; inventory unreachable (closed server). Every case except unreachable makes exactly one inventory request.
  - Matching uses `id == "tts"` only: an entry with `name: "tts"` and another id counts as absent.
  - Given any error other than `*hub.TTSNotOfferedError` (for example a `*hub.TTSError`), assert zero inventory requests.
  - With `verbose=true`, a `detail:` line is printed. Without it, the lookup's own error never appears.
  - Output format is `error: <message> (hub URL: <url>)`, except the inventory-404 case, which is `error: <message>` with no suffix (the URL appears exactly once).

### Implementation for Phase 2

- [X] T005 Add the TTS error types. In internal/hub/tts.go (new file, `package hub`):
  - `TTSError{StatusCode int; Code, Message string}`;
  - `TTSNotOfferedError{}`;
  - `TTSDiagnosis` (`DiagnosisUnknown`, `DiagnosisNotInstalled`, `DiagnosisDisabled`, `DiagnosisRejected`, `DiagnosisInert`, `DiagnosisVersionMismatch`, `DiagnosisHubAddress`);
  - `TTSUnavailableError{Diagnosis TTSDiagnosis; Reason string; LoadingDisabled bool; BaseURL string; Cause error}`, each with `Error()`, and `Unwrap()` returning `Cause`.
  In internal/hub/errors.go:
  - add `ClassTTSUnavailable` to the `ErrorClass` enum with `ExitCode() == 13`, and update the ExitCode doc comment;
  - add `ClassifyError` branches for `*TTSUnavailableError` and `*TTSError`, **before** the existing branches, producing the classes and messages in data-model.md (Exit codes) and contracts/cli-tts.md (Failure messages). Hub messages are collapsed to a single line.
  - Leave every existing branch unchanged. Makes T002 (a)–(c) pass.
- [X] T006 [P] In internal/hub/client.go, add `NewClientWithTimeout(d time.Duration) *http.Client` (same `http.DefaultTransport`) and make `NewClient()` return `NewClientWithTimeout(requestTimeout)`. Makes T002 (d) pass.
- [X] T007 [P] Create internal/hub/extensions.go:
  - `ExtensionInventory{LoadingEnabled *bool; Extensions []Extension}` and `Extension{ID, Status string; RejectionReason *string}`, with JSON tags per data-model.md;
  - `ListExtensions(ctx, client, baseURL) (*ExtensionInventory, error)`: `GET {baseURL}/api/v2/extensions`. Any non-2xx is a `*StatusError`, and an undecodable body is a `*DecodeError`.
  Makes T003 pass.
- [X] T008 Create internal/cli/tts/report.go with `ReportError(stderr io.Writer, err error, baseURL string, verbose bool) int`:
  - On `*hub.TTSNotOfferedError`, call `hub.ListExtensions` exactly once with a fresh `hub.NewClient()` and build a `*hub.TTSUnavailableError` from the result per research.md §5, setting `BaseURL` for `DiagnosisHubAddress` and `Cause` for lookup failures.
  - Then, for any error, classify with `hub.ClassifyError`, print `error: <msg> (hub URL: <baseURL>)` (just `error: <msg>` for `DiagnosisHubAddress`, whose message already starts with the URL), add `detail: <err>` when verbose, and return `class.ExitCode()`.
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

- [X] T009 [P] [US1] Create tests/contract/tts_test.go with `hub.Speak(ctx, client, baseURL, hub.SpeakRequest)` contract tests:
  - Request: `POST /api/tts/speak`, `Content-Type: application/json`, and a body with exactly `text`, `targetName`, `targetType` when the optional pointers are nil (no `providerName`/`voice`/`language` keys).
  - 202: decodes `announcementId`, `cacheHit` and `queueDepth`. Each missing field (three cases), an empty `announcementId` (`""`), and a non-JSON body, return `*hub.DecodeError` (FR-005a).
  - 400 and 503: each documented code returns `*hub.TTSError` with the status, code and message.
  - An unknown code (`SOMETHING_NEW`) is kept verbatim.
  - A core-shaped body (`{"title","detail"}`), an HTML body and an empty body return a `*hub.TTSError` with an empty code, the message taken from `detail`, then `title`, else empty.
  - A 1 MiB error body is handled without error. At most 64 KiB is read, and the message is a single line.
  - 404 returns `*hub.TTSNotOfferedError`.
  - `hub.SpeakTimeout == 15*time.Second`.
- [X] T010 [P] [US1] Create tests/unit/cli_tts_speak_test.go for `tts.RunSpeak(args, stdin io.Reader, stdout, stderr) int`, using an `httptest.Server` fake hub. Cover:
  - success with the YAML field order and values, and `--json` producing strict one-line JSON;
  - flags before, between and after positionals;
  - `-` reading text from an injected `strings.Reader`, trimming trailing `\n` and `\r\n` while keeping inner newlines;
  - `-` with empty or whitespace-only input (exit 2, no request);
  - a failing reader (exit 2, `could not read text from standard input`);
  - a literal `""` and `"   "` text (exit 2, `text must not be empty`, no request);
  - `--` followed by `"-5 degrees"` spoken as text, and a later `--json` after `--` treated as a positional (so exit 2, unexpected argument);
  - missing text, missing target and extra positional (exit 2, messages per contracts/cli-tts.md);
  - positionals in the wrong order (`speak outputs/kitchen "Hi"`) → exit 2 with the respath error for `Hi`, no request;
  - a table test of non-group targets: `out/kitchen` → `SINGLE_OUTPUT`/`kitchen`; `routes/x` and `inputs/x` → exit 2 with `speak target must be outputs/<id> or groups/<id>`; `outputs` and `gr` (no id) → exit 2 with `must include an id`; `bogus/x` → exit 2 with the respath error. Every usage case asserts zero hub requests. (Group targets are US2: T018.)
  - `--help` (stdout, exit 0);
  - a 404 from speak followed by a diagnosis, confirming `ReportError` is used (exit 13).
- [X] T011 [P] [US1] Create tests/unit/render_tts_test.go for `render.RenderSpeakYAML(hub.SpeakAccepted)` and `render.RenderSpeakJSON`: exact YAML `announcementId: "<id>"\ncacheHit: <bool>\nqueueDepth: <n>\n`, and JSON with exactly those three keys, newline-terminated.
- [X] T012 [P] [US1] Create tests/integration/tts_test.go. Add:
  - a `runCLIWithStdin(t, stdin string, args ...string) cliResult` helper, same env handling as `runCLI` in tests/integration/outputs_list_test.go, with `cmd.Stdin` set;
  - a `mockTTSHub(t)` fake that serves `/api/tts/speak` and `/api/v2/extensions`, lets each test script responses, and records per-path request counts plus the last speak body.
  Tests:
  - speak success: YAML, then `--json`, with exactly 1 request total and 0 to `/api/v2/extensions` (SC-005a);
  - `echo`-style piped text via `runCLIWithStdin` with `-`;
  - `TARGET_NOT_FOUND` → exit 12, stderr contains `TARGET_NOT_FOUND:` and the hub message;
  - `INVALID_REQUEST` (text too long) → 6;
  - `PROVIDER_TIMEOUT` → 10;
  - no retries (FR-014): each of the three failure cases above records exactly 1 request to `/api/tts/speak`;
  - `routes/x` → exit 2 with 0 requests recorded;
  - a 404 with the inventory listing `tts` as `DISABLED` → 13 and exactly 1 inventory request;
  - a 404 with the inventory also 404 → 4, and the message names the URL;
  - an unreachable hub (closed server URL) → 4;
  - a malformed 202 → 3;
  - `TestSpeak_SlowProviderStillReceived`: the handler sleeps 11 s and then answers 503 `PROVIDER_TIMEOUT`, expecting exit 10, not 4 (SC-003). Call `t.Skip` under `testing.Short()`.
- [X] T013 [P] [US1] Extend cmd/sonora/main_test.go: `run([]string{"help"}, …)` output contains the `speak <text|-> <target>` line and a `sonora speak` example; `run([]string{"speak"}, …)` exits 2 with the speak usage line on stderr.

### Implementation for User Story 1

- [X] T014 [US1] In internal/hub/tts.go, add:
  - `SpeakTimeout = 15 * time.Second`;
  - `SpeakRequest{Text, TargetName, TargetType string; ProviderName, Voice, Language *string}` with `omitempty` on the pointers;
  - `SpeakAccepted{AnnouncementID string; CacheHit bool; QueueDepth int32}`;
  - an unexported `decodeTTSError(resp *http.Response) error`, which reads at most 64 KiB, handles the `{error,message}` shape with a fallback to `detail`/`title`, and returns a `*TTSError`, per research.md §4;
  - `Speak(ctx, client, baseURL, req) (*SpeakAccepted, error)`. It posts JSON to `/api/tts/speak`. A 404 returns `*TTSNotOfferedError`. Any other non-2xx goes through `decodeTTSError`. A 2xx decodes into an unexported wire struct with `*string`/`*bool`/`*int32` fields, and a missing or empty field returns `*DecodeError`.
  Depends on T005. Makes T009 pass.
- [X] T015 [P] [US1] Create internal/render/tts.go with `RenderSpeakYAML(hub.SpeakAccepted) string` and `RenderSpeakJSON(hub.SpeakAccepted) string`, in the style of internal/render/play.go. Makes T011 pass.
- [X] T016 [US1] Create internal/cli/tts/speak.go with `RunSpeak(args []string, stdin io.Reader, stdout, stderr io.Writer) int`:
  - Define the flags `--json`, `--verbose` and `--hub-url`, and set up `clihelp.SetUsage`/`clihelp.Requested` with the usage line `usage: sonora speak <text|-> <outputs|groups>/<id> [flags]`.
  - Parse with `play`'s re-parse loop, except that once `--` has been consumed all remaining args are positional (research.md §7).
  - Require exactly 2 positionals. Parse the target with `respath.Parse`, require an id, and accept only `respath.Outputs` (→ `SINGLE_OUTPUT`); anything else is exit 2 with `speak target must be outputs/<id> or groups/<id>`. Group targets are added in US2 (T020), after their tests fail.
  - Validate the target before reading standard input. When the text is `-`, read with `io.ReadAll(stdin)` and trim trailing CR/LF.
  - Reject empty or whitespace-only text (exit 2).
  - Resolve the URL with `config.ResolveHubURL`, build the client with `hub.NewClientWithTimeout(hub.SpeakTimeout)` and call `hub.Speak`. On error, return `ReportError(...)`. On success, render YAML or JSON.
  Depends on T008, T014 and T015. Makes T010 pass.
- [X] T017 [US1] In cmd/sonora/main.go, add `if args[0] == "speak" { return tts.RunSpeak(args[1:], os.Stdin, stdout, stderr) }` next to the `play`/`route`/`transfer` special cases. Add the helpText command line `speak <text|-> <target>` ("Speak text on an output or group via the hub's TTS extension") and the example `sonora speak "Dinner is ready" outputs/kitchen`. Depends on T016. Makes T012 and T013 pass.

**Checkpoint**: US1 is fully functional: the MVP. `go test ./...` is green, and `go test -short ./...` skips only the slow test.

---

## Phase 4: User Story 2 - Announce text on an output group (Priority: P2)

**Goal**: The same command works with `groups/<id>` and its alias `gr/<id>`. Rejecting other
kinds and bare resources, and the `out/` alias, are already covered by US1 (T010, T012).

**Independent Test**: `sonora speak "Motion detected in garden" groups/all-rooms` exits 0 and
the fake hub receives `targetType: OUTPUT_GROUP`, `targetName: all-rooms`.

### Tests for User Story 2 (write first, must fail) ⚠️

- [X] T018 [P] [US2] Extend tests/unit/cli_tts_speak_test.go with a table test of group targets: `groups/all-rooms` and `gr/all-rooms` → exit 0, recorded body `targetType: OUTPUT_GROUP`, `targetName: all-rooms`. These fail against T016, which accepts outputs only.
- [X] T019 [P] [US2] Extend tests/integration/tts_test.go: a `gr/all-rooms` target succeeds with the same YAML fields as US1, and the recorded body has `targetType: OUTPUT_GROUP`.

### Implementation for User Story 2

- [X] T020 [US2] In internal/cli/tts/speak.go, also accept `respath.Groups` (→ `OUTPUT_GROUP`) in the target-kind check. Other kinds keep the T016 error. Depends on T016. Makes T018 and T019 pass, and the T010 target table still passes.

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

- [X] T021 [P] [US3] Extend tests/contract/tts_test.go: a `hub.SpeakRequest` with `ProviderName`, `Voice` and `Language` set produces exactly those three extra keys with the given values, and a request with only `Voice` set produces only `voice`.
- [X] T022 [P] [US3] Extend tests/unit/cli_tts_speak_test.go:
  - each of `--provider`, `--voice` and `--language`, given alone, is forwarded, and the other keys are absent from the recorded request body;
  - `--provider ""`, `--voice ""`, `--language "  "` → exit 2 with `--<flag> must not be empty` and zero requests;
  - `--help` lists all six flags.
- [X] T023 [P] [US3] Extend tests/integration/tts_test.go: `--provider no-such` with the hub answering 400 `PROVIDER_NOT_FOUND` → exit 5, stderr contains `PROVIDER_NOT_FOUND:`; the same target with `TARGET_NOT_FOUND` → exit 12, confirming the two are distinct (US3 scenario 3).

### Implementation for User Story 3

- [X] T024 [US3] In internal/cli/tts/speak.go, add the `--provider NAME`, `--voice ID` and `--language TAG` string flags, with usage text per contracts/cli-tts.md. Detect whether each was supplied with `fs.Visit`. A supplied value that is empty after `strings.TrimSpace` is exit 2 (`error: --<flag> must not be empty`), checked before any I/O. Set the matching `*string` field on `hub.SpeakRequest` only when supplied. Makes T021–T023 pass.

**Checkpoint**: US1–US3 pass. `speak` is feature-complete.

---

## Phase 6: User Story 4 - Inspect the TTS cache (Priority: P3)

**Goal**: `sonora get tts-cache` prints total entries, total size, max size and the
per-provider breakdown, in YAML (default) or JSON.

**Independent Test**: Against a fake hub serving `GET /api/tts/cache/stats`,
`sonora get tts-cache` exits 0 with the four fields and providers sorted. An empty cache (no
`entriesByProvider`) prints `entriesByProvider: {}`.

### Tests for User Story 4 (write first, must fail) ⚠️

- [X] T025 [P] [US4] Extend tests/contract/tts_test.go (create it if US1's T009 has not run) with `hub.GetTTSCacheStats(ctx, client, baseURL)`:
  - `GET /api/tts/cache/stats`, decoding all four fields;
  - `entriesByProvider` absent and `null` both yield a non-nil empty map;
  - each missing required field (`totalEntries`, `totalSizeBytes`, `maxSizeBytes`) returns `*hub.DecodeError`;
  - 404 returns `*hub.TTSNotOfferedError`;
  - 500 with a `{error,message}` body returns `*hub.TTSError`.
- [X] T026 [P] [US4] Extend tests/unit/render_tts_test.go (create it if US1's T011 has not run) for `render.RenderTTSCacheStatsYAML`/`JSON`:
  - exact YAML for a populated map, with providers sorted and names quoted, and for an empty map (`entriesByProvider: {}`);
  - JSON always contains an `entriesByProvider` object (`{}` when empty, never `null`);
  - `totalSizeBytes` above 2^31 renders exactly.
- [X] T027 [P] [US4] Create tests/unit/cli_tts_cache_test.go for `tts.RunGetCache(args, stdout, stderr) int`: success YAML and `--json`; an extra positional → exit 2; `--help` → stdout, exit 0; a 404 → diagnosis via `ReportError` (exit 13).
- [X] T028 [P] [US4] Extend tests/integration/tts_test.go and `mockTTSHub` with `/api/tts/cache/stats`. If US1's T012 has not run, first create the file with `runCLIWithStdin` and `mockTTSHub` as T012 describes (request counting per path, scriptable responses, `/api/v2/extensions`).
  - `get tts-cache` success with exactly 1 request;
  - `--json`;
  - an empty cache;
  - a 404 with the inventory showing `REJECTED` and a reason → 13, with stderr containing the reason;
  - `list tts-cache` → exit 2 with `use 'sonora get tts-cache'`.
- [X] T029 [P] [US4] Extend cmd/sonora/main_test.go: the help text contains `get tts-cache`; `run([]string{"list", "tts-cache"}, …)` exits 2 with the contract's message; `run([]string{"get", "tts-cache/x"}, …)` exits 2 (not recognised as a resource, contracts/cli-tts.md).

### Implementation for User Story 4

- [X] T030 [US4] In internal/hub/tts.go, add `TTSCacheStats{TotalEntries int32; TotalSizeBytes, MaxSizeBytes int64; EntriesByProvider map[string]int32}` and `GetTTSCacheStats(ctx, client, baseURL) (*TTSCacheStats, error)`.
  - A 404 returns `*TTSNotOfferedError`, and any other non-2xx goes through `decodeTTSError`.
  - Decode into a pointer-field wire struct, enforce the three required fields, and normalise a nil map to an empty map.
  Makes T025 pass.
- [X] T031 [P] [US4] In internal/render/tts.go, add `RenderTTSCacheStatsYAML` and `RenderTTSCacheStatsJSON` per data-model.md (Rendered views): sorted provider keys via `sort.Strings`, and `{}` when empty. Makes T026 pass.
- [X] T032 [US4] Create internal/cli/tts/cache.go with `RunGetCache(args, stdout, stderr) int`:
  - flags `--json`, `--verbose` and `--hub-url`, usage line `usage: sonora get tts-cache [flags]`;
  - reject extra positionals;
  - `hub.NewClient()` (5 s), `hub.GetTTSCacheStats`, `ReportError` on failure, then render.
  Depends on T008, T030 and T031. Makes T027 pass.
- [X] T033 [US4] In cmd/sonora/main.go `dispatchGetList`, add a `tts-cache` keyword check next to `master-mute`. `get` calls `tts.RunGetCache(args[1:], …)`. `list` prints the usage line and `error: list does not support tts-cache; use 'sonora get tts-cache' instead`, then exits 2. Add the helpText line `get tts-cache` ("Fetch TTS audio-cache statistics") and the example `sonora get tts-cache`. Depends on T032. Makes T028 and T029 pass.

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

- [X] T034 [P] [US5] Extend tests/contract/tts_test.go (create it if neither T009 nor T025 has run) with `hub.ClearTTSCache(ctx, client, baseURL, provider *string)`:
  - `DELETE /api/tts/cache`, with no query string when `provider` is nil;
  - `providerName=openai` when set, with a value containing a space or `&` URL-encoded;
  - 204 returns nil;
  - 400 `PROVIDER_NOT_FOUND` returns `*hub.TTSError`;
  - 404 returns `*hub.TTSNotOfferedError`.
- [X] T035 [P] [US5] Extend tests/unit/render_tts_test.go (create it if neither T011 nor T026 has run) for `render.RenderTTSCacheClearedYAML(provider *string)`/`JSON`: `cleared: all\n` and `{"cleared":"all"}`; `cleared: provider\nprovider: "openai"\n` and `{"cleared":"provider","provider":"openai"}`.
- [X] T036 [P] [US5] Extend tests/unit/cli_tts_cache_test.go (create it if US4's T027 has not run) for `tts.RunClearCache(args, stdout, stderr) int`: with and without `--provider`, the recorded query and output match; `--provider ""` and `--provider "  "` → exit 2 with `--provider must not be empty` and zero requests; an extra positional → 2; `--json`; `--help` lists `--provider`.
- [X] T037 [P] [US5] Extend tests/integration/tts_test.go and `mockTTSHub` with `DELETE /api/tts/cache` (create the file and helpers as T012 describes if neither T012 nor T028 has run):
  - clear all;
  - clear with a provider;
  - clear twice, both exit 0 (FR-008);
  - an unknown provider → 5;
  - a 404 with the inventory showing `INERT` → 13;
  - `clear` with no argument, `clear outputs/x` and `clear routes` → exit 2 with `clear supports only tts-cache` and 0 requests.
- [X] T038 [P] [US5] Extend cmd/sonora/main_test.go: the help text contains `clear tts-cache`; `run([]string{"clear"}, …)` and `run([]string{"clear", "master-mute"}, …)` exit 2 with the usage line `usage: sonora clear tts-cache [flags]`; `run([]string{"clear", "--help"}, …)` prints that usage on stdout and exits 0.

### Implementation for User Story 5

- [X] T039 [US5] In internal/hub/tts.go, add `ClearTTSCache(ctx, client, baseURL string, provider *string) error`. It sends `DELETE /api/tts/cache`, adding `providerName` via `url.Values` only when `provider` is non-nil. A 404 returns `*TTSNotOfferedError`, any other non-2xx goes through `decodeTTSError`, and 2xx returns nil. Makes T034 pass.
- [X] T040 [P] [US5] In internal/render/tts.go, add `RenderTTSCacheClearedYAML(provider *string)` and `RenderTTSCacheClearedJSON(provider *string)`, using a payload struct `{Cleared string "json:cleared"; Provider *string "json:provider,omitempty"}`. Makes T035 pass.
- [X] T041 [US5] In internal/cli/tts/cache.go, add `RunClearCache(args, stdout, stderr) int`:
  - flags `--json`, `--verbose`, `--hub-url` and `--provider NAME`, usage line `usage: sonora clear tts-cache [flags]`;
  - a supplied `--provider` that is empty after `strings.TrimSpace` is exit 2 (`error: --provider must not be empty`), the same rule as `speak` (T024); extra positionals are exit 2;
  - `hub.NewClient()`, `hub.ClearTTSCache`, `ReportError` on failure, then render.
  Depends on T008, T039 and T040. Makes T036 pass.
- [X] T042 [US5] In cmd/sonora/main.go, add `case "clear": return dispatchClear(args[1:], stdout, stderr)` and write `dispatchClear`:
  - `--help` with no resource prints usage to stdout and exits 0;
  - a first argument of `tts-cache` calls `tts.RunClearCache(args[1:], …)`;
  - no argument, or anything else, prints the usage line plus `error: clear supports only tts-cache` and exits 2.
  Add the helpText line `clear tts-cache` ("Clear the TTS audio cache (all, or one provider's)") and the example `sonora clear tts-cache --provider openai`. Depends on T041. Makes T037 and T038 pass.

**Checkpoint**: All five stories pass independently.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T043 [P] Update README.md: add `speak`, `get tts-cache` and `clear tts-cache` to the command table and examples, note that they need the hub's TTS extension. The README has no exit-code list yet, so add an "Exit codes" section listing every code from `hub.ErrorClass.ExitCode` (0, 2, 3, 4, 5, 6, 8, 9, 10, 11, 12) plus the new 13 "TTS not available", with the TTS sources from data-model.md (Exit codes) (FR-015, SC-004).
- [X] T044 [P] Update docs/cli-command-landscape.md: add a "tts (extension)" section with ✅ rows mapping `sonora speak <text|-> <outputs|groups>/<id> [--provider] [--voice] [--language]` → `POST /api/tts/speak`, `sonora get tts-cache` → `GET /api/tts/cache/stats` and `sonora clear tts-cache [--provider]` → `DELETE /api/tts/cache`. Note that `listExtensions` is used internally only, and update the operationId count in the intro.
- [X] T045 Review cmd/sonora/main.go `helpText` as a whole: the new lines are aligned with existing entries, and the `Commands:` column widths are unchanged (FR-015). Re-run `go test ./cmd/...`.
- [X] T046 Run `make check` (gofmt, go vet, the full `go test ./...` including the slow SC-003 test) and fix any findings in files touched by this feature.
- [X] T047 Regression (SC-006): diff the `sonora help` output and the exit codes of existing commands against `main`, using `go test ./tests/... ./cmd/...` on both branches. Confirm no existing test changed expectation apart from the additive help-text assertions. Also measure cold start on both branches (for example, the median of 20 runs of `sonora get outputs --hub-url http://127.0.0.1:1`, which fails fast) and record the before/after timings under Notes (Principle I, Performance Standards).
- [X] T048 Walk through specs/009-tts-commands/quickstart.md sections 1–2. Run section 3 against a real hub if one is available, including row 2 (cache-hit `speak` under 1 s, SC-002). Otherwise record under Notes in this file which manual rows were not run.
- [X] T049 Write the PR self-review against Constitution Principles I, III, IV and VI. It is required because this feature changes HTTP client construction (`hub.NewClientWithTimeout`). Cover: startup path unchanged (T047 timings); no new dependency; every request bounded (5 s / 15 s), with no retries and connection reuse through `http.DefaultTransport`; every behaviour had a failing test first. Put it in the PR description.

---

## Phase 9: User Story 6 - Override the response wait for a slow TTS provider (Priority: P4)

**Added**: 2026-09-22, amending spec.md (FR-013a, Session 2026-09-22), plan.md, and
research.md §6 — a real hub was found whose TTS provider timeout exceeds the 10 s default
FR-013's fixed 15 s bound assumed, so the "`--timeout` is out of scope" call was reversed.

**Goal**: `sonora speak ... --timeout <duration>` overrides the default 15 s response bound
per invocation (FR-013a), for hubs whose TTS provider timeout exceeds the hub's assumed
10 s default.

**Independent Test**: Against a fake hub whose `/api/tts/speak` handler sleeps 300 ms before
responding, `sonora speak "Hi" outputs/kitchen --timeout 50ms` exits 4 (network timeout) in
well under 1 s, proving the flag actually shortens the bound below the 15 s default.
`--timeout 0`/`-5s`/`notaduration` exit 2 with zero requests sent.

### Tests for User Story 6 (write first, must fail) ⚠️

- [ ] T050 [P] [US6] Extend tests/unit/cli_tts_speak_test.go with `--timeout` cases for `tts.RunSpeak`:
  - a slow fake hub (sleeps e.g. 300 ms) with `--timeout 50ms` exits 4 in well under 1 s (proves the flag actually overrides the 15 s default, not just accepted and ignored);
  - the same slow fake hub (300 ms) with no `--timeout` still succeeds (the default remains long enough for an ordinary slow response);
  - `--timeout 0`, `--timeout -5s`, and `--timeout notaduration` each exit 2 with `error: --timeout must be a positive duration` and zero requests sent;
  - `--help` lists `--timeout` alongside the other six flags (extends T022's flag-listing test).
- [ ] T051 [P] [US6] Extend tests/integration/tts_test.go with `TestSpeak_TimeoutOverride_SlowProviderStillReceived`: set `mockTTSHub.speakDelay` past the 15 s default (e.g. 20 s), `speakStatus`/`speakBody` to a 503 `PROVIDER_TIMEOUT` response as in `TestSpeak_SlowProviderStillReceived`, and invoke with `--timeout 25s`; assert exit 10 (the hub's own response, not a CLI network timeout) and that elapsed time is under 25 s. `t.Skip` under `testing.Short()`, matching the existing slow test.

### Implementation for User Story 6

- [ ] T052 [US6] In internal/cli/tts/speak.go: add a `--timeout DURATION` string flag ("override the default 15s response-wait bound", per contracts/cli-tts.md). Detect "supplied" with `fs.Visit`, the same pattern as `--provider`/`--voice`/`--language` (T024). When supplied, parse with `time.ParseDuration`; a parse error or a value `<= 0` is `error: --timeout must be a positive duration`, exit 2, checked before any I/O (alongside the other empty-value checks). Build the client with `hub.NewClientWithTimeout(d)` where `d` is the parsed duration when supplied, else `hub.SpeakTimeout` (the existing default). Makes T050 and T051 pass.

**Checkpoint**: US1–US6 all pass independently. `speak` supports a configurable response bound.

### Polish for User Story 6

- [ ] T053 [P] Update README.md: mention `--timeout` in the `speak` paragraph (FR-015 already requires the command table and exit codes to stay current; this extends that coverage to the new flag).
- [ ] T054 [P] Update docs/cli-command-landscape.md: add `--timeout` to the `speak` row's flag list in the "tts (extension)" section.
- [ ] T055 Run `make check` (gofmt, go vet, full `go test ./...`, including both slow tests — T051 and the original SC-003 test) and fix any findings. Re-run T047's regression method (diff `sonora help` output, exit codes, and cold-start timing against `main`) to confirm no existing command or the original five TTS user stories regressed.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: depends on Setup. **Blocks all stories.** Every command uses
  `ReportError`, the TTS error types and exit 13.
- **US1 (Phase 3)**: depends on Phase 2.
- **US2 (Phase 4)**: depends on US1's `RunSpeak` (T016) and dispatch (T017). It adds group
  targets (T020).
- **US3 (Phase 5)**: depends on US1 (T016). Its tests can be written in parallel with US2's,
  but T020 and T024 both edit internal/cli/tts/speak.go, so run those two one after the other.
- **US4 (Phase 6)**: depends on Phase 2 only. It is independent of US1–US3. `internal/hub/tts.go`
  already exists from T005. `internal/render/tts.go` is created by T015, or by T031 if US4
  runs first. The same applies to the test files US1 creates (tests/contract/tts_test.go,
  tests/unit/render_tts_test.go, tests/integration/tts_test.go with `mockTTSHub` and
  `runCLIWithStdin`): whichever story lands first creates them (T025, T026, T028).
- **US5 (Phase 7)**: depends on Phase 2. It shares `internal/cli/tts/cache.go` with US4
  (T041 edits the file T032 creates) and tests/unit/cli_tts_cache_test.go (T036 extends the
  file T027 creates), so run it after US4 or create the files in whichever lands first. The
  US1 test files follow the same rule (T034, T035, T037).
- **Polish (Phase 8)**: after all stories. T049 comes after T047, whose timings it cites.
- **US6 (Phase 9, added 2026-09-22)**: depends on US1's `RunSpeak` (T016). T052 edits
  internal/cli/tts/speak.go, the same file as T020/T024, so it must run after those two (all
  three land in Phase 8 already, well before this phase starts). T050/T051 can be written any
  time after their target files exist (T010, T012).

### Within Each Story

- Test tasks first. They must compile-fail or fail before implementation starts
  (Principle VI).
- hub function → render → `Run*` → main.go dispatch.
- T005 → T008. T006 and T007 are independent of T005. T008 needs T005 and T007.

### Parallel Opportunities

- Phase 2 tests: T002, T003 and T004 are different files.
- Phase 2 implementation: T006 and T007 in parallel with T005. T008 comes last.
- US1 tests: T009–T013 are five different files. US1 implementation: T015 in parallel with T014.
- US2 (T018, T019) and US3 (T021–T023) tests after US1. Only T021 runs in parallel with the
  US2 tests: T018 and T022 both edit tests/unit/cli_tts_speak_test.go, and T019 and T023 both
  edit tests/integration/tts_test.go, so run each pair one after the other.
- US4 tests T025–T029 and implementation T031 in parallel with T030.
- US5 tests T034–T038 and implementation T040 in parallel with T039.
- Polish: T043 and T044 in parallel.
- US6 tests: T050 and T051 are different files, run in parallel. Polish for US6: T053 and
  T054 in parallel.

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
4. US6 (`--timeout`, added 2026-09-22): optional post-MVP amendment, done any time after US1
   exists (it only needs `RunSpeak`, T016). Not part of the original MVP scope.

## Notes

- [P] = different files, no dependency on incomplete tasks.
- Verify each test fails before implementing. Commit after each task or logical group, using
  Conventional Commits (`feat:`, `test:`, `docs:`), with no attribution trailers (AGENTS.md).
- No new dependency (Principle III). No change to existing `ClassifyError` branches or
  existing exit codes (SC-006).
- T001: no pre-existing failure found; `gofmt`/`go vet`/`go test ./...` were green on
  `009-tts-commands` before any change.
- T047 regression: compared against `main` (994bd4f) via a disposable `git worktree`.
  `go test ./tests/... ./cmd/...` is green on both branches; every pre-existing test's
  expectation is unchanged (only additive TTS assertions were added). Cold start (median of
  20 runs of `sonora get outputs --hub-url http://127.0.0.1:1`): main 15.79 ms, branch
  15.21 ms — no regression (Principle I).
- T048: sections 1–2 of quickstart.md were run (the `go test` commands in this Notes/T046
  entry and the full suite). Section 3's manual scenarios (rows 1–15) were **not** run — no
  real Multiroom Audio Hub with the TTS extension was available in this environment. All 15
  rows remain to be manually verified against a real hub before release, in particular row 2
  (cache-hit `speak` under 1 s, SC-002), which automated tests cannot exercise against a real
  TTS provider.
- docs/cli-command-landscape.md is excluded from git tracking by this repo's local
  `.git/info/exclude` (not the committed `.gitignore`) — a deliberate, pre-existing local-only
  exclusion. T044's edit was made on disk as instructed, but it will not appear in `git
  status`/diffs and won't be included in any commit from this branch.

### T049: PR self-review (Constitution Principles I, III, IV, VI)

This feature changes HTTP client construction (`hub.NewClientWithTimeout`), so this review is
required before merge.

- **I. Instant Startup**: no I/O runs before argument parsing. `hub.NewClient()`/
  `NewClientWithTimeout()` are called inside `RunSpeak`/`RunGetCache`/`RunClearCache` only
  after flags, positionals, the target path, and (for `speak`) standard input are all
  validated — a bad target or empty text fails before any client is built or any pipe is read.
  Dispatch in `cmd/sonora/main.go` adds three plain string comparisons (`speak`, `tts-cache`,
  `clear`). Measured cold start (T047, median of 20 runs of a fast-failing request): 15.79 ms
  on `main` vs. 15.21 ms on this branch — no regression, both well under the 50 ms budget.
- **III. Minimal Dependencies**: no new module; `go.mod`/`go.sum` are unchanged. All four new
  files (`hub/tts.go`, `hub/extensions.go`, `cli/tts/{speak,cache,report}.go`,
  `render/tts.go`) use only the standard library already imported elsewhere in this repo
  (`net/http`, `encoding/json`, `flag`, `net/url`, `io`, `sort`, `strings`, `time`).
- **IV. Resilient HTTP Client**: every request has an explicit bound — 5 s via `NewClient()`
  for the cache commands and the inventory lookup, 15 s via `NewClientWithTimeout(hub.
  SpeakTimeout)` for `speak` (margin over the hub's own ~10 s provider timeout, verified by
  `TestSpeak_SlowProviderStillReceived`, SC-003). No retries anywhere (asserted by per-path
  request counters in the integration tests). Every failure mode — TTS error codes, a
  non-TTS-shaped body, a 404 (missing/disabled/rejected/inert/version-mismatched extension, or
  a wrong hub address), a malformed success body, and network/timeout — maps to a
  distinguishable exit code with a plain `error: ...` message; the raw error is available via
  `--verbose`. Error bodies are read through a 64 KiB `io.LimitReader`, so a large or slow
  proxy error page can't exhaust memory or hang decoding. No panics: every decode path returns
  a `*DecodeError` rather than indexing into an unchecked value. Connection reuse holds because
  every client (5 s and 15 s) shares `http.DefaultTransport`.
- **VI. Test-First**: every behavior in this feature has a contract, unit, or integration test
  that was written and run to a compile/assertion failure before its implementation existed —
  see the phase-by-phase red/green checkpoints in this file's task list (T002–T049). The full
  suite (`go test ./...`, including the ~11 s slow-provider test) is green; the pre-existing
  suite's expectations are unchanged (T047).
