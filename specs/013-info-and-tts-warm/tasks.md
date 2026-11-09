# Tasks: Hub Info and TTS Cache Warm-up

**Input**: Design documents from `/specs/013-info-and-tts-warm/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, contracts/hub-package.md

**Tests**: Required. Constitution Principle VI: every test task comes before the code it covers and must
be seen to fail first.

**Organization**: grouped by user story. US1 (`get info`) and US2 (`warm tts-cache`) are independent after
Phase 2. US3 (Go callers) is the `hub` package surface of both; its tests are written inside US1 and US2,
and Phase 5 only checks the public-package rules.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: different files, no dependency on an unfinished task
- Paths are from the repository root `D:\projects-sonora\sonora-cli`
- Go checks run directly, not through `make` (this machine has no `make`):
  `gofmt -l .`, `go vet ./...`, `go test ./...`

---

## Phase 1: Setup

- [ ] T001 Confirm the baseline: run `gofmt -l .`, `go vet ./...` and `go test ./...` on branch
  `013-info-and-tts-warm` and record the result in `specs/013-info-and-tts-warm/verification.md` (create it
  with a dated "Baseline" section). The baseline already includes the committed, unrelated `Group.system` mirror
  (`hub/groups.go`, `internal/render/groups.go`, their tests); do not touch it.
- [ ] T002 Confirm `api/openapi.json` declares what the feature mirrors: `HubInfoResponse` (five optional
  properties, no `required` list), `WarmRequest` (`text` plus seven optional fields) and `WarmResponse`
  (`cacheHit`), and the paths `/api/v2/info` and `/api/tts/cache/warm`. Note any difference from
  `data-model.md` in `verification.md`; fix `data-model.md` if the spec disagrees.

---

## Phase 2: Foundational (shared hub errors)

**Purpose**: the two error changes both stories rely on. Blocks Phase 3 and Phase 4.

- [ ] T003 Write failing `ClassifyError` cases in `tests/unit/hub_client_test.go`:
  `*hub.InfoNotOfferedError` gives `hub.ClassHub` and the message `this hub does not offer hub info (needs
  hub 0.1.33 or later)`; a `*hub.TTSError` with code `CACHE_WRITE_FAILED` gives
  `hub.ClassServiceUnavailable`; the existing TTS cases are unchanged.
- [ ] T004 Add `InfoNotOfferedError` (empty struct, `Error()` method, self-contained godoc) and the two
  `ClassifyError` mappings from T003 in `hub/errors.go`. Make T003 pass.

**Checkpoint**: `go test ./tests/unit/...` passes; `hub` still builds.

---

## Phase 3: User Story 1 - See which hub I am talking to (P1) MVP

**Goal**: `sonora get info` and `hub.GetInfo`.

**Independent Test**: run `sonora get info` against the mock hub and compare the output with its response;
against the live `multiroom.lan` it is read-only and allowed.

### Tests first (write, run, watch fail)

- [ ] T005 [P] [US1] Contract tests in `tests/contract/info_test.go` against the mock server:
  - the request is `GET /api/v2/info` with no body;
  - a reply with all five fields decodes to `hub.Info`, and `json.Marshal` of it gives the property names
    `product`, `version`, `apiVersion`, `startedAt`, `uptimeSeconds`;
  - the fixture is `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33","startedAt":"2026-11-06T10:06:59.512Z","uptimeSeconds":16981}`
    and is checked against `HubInfoResponse` in `api.Spec`, as the other contract tests do;
  - `startedAt` or `uptimeSeconds` missing still decodes, and `UptimeSeconds` is nil rather than 0;
  - `product` other than `multiroom`, and `version`/`apiVersion` equal to `unknown`, are returned as sent;
  - 404 gives `*hub.InfoNotOfferedError`; 500 gives `*hub.StatusError`; both classified by
    `hub.ClassifyError` as `ClassHub`;
  - a 200 body that is not JSON, not an object, or lacks `product`, `version` or `apiVersion` (required
    and non-empty, the client's rule) gives `*hub.DecodeError`.
- [ ] T006 [P] [US1] Renderer tests in `tests/unit/render_info_test.go`: YAML shows the five keys in the
  order of `contracts/cli.md` with string values quoted; `startedAt` and `uptimeSeconds` are left out
  when absent; JSON is one object with the same names; a zero `uptimeSeconds` that the hub sent (non-nil)
  is still shown.
- [ ] T007 [P] [US1] Command tests in `tests/integration/info_test.go` (mock hub, run through the real
  dispatcher like `master_mute_test.go`): YAML output; `--json`; hub 404 prints
  `error: this hub does not offer hub info (needs hub 0.1.33 or later) (hub URL: <url>)` with the hub-error
  exit code; unreachable hub gives the network message and exit code; a malformed body gives `hub returned
  an unexpected or malformed response`; `get info extra` is a usage error; `list info` is a usage error
  naming `get info`; `get info --help` prints usage and exits 0; `--verbose` keeps the cause.

### Implementation

- [ ] T008 [US1] Add `Info` and `GetInfo(ctx, client, baseURL)` in `hub/info.go`, shaped like
  `GetMasterMute` in `hub/mastermute.go`: 404 returns `&InfoNotOfferedError{}`; other non-2xx returns
  `*StatusError`; decode into a wire type and return `*DecodeError` when `product`, `version` or
  `apiVersion` is missing or empty (quote `data-model.md`: "yes, non-empty (client rule)"); `StartedAt`
  is a string omitted when empty, `UptimeSeconds` is `*int64` omitted when nil. Godoc must make sense
  without this repository's specs. Make T005 pass.
- [ ] T009 [P] [US1] Add the YAML and JSON renderers in `internal/render/info.go`, following the
  conventions in `internal/render/mastermute.go`. Make T006 pass.
- [ ] T010 [US1] Add `RunGet` in `internal/cli/hubinfo/info.go` (new package, modelled on
  `internal/cli/mastermute/mastermute.go`): flags `--json`, `--verbose`, `--hub-url`, `--help`; no
  positional arguments; 5 s request timeout; errors through `hub.ClassifyError` and the shared exit codes.
  Depends on T008, T009.
- [ ] T011 [US1] In `cmd/sonora/main.go`: in `dispatchGetList` route `info` with verb `get` to
  `hubinfo.RunGet`, and answer `list info` with the usage error `list does not support info; use 'sonora
  get info' instead` (same shape as `master-mute`); add `get info  Show which hub is running and since
  when` to `helpText` and `sonora get info` to its examples. Make T007 pass.

**Checkpoint**: `go test ./tests/...` passes. `sonora get info` against the live hub (read-only) shows
the five fields; record it in `verification.md`.

---

## Phase 4: User Story 2 - Warm the TTS cache before it is needed (P2)

**Goal**: `sonora warm tts-cache <text|->` and `hub.WarmTTSCache`.

**Independent Test**: warm a phrase against the mock hub, see `cacheHit` for both outcomes and nothing
played. Live: needs the user's permission (writes one cache entry); see `quickstart.md`.

### Tests first (write, run, watch fail)

- [ ] T012 [P] [US2] Add warm cases to `tests/contract/tts_test.go`:
  - the request is `POST /api/tts/cache/warm`, `Content-Type: application/json`, and the body is checked
    against `WarmRequest` in `api.Spec`;
  - only the set fields are sent: a request with `Text` only sends just `text`; one with all eight fields
    sends all eight with the spec's names (`providerName`, `voice`, `language`, `engine`, `pitch`,
    `speakingRate`, `stylePrompt`);
  - `{"cacheHit":false}` and `{"cacheHit":true}` decode to `hub.WarmResult`;
  - a 2xx body without `cacheHit` is `*hub.DecodeError` (never `CacheHit: false`);
  - 404 gives `*hub.TTSNotOfferedError`; 400 with `INVALID_REQUEST` is `ClassValidation`; 400 with
    `PROVIDER_NOT_FOUND` is `ClassNotFound`; 503 with a provider code and with `CACHE_WRITE_FAILED` is
    `ClassServiceUnavailable`; errors decode through the existing `TtsErrorResponse` path.
