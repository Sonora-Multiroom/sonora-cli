---

description: "Task list for 011-hub-0128-alignment"
---

# Tasks: Align the CLI with Hub 0.1.25–0.1.28

**Input**: Design documents from `specs/011-hub-0128-alignment/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli.md](contracts/cli.md),
[contracts/hub-package.md](contracts/hub-package.md), [quickstart.md](quickstart.md)

**Tests**: Required. The constitution's Principle VI is test-first and non-negotiable, and FR-015 asks
for it. In every story the test tasks come first and MUST fail before the matching implementation task
starts.

**Organization**: One phase per user story. The three stories touch disjoint files, apart from
`README.md`, and can be built in any order.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1, US2 or US3 from spec.md

## Conventions for every task

- Git Bash on Windows, no `make`: run `gofmt -l .`, `go vet ./...` and `go test -count=1 ./...`
  directly, and read a command's own exit code (`cmd > "$TEMP/x.log" 2>&1; echo $?`), never through a
  pipe.
- Keep the surrounding style: test names `TestXxx_Scenario`, table tests where the file already uses
  them, and doc comments with no `FR-`/`SC-`/`US` numbers (`tests/unit/stray_refs_test.go` enforces
  this for `hub/`).
- Files are stored with LF line endings.

---

## Phase 1: Setup

**Purpose**: Confirm the baseline the feature builds on.

- [X] T001 Confirm `api/openapi.json` reads `"version": "0.1.28"`, and that `go test -count=1 ./...` passes on branch `011-hub-0128-alignment` before any change. Record the result as the baseline in `specs/011-hub-0128-alignment/verification.md`, with the date, the command and the package counts.

---

## Phase 2: Foundational

None. No story depends on another, and no shared code is introduced.

---

## Phase 3: User Story 1 - Know that a disable stops playback before running it (Priority: P1) 🎯 MVP

**Goal**: The help for each of the six enable/disable commands, and the README, say what that command
does to playback that is already running.

**Independent Test**: `sonora <enable|disable> <inputs|outputs|groups>/x --help` prints the
`usage: …` line, a blank line, the description line from [contracts/cli.md](contracts/cli.md), and then
`Flags:`. It exits 0 without contacting a hub.

### Tests for User Story 1 ⚠️ write first, watch them fail

- [X] T002 [P] [US1] In `tests/unit/cli_inputs_enabled_test.go`, add `TestInputsRunEnable_HelpDescribesPlayback` and `TestInputsRunDisable_HelpDescribesPlayback`. Each calls `inputs.RunEnable` or `inputs.RunDisable` with `[]string{"--help"}` and asserts exit 0, that line 1 is `usage: sonora <verb> inputs/<input-id> [flags]`, that line 2 is empty, and that line 3 is the verb's exact sentence from the `enable`/`disable` `inputs` rows of `specs/011-hub-0128-alignment/contracts/cli.md`. Also add `TestInputsRunDisable_UsageErrorOmitsDescription`: with no arguments, stderr contains the `usage: …` line and does not contain the description sentence.
- [X] T003 [P] [US1] Add the same two tests for outputs in `tests/unit/cli_outputs_enabled_test.go`: `TestOutputsRunEnable_HelpDescribesPlayback` and `TestOutputsRunDisable_HelpDescribesPlayback`, using the `outputs` rows, plus `TestOutputsRunDisable_UsageErrorOmitsDescription`.
- [X] T004 [P] [US1] Add the same two tests for groups in `tests/unit/cli_groups_enabled_test.go`: `TestGroupsRunEnable_HelpDescribesPlayback` and `TestGroupsRunDisable_HelpDescribesPlayback`, using the `groups` rows, plus `TestGroupsRunDisable_UsageErrorOmitsDescription`.

### Implementation for User Story 1

- [X] T005 [P] [US1] In `internal/cli/inputs/enabled.go`, `runSetEnabled`: keep `usage` as the one-line `usage: …` string, and pass `usage + "\n\n" + description` only to the `clihelp.PrintUsage` call that serves `--help`. Take the verb's description from contracts/cli.md, as two unexported string constants, one per verb. `clihelp.SetUsage` and the missing-argument message keep the one-line `usage`, so error output is unchanged. Do not change `clihelp`.
- [X] T006 [P] [US1] Make the same change, with the `outputs` sentences, in `internal/cli/outputs/enabled.go`.
- [X] T007 [P] [US1] Make the same change, with the `groups` sentences, in `internal/cli/groups/enabled.go`.
- [X] T008 [US1] In `README.md`, replace the paragraph that starts "`enable <resource>/<id>` and `disable <resource>/<id>` set the input's…" (around line 87) with the five rules of FR-001, keeping its last sentence about which resources are supported:
  - disabling an input stops every route playing it, including paused routes and routes sharing its hardware input, and removes an input registered with auto-remove;
  - disabling a group stops the routes addressed to it, while routes addressed to a member keep playing;
  - disabling an output takes it out of every route playing on it: a route addressed to it stops, and a group route keeps playing on its other members;
  - enabling an output puts it back into the live group routes of its enabled groups, under the same route ID, and a route stopped by the disable does not come back;
  - enabling an input or a group starts nothing.

  Also, if `sonora --help` in `cmd/sonora/main.go` (lines 39–40) says anything contradicting this, fix it. If it is neutral, leave it as is.

**Checkpoint**: T002–T004 pass, `go test -count=1 ./tests/unit/` passes, and the README section reads correctly.

---

## Phase 4: User Story 2 - See what `sonora play` started (Priority: P2)

**Goal**: The hub client's play result carries `displayName`, and `sonora play` shows it as the second
field when it is non-empty.

**Independent Test**: The render tests and the play integration tests pass both with and without
`displayName` in the hub's answer. Without it, the output is byte-identical to the old output.

### Tests for User Story 2 ⚠️ write first, watch them fail

- [X] T009 [P] [US2] In `tests/contract/play_test.go`, add two tests:
  - `TestPlayback_Success_DecodesDisplayName`: a 200 body shaped like `#/components/schemas/PlaybackResponse` at 0.1.28, with `"displayName": "Flickermood — Forss"`. Assert `resp.DisplayName == "Flickermood — Forss"`.
  - `TestPlayback_Success_MissingDisplayNameDecodesEmpty`: the same body without the key. Assert no error and `resp.DisplayName == ""`.
