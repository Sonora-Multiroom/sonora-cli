---

description: "Task list for 012-join-modes"
---

# Tasks: Join Modes in the CLI

**Input**: Design documents from `specs/012-join-modes/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli.md](contracts/cli.md),
[contracts/hub-package.md](contracts/hub-package.md), [quickstart.md](quickstart.md)

**Tests**: Required. Principle VI of the constitution (test-first, non-negotiable) and FR-017 both
demand them. In every story the test tasks come first, and each MUST fail before its implementation
task starts.

**Organization**: One phase per user story, in priority order. The three stories share only
`README.md`, `hub/routes.go` (US1 adds a request field; US2 adds a shared validation helper) and
`internal/render/routes.go` (US3 only). They can be built in any order.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel with the others (a different file, and no dependency on an unfinished task)
- **[Story]**: US1, US2 or US3 from spec.md

## Conventions for every task

- **Shell**: Git Bash on Windows, with no `make`. Run `gofmt -l .`, `go vet ./...` and
  `go test -count=1 ./...` directly. Read each command's own exit code
  (`cmd > "$TEMP/x.log" 2>&1; echo $?`), never one from the far end of a pipe.
- **Style**: keep what the surrounding code does. Test names are `TestXxx_Scenario`; use a table test
  where the file already uses them; extend a file's existing mock rather than writing a new one.
- **Doc comments**: no `FR-`, `SC-` or `US` numbers (`tests/unit/stray_refs_test.go`). Every new
  exported `hub` identifier needs godoc (`tests/unit/hub_godoc_test.go`).
- **Hub answers in tests**: fake-hub bodies follow `api/openapi.json` at 0.1.28.
- **Line endings**: files are stored with LF.

---

## Phase 1: Setup

**Purpose**: Confirm the baseline the feature starts from.

- [ ] T001 Confirm that `api/openapi.json` reads `"version": "0.1.28"`, and that `gofmt -l .`, `go vet ./...` and `go test -count=1 ./...` pass on branch `012-join-modes` before any change. Create `specs/012-join-modes/verification.md` with a `## Baseline` section: the date, the commands, the exit codes and the package count.

---

## Phase 2: Foundational

None. No story depends on another, and the only shared code (the `JoinMode*` constants) belongs to
US1, the first story.

---

## Phase 3: User Story 1 - Announce over music, or play alongside it (Priority: P1) 🎯 MVP

**Goal**: The hub client's route, play and create-input requests carry an optional join mode.
`sonora play` and `sonora route` take `--join-mode`, and `sonora create inputs/<id>` takes
`--default-join-mode`. Each command checks the value before sending anything.

**Independent Test**:
- Run each of the three commands against a fake hub, with and without the flag. The request body
  carries the upper-cased mode only when the flag is given; stdout and the exit code match a run
  without the flag.
- With `fade` or `""`, the command exits 2, prints the valid modes, and the fake hub receives no
  request.

### Tests for User Story 1 ⚠️ write first, watch them fail

- [ ] T002 [P] [US1] In `tests/contract/route_test.go`, add two tests:
  - `TestCreateRoute_RequestBody_OmitsJoinModeWhenUnset`: decode the received body into a `map[string]any` and assert that it has no `joinMode` key and exactly the keys `inputId`, `targetId` and `targetType`.
  - `TestCreateRoute_RequestBody_IncludesJoinModeWhenSet`: with `JoinMode: &mode`, where `mode := hub.JoinModeDuckOthers`, assert `body["joinMode"] == "DUCK_OTHERS"`.
- [ ] T003 [P] [US1] In `tests/contract/play_test.go`:
  - extend `TestPlayback_RequestBody_OmitsOptionalFieldsWhenUnset` so it also asserts there is no `joinMode` key;
  - extend `TestPlayback_RequestBody_IncludesOptionalFieldsWhenSet` with `JoinMode: &mode` (`hub.JoinModeMix`), and assert `"joinMode": "MIX"` next to `volume` and `displayName`.
