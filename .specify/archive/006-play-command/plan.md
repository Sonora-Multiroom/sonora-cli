# Implementation Plan: Instant Playback (`play`)

**Branch**: `006-play-command` | **Date**: 2026-08-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-play-command/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora play <uri> <target-id>` — a single `POST /api/v2/play` call (operationId
`playback`) that creates an ephemeral input and a route in one hub round trip. Unlike every
prior feature (`outputs`/`inputs`/`routes`/`groups` `list`/`get`, all read-only `GET`s), this
is the CLI's first mutating command and the first command whose target identifier is
polymorphic: by default the command resolves `<target-id>` against both `GET
/api/v2/outputs/{id}` and `GET /api/v2/groups/{id}` (reusing the existing `hub.GetOutput`/
`hub.GetGroup`) to decide whether it names a single output or a group, with `--group`/
`--output` as mutually exclusive escape hatches when both exist under the same identifier.
The command returns as soon as the hub responds — no polling for a terminal route state — and
introduces five new exit-code classes (validation, ambiguous target, route-creation-failed,
source-unreachable, service-unavailable) alongside the four already established, because
`POST /api/v2/play`'s documented 400/404/422/502/503 responses are semantically distinct
failure modes the constitution's exit-code requirement (Principle V) requires the CLI to tell
apart. See [research.md](research.md) for the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`;
no new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `strconv` (tests: `net/http/httptest`). No third-party module is introduced; this
feature adds a new `internal/cli/play` package and a new `internal/hub/play.go` /
`internal/render/play.go` pair. See
[research.md §1](research.md#1-target-resolution-reuse-getoutputgetgroup-sequentially-not-list-not-concurrent).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read (not
written) via the existing `config.ResolveHubURL` — unchanged from prior features.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against
`httptest.Server`s shaped from `playback`'s request/response/error bodies in
`api/openapi.json`), and integration (`tests/integration`, full CLI invocation, using a fake
hub server that also serves `/api/v2/outputs/{id}` and `/api/v2/groups/{id}` for target
resolution). Written before implementation per Principle VI — see
[research.md §5](research.md#5-testing-strategy-a-fake-hub-serving-three-endpoints-per-scenario).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a new top-level noun, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from prior features).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); a full playback round trip (up to two target-resolution
requests plus the `play` request) reported in under 5s under normal network conditions
(SC-003) — this is an accept/reject confirmation, not confirmation that audio is playing.