- [X] T010 [P] [US2] In `tests/unit/render_play_test.go`, add:
  - `TestRenderPlaybackYAML_DisplayNameSecondWhenSet`: with `DisplayName: "Flickermood — Forss"`, assert the line `displayName: "Flickermood — Forss"` appears, unescaped, and that the field order is inputId, displayName, routeId, status, message.
  - `TestRenderPlaybackJSON_DisplayNameWhenSet`: assert 5 keys, `displayName` equal to the name, and that the raw output contains `—` literally.
  - `TestRenderPlayback_CyrillicDisplayNameUnescaped`: with `"Кино — Группа крови"`, assert the YAML and JSON both contain it verbatim.
  - `TestRenderPlayback_EmptyDisplayNameOmitted`: with `DisplayName: ""`, assert the YAML and JSON equal exactly what `testPlaybackResponse()` rendered before this change. Write that expected YAML and JSON out as literal strings in the test. Also assert `len(decoded) == 4` and that `displayName` is absent.
  The existing `TestRenderPlaybackJSON_RoundTripsWithExactFields` (4 keys) must keep passing unchanged.
- [X] T011 [P] [US2] In `tests/integration/play_test.go`, add:
  - `TestPlay_Success_ShowsDisplayName`: the mock hub's playback 200 body includes `"displayName": "Flickermood — Forss"`. Assert the YAML stdout has `displayName: "Flickermood — Forss"` as its second line, and the `--json` stdout has the key.
  - `TestPlay_Success_NoDisplayNameFromOlderHub`: the mock omits the key. Assert the stdout has no `displayName`.
  Extend the existing mock in that file with an optional display-name field; do not write a new mock.