- [ ] T013 [P] [US2] Renderer tests in `tests/unit/render_tts_warm_test.go`: YAML `cacheHit: false` /
  `cacheHit: true`; JSON `{"cacheHit":false}` with no extra fields.
- [ ] T014 [P] [US2] Command tests in `tests/integration/tts_warm_test.go` (mock hub; follow
  `tests/integration/tts_test.go`):
  - `warm tts-cache "dinner"` prints `cacheHit: false`, and the mock records that nothing was sent to any
    play or route endpoint;
  - a repeat prints `cacheHit: true`; `--json` prints the object;
  - `--provider`, `--voice`, `--language` reach the hub body; flags work before and after the text;
  - `-` reads standard input and trims trailing line breaks, as `speak` does;
  - no text, empty text, blank-after-trim stdin, an empty `--provider`/`--voice`/`--language`, and a bad
    `--timeout` are usage errors (messages as in `contracts/cli.md`);
  - hub 400, unknown provider, extension absent (the diagnosed "text-to-speech is not available" message)
    and unreachable hub give the messages and exit codes of `speak`;
  - 503 prints the hub's message followed by `; nothing was stored, repeating the request is safe`;
  - `--timeout` changes the request bound from the 15 s default;
  - cache round trip (SC-002): the mock keeps a cache keyed by text and voice options; warming a phrase and
    then running `speak` with the same text, `--provider`, `--voice` and `--language` reports a cache hit,
    and `speak` with different options does not.