- [ ] T004 [P] [US1] In `tests/contract/inputs_create_test.go`:
  - extend `TestCreateInput_RequestBody_OmitsUnsetEnabledAndAutoRemove` so it also asserts there is no `defaultJoinMode` key;
  - add `TestCreateInput_RequestBody_IncludesDefaultJoinModeWhenSet`, asserting `"defaultJoinMode": "DUCK_OTHERS"`.
- [ ] T005 [P] [US1] Create `tests/unit/joinmode_test.go` with a table test `TestJoinModeParse` for `joinmode.Parse`:
  - accepted, and returned upper-cased: `REPLACE`→`REPLACE`, `mix`→`MIX`, `Duck_Others`→`DUCK_OTHERS`, `duck_others`→`DUCK_OTHERS`;
  - rejected (`ok == false`): `""`, `fade`, `duck-others`, `DUCK OTHERS`, ` MIX`.

  Also add `TestJoinModeValues`, asserting that `joinmode.Values` equals `[]string{hub.JoinModeReplace, hub.JoinModeMix, hub.JoinModeDuckOthers}`, in that order.
- [ ] T006 [P] [US1] In `tests/unit/cli_play_test.go`, add three tests:
  - `TestPlayRun_HelpDescribesJoinMode`: the `--help` stdout contains `-join-mode MODE` and the text from the `play` row of `specs/012-join-modes/contracts/cli.md`, including "without it, REPLACE".
  - `TestPlayRun_InvalidJoinMode_UsageErrorNoRequest`: a table over `fade`, `""` and `duck-others`, each run against `countingHub`. Assert exit 2; that stderr's first line is the play usage line; that stderr contains `error: unknown join mode "<v>"; valid modes: REPLACE, MIX, DUCK_OTHERS`; and that the hub count is 0.
  - `TestPlayRun_InvalidJoinMode_CheckedAfterArguments`: a missing `<target-path>` combined with `--join-mode fade` reports the missing argument, not the join mode. This fixes the order of the checks.
- [ ] T007 [P] [US1] In `tests/unit/cli_route_test.go`, add `TestRouteRun_HelpDescribesJoinMode` (the `route` row text, including "the input's default join mode, else REPLACE") and `TestRouteRun_InvalidJoinMode_UsageErrorNoRequest`. The second follows T006's pattern and uses `countingRouteHub`.
- [ ] T008 [P] [US1] In `tests/unit/cli_inputs_create_test.go`, add:
  - `TestInputsRunCreate_HelpDescribesDefaultJoinMode`, with the `create inputs` row text, including "its routes use REPLACE";
  - `TestInputsRunCreate_InvalidDefaultJoinMode_UsageErrorNoRequest`: values `fade` and `""`, against an `httptest` server that counts requests (the same shape as `countingHub` in `cli_play_test.go`). Assert exit 2, the valid modes on stderr, and 0 requests.
- [ ] T009 [P] [US1] In `tests/integration/play_test.go`, using `mockPlayHub.playRequest`:
  - `TestPlay_JoinMode_SentUpperCased`: `--join-mode duck_others` gives `playRequest["joinMode"] == "DUCK_OTHERS"`; stdout and exit 0 match `TestPlay_Success_SingleOutput_YAML`.
  - `TestPlay_NoJoinMode_FieldAbsent`: without the flag, `playRequest` has no `joinMode` key.
  - `TestPlay_JoinModeWithVolumeAndDisplayName`: `--join-mode mix --volume 20 --display-name Chime` sends all three fields.
  - `TestPlay_JoinMode_GroupTarget`: `groups/<id> --join-mode MIX` sends `"joinMode": "MIX"` with `"targetType": "OUTPUT_GROUP"`.
- [ ] T010 [P] [US1] In `tests/integration/route_test.go`, using `mockRouteHub.routesRequest`:
  - `TestRoute_JoinMode_SentUpperCased`: `inputs/<id> groups/<id> --join-mode MIX` sends `"joinMode": "MIX"`, and the output matches the existing group-target success test.
  - `TestRoute_NoJoinMode_FieldAbsent`: without the flag, there is no `joinMode` key.