- [X] T012 [P] [US2] In `tests/unit/cli_play_test.go`, add `TestPlayRun_HelpDescribesDisplayName`: `--help` output contains the `--display-name` description from contracts/cli.md (the substring "without it the hub names a SoundCloud or YouTube link after its track").

### Implementation for User Story 2

- [X] T013 [US2] In `hub/play.go`, add ``DisplayName string `json:"displayName"` `` to `PlaybackResponse`. Its godoc: "the input's display name once the route exists: the name sent with the request, the title a SoundCloud or YouTube link resolved to, or a generated placeholder. Empty when the hub does not report it (before hub 0.1.28)." No validation: an empty name is not an error. This makes T009 pass.
- [X] T014 [US2] In `internal/render/play.go`:
  - add ``DisplayName string `json:"displayName,omitempty"` `` to `playbackPayload`, between `InputID` and `RouteID`, and fill it in `toPlaybackPayload`;
  - in `RenderPlaybackYAML`, write `displayName: %q` after `inputId` only when it is non-empty;
  - update the comments that say "exactly inputId, routeId, status, message", so they describe the optional fifth field without citing FR numbers.
  This makes T010 and T011 pass.
- [X] T015 [US2] In `internal/cli/play/play.go`, change the `--display-name` flag's usage string to: "display `NAME` for the input, kept as given; without it the hub names a SoundCloud or YouTube link after its track and anything else with a generated name". This makes T012 pass.
- [X] T016 [US2] In the `play` paragraph of `README.md` (around line 113, "`--display-name NAME` sets the ephemeral input's display name"):
  - say the flag is optional;
  - say that without it the hub names a SoundCloud or YouTube link "Title — Uploader", and anything else with a generated name;
  - say that a name the user gives is kept;
  - say that the output shows the name as `displayName` when the hub reports it (0.1.28 and later).

**Checkpoint**: T009–T012 pass. Every other play test is unchanged and passes.

---

## Phase 5: User Story 3 - A transfer right after enabling an output just works (Priority: P3)

**Goal**: `hub` recognises the transient refusal and gives it its own message. `sonora transfer`
retries only that refusal: it waits 100, 200, 400 and then 800 ms between attempts, makes at most 5
attempts in total, then exits 14.

**Independent Test**: Run the integration tests for a transfer against a mock hub that refuses N times
with `ROUTE_CHANGING_OUTPUTS` and then accepts. Check the exit code, the stdout, the stderr and the
number of transfer requests, as in the table in [contracts/cli.md](contracts/cli.md).

### Tests for User Story 3 ⚠️ write first, watch them fail

- [X] T017 [P] [US3] In `tests/contract/route_refusal_test.go`, add a case to `TestRouteRefusal_409_DecodesReasonAndOutputID`'s table: the transfer endpoint, reason `ROUTE_CHANGING_OUTPUTS`, no `outputId`, and detail `Route is changing its outputs, try again: route-3f9a1c07b2e4`. Expect `Reason == "ROUTE_CHANGING_OUTPUTS"` and `OutputID == ""`. It may already pass, because decoding is generic; that is acceptable. It documents the 0.1.28 shape.
- [X] T018 [P] [US3] In `tests/unit/hub_client_test.go`, add:
  - `TestIsTransientConflict`, a table that expects true only for `&hub.APIError{StatusCode: 409, Reason: hub.ReasonRouteChangingOutputs}`, including when it is wrapped with `fmt.Errorf("%w")`. It expects false for: a 409 with reason `OUTPUT_DISABLED`; a 409 with no reason; a 422 carrying that reason; `&hub.StatusError{StatusCode: 409}`; a `*hub.NotFoundError`; a network error; and `nil`.
  - In `TestClassifyError_Conflict_Messages`, a case for the transient reason that expects `hub.ClassConflict` and the exact message `the route is still adding an output; retry in a moment`, ignoring the hub's detail. Existing cases must stay unchanged.