**Constraints**: 5s total request timeout per HTTP call, single attempt with no automatic
retry on network/timeout failure (FR-010), reusing `hub.NewClient()` unchanged; no unbounded
waits (Principle IV); exit codes MUST distinguish nine classes: success, usage error,
validation error, target not found, ambiguous target, route creation failed, source
unreachable, upstream service unavailable, and connectivity/network error (FR-011) — five new
classes beyond the existing four, see
[research.md §3](research.md#3-exit-code-scheme-five-new-classes-appended-nothing-renumbered).

**Scale/Scope**: One new command (`sonora play <uri> <target-id>`) under a new `play` noun
(verb-less, unlike prior two-verb nouns). New flags: `--group`/`--output` (mutually exclusive
boolean type-forcing switches), `--volume` (0-100 int), `--name` (display name string), plus
the existing `--json`/`--verbose`/`--hub-url` flag surface reused unchanged. See
[data-model.md](data-model.md) and [contracts/cli-play.md](contracts/cli-play.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `hub.NewClient()` is constructed inside the command handler, after `flag.Parse()` completes, identical in shape to prior commands — no disk/network work occurs before argument parsing; target resolution and the `play` call happen only after flags/positionals are validated. | PASS |
| II. API Contract Fidelity | `PlaybackRequest`/`PlaybackResponse` decode/encode 1:1 against `#/components/schemas/PlaybackRequest`/`PlaybackResponse`; `POST /api/v2/play` (operationId `playback`) and its documented 400/404/422/502/503 `ErrorResponse` bodies are all mapped explicitly — no undocumented fields/endpoints are invented, no status code is silently ignored. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature adds one new package (`internal/cli/play`) and one new file each to the existing stdlib-only `internal/hub`/`internal/render` packages. | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged for all three possible requests (two lookups + one play call); every non-2xx/network/decode failure is classified and translated to a plain message by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora play <uri> <target-id>` follows `<noun> [args]` (this noun has no verb, unlike prior two-verb nouns, because the API exposes exactly one operation); default output is YAML, `--json` switches to JSON; exit codes distinguish nine classes (research.md §3); flag names (`--json`/`--verbose`/`--hub-url`) are reused unchanged from prior commands, per Principle V's "no synonyms" rule; `--group`/`--output` deliberately mirror each other symmetrically per the spec's clarification. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §5) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`); `play` is called out by the constitution by name as a core flow requiring contract/integration coverage before merge. | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/006-play-command/
├── plan.md               # This file (/speckit-plan command output)
├── research.md           # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   └── cli-play.md
├── checklists/
│   └── requirements.md   # Pre-existing (speckit-checklist output)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: adds an early special case for `play`, checked before the existing len(args)<2 gate and noun/verb split (both of which assume a verb token and would otherwise swallow <uri>) — see research.md §2 — routing directly to play.Run(args[1:], ...)

internal/
├── hub/                                # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                       # UNCHANGED — hub.NewClient() reused as-is
│   ├── outputs.go                      # UNCHANGED — GetOutput reused for target resolution
│   ├── inputs.go                       # UNCHANGED
│   ├── routes.go                       # UNCHANGED — Route struct reused for the nested route in PlaybackResponse
│   ├── groups.go                       # UNCHANGED — GetGroup reused for target resolution
│   ├── play.go                         # NEW: PlaybackRequest, PlaybackResponse structs; ResolveTarget(ctx, client, baseURL, id, forceGroup, forceOutput) (targetType string, err error); Playback(ctx, client, baseURL, req) (*PlaybackResponse, error)
│   └── errors.go                       # MODIFIED: adds ClassValidation/ClassAmbiguous/ClassRouteFailed/ClassSourceUnreachable/ClassServiceUnavailable, APIError{StatusCode,Title,Detail} (decoded from ErrorResponse for 400/422/502/503), AmbiguousTargetError{ID}; ClassifyError extended to classify all of the above
├── config/
│   └── config.go                       # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   ├── outputs/                        # UNCHANGED
│   ├── inputs/                         # UNCHANGED
│   ├── routes/                         # UNCHANGED
│   ├── groups/                         # UNCHANGED
│   └── play/
│       └── play.go                     # NEW: `sonora play <uri> <target-id>` — flag definitions (--group/--output/--volume/--name/--json/--verbose/--hub-url), two-positional parsing, client-side volume range check, target resolution call, playback call, dispatch, exit code mapping
└── render/
    ├── outputs.go                      # UNCHANGED
    ├── inputs.go                       # UNCHANGED
    ├── routes.go                       # UNCHANGED
    ├── groups.go                       # UNCHANGED
    └── play.go                         # NEW: RenderPlaybackYAML(hub.PlaybackResponse), RenderPlaybackJSON(hub.PlaybackResponse) — both expose exactly inputId/routeId/status/message (FR-006), never the full nested Route

tests/
├── contract/
│   └── play_test.go                    # NEW: httptest.Server shaped from playback's 200/400/404/422/502/503 response bodies; request body + decode contract
├── integration/
│   └── play_test.go                    # NEW: full `sonora play` invocation against a fake hub serving /api/v2/play + /api/v2/outputs/{id} + /api/v2/groups/{id} — single-output target, group target, ambiguous target, --group/--output forcing, volume range validation, missing args, --json, every exit code class
└── unit/
    ├── cli_play_test.go                 # NEW: play.Run flag/positional parsing, mutually-exclusive --group/--output usage error, volume range usage/validation error, dispatch
    └── render_play_test.go              # NEW: YAML/JSON rendering of a PlaybackResponse, field set matches FR-006 exactly
```

**Structure Decision**: New `internal/cli/play` package (mirroring the existing per-noun
package structure, but with a single `Run` entry point instead of `RunList`/`RunGet` since the
API exposes one operation) plus one new file each in the existing `internal/hub` and
`internal/render` packages, plus a targeted extension of `internal/hub/errors.go` for the five
new failure classes `POST /api/v2/play` introduces. No new top-level directories, no changes to
existing `outputs`/`inputs`/`routes`/`groups` code beyond reusing their existing `GetOutput`/
`GetGroup` functions unchanged.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