- [ ] T011 [P] [US1] In `tests/integration/inputs_create_test.go`, using the body pointer that `mockInputCreateServer` returns:
  - `TestInputsCreate_DefaultJoinMode_SentUpperCased`: `--default-join-mode Duck_Others` sends `"defaultJoinMode": "DUCK_OTHERS"`.
  - `TestInputsCreate_NoDefaultJoinMode_FieldAbsent`: without the flag, there is no `defaultJoinMode` key.

### Implementation for User Story 1

- [ ] T012 [US1] In `hub/routes.go`:
  - add the `JoinModeReplace`, `JoinModeMix` and `JoinModeDuckOthers` constants, with the godoc from `specs/012-join-modes/contracts/hub-package.md`;
  - add to `CreateRouteRequest` the field ``JoinMode *string `json:"joinMode,omitempty"` ``, documented as "how the route joins its target (a JoinMode* value). Nil leaves it to the hub: the input's default join mode, else REPLACE."

  This makes T002 pass.
- [ ] T013 [P] [US1] In `hub/play.go`, add to `PlaybackRequest` the field ``JoinMode *string `json:"joinMode,omitempty"` ``, documented as "how the playback joins its target (a JoinMode* value). Nil means REPLACE." This makes T003 pass.
- [ ] T014 [P] [US1] In `hub/inputs.go`, add to `CreateInputRequest` the field ``DefaultJoinMode *string `json:"defaultJoinMode,omitempty"` ``, documented as "the join mode for routes from this input whose request names none (a JoinMode* value). Nil declares none, and such routes use REPLACE." Update the type's comment, which names `Enabled` and `AutoRemove` as the optional fields. This makes T004 pass.
- [ ] T015 [US1] Create `internal/cli/joinmode/joinmode.go` with:
  - a package comment;
  - `var Values = []string{hub.JoinModeReplace, hub.JoinModeMix, hub.JoinModeDuckOthers}`;
  - `func Parse(s string) (string, bool)`, which returns `strings.ToUpper(s)` and `slices.Contains(Values, …)`;
  - `func ErrorLine(v string) string`, which returns `fmt.Sprintf("error: unknown join mode %q; valid modes: %s", v, strings.Join(Values, ", "))`, so all three commands print the same text.

  Keep the comment style of `routeStatuses` in `internal/cli/routes/list.go`. This makes T005 pass. It depends on T012.
- [ ] T016 [US1] In `internal/cli/play/play.go`:
  - declare `joinModeFlag := fs.String("join-mode", "", …)` with the `play` help text from contracts/cli.md (`MODE` in backquotes);
  - after the target-path checks and before `config.ResolveHubURL`, detect the flag with the existing `fs.Visit` loop (extend it next to `volume`);
  - if the flag was given and `joinmode.Parse` fails, print `usage` and `joinmode.ErrorLine(*joinModeFlag)` to stderr and return `exitcode.Usage`;
  - if it is valid, set `req.JoinMode` to the upper-cased value.

  This makes T006 and T009 pass. It depends on T013 and T015.
- [ ] T017 [US1] In `internal/cli/route/route.go`, make the same change with the `route` help text. Check the flag after the target path is checked and before `config.ResolveHubURL`; set `req.JoinMode` on the `hub.CreateRouteRequest`. This makes T007 and T010 pass. It depends on T012 and T015.
- [ ] T018 [US1] In `internal/cli/inputs/create.go`:
  - add `--default-join-mode` with the `create inputs` help text, checked after the `--display-name` check and before `config.ResolveHubURL`;
  - set `req.DefaultJoinMode` when the flag was given;
  - update `RunCreate`'s doc comment, which lists the flags.

  This makes T008 and T011 pass. It depends on T014 and T015.
- [ ] T019 [US1] In `README.md`:
  - add a short "Join modes" paragraph defining `REPLACE` (stops what plays on the target, except ducking routes), `MIX` (plays alongside) and `DUCK_OTHERS` (plays alongside and lowers the rest while it plays). Say that the value is accepted in any case and needs hub 0.1.21 or later.
  - document `--join-mode` in the `play` and `route` sections, with what happens without it, and give one example each (`sonora play <uri> outputs/kitchen --join-mode duck_others`);
  - document `--default-join-mode` in the `create inputs` section.