- [X] T019 [US3] In `tests/integration/route_test.go`, give `mockRouteHub` a field `transferTransient int`. While `transferRequests` is less than or equal to it, the transfer endpoint answers 409 with `{"title":"Conflict","status":409,"detail":"Route is changing its outputs, try again: <id>","reason":"ROUTE_CHANGING_OUTPUTS"}`, and after that it uses the existing success or `transferStatus` path. Then add:
  - `TestRouteTransfer_TransientRefusalThenSuccess`: `transferTransient: 1`. Assert exit 0, stdout equal to what `TestRouteTransfer_Success_Output_YAML` expects, empty stderr, and `transferRequests == 2`.
  - `TestRouteTransfer_TransientRefusalFourTimesThenSuccess`: `transferTransient: 4`. Assert exit 0 and 5 requests.
  - `TestRouteTransfer_TransientRefusalRunsOut`: `transferTransient: 99`. Assert exit 14; stderr containing `the route is still adding an output; retry in a moment`; `transferRequests == 5`; and an elapsed time between 1.4 s and 3 s.
  - `TestRouteTransfer_TransientRefusalRunsOut_Verbose`: the same with `--verbose`. Assert stderr also contains `detail: Route is changing its outputs`.
  - `TestRouteTransfer_OtherRefusalNotRetried`: `transferStatus: 409` with reason `OUTPUT_DISABLED`. Assert exit 14, `transferRequests == 1`, and the hub's detail on stderr.
  - `TestRouteTransfer_TransientThenOtherRefusal`: `transferTransient: 1`, then `transferStatus: 409` with a `transferBody` carrying reason `OUTPUT_DISABLED`, as in `TestRouteTransfer_409_RefusedRoute`. Assert exit 14, the hub's detail on stderr, no transient message, and `transferRequests == 2`.
  - `TestRouteTransfer_TransientThenRouteGone`: `transferTransient: 1`, then `transferStatus: 404` with a `transferBody` whose detail is `Route not found: <id>` (the transient check runs before `transferStatus`, so the first request is still refused as transient). Assert the existing route-not-found exit code and 2 requests.
  - `TestRouteTransfer_TransientThenNetworkFailure`: give `mockRouteHub` a field `transferHangUp bool`; when set, a transfer request after the transient ones hijacks the connection (`http.Hijacker`) and closes it without answering. With `transferTransient: 1`, assert exit 4 (`exitcode.For(hub.ClassNetwork)`), stderr containing `could not reach the hub`, and `transferRequests == 2`. No existing test hijacks a connection; keep the hijack inside the mock's transfer case.
  The existing `TestRouteTransfer_409_RefusedRoute` must still assert exactly one request; add that assertion if it is missing.

### Implementation for User Story 3

- [X] T020 [US3] In `hub/errors.go`:
  - add `const ReasonRouteChangingOutputs = "ROUTE_CHANGING_OUTPUTS"`, with the godoc from `specs/011-hub-0128-alignment/contracts/hub-package.md`;
  - add `func IsTransientConflict(err error) bool`, which is true only when `errors.As` finds an `*APIError` with `StatusCode == 409` and `Reason == ReasonRouteChangingOutputs`;
  - in `ClassifyError`'s `case 409`, return `ClassConflict, "the route is still adding an output; retry in a moment"` for that reason, before the `Detail`/`Title` fallback;
  - update the `APIError.Reason` doc comment to list all seven reasons, mark this one transient, and keep "treat an unknown one as a generic conflict";
  - update the `OutputID` doc comment to say it is empty for `INPUT_DISABLED`, `GROUP_DISABLED` and `ROUTE_CHANGING_OUTPUTS`.
  `Reason` stays a `string`. This makes T018 pass.
