# Main Implementation Plan

> **Revision**: 2026-11-23 — Seeded from .specify/archive/001-list-outputs (first archival): summary, Technical Context, project structure, hub-URL configuration, exit codes, testing strategy. Describes the layout as planned by that feature.
> **Revision**: 2026-11-23 — Archived .specify/archive/002-outputs-get: added the `outputs get` command to Summary, Technical Context, Project Structure, Routing & Navigation (fifth exit code `5`) and Testing Strategy; `outputs.Run` is now `outputs.RunList`. No Technical Context value conflicted with the 001 values.
> **Revision**: 2026-11-23 — Archived .specify/archive/003-inputs-list-get: added `inputs list` and `inputs get` to Summary, Technical Context (Scale/Scope now composes four commands), Project Structure, Configuration, Routing & Navigation and Testing Strategy. `NotFoundError` is now resource-generic (`Resource`, `ID`), so `outputs get` and `inputs get` share one not-found type. No Technical Context value conflicted; the remaining values are identical to the 001/002 ones and each now carries a 003 ref. The main plan has no legacy per-feature blocks, so the layout is unmixed.
> **Revision**: 2026-11-23 — Archived .specify/archive/004-routes-list-get (moved from specs/ in the same run): added `routes list` and `routes get` to Summary, Technical Context (Scale/Scope now composes six commands; Language/Version, Primary Dependencies, Storage, Target Platform, Project Type, Performance Goals and Constraints each take a 004 ref, stating the same values), Project Structure, Configuration, Routing & Navigation (exit `5` now covers route identifiers) and Testing Strategy. No Technical Context value conflicted with the 001–003 values.
> **Revision**: 2026-11-23 — Archived .specify/archive/005-groups-list-get (moved from specs/ in the same run): added `groups list` and `groups get` to Summary, Technical Context (Scale/Scope now composes eight commands; Language/Version, Primary Dependencies, Storage, Testing, Target Platform, Project Type and Constraints each take a 005 ref for the same value; Performance Goals now names the groups commands as well), Project Structure, Configuration, Routing & Navigation (exit `5` now covers group identifiers) and Testing Strategy. No Technical Context value conflicted with the 001–004 values, and the main plan has no legacy per-feature blocks, so the layout is unmixed.
> **Revision**: 2026-11-23 — Archived .specify/archive/006-play-command (moved from specs/ in the same run): added `play` to Summary, Technical Context (Scale/Scope now composes nine commands; Primary Dependencies adds `strconv`; Performance Goals adds the playback round trip; Constraints adds the play exit classes; Language/Version, Storage, Testing, Target Platform and Project Type each take a 006 ref for the same value), Configuration, Routing & Navigation (exit codes `6`–`10` and the play routing entry), Testing Strategy and the Structure Decision. The verb-less `play` noun needed an early `main.go` special case; that gotcha is recorded in the agent file. No Technical Context value conflicted with the 001–005 values, and the main plan has no legacy per-feature blocks, so the layout is unmixed.
> **Revision**: 2026-11-23 — tinyspec help-command (.specify/archive/tiny/2026-08-26-help-command/help-command.md): `help`, `-h`, `--help` and bare invocation print the usage text to stdout with exit `0`, dispatched in `cmd/sonora/main.go` after the `--version`/`-v` check and before the noun/verb split; the help text is a constant there, and `cmd/sonora/main_test.go` covers the `run()` help and usage paths. The zero-argument case no longer exits `2`; missing-verb and unknown-command errors are unchanged.
> **Revision**: 2026-11-23 — Archived .specify/archive/007-refactor-cli-commands (moved from specs/ in the same run): the read commands and `play` moved to verb-first, path-style invocation (`get`/`list <resource>[/<id>]`, `play <uri> <outputs|groups>/<id> --display-name`). Added `internal/cli/respath` and a verb-first `main.go` dispatcher; updated Summary, Technical Context, Project Structure, Routing & Navigation and Testing Strategy, and rewrote the old command shapes in them in place (user-confirmed). Technical Context conflicts resolved to 007's values: Constraints dropped `play`'s ambiguous-target exit code `7` (ref 006, `play`'s ambiguous-target clause only) and the usage error for `--group` with `--output`, both removed by 007; Scale/Scope replaced the nine noun-verb command shapes and `play`'s `--group`/`--output`/`--name` flags with the `get`/`list`/`play` shapes (the old text was carried under the 001–006 refs); Performance Goals now states `play`'s round trip as one target-existence request plus the play request instead of up to two target-resolution requests. The main plan has no legacy per-feature blocks, so the layout is unmixed.
> **Revision**: 2026-11-23 — Archived .specify/archive/008-route-command (moved from specs/ in the same run): added `route` to Summary, Technical Context (Scale/Scope now composes ten commands; Primary Dependencies adds `errors`; Constraints adds exit codes `11` and `12`; Performance Goals adds the routing round trip; Language/Version, Storage, Testing, Target Platform and Project Type each take an 008 ref for the same value), Project Structure, Routing & Navigation, exit codes and Testing Strategy. The feature's plan names `hub.ResolveTarget` for its target check; 007 had already replaced that with the per-type existence check, which `route` reuses. No Technical Context value conflicted with the 001–007 values, and the main plan has no legacy per-feature blocks, so the layout is unmixed.

## Summary

`sonora outputs list` is the CLI's first command: a read-only `GET /api/v2/outputs` call
(optionally with `?includeDisabled=true`) rendered as YAML by default or JSON with
`--json`, each output showing identifier, display name, volume, mute state, enabled state,
and hardware availability. It also establishes the reusable skeleton later commands build
on: a timeout-bound, single-attempt HTTP client (Principle IV), openapi.json-validated
request/response types (Principle II), and dual human/`--json` rendering with
class-distinguishing exit codes (Principle V), all on the Go standard library only.
[Source: .specify/archive/001-list-outputs/plan.md -> "Summary"]

`sonora outputs get <output-id>` is the single-item counterpart: a read-only
`GET /api/v2/outputs/{outputId}` call (operationId `getOutput`) rendered as one YAML record
(or JSON with `--json`), returned regardless of the output's enabled state. It reuses the
`Output` type, `hub.NewClient()`, hub-URL resolution, flag names and rendering approach
unchanged; the new pieces are `GetOutput`, a `NotFoundError`/`ClassNotFound` (exit `5`) class
for 404, and single-record rendering.
[Source: .specify/archive/002-outputs-get/plan.md -> "Summary"]

`sonora inputs list` and `sonora inputs get <input-id>` mirror the outputs pair: read-only
`GET /api/v2/inputs` (operationId `listInputs`) and `GET /api/v2/inputs/{inputId}`
(operationId `getInput`) calls with the same flags, hub-URL resolution, exit-code scheme and
rendering. The new pieces are the `Input` type, whose attributes differ from `Output` (no
volume, mute or availability; instead `uri`, `source`, `autoRemove`, `pauseable` and a
nullable `createdAt`), the `inputs` noun dispatch, and `NotFoundError` now carrying a
`Resource` name so that `inputs get`'s 404 reuses `ClassNotFound` (exit `5`).
[Source: .specify/archive/003-inputs-list-get/plan.md -> "Summary"]

`sonora routes list` and `sonora routes get <route-id>` are read-only `GET /api/v2/routes`
(operationId `listRoutes`) and `GET /api/v2/routes/{routeId}` (operationId `getRoute`) calls,
rendered as YAML by default or JSON with `--json`, reusing `hub.NewClient()`, hub-URL
resolution, the exit-code scheme and the stdout/stderr separation unchanged. The new pieces are
the `Route` type (the list view shows five fields, the get view ten), three AND-combined list
filters (`--status`, `--input-id`, `--target-id`) that the hub validates and applies, and the
`routes` noun dispatch.
[Source: .specify/archive/004-routes-list-get/plan.md -> "Summary"]

`sonora groups list` and `sonora groups get <group-id>` are read-only `GET /api/v2/groups` (operationId `listGroups`) and `GET /api/v2/groups/{groupId}` (operationId `getGroup`) calls, rendered as YAML by default or JSON with `--json`, reusing `hub.NewClient()`, hub-URL resolution, the exit-code scheme and the stdout/stderr separation unchanged. The new pieces are the `Group` type, whose list and get views show the same five fields (its `outputIds` member list renders as `[]` when a group has no member outputs), and the `groups` noun dispatch.
[Source: .specify/archive/005-groups-list-get/plan.md -> "Summary"]

`sonora play <uri> <target-id>` is the CLI's first mutating command: one `POST /api/v2/play` (operationId `playback`) creates an ephemeral input and a route in a single hub round trip, and the command returns as soon as the hub responds, without polling. Four new exit codes (6, 8, 9, 10) separate the failure modes the hub's `400`, `422`, `502` and `503` responses describe; the ambiguous-target code `7` was removed by 007, which also replaced its target-id and `--group`/`--output` addressing with a resource path.
[Source: .specify/archive/006-play-command/plan.md -> "Summary"]

`sonora help`, `sonora -h`, `sonora --help` and bare `sonora` print the command table, the global flags and example invocations to stdout with exit `0`, without contacting the hub. The usage text is a constant in `cmd/sonora/main.go`, dispatched after the `--version`/`-v` check and ahead of the noun/verb split.
[Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]

The verb-first refactor moved every command above onto one grammar: `sonora get <resource>[/<id>]` and `sonora list <resource>` for inputs, outputs, groups and routes (aliases `in`/`out`/`gr`/`rt`), and `sonora play <uri> <outputs|groups>/<id>` with `--display-name`. The `sonora <resource> list|get` forms named in the paragraphs above no longer exist and fail with a standard usage error. It is a dispatch and argument-parsing change: the `internal/cli/<resource>` `RunList`/`RunGet` functions are reused with the same argument shape, so requests, rendering and exit codes for the read commands are unchanged. A new `internal/cli/respath` package parses the resource path once for both `get`/`list` and `play`.
[Source: .specify/archive/007-refactor-cli-commands/plan.md -> "Summary"]

`sonora route inputs/<input-id> <outputs|groups>/<target-id>` connects an existing, already-defined input to an existing output or output group with one `POST /api/v2/routes` call (operationId `createRoute`) and creates no input. Unlike `play`, which always mints a new input and resolves a bare target identifier, `route` takes path-style arguments only, so the target type comes from the prefix. Before creating the route it checks that the input exists (`GET /api/v2/inputs/{inputId}`) and that the target exists as the named type. It adds two exit classes, input not found (`11`) and target not found (`12`), so those failures do not collide with the generic not-found exit `5`.
[Source: .specify/archive/008-route-command/plan.md -> "Summary"]

## Technical Context

**Language/Version**: Go 1.27.0 (module-based; `go.mod` created by the list-outputs feature).
[Source: .specify/archive/001-list-outputs/plan.md -> "Language/Version"; .specify/archive/002-outputs-get/plan.md -> "Language/Version"; .specify/archive/003-inputs-list-get/plan.md -> "Language/Version"; .specify/archive/004-routes-list-get/plan.md -> "Language/Version"; .specify/archive/005-groups-list-get/plan.md -> "Language/Version"; .specify/archive/006-play-command/plan.md -> "Language/Version"; .specify/archive/007-refactor-cli-commands/plan.md -> "Language/Version"; .specify/archive/008-route-command/plan.md -> "Language/Version"]

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `strconv`, `strings`, `errors`, `net/http/httptest` (tests). No third-party module; a small fixed-shape YAML emitter in
`internal/render` replaces a YAML library because every output field is a flat scalar.
[Source: .specify/archive/001-list-outputs/plan.md -> "Primary Dependencies"; .specify/archive/002-outputs-get/plan.md -> "Primary Dependencies"; .specify/archive/003-inputs-list-get/plan.md -> "Primary Dependencies"; .specify/archive/004-routes-list-get/plan.md -> "Primary Dependencies"; .specify/archive/005-groups-list-get/plan.md -> "Primary Dependencies"; .specify/archive/006-play-command/plan.md -> "Primary Dependencies"; .specify/archive/007-refactor-cli-commands/plan.md -> "Primary Dependencies"; .specify/archive/008-route-command/plan.md -> "Primary Dependencies"]

**Storage**: A single optional local config file, `~/.config/sonora/config.json` (JSON), read
(never written) as the third-priority source for the hub base URL. No database.
[Source: .specify/archive/001-list-outputs/plan.md -> "Storage"; .specify/archive/002-outputs-get/plan.md -> "Storage"; .specify/archive/003-inputs-list-get/plan.md -> "Storage"; .specify/archive/004-routes-list-get/plan.md -> "Storage"; .specify/archive/005-groups-list-get/plan.md -> "Storage"; .specify/archive/006-play-command/plan.md -> "Storage"; .specify/archive/007-refactor-cli-commands/plan.md -> "Storage"; .specify/archive/008-route-command/plan.md -> "Storage"]

**Testing**: `go test` (stdlib `testing`) in three layers — unit (`tests/unit`), contract
(`tests/contract`, against an `httptest.Server` shaped from `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation). Written before implementation per
Principle VI. [Source: .specify/archive/001-list-outputs/plan.md -> "Testing"; .specify/archive/002-outputs-get/plan.md -> "Testing"; .specify/archive/003-inputs-list-get/plan.md -> "Testing"; .specify/archive/005-groups-list-get/plan.md -> "Testing"; .specify/archive/006-play-command/plan.md -> "Testing"; .specify/archive/007-refactor-cli-commands/plan.md -> "Testing"; .specify/archive/008-route-command/plan.md -> "Testing"]

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), single static Go binary
via `go build`. [Source: .specify/archive/001-list-outputs/plan.md -> "Target Platform"; .specify/archive/002-outputs-get/plan.md -> "Target Platform"; .specify/archive/003-inputs-list-get/plan.md -> "Target Platform"; .specify/archive/004-routes-list-get/plan.md -> "Target Platform"; .specify/archive/005-groups-list-get/plan.md -> "Target Platform"; .specify/archive/006-play-command/plan.md -> "Target Platform"; .specify/archive/007-refactor-cli-commands/plan.md -> "Target Platform"; .specify/archive/008-route-command/plan.md -> "Target Platform"]

**Project Type**: Single project — CLI plus internal packages.
[Source: .specify/archive/001-list-outputs/plan.md -> "Project Type"; .specify/archive/002-outputs-get/plan.md -> "Project Type"; .specify/archive/003-inputs-list-get/plan.md -> "Project Type"; .specify/archive/004-routes-list-get/plan.md -> "Project Type"; .specify/archive/005-groups-list-get/plan.md -> "Project Type"; .specify/archive/006-play-command/plan.md -> "Project Type"; .specify/archive/007-refactor-cli-commands/plan.md -> "Project Type"; .specify/archive/008-route-command/plan.md -> "Project Type"]

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms; full
output, input, route and group list, and a single output, input, route or group, rendered in under 1s under normal network conditions; for `play`, a full playback round trip (one target-existence request plus the play request) in under 5s, as an accept/reject confirmation rather than confirmation that audio is playing. For `route`, a full routing round trip (one input-existence request, one target-existence request and the create request) in under 5s, as an accept/reject confirmation.
[Source: .specify/archive/001-list-outputs/plan.md -> "Performance Goals"; .specify/archive/002-outputs-get/plan.md -> "Performance Goals"; .specify/archive/003-inputs-list-get/plan.md -> "Performance Goals"; .specify/archive/004-routes-list-get/plan.md -> "Performance Goals"; .specify/archive/005-groups-list-get/plan.md -> "Performance Goals"; .specify/archive/006-play-command/plan.md -> "Performance Goals"; .specify/archive/007-refactor-cli-commands/plan.md -> "Performance Goals"; .specify/archive/008-route-command/plan.md -> "Performance Goals"]

**Constraints**: 5s total request timeout, single attempt, no automatic retry on
network/timeout failure; no unbounded waits; exit codes distinguish success / usage error /
hub error / network error and, for lookups by identifier, not-found (exit `5`). For `play`, exit codes also distinguish validation (`6`), route creation failed (`8`), source unreachable (`9`) and upstream service unavailable (`10`); exit `7` is unused. For `route`, input not found (`11`) and target not found (`12`) are added, and route creation failed reuses `8`. Old-style command forms, an unrecognized resource or alias, a malformed resource path, and the removed `play` flags (`--group`, `--output`, `--name`) are all usage errors (`2`), with no migration hint (hard cutover). Displayed data of the read commands is byte-for-byte unchanged by the refactor.
[Source: .specify/archive/001-list-outputs/plan.md -> "Constraints"; .specify/archive/002-outputs-get/plan.md -> "Constraints"; .specify/archive/003-inputs-list-get/plan.md -> "Constraints"; .specify/archive/004-routes-list-get/plan.md -> "Constraints"; .specify/archive/005-groups-list-get/plan.md -> "Constraints"; .specify/archive/006-play-command/plan.md -> "Constraints"; .specify/archive/007-refactor-cli-commands/plan.md -> "Constraints"; .specify/archive/008-route-command/plan.md -> "Constraints"]

**Scale/Scope**: Ten entry points: `sonora get <resource>` and `sonora list <resource>` (collection form, the latter an exact synonym) and `sonora get <resource>/<id>` (single item) for each of `inputs`/`in`, `outputs`/`out`, `groups`/`gr` and `routes`/`rt`, plus `sonora play <uri> <outputs|groups>/<id>` (`--volume`, `--display-name`, `--json`, `--verbose`, `--hub-url`), and `sonora route inputs/<input-id> <outputs|groups>/<target-id>` (`--json`, `--verbose`, `--hub-url`; no `--volume` or `--display-name`). The collection form of inputs, outputs and groups takes `--include-disabled`; routes takes `--status`, `--input-id` and `--target-id`; all take `--json`, `--verbose` and `--hub-url`, and the single-item form takes only those three. [Source: .specify/archive/001-list-outputs/plan.md -> "Scale/Scope"; .specify/archive/002-outputs-get/plan.md -> "Scale/Scope"; .specify/archive/003-inputs-list-get/plan.md -> "Scale/Scope"; .specify/archive/004-routes-list-get/plan.md -> "Scale/Scope"; .specify/archive/005-groups-list-get/plan.md -> "Scale/Scope"; .specify/archive/006-play-command/plan.md -> "Scale/Scope"; .specify/archive/007-refactor-cli-commands/plan.md -> "Scale/Scope"; .specify/archive/008-route-command/plan.md -> "Scale/Scope"]

## Project Structure

### Source Code (repository root)

```text
cmd/
└── sonora/
    ├── main.go                    # Entrypoint: verb-first dispatch (007: `get`, `list`, `play`; the resource path is resolved via internal/cli/respath), exit code translation, usage text for `help`/`-h`/`--help`/bare invocation
    └── main_test.go               # Unit tests for `run()`'s help and usage paths (tinyspec `help-command`)

internal/
├── hub/                           # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                  # HTTP client construction (timeout, single attempt) — Principle IV
│   ├── outputs.go                 # ListOutputs(ctx, includeDisabled) — GET /api/v2/outputs; GetOutput(ctx, client, baseURL, outputID) — GET /api/v2/outputs/{outputId}
│   ├── inputs.go                  # NEW (003): ListInputs(ctx, client, baseURL, includeDisabled) — GET /api/v2/inputs; GetInput(ctx, client, baseURL, inputID) — GET /api/v2/inputs/{inputId}
│   ├── routes.go                  # NEW (004): ListRoutes(ctx, client, baseURL, status, inputID, targetID) — GET /api/v2/routes; GetRoute(ctx, client, baseURL, routeID) — GET /api/v2/routes/{routeId}
│   ├── groups.go                  # NEW (005): ListGroups(ctx, client, baseURL, includeDisabled) — GET /api/v2/groups; GetGroup(ctx, client, baseURL, groupID) — GET /api/v2/groups/{groupId}
│   ├── play.go                    # NEW (006): PlaybackRequest, PlaybackResponse, Playback; 007 replaced ResolveTarget with a per-type existence check (the path prefix already fixes the target type)
│   └── errors.go                  # Error classification: usage / hub / network / not found (NotFoundError {Resource, ID}, ClassNotFound); 007 removed ClassAmbiguous and AmbiguousTargetError (exit 7)
├── config/
│   └── config.go                  # Read-only load of ~/.config/sonora/config.json; hub URL precedence resolution
├── cli/
│   ├── respath/                   # NEW (007): shared resource-path parsing — alias table (in/out/gr/rt), `resource/id` split on the first `/`, id pattern validation; used by main.go and play
│   │   └── respath.go
│   ├── clihelp/
│   │   └── usage.go               # Help/usage text for the verb-first grammar (007)
│   ├── outputs/
│   │   ├── list.go                # `sonora get|list outputs` (RunList): flag definitions, dispatch, exit code mapping
│   │   └── get.go                 # `sonora get outputs/<output-id>` (RunGet): id supplied by the dispatcher, flag definitions, exit code mapping
│   ├── inputs/                    # NEW (003): same shape as outputs/
│       ├── list.go                # `sonora get|list inputs` (RunList): flag definitions, dispatch, exit code mapping
│       └── get.go                 # `sonora get inputs/<input-id>` (RunGet): id supplied by the dispatcher, flag definitions, exit code mapping
│   ├── routes/                    # NEW (004): same shape as outputs/ and inputs/
│       ├── list.go                # `sonora get|list routes` (RunList): filter flags, dispatch, exit code mapping
│       └── get.go                 # `sonora get routes/<route-id>` (RunGet): id supplied by the dispatcher, flag definitions, exit code mapping
│   ├── groups/                    # NEW (005): same shape as outputs/, inputs/ and routes/
│       ├── list.go                # `sonora get|list groups` (RunList): flag definitions, dispatch, exit code mapping
│       └── get.go                 # `sonora get groups/<group-id>` (RunGet): id supplied by the dispatcher, flag definitions, exit code mapping
│   └── play/
│       └── play.go                # NEW (006): `sonora play <uri> <outputs|groups>/<id>`; 007: target parsed via respath, `--group`/`--output` removed, `--name` renamed `--display-name`
└── render/
    ├── outputs.go                 # YAML (default, fixed-shape emitter) / JSON (--json) rendering; RenderOutputYAML/RenderOutputJSON for a single unwrapped record
    ├── inputs.go                  # NEW (003): RenderYAML/RenderJSON for input lists; RenderInputYAML/RenderInputJSON for a single record
    ├── routes.go                  # NEW (004): RenderRoutesYAML/RenderRoutesJSON for list views (five fields); RenderRouteYAML/RenderRouteJSON for a single record (ten fields)
    └── groups.go                 # NEW (005): RenderGroupsYAML/RenderGroupsJSON for list views; RenderGroupYAML/RenderGroupJSON for a single record (the same five fields in both)

tests/
├── contract/
│   ├── outputs_list_test.go       # httptest.Server shaped from openapi.json OutputResponse
│   └── outputs_get_test.go        # httptest.Server shaped from getOutput's 200/404 responses; request path + decode contract
├── integration/
│   ├── outputs_list_test.go       # Full `sonora get outputs` / `sonora list outputs` invocation against the mock server
│   └── outputs_get_test.go        # Full `sonora get outputs/<id>` invocation: found, not-found, malformed-path usage error, --json, exit codes
└── unit/
    ├── cli_outputs_test.go        # Command-level tests (call sites use outputs.RunList)
    ├── hub_client_test.go         # Timeout, single-attempt, error classification
    ├── config_test.go             # Hub URL precedence, malformed config file handling
    ├── render_outputs_test.go     # YAML/JSON rendering, zero-outputs, unavailable-output distinguishability
    └── render_output_get_test.go  # Single-output YAML/JSON rendering, unavailable-output distinguishability
```

Go layout: `cmd/` for the entrypoint, `internal/` for non-exported packages. `internal/hub`
is the service layer, `internal/render` the reusable formatting layer, `internal/config`
the read-only config pair, `internal/cli/outputs` the per-noun flag parsing and dispatch.
[Source: .specify/archive/001-list-outputs/plan.md -> "Structure Decision"]

Adding a second verb to the `outputs` noun required a distinct exported name per verb, so
`outputs.Run` became `outputs.RunList` (a pure rename) beside `outputs.RunGet`.
[Source: .specify/archive/002-outputs-get/plan.md -> "Structure Decision"]

The `inputs` noun (003) mirrors `outputs` in its own packages: `internal/cli/inputs` (`RunList`,
`RunGet`), `internal/hub/inputs.go` (`ListInputs`, `GetInput`) and `internal/render/inputs.go`.
Its `NotFoundError` change is additive: `outputs get`'s message text and exit code are unchanged.
[Source: .specify/archive/003-inputs-list-get/research.md -> "6. CLI dispatch: new `internal/cli/inputs` package, same shape as `internal/cli/outputs`"; .specify/archive/003-inputs-list-get/research.md -> "4. Not-found error type: generalize rather than duplicate"]

The `routes` noun (004) follows the same layout: `internal/cli/routes` (`RunList`, `RunGet`),
`internal/hub/routes.go` (`ListRoutes`, `GetRoute`) and `internal/render/routes.go`. Its
`NotFoundError` use passes `Resource: "route"`, so no shared error type changed.
[Source: .specify/archive/004-routes-list-get/plan.md -> "Structure Decision"]

The `groups` noun (005) follows the same layout: `internal/cli/groups` (`RunList`, `RunGet`), `internal/hub/groups.go` (`ListGroups`, `GetGroup`) and `internal/render/groups.go`. Its `NotFoundError` use passes `Resource: "group"`, so no shared error type changed.
[Source: .specify/archive/005-groups-list-get/plan.md -> "Structure Decision"]

The `play` command (006) has no `RunList`/`RunGet` pair: `internal/cli/play` exposes a single `Run`, and `cmd/sonora/main.go` routes `play` to it by an early special case ahead of the verb dispatch, because `<uri>` would otherwise be taken for the resource. `internal/hub/play.go` (`PlaybackRequest`, `PlaybackResponse`, `Playback`) and `internal/render/play.go` are new; `internal/hub/errors.go` gains four new exit classes and `APIError`. No other command's code changes beyond reusing `GetOutput` and `GetGroup`.
[Source: .specify/archive/006-play-command/plan.md -> "Structure Decision"; .specify/archive/006-play-command/research.md -> "2. play is verb-less"]

The verb-first refactor (007) is concentrated in a new `internal/cli/respath` package, `cmd/sonora/main.go`'s dispatcher and `internal/cli/play`; the four resource packages' `RunList`/`RunGet` keep their signatures and are called with the same argument shape the old noun-verb switch gave them, with the dispatcher resolving the resource path first. `list` is a pre-dispatch guard (an id after `list` is a usage error), not a second code path. `internal/hub/play.go`'s `ResolveTarget` became a per-type existence check, and `ClassAmbiguous`/`AmbiguousTargetError` were removed from `internal/hub/errors.go`.
[Source: .specify/archive/007-refactor-cli-commands/plan.md -> "Structure Decision"; .specify/archive/007-refactor-cli-commands/plan.md -> "Project Structure"; .specify/archive/007-refactor-cli-commands/research.md -> "2. Reuse existing `RunList`/`RunGet` functions unchanged; translate at the dispatch boundary"; .specify/archive/007-refactor-cli-commands/research.md -> "3. `list` is a pre-dispatch guard, not a second code path"]

The `route` command (008) follows `play`'s verb-less shape. It adds `internal/cli/route` (a single `Run`; `cmd/sonora/main.go` sends `route` to it by an early special case beside `play`), `CreateRouteRequest` and `CreateRoute` in `internal/hub/routes.go` (one `POST /api/v2/routes`, operationId `createRoute`), `ClassInputNotFound` (11) and `ClassTargetNotFound` (12) in `internal/hub/errors.go`, and `internal/render/route.go`. Its input check reuses `hub.GetInput`, and its target check reuses `play`'s per-type existence check (007), so neither `outputs`, `inputs`, `groups` nor `routes` changes behaviour.
[Source: .specify/archive/008-route-command/plan.md -> "Structure Decision"; .specify/archive/008-route-command/research.md -> "5. `route` is a verb-less top-level noun, dispatched exactly like `play`"]

## Configuration

- **Hub base URL**, highest precedence first: `--hub-url` flag; `MULTIROOM_URL` environment
  variable; `hubUrl` field of `~/.config/sonora/config.json`; built-in default
  `http://localhost:8080` (the `servers[0].url` in `api/openapi.json`). The config file is
  read lazily, only when the URL must be resolved, after flag parsing. A missing file falls
  through silently; malformed JSON or a non-string `hubUrl` is a usage error (exit `2`)
  naming the file. Writing the file is out of scope. `get`/`list` on every resource and `play` use the same precedence.
  [Source: .specify/archive/001-list-outputs/research.md -> "5. Hub location resolution"; .specify/archive/002-outputs-get/data-model.md -> "--hub-url"; .specify/archive/003-inputs-list-get/data-model.md -> "Config file"; .specify/archive/004-routes-list-get/data-model.md -> "Config file"; .specify/archive/005-groups-list-get/data-model.md -> "Config file"; .specify/archive/006-play-command/plan.md -> "Storage"; .specify/archive/007-refactor-cli-commands/plan.md -> "Storage"; .specify/archive/008-route-command/plan.md -> "Storage"]
- **Config file** `~/.config/sonora/config.json`: optional; field `hubUrl` (string).
  [Source: .specify/archive/001-list-outputs/data-model.md -> "Config file"]

## Routing & Navigation

- `sonora get outputs` / `sonora list outputs [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs?includeDisabled={true|false}`; single attempt, 5s timeout, no retries. Unknown flags or extra positionals are a usage error (exit `2`), as is an id in the path for `list`.
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Invocation"]
- `sonora get outputs/<output-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs/{output-id}` (path-escaped; operationId `getOutput`); single attempt, 5s timeout, no retries. The identifier is the path segment after the first `/` and must match `^[a-zA-Z0-9_-]{1,255}$` (a malformed path is a usage error, exit `2`); no `--include-disabled`, since the output is returned regardless of enabled state. Success output is a single unwrapped record (YAML by default, strict JSON with `--json`), every field always present.
  [Source: .specify/archive/002-outputs-get/contracts/cli-outputs-get.md -> "Invocation"]
- `sonora get inputs` / `sonora list inputs [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/inputs?includeDisabled={true|false}` (operationId `listInputs`); single attempt, 5s timeout, no retries. Same flags, usage errors and exit codes as `get outputs`. Success output is a top-level `inputs:` list (YAML) or `{"inputs": [...]}` (`--json`); `createdAt` is `null` for static inputs.
  [Source: .specify/archive/003-inputs-list-get/contracts/cli-inputs-list.md -> "Invocation"]
- `sonora get inputs/<input-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/inputs/{input-id}` (path-escaped; operationId `getInput`); single attempt, 5s timeout, no retries. The path identifier must match the id pattern (exit `2` otherwise); no `--include-disabled`, since the input is returned regardless of its enabled state. Success output is one unwrapped record, as for `get outputs/<id>`; a 404 exits `5` with an "input not found" message.
  [Source: .specify/archive/003-inputs-list-get/contracts/cli-inputs-get.md -> "Invocation"]
- `sonora get routes` / `sonora list routes [--status STATUS] [--input-id ID] [--target-id ID] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/routes`, sending `status`, `inputId` and `targetId` as query parameters only when set (operationId `listRoutes`); the hub applies them with AND logic and answers an invalid `status` with `400`, which exits `3`. Single attempt, 5s timeout, no retries. Success output is a top-level `routes:` list (YAML) or `{"routes": [...]}` (`--json`) with five fields per route; zero routes exits `0`.
  [Source: .specify/archive/004-routes-list-get/data-model.md -> "CLI invocation shape"]
- `sonora get routes/<route-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/routes/{route-id}` (path-escaped; operationId `getRoute`); single attempt, 5s timeout, no retries. The path identifier must match the id pattern (exit `2` otherwise); a 404 exits `5` with a "route not found" message. Success output is one unwrapped record with all ten fields; `startedAt` is `null` until playback starts.
  [Source: .specify/archive/004-routes-list-get/data-model.md -> "CLI invocation shape"]
- `sonora get groups` / `sonora list groups [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/groups?includeDisabled={true|false}` (operationId `listGroups`); single attempt, 5s timeout, no retries. Same flags, usage errors and exit codes as `get outputs`. A group with no member outputs shows `outputIds: []`, never omitted or `null`.
  [Source: .specify/archive/005-groups-list-get/data-model.md -> "CLI invocation shape"; .specify/archive/005-groups-list-get/plan.md -> "Summary"]
- `sonora get groups/<group-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/groups/{group-id}` (path-escaped; operationId `getGroup`); single attempt, 5s timeout, no retries. The path identifier must match the id pattern (exit `2` otherwise); no `--include-disabled`, since the group is returned regardless of its enabled state. A 404 exits `5` with a "group not found" message. Success output is one unwrapped record with the same five fields as `get groups`.
  [Source: .specify/archive/005-groups-list-get/data-model.md -> "CLI invocation shape"; .specify/archive/005-groups-list-get/plan.md -> "Summary"]
- `sonora play <uri> <outputs|groups>/<id> [--volume N] [--display-name NAME] [--json] [--verbose] [--hub-url URL]` → `POST {hub-url}/api/v2/play` (operationId `playback`), one request body carrying `uri`, `targetId`, `targetType` and, when given, `volume` and `displayName`; single attempt, 5s timeout, no retries, no polling. Exactly two positional arguments are required (exit `2` otherwise). The target is a resource path (aliases `out/` and `gr/` accepted) whose prefix fixes the target type; a path naming another resource, one without an id, or a malformed one is a usage error (exit `2`), as are the removed `--group`, `--output` and `--name` flags (the standard "flag provided but not defined" error). A per-type existence check runs first, reusing `hub.GetOutput` or `hub.GetGroup`; an identifier absent from that type is not found (exit `5`), even if it exists as the other type. An out-of-range `--volume` is a validation error (exit `6`) before any request. Success output is four fields — `inputId`, `routeId`, `status`, `message` — as YAML (default) or JSON (`--json`). Hub `400` exits `6`, `404` exits `5`, `422` exits `8`, `502` exits `9`, `503` exits `10`; any other non-2xx status or a malformed response exits `3`; unreachable or timed out exits `4`.
  [Source: .specify/archive/006-play-command/data-model.md -> "Exit Code Table (full, this feature's additions marked NEW)"; .specify/archive/006-play-command/research.md -> "4. Client-side vs. hub-side validation: volume range checked before any request"; .specify/archive/007-refactor-cli-commands/data-model.md -> "CLI invocation shape — `play` (changed fields only)"; .specify/archive/007-refactor-cli-commands/research.md -> "4. `play`'s target becomes a required second resource path; `--group`/`--output` are deleted from its flag set"]
- `sonora route inputs/<input-id> <outputs|groups>/<target-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/inputs/{input-id}` (operationId `getInput`; 404 exits `11`), then the per-type target check (`GET .../outputs/{id}` or `.../groups/{id}`; 404 exits `12`), then `POST {hub-url}/api/v2/routes` (operationId `createRoute`) carrying the input identifier, the target identifier and the target type. Single attempt per request, 5s timeout, no retries, no polling. Exactly two positional arguments are required (exit `2` otherwise); the input prefix must be `inputs` or `in` and the target prefix `outputs`/`out` or `groups`/`gr` (exit `2` otherwise). Success output is three fields — `routeId`, `status`, `message` — as YAML (default) or JSON (`--json`). Hub `400` exits `6`, `422` exits `8`; any other non-2xx status or a malformed response exits `3`; unreachable or timed out exits `4`. There are no `--volume` or `--display-name` flags.
  [Source: .specify/archive/008-route-command/contracts/cli-route.md -> "Invocation"; .specify/archive/008-route-command/data-model.md -> "Exit Code Table"]
- `sonora get <resource>[/<id>]` and `sonora list <resource>` dispatch: the argument is split on the first `/` into a resource name or alias (`inputs`/`in`, `outputs`/`out`, `groups`/`gr`, `routes`/`rt`) and an optional id matching `^[a-zA-Z0-9_-]{1,255}$`, then handed to that resource's `RunList` (no id) or `RunGet` (id). `list` is an exact synonym of `get` for the collection form and rejects an id. A missing resource argument, an unrecognized resource or alias, a second `/`, or an id outside the pattern is a usage error (exit `2`); only the missing-argument case lists the valid resources. The old `sonora <resource> list|get` forms get the same generic unknown-command usage error as any unrecognized first token, with no migration hint. `--json`, `--verbose` and `--hub-url` behave as before.
  [Source: .specify/archive/007-refactor-cli-commands/data-model.md -> "CLI invocation shape — `get` / `list`"; .specify/archive/007-refactor-cli-commands/research.md -> "5. Old noun-verb commands are removed from the dispatch switch entirely"]
- `sonora help`, `sonora -h`, `sonora --help` and bare `sonora` → usage text on stdout, exit `0`, no hub request; checked after `--version`/`-v`. An unknown verb, including the removed noun-verb forms such as `sonora outputs list`, still exits `2` with the message on stderr; a bare `sonora get` or `sonora list` exits `2` with a message listing the valid resources.
  [Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]
- **Output**: YAML by default (a top-level `outputs:` list; string scalars always double-quoted; `available` always printed; zero outputs prints `outputs: []` plus a plain-text note), strict JSON `{"outputs": [...]}` with `--json`. Success payload on stdout, failures on stderr; `--verbose` appends the raw underlying error.
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Success output"]
- **Exit codes**: `0` success (including zero outputs), `2` usage error, `3` hub-reported error (non-2xx other than a get's 404, or 2xx with a malformed body), `4` network/connectivity error (unreachable, DNS, timeout), `5` not found (hub answered `404` for the given output, input, route or group identifier; appended, nothing renumbered). `play` (006) appends `6` validation, `8` route creation failed, `9` source unreachable and `10` upstream service unavailable; `7` (ambiguous target) was removed by 007 and is unused. `route` (008) appends `11` input not found and `12` target not found; its hub `422` reuses `8`.
  [Source: .specify/archive/001-list-outputs/research.md -> "6. Exit code scheme"; .specify/archive/002-outputs-get/research.md -> "2. The 404 \"not found\" case needs its own exit-code class"; .specify/archive/004-routes-list-get/data-model.md -> "Exit code classes"; .specify/archive/005-groups-list-get/data-model.md -> "Exit code classes (unchanged from prior features)"; .specify/archive/006-play-command/data-model.md -> "Exit Code Table (full, this feature's additions marked NEW)"; .specify/archive/008-route-command/data-model.md -> "Exit Code Table"]
- **Errors**: `internal/hub` classifies stdlib errors into the failure classes (usage, hub, network, and for `get` not-found) with short actionable messages; raw Go errors appear only with `--verbose`. For `play` (006), `APIError` decodes the hub's error body for `400`, `422`, `502` and `503`; a body that does not decode falls back to the generic status error, so a malformed error body never crashes error handling.
  [Source: .specify/archive/001-list-outputs/research.md -> "7. Error message translation"; .specify/archive/006-play-command/data-model.md -> "APIError (client-side, internal/hub.APIError)"]

## Testing Strategy

- **Unit** (`tests/unit`): `internal/hub` request construction and decoding; `internal/render` YAML/JSON output including the zero-outputs and unavailable-output cases.
- **Contract** (`tests/contract`): an `httptest.Server` serving `GET /api/v2/outputs` shaped from `api/openapi.json` (including `includeDisabled` and the `OutputResponse` fields); asserts the request and rejection of a malformed body.
- **Integration** (`tests/integration`): full `sonora get outputs` invocation against the same mock server, asserting stdout format, exit codes and include-disabled behavior.
- **Get command**: a unit test for single-output YAML/JSON rendering (including the unavailable-output case); a contract test whose mock server serves `getOutput`'s `200 OutputResponse` and `404 ErrorResponse` shapes and asserts the request path and both decode paths; an integration test of `sonora get outputs/<id>` covering found, not-found, malformed-path usage error, `--json` and exit codes.
  [Source: .specify/archive/002-outputs-get/research.md -> "6. Testing strategy — same three layers, extended for get"]

- **Inputs** (003): unit `tests/unit/cli_inputs_test.go`, `cli_inputs_get_test.go` and `render_inputs_test.go` (list and single-record YAML/JSON, zero inputs, disabled included, null-vs-populated `createdAt`), with `hub_client_test.go` extended for `NotFoundError`'s `Resource`/`ID`; contract `tests/contract/inputs_list_test.go` and `inputs_get_test.go` (`httptest.Server`s shaped from `listInputs`/`getInput`); integration `tests/integration/inputs_list_test.go` and `inputs_get_test.go`.
  [Source: .specify/archive/003-inputs-list-get/research.md -> "9. Testing strategy — same three layers, extended for `inputs`"]

- **Routes** (004): unit `tests/unit/cli_routes_test.go`, `cli_routes_get_test.go` and `render_routes_test.go` (list and single-record YAML/JSON, zero routes, the nullable `startedAt`); contract `tests/contract/routes_list_test.go` and `routes_get_test.go` (`httptest.Server`s shaped from `listRoutes`/`getRoute`, including the filter query parameters and the `400` for an invalid status); integration `tests/integration/routes_list_test.go` and `routes_get_test.go`.
  [Source: .specify/archive/004-routes-list-get/research.md -> "7. Testing strategy — same three layers, extended for `routes`"]

- **Groups** (005): unit `tests/unit/cli_groups_test.go`, `cli_groups_get_test.go` and `render_groups_test.go` (list and single-record YAML/JSON, the same five fields in both views, zero groups, and a zero-member group rendered as `outputIds: []`); contract `tests/contract/groups_list_test.go` and `groups_get_test.go` (`httptest.Server`s shaped from `listGroups`/`getGroup`, including the always-sent `includeDisabled` parameter and the `404`); integration `tests/integration/groups_list_test.go` and `groups_get_test.go`.
  [Source: .specify/archive/005-groups-list-get/research.md -> "6. Testing strategy — same three layers, extended for groups"]

- **Play** (006): unit `tests/unit/cli_play_test.go` and `render_play_test.go` (flag and positional parsing, the target-path parsing, the removed `--group`/`--output`/`--name` usage errors, the volume range check, and the four-field YAML/JSON rendering); contract `tests/contract/play_test.go` (`httptest.Server` shaped from `playback`'s 200, 400, 404, 422, 502 and 503 bodies); integration `tests/integration/play_test.go` against a fake hub serving `/api/v2/play`, `/api/v2/outputs/{id}` and `/api/v2/groups/{id}`, covering outputs and groups targets and every exit code.
  [Source: .specify/archive/006-play-command/research.md -> "5. Testing strategy: a fake hub serving three endpoints per scenario"]

- **Route** (008): unit `tests/unit/cli_route_test.go` and `render_route_test.go` (flag and positional parsing, prefix usage errors, the not-found dispatch to exit `11`/`12`, and the YAML/JSON rendering of the three-field result); contract `tests/contract/route_test.go` (`httptest.Server` shaped from `createRoute`'s 201, 400, 404, 422 bodies); integration `tests/integration/route_test.go` against a fake hub serving `/api/v2/inputs/{id}`, `/api/v2/outputs/{id}`, `/api/v2/groups/{id}` and `/api/v2/routes`, covering single-output and group targets, colliding identifiers, a missing input, a missing target, missing arguments, invalid prefixes, `--json` and every exit class.
  [Source: .specify/archive/008-route-command/plan.md -> "Testing"; .specify/archive/008-route-command/research.md -> "6. Testing strategy: contract test for `createRoute` alone; integration test for the full chain"]

- **Verb-first refactor** (007): written test-first per Principle VI. Unit `tests/unit/respath_test.go` (alias table, the first-`/` split, id pattern) plus usage-line assertions in the existing `cli_*_test.go` files, with `hub_client_test.go` dropping its `ClassAmbiguous` cases; contract tests mostly unchanged, since they call `internal/hub` directly, apart from `play_resolve_test.go` losing its ambiguous-target case; integration tests updated to the new invocation shape, since they run the built binary.
  [Source: .specify/archive/007-refactor-cli-commands/plan.md -> "Project Structure"; .specify/archive/007-refactor-cli-commands/plan.md -> "Testing"]

- **Help** (tinyspec `help-command`): `cmd/sonora/main_test.go` calls `run()` directly. `help`, `-h`, `--help` and bare invocation print the command names to stdout with exit `0`; an unknown noun and a missing verb still exit `2` on stderr.
  [Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]

All written before implementation (Principle VI). `gofmt` and `go vet` are the baseline checks.
[Source: .specify/archive/001-list-outputs/research.md -> "8. Testing strategy"]