**Checkpoint**: T002–T011 pass, and every existing play, route and create-input test is unchanged and
still passes (SC-004, SC-005).

---

## Phase 4: User Story 2 - See what an output is playing (Priority: P2)

**Goal**: The hub client lists the routes on an output. `sonora get outputs/<id> --include-routes`
prints the output's record with a `routes` list. An older hub gets a clear message with exit 3.
`--include-routes` anywhere else is a usage error.

**Independent Test**: Run `get outputs/<id> --include-routes` against a fake hub that returns:
- two routes;
- `[]`;
- an unknown duck state;
- a 404 for the output;
- a 404, a 405 or a 500 for the routes only.

Check the YAML, the JSON, the exit code, and that stdout is empty on every failure. Without the flag,
exactly one request is sent.

### Tests for User Story 2 ⚠️ write first, watch them fail

- [ ] T020 [P] [US2] Create `tests/contract/outputs_routes_test.go` for `hub.GetRoutesForOutput`, following `tests/contract/outputs_get_test.go`:
  - `TestGetRoutesForOutput_RequestAndDecodeContract`: assert the method is `GET` and the path is `/api/v2/outputs/kitchen/routes`. The server returns two 0.1.28 `RouteOnOutputResponse` entries: a group route with `duckState` `LOWERED` and `main: true`, then a direct `DUCK_OTHERS` route with `FULL`, `main: false` and `startedAt: null`. Assert that every field decodes, in the hub's order.
  - `TestGetRoutesForOutput_EmptyList` and `TestGetRoutesForOutput_NullBody`: both give a non-nil empty slice and no error.
  - `TestGetRoutesForOutput_NotFound`: a 404 gives a `*hub.NotFoundError` with `Resource == "output"` and `ID == "kitchen"`.
  - `TestGetRoutesForOutput_MethodNotAllowed`: a 405 gives a `*hub.StatusError` with `StatusCode == 405`.
  - `TestGetRoutesForOutput_UnknownDuckState`: `"duckState": "MUFFLED"` decodes unchanged, with no error, and `Lowered()` is false.
  - `TestGetRoutesForOutput_UnknownJoinMode`: `"joinMode": "CROSSFADE"` decodes unchanged.
  - `TestGetRoutesForOutput_MalformedBodyRejected`: a table of a non-array body, an entry missing `routeId`, an unknown `targetType` and an unknown `status`, each giving a `*hub.DecodeError`.
- [ ] T021 [P] [US2] In `tests/unit/hub_client_test.go`, add `TestRouteOnOutput_Lowered`, a table: `LOWERED` gives true; `FULL`, `""` and `MUFFLED` give false. Also assert `hub.DuckStateFull == "FULL"` and `hub.DuckStateLowered == "LOWERED"`.
- [ ] T022 [P] [US2] In `tests/unit/render_output_get_test.go`, add:
  - `TestRenderOutputWithRoutesYAML_RoutesAfterRecord`: the exact expected string. It is the six output lines, then `routes:`, then two entries of nine fields each in the order `routeId, inputId, targetId, targetType, status, joinMode, duckState, main, paused`: strings quoted, booleans bare, two-space indentation with `  - ` on the first field, as in `RenderRoutesYAML`.
  - `TestRenderOutputWithRoutesYAML_NoRoutes`: the six output lines, then `routes: []`, with no comment line.
  - `TestRenderOutputWithRoutesJSON_FlatObjectWithRoutes`: unmarshal into `map[string]any`. Assert the six output keys and `routes`, and that each entry has all 13 `RouteOnOutputResponse` keys (including `createdAt`, `startedAt`, `transferable` and `pauseable`).
  - `TestRenderOutputWithRoutesJSON_NoRoutesIsEmptyArray`: with a nil slice, the raw output contains `"routes":[]`.
  - `TestRenderOutputWithRoutes_UnknownDuckStateShownAsSent`: `MUFFLED` appears unchanged in the YAML and the JSON.