### Implementation

- [ ] T015 [US2] In `hub/tts.go` add `WarmRequest`, `WarmResult` and `WarmTTSCache(ctx, client, baseURL,
  req)`, copying `Speak`'s flow (research R4): pointer-typed optional fields with `omitempty`; 404 returns
  `*TTSNotOfferedError`; other non-2xx go through `decodeTTSError`; decode 2xx through a wire type with
  `*bool` so a missing `cacheHit` is `*DecodeError`. Godoc self-contained. Make T012 pass.
- [ ] T016 [P] [US2] Add the warm result renderers (YAML and JSON) in `internal/render/tts.go`. Make T013
  pass.
- [ ] T017 [US2] Add `RunWarm` in `internal/cli/tts/warm.go`: reuse `speak`'s text and `-` handling and
  its `--provider`, `--voice`, `--language`, `--timeout` validation (a given flag must not be empty;
  `--timeout` a positive duration; default `hub.SpeakTimeout`); report errors through `tts.ReportError`,
  appending `; nothing was stored, repeating the request is safe` for `ClassServiceUnavailable` only
  (that wording is CLI-side, not in `hub/`). Depends on T015, T016.
- [ ] T018 [US2] In `cmd/sonora/main.go` add a `warm` case to `run` and a `dispatchWarm` (like
  `dispatchClear`) that accepts only `tts-cache` and calls `tts.RunWarm`; anything else is `usage: sonora
  warm tts-cache <text|-> [flags]` with `error: warm supports only tts-cache`, exit 2; `warm --help`
  prints the usage and exits 0. Add `warm tts-cache <text|->  Synthesize a phrase into the TTS cache
  without playing it` to `helpText` and `sonora warm tts-cache "Dinner is ready" --provider piper-local`
  to its examples. Make T014 pass.

**Checkpoint**: `go test ./tests/...` passes. Live check from `quickstart.md` only with the user's
permission, `--provider piper-local`, and without `clear tts-cache`.

---

## Phase 5: User Story 3 - Use both from Go code (P2)