- [X] T021 [US3] In `internal/cli/route/transfer.go`, `RunTransfer`: wrap the `hub.TransferRoute` call in a loop over the delays `[]time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}`. On `hub.IsTransientConflict(err)` with a delay left, `time.Sleep` that delay and retry. On success, or on any other error, leave the loop and continue through the existing code paths. The target pre-check (`hub.ResolveTarget`) runs once, before the loop. Print nothing while retrying. Add a short comment that the hub reports this refusal while a route is still adding an output that was just enabled, and that it clears within moments. Update `RunTransfer`'s doc comment to mention the retry. This makes T019 pass.
- [X] T022 [US3] In `README.md`, extend exit-code row 14 (around line 173) with: "; a refusal because a route is still adding an output clears by itself, and `transfer` retries it for a moment". In the `transfer` paragraph (around line 104), add one sentence saying the same.

**Checkpoint**: T017–T019 pass, and every existing transfer, route and play test is unchanged and passes.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T023 Run `gofmt -l .` (it must print nothing), `go vet ./...` and `go test -count=1 ./...`, reading each command's own exit code. Fix anything that fails, including `tests/unit/hub_godoc_test.go` (every new exported identifier needs godoc) and `tests/unit/stray_refs_test.go`.
- [X] T024 Work through every check in `specs/011-hub-0128-alignment/quickstart.md`:
  - the six `--help` outputs, via `go run ./cmd/sonora …`, with no request to any hub;
  - a review of the README against the 0.1.28 descriptions in `api/openapi.json` (SC-001).
  Do not make audible requests to `multiroom.lan`.
- [X] T025 Record the T023 and T024 results, with dates, commands and test counts, in `specs/011-hub-0128-alignment/verification.md`. Mark tasks `[X]` in this file as they are done.
- [X] T026 Update `docs/backlog/`:
  - delete `enable-disable-stops-playback-docs.md`, `play-shows-resolved-name.md` and `transient-route-changing-outputs.md`;
  - remove their three rows from `docs/backlog/INDEX.md`.
  The `docs/` folder is uncommitted user work: change only those four files.
- [X] T027 Run `git ls-files --eol` over the changed files and confirm none shows `w/crlf`.

---

## Dependencies & Execution Order

- **Setup (T001)** → everything.
- **US1 (T002–T008)**, **US2 (T009–T016)** and **US3 (T017–T022)** are independent of each other. Within
  a story, each test task comes before its implementation task:
  - US1: T002→T005, T003→T006, T004→T007.
  - US2: T009→T013, T010 and T011→T014, T012→T015.
  - US3: T017 and T018→T020, then T019→T021.
- **T020 before T021**: the transfer loop uses `hub.IsTransientConflict`.
- **README tasks (T008, T016, T022)** edit the same file: do them one after another, never in parallel.
- **Polish (T023–T027)** comes after every story it covers.

## Parallel Opportunities

- US1: T002, T003 and T004 together, then T005, T006 and T007 together.
- US2: T009, T010, T011 and T012 together; T013 and T015 touch different files and can run together.
- US3: T017 and T018 together, while T019 is being written.
- Across stories: once T001 is done, the three stories can be built by different people, except for the
  README tasks.

## Implementation Strategy

1. **MVP = US1** (T001–T008): documentation only, so no release of the hub client package is needed. It
   removes the one wrong statement that can stop playback by surprise.
2. **US2**: restores the field-for-field mirror of the play response. It is an additive change to `hub/`,
   and the next tag must be a minor version.
3. **US3**: adds the retry. It is additive to `hub/`; ship it in the same minor tag as US2 if both are
   ready.
4. Polish, then open a PR. After the tag is released, sonora-mcp can update with
   `go get github.com/Sonora-Multiroom/sonora-cli@vX.Y.Z`. That is outside this feature.