- [ ] T023 [P] [US2] In `tests/unit/cli_outputs_get_test.go`, add `TestOutputsRunGet_HelpDescribesIncludeRoutes`: the `--help` stdout contains `-include-routes` and `also list the routes playing on this output, in join order`.
- [ ] T024 [US2] In `tests/integration/outputs_get_test.go`, add a mock `mockOutputRoutesServer(t, output map[string]any, routesStatus int, routesBody string)`. It serves `GET /api/v2/outputs/<id>` from `output` (a 404 for any other ID) and `GET /api/v2/outputs/<id>/routes` with `routesStatus` and `routesBody`, and it counts requests per path. Keep `mockOutputServer` unchanged. Then add these tests (expected strings as in T022):
  - `TestOutputsGet_IncludeRoutes_YAML`: two routes, exit 0, stdout is the record plus `routes`, and 1 request to each path.
  - `TestOutputsGet_IncludeRoutes_JSON`: the same with `--json`; one object with `routes`.
  - `TestOutputsGet_IncludeRoutes_Empty`: `[]`, so `routes: []` in YAML and `"routes":[]` in JSON.
  - `TestOutputsGet_IncludeRoutes_OutputNotFound`: exit 5, stderr `output not found: <id>`, stdout empty, and 0 requests to `/routes`.
  - `TestOutputsGet_IncludeRoutes_OlderHub`: a table over routes 404 and 405. Each gives exit 3; stderr containing `this hub cannot list an output's routes; it needs hub 0.1.21 or later`; stderr NOT containing `not found`; and stdout empty.
  - `TestOutputsGet_IncludeRoutes_OlderHub_Verbose`: with `--verbose`, stderr also has a `detail:` line.
  - `TestOutputsGet_IncludeRoutes_RoutesFailOtherwise`: a table of a 500, and a 200 with the body `{"not":"an array"}`. Each gives exit 3, stdout empty, and no older-hub message.
  - `TestOutputsGet_WithoutIncludeRoutes_OneRequest`: no flag gives exactly 1 request, to `/api/v2/outputs/<id>`, and stdout byte-identical to `TestOutputsGet_SuccessYAML`'s expectation.
  - `TestOutputsGet_IncludeRoutes_FlagAfterID` and `…_FlagBeforeID`: both orders work.
- [ ] T025 [P] [US2] In `cmd/sonora/main_test.go`, add `TestRunGetList_IncludeRoutesElsewhere_UsageError`. It is a table over the argument lists `get outputs --include-routes`, `list outputs --include-routes`, `get inputs/x --include-routes`, `get groups/x --include-routes`, `get routes/x --include-routes`, `get routes --include-routes`, `get master-mute --include-routes`, `get tts-cache --include-routes`, `get outputs --include-routes=true` and `get --include-routes outputs/x` (flags follow the resource, as for every other `get` flag), each with `--hub-url` pointing at an `httptest` server that counts requests. For each, assert exit 2; that stderr contains `error: --include-routes applies only to 'sonora get outputs/<id>'`; and 0 requests. Also assert that `get outputs/x -- --include-routes` is not caught by the dispatcher (anything after `--` is not scanned).

### Implementation for User Story 2

- [ ] T026 [US2] In `hub/routes.go`, extract `validateRouteFields(routeID, inputID, targetID, targetType, status string) error` from `validateRoute`, keeping its exact messages, and make `validateRoute` call it. Do not change any behaviour; the existing route contract tests must still pass.
- [ ] T027 [US2] In `hub/outputs.go`:
  - add the `DuckStateFull` and `DuckStateLowered` constants (godoc: "the hub may add values; treat an unknown one as DuckStateFull");
  - add the `RouteOnOutput` type, mirroring `#/components/schemas/RouteOnOutputResponse` field for field, in the order and with the Go types of `specs/012-join-modes/data-model.md`. `StartedAt` is a `*string`. Write godoc on the type, and on `JoinMode`, `DuckState` and `Main`.
  - add `func (r RouteOnOutput) Lowered() bool { return r.DuckState == DuckStateLowered }`, with godoc;
  - add `GetRoutesForOutput(ctx context.Context, client *http.Client, baseURL, outputID string) ([]RouteOnOutput, error)`, built like `GetOutput` and `ListRoutes`. It calls `GET {baseURL}/api/v2/outputs/{url.PathEscape(outputID)}/routes`. A 404 gives `&NotFoundError{Resource: "output", ID: outputID}`, and any other non-2xx gives `&StatusError`. It decodes into `[]RouteOnOutput`, turns nil into `[]RouteOnOutput{}`, and checks each entry with `validateRouteFields`, wrapping a failure in `&DecodeError`. It does not check `JoinMode` or `DuckState`.

  The godoc on `GetRoutesForOutput` says it lists every route playing on the output, addressed to it or to a group it belongs to, in join order. This makes T020 and T021 pass. It depends on T026.