**Goal**: the new `hub` identifiers satisfy the public-package rules. Behaviour tests are already in T005
and T012; this phase adds only what they do not cover.

- [ ] T019 [US3] Run `tests/unit/hub_godoc_test.go` and `tests/unit/api_spec_test.go` over the new
  exported identifiers (`Info`, `GetInfo`, `InfoNotOfferedError`, `WarmRequest`, `WarmResult`,
  `WarmTTSCache`). If the godoc test does not enumerate exports by itself, extend it so each new
  identifier has a comment with no `research.md`, `FR-` or `Principle` reference. Fix any failure in
  `hub/info.go`, `hub/tts.go` or `hub/errors.go`.
- [ ] T020 [US3] Check `hub/` stays protocol-only and stdlib-only: `go list -deps -f '{{if not
  .Standard}}{{.ImportPath}}{{end}}' ./hub` lists only this module's own packages, and no new file in `hub/` imports anything from `internal/`.
  Record the result in `verification.md`.

**Checkpoint**: sonora-mcp can use `hub.GetInfo` and `hub.WarmTTSCache` from the next minor tag.

---

## Phase 6: Polish

- [ ] T021 [P] Update `README.md`: document `sonora get info` and `sonora warm tts-cache <text|->` with
  an example each, the flags `warm` accepts, the YAML and JSON output, the failure cases from
  `contracts/cli.md`, and that both need hub 0.1.33 or later (older hubs keep working for every other
  command).
- [ ] T022 [P] Add a "Startup" line to `verification.md`: the new `case` arms add no work before argument
  parsing (Principle I); no benchmark is needed because the startup path is unchanged. Confirm by reading
  `run` in `cmd/sonora/main.go`.
- [ ] T023 Run `gofmt -l .` (no output), `go vet ./...`, `go test ./...`, and `git ls-files --eol` (no
  tracked file `w/crlf`). Record commands, date and counts in `verification.md`.
- [ ] T024 Run the `quickstart.md` live steps. `get info` is read-only. `warm tts-cache` writes one entry
  to the production cache: ask the user first, use `--provider piper-local`, note the entry count before
  and after (`sonora get tts-cache`), and do not run `clear tts-cache`. Record the results.
  The cache round trip (SC-002) is audible: with the user's permission, record the volume and mute state of
  the target output first, keep the volume low, warm a `piper-local` phrase, then `speak` the same phrase
  with the same options and expect `cacheHit: true`; restore state and run the ALSA/metrics checks from the
  repository-parent `AGENTS.md`.
- [ ] T025 Mark every finished task `[X]` in this file.
- [ ] T026 Release (only when the user asks): after merge, cut a minor tag for the additive `hub` API so
  sonora-mcp can bump to it; never move or re-create a published tag.

---

## Dependencies & Execution Order

- Phase 1 first. Phase 2 blocks Phases 3 and 4. Phase 3 and Phase 4 are independent of each other, apart
  from both editing `cmd/sonora/main.go` (T011, T018) and `README.md` (T021): do those edits one after
  the other. Phase 5 needs Phases 3 and 4. Phase 6 last.
- Inside each story: tests, then `hub`, then renderers, then the command, then `main.go`. Watch each
  test fail before its implementation task.
- MVP is Phases 1 to 3: `get info` ships alone as a minor tag if needed.

### Parallel opportunities

- Phase 3 tests: T005, T006, T007 together. After T008: T009 beside T010's prerequisites.
- Phase 4 tests: T012, T013, T014 together; T016 beside T015.
- A second developer can take all of Phase 4 while the first takes Phase 3, merging `main.go` last.
- T021 and T022 together.

## Implementation Strategy

1. T001 to T004: baseline and shared errors.
2. Phase 3 (MVP): validate with the mock hub, then read-only against the live hub.
3. Phase 4: validate with the mock hub; live only with permission.
4. Phase 5 and 6: public-package rules, README, final checks.