- [ ] T028 [US2] In `internal/render/outputs.go`, add:
  - `RenderOutputWithRoutesYAML(o hub.Output, routes []hub.RouteOnOutput) string`, which returns `RenderOutputYAML(o)` followed by `routes: []\n` when there are no routes. Otherwise it writes `routes:\n` and, per route, the nine fields in the order of T022, using `%q` for strings and `%t` for booleans.
  - `RenderOutputWithRoutesJSON(o hub.Output, routes []hub.RouteOnOutput) string`, which marshals an unexported `outputWithRoutesPayload struct { hub.Output; Routes []hub.RouteOnOutput \`json:"routes"\` }` with nil turned into an empty slice, and appends `"\n"`. It follows the panic-on-impossible-error pattern of `RenderOutputJSON`.

  This makes T022 pass. It depends on T027.
- [ ] T029 [US2] In `internal/cli/outputs/get.go`, `RunGet`:
  - add `includeRoutes := fs.Bool("include-routes", false, "also list the routes playing on this output, in join order")`;
  - after `hub.GetOutput` succeeds, and only when the flag is set, call `hub.GetRoutesForOutput` with the same `client` and context;
  - on a `*hub.NotFoundError`, or a `*hub.StatusError` with `StatusCode == 405`, print `error: this hub cannot list an output's routes; it needs hub 0.1.21 or later (hub URL: <url>)`, plus `detail: <err>` under `--verbose`, and return `exitcode.For(hub.ClassHub)`. Add a comment saying the output request has just succeeded, so a 404 here means the hub lacks the operation, and that a hub with the stop-routes operation on the same path answers 405;
  - on any other error, use the existing classify-and-report path;
  - print nothing to stdout until both calls have succeeded, then render with `RenderOutputWithRoutesYAML` or `RenderOutputWithRoutesJSON`. Without the flag, keep the current code path byte for byte.
  - update `RunGet`'s doc comment.

  This makes T023 and T024 pass. It depends on T027 and T028.
- [ ] T030 [US2] In `cmd/sonora/main.go`, `dispatchGetList`:
  - at the top, after `usage` is chosen and before the `master-mute` special case, scan `args` up to the first `--` for an element equal to `--include-routes` or starting with `--include-routes=`;
  - if one is found, unless `verb == "get"`, `args[0]` parses (via `respath.Parse`) as `respath.Outputs` and the path has a non-empty ID, print `usage` and `error: --include-routes applies only to 'sonora get outputs/<id>'` to stderr and return 2;
  - put the scan in a small helper `hasIncludeRoutes(args []string) bool` beside `startsWithResource`;
  - update the function's doc comment;
  - add `sonora get outputs/kitchen --include-routes` to the examples in the top-level help text.

  This makes T025 pass.
- [ ] T031 [US2] In `README.md`, document `--include-routes` in the `get outputs/<id>` section:
  - what it lists: every route playing on the output, including group routes, in join order, with `joinMode`, `duckState` and `main`;
  - that YAML shows nine fields per route and JSON shows all of them;
  - that it needs hub 0.1.21 or later, and that an older hub exits 3 with a message saying so;
  - one example. Do it after T019 (same file).

**Checkpoint**: T020–T025 pass, and every existing outputs, dispatch and route test is unchanged and
still passes.

---

## Phase 5: User Story 3 - See a route's join mode and outputs, and an input's default (Priority: P3)

**Goal**: The YAML of `get routes/<id>` shows `joinMode` and `outputs`, each left out when the hub did
not send it. Every one-input YAML record shows `defaultJoinMode`, or null. The JSON and the list views
are unchanged.

**Independent Test**: Run `get routes/<id>` and `get inputs/<id>` against a fake hub that sends the
fields and against one that leaves them out, and compare the YAML and JSON. Check that `list routes`
and `list inputs` are unchanged.

### Tests for User Story 3 ⚠️ write first, watch them fail

- [ ] T032 [P] [US3] In `tests/unit/render_routes_test.go`, which already tests `RenderRouteYAML`, add:
  - `TestRenderRouteYAML_JoinModeAndOutputsAfterPaused`: `JoinMode: "MIX"` and `Outputs: []string{"kitchen", "office"}` give the ten existing lines, then `joinMode: "MIX"`, `outputs:`, `  - "kitchen"` and `  - "office"`, as an exact string.
  - `TestRenderRouteYAML_EmptyOutputsList`: `Outputs: []string{}` gives `outputs: []`.
  - `TestRenderRouteYAML_OlderHubOmitsBoth`: `JoinMode: ""` and `Outputs: nil` give exactly the ten existing lines.
  - `TestRenderRouteYAML_UnknownJoinModeShownAsSent`: `CROSSFADE` appears unchanged.
  - `TestRenderRouteJSON_ShapeUnchanged`: the JSON still marshals `hub.Route` as is: 12 keys, `outputs` in the hub's order.
- [ ] T033 [P] [US3] In `tests/unit/render_inputs_test.go`:
  - extend `TestRenderInputYAML_AllFieldsAsBareRecord` so that `defaultJoinMode: "DUCK_OTHERS"` is the last line when it is set;
  - add `TestRenderInputYAML_NoDefaultJoinModeIsNull`, where a nil value gives a bare `defaultJoinMode: null` as the last line;
  - add `TestRenderInputsYAML_ListUnchanged`, asserting that `RenderInputsYAML` output contains no `defaultJoinMode`.
- [ ] T034 [P] [US3] In `tests/integration/routes_get_test.go`, add `TestRoutesGet_ShowsJoinModeAndOutputs` (a 0.1.28 body with `joinMode` and `outputs`) and `TestRoutesGet_OlderHubBodyOmitsFields` (a body without both keys, whose YAML equals `TestRoutesGet_SuccessYAML`'s expectation). In `tests/integration/inputs_get_test.go`, add `TestInputsGet_ShowsDefaultJoinMode` and `TestInputsGet_NullDefaultJoinMode`. For each, check that `--json` contains the same keys.
- [ ] T035 [P] [US3] In the existing integration tests for `create`, `rename` and `enable|disable inputs/<id>` (`tests/integration/inputs_create_test.go`, `inputs_rename_test.go` and `inputs_enabled_test.go`), wherever one compares the exact YAML, add the trailing `defaultJoinMode: null` line, or the value its fake hub sends. Change nothing else in those tests. They fail until T037. This is the deliberate change recorded in research.md §2 and FR-013.

### Implementation for User Story 3

- [ ] T036 [P] [US3] In `internal/render/routes.go`, `RenderRouteYAML`:
  - after `paused`, write `joinMode: %q` when `r.JoinMode != ""`;
  - when `r.Outputs != nil`, write `outputs: []` for an empty slice, otherwise `outputs:` followed by one `  - %q` line per output;
  - update the doc comment: say the ten fields are always present and the two join-mode fields are left out when the hub did not send them;
  - leave `RenderRoutesYAML`, `RenderRoutesJSON` and `RenderRouteJSON` unchanged.

  This makes T032 and the route half of T034 pass.
- [ ] T037 [P] [US3] In `internal/render/inputs.go`, `RenderInputYAML`: after `writeCreatedAt`, write `defaultJoinMode: %q`, or a bare `defaultJoinMode: null` when it is nil. Add a small `writeDefaultJoinMode` helper next to `writeCreatedAt`. Leave `RenderInputsYAML` unchanged. Update the doc comment: this record is also printed by `create`, `rename` and `enable|disable inputs/<id>`. This makes T033, T035 and the input half of T034 pass.
- [ ] T038 [US3] In `README.md`, add a line to the `get routes/<id>` section saying it shows `joinMode` and the `outputs` the route plays on right now (hub 0.1.21 or later), and that a group route may list fewer members than its group has. Add a line to the `get inputs/<id>` section about `defaultJoinMode`. Do it after T031 (same file).

**Checkpoint**: T032–T035 pass. `list routes` and `list inputs` are unchanged. The tests changed in
T035 differ only by the new last line.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T039 Run `gofmt -l .` (it must print nothing), `go vet ./...` and `go test -count=1 ./...`, reading each command's own exit code. Fix any failure, including in `tests/unit/hub_godoc_test.go` and `tests/unit/stray_refs_test.go`. Also confirm that `git diff main -- go.mod go.sum` prints nothing: the feature adds no dependency.
- [ ] T040 Work through `specs/012-join-modes/quickstart.md`:
  - the four `--help` outputs, via `go run ./cmd/sonora …`;
  - the **read-only** checks against `http://multiroom.lan:8080`: `get outputs/<id> --include-routes`, the same with `--json`, `get routes/<id>` and `get inputs/<id>`, comparing against `curl -s …/api/v2/outputs/<id>/routes`.

  Make no request that plays, stops or changes anything on `multiroom.lan`. An audible `--join-mode` check needs the user's permission first.
- [ ] T041 Record the T039 and T040 results in `specs/012-join-modes/verification.md`: the dates, the commands, the test counts, and the production hub version from `/api-docs`. Mark the tasks `[X]` in this file as they are done.
- [ ] T042 Update `docs/backlog/`: delete `join-modes-in-cli.md`, and remove its row from `docs/backlog/INDEX.md`.
- [ ] T043 Run `git ls-files --eol` over the changed and new files, and confirm that none shows `w/crlf`.

---

## Dependencies & Execution Order

- **Setup (T001)** comes before everything.
- **The stories**: US1 (T002–T019), US2 (T020–T031) and US3 (T032–T038) do not depend on each other.
  Within each story, every test task comes before the implementation task it drives:
  - **US1**: T002→T012; T003→T013; T004→T014; T005→T015; T006 and T009→T016; T007 and T010→T017;
    T008 and T011→T018.
    - T012 comes before T015 (`joinmode` uses the `hub` constants).
    - T015 comes before T016, T017 and T018.
  - **US2**: T020 and T021→T026→T027; T022→T028; T023 and T024→T029; T025→T030.
    - T027 comes before T028, which comes before T029.
  - **US3**: T032→T036; T033 and T035→T037; T034→T036 and T037. Each test task fails before its
    implementation task starts.
- **Shared files**:
  - `hub/routes.go`: T012 (US1) and T026 (US2) edit it, so do them one after the other.
  - `README.md`: T019, T031 and T038 edit it, so do them in that order, never in parallel.
- **Polish (T039–T043)** comes after every story it covers.

## Parallel Opportunities

- **US1**:
  - the tests T002–T011 together (ten different files);
  - then T013 and T014 together, alongside T012;
  - then T016, T017 and T018 together, once T015 is done.
- **US2**:
  - the tests T020–T023 and T025 together, and T024 after T023 (both are outputs-get files, but in
    different suites);
  - T030 can run in parallel with T026–T029.
- **US3**: T032–T035 together, then T036 and T037 together.
- **Across stories**: once T001 is done, the three stories can be built side by side, apart from the
  two shared files above.

## Implementation Strategy

1. **MVP = US1** (T001–T019). It is the one thing the CLI cannot do at all today: announce over music,
   or mix two sources. It is additive to `hub/` (three fields and three constants).
2. **US2** adds the new `hub` operation and type, and `--include-routes`. It is the only way to see
   the effect of US1, and to see group routes on an output.
3. **US3** is a YAML-only change.
4. Ship all three in one minor tag, `v0.3.0` (the last tag is `v0.2.1`), with release notes listing
   the `hub` additions and the new `defaultJoinMode` line in the input YAML. Then open the PR. Bumping
   sonora-mcp afterwards (`go get github.com/Sonora-Multiroom/sonora-cli@v0.3.0`) is outside this
   feature.
