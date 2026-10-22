# Implementation Plan: Join Modes in the CLI

**Branch**: `012-join-modes` | **Date**: 2026-10-23 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/012-join-modes/spec.md`

## Summary

The CLI gains three join-mode flags, one way to list an output's routes, and three more fields in its
YAML. There are no new exit codes and no new dependencies.

1. **Choosing the join mode**:
   - `hub`: `CreateRouteRequest`, `PlaybackRequest` and `CreateInputRequest` gain an optional
     `*string` join-mode field (`omitempty`), and `JoinMode*` constants are added.
   - `play` and `route` get `--join-mode`, and `create inputs` gets `--default-join-mode`. A new
     `internal/cli/joinmode` package checks the value: it accepts any case, sends it upper-cased, and
     treats a bad value as exit 2 with no request sent.
2. **Showing an output's routes**:
   - `hub`: `GetRoutesForOutput`, the `RouteOnOutput` type with `Lowered()`, and `DuckState*`
     constants.
   - `get outputs/<id> --include-routes` sends a second request and prints the output's record with a
     `routes` list.
   - A 404 or 405 on the second request means the hub is older than 0.1.21. The CLI says so and exits
     3. Any failure leaves stdout empty.
   - `dispatchGetList` rejects the flag anywhere else, with exit 2.
3. **Showing join modes on routes and inputs**:
   - `RenderRouteYAML` adds `joinMode` and `outputs`, leaving out each one the hub did not send.
   - `RenderInputYAML` adds `defaultJoinMode`, or null when the input has none. That renderer is
     shared by `get`, `create`, `rename` and `enable|disable inputs/<id>`, so all four show the field.
   - The JSON and the list views are unchanged.

Research changed two points in the spec, and both are recorded in [research.md](research.md):

- An older hub answers 405, not 404, for the routes request.
- Every command that prints one input shows `defaultJoinMode`.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`)

**Primary Dependencies**: standard library only (`net/http`, `encoding/json`, `flag`, `strings`, `slices`)

**Storage**: N/A

**Testing**: `go test ./...`. The suites are:
- `tests/unit`: render, the join-mode check, help text, godoc.
- `tests/contract`: `hub` against `httptest` servers shaped by `api/openapi.json` 0.1.28.
- `tests/integration`: whole commands against a fake hub.
- `cmd/sonora/main_test.go`: dispatch.

**Target Platform**: Windows and Linux (amd64, and arm64 for the Pi)

**Project Type**: CLI plus a public Go client package (`hub/`)

**Performance Goals**: startup unchanged (under 50 ms to the first request); `--include-routes` adds
one round trip on the same keep-alive connection

**Constraints**:
- Without the new flags, the request bodies are byte-identical to today's.
- Hubs older than 0.1.21 keep working for every command that does not use the new flags.
- Each `hub` call sends one request.

**Scale/Scope**: about 15 source files, the README, and about 14 test files (2 new)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant startup | Nothing runs before argument parsing. The join-mode check is a slice lookup after parsing. The second request runs only with `--include-routes` | Pass |
| II. Contract fidelity | Request fields, `RouteOnOutput` and the new operation follow 0.1.28 field for field. Tests use 0.1.28 shapes. Unknown join modes and duck states pass through, as the contract asks | Pass |
| III. Minimal dependencies | Standard library only | Pass |
| IV. Resilient HTTP | One client and the same timeouts. Every failure is classified, and `--verbose` shows the underlying error. An older hub gets an actionable message, not a raw HTTP status | Pass |
| V. CLI UX consistency | One flag name (`--join-mode`) on both commands that start playback. `--default-join-mode` is the input's own property, named as in the contract. YAML is the default and `--json` keeps its shape. The exit-code classes are unchanged | Pass |
| VI. Test-first | Every task pair starts with a failing contract, unit or integration test ([quickstart.md](quickstart.md)) | Pass (enforced in tasks) |
| VII. Public `hub/` package | Only protocol pieces are added: fields, a type, an operation and constants. Reading a 404 or 405 as "older hub", and the flag checks, stay in `internal/`. The godoc reads on its own. The change is additive, so it ships as `v0.3.0` | Pass |

Post-design re-check: unchanged, all pass. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/012-join-modes/
├── plan.md              # This file
├── research.md          # Phase 0: decisions and two spec corrections
├── data-model.md        # Phase 1: hub types, flag values, rendered records
├── quickstart.md        # Phase 1: validation scenarios
├── contracts/
│   ├── cli.md           # Flags, help text, errors, requests, output
│   └── hub-package.md   # Additions to the public hub/ API
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
hub/
├── routes.go            # CreateRouteRequest.JoinMode; JoinMode* constants; validateRouteFields
├── play.go              # PlaybackRequest.JoinMode
├── inputs.go            # CreateInputRequest.DefaultJoinMode
└── outputs.go           # RouteOnOutput, Lowered, DuckState* constants, GetRoutesForOutput

internal/
├── cli/joinmode/joinmode.go   # NEW: Values, Parse
├── cli/play/play.go           # --join-mode
├── cli/route/route.go         # --join-mode
├── cli/inputs/create.go       # --default-join-mode
├── cli/outputs/get.go         # --include-routes, second request, older-hub message
└── render/
    ├── outputs.go       # RenderOutputWithRoutesYAML / JSON
    ├── routes.go        # RenderRouteYAML: joinMode, outputs
    └── inputs.go        # RenderInputYAML: defaultJoinMode

cmd/sonora/main.go       # dispatchGetList: reject --include-routes elsewhere

README.md                # join-mode paragraph; three flags documented

tests/
├── contract/outputs_routes_test.go   # NEW: GetRoutesForOutput
├── contract/{route,play,inputs_create}_test.go  # field sent only when set
├── unit/joinmode_test.go             # NEW
├── unit/render_{route,inputs,output_get}_test.go
├── unit/cli_{play,route,inputs_create,outputs_get}_test.go  # help text
├── integration/{play,route,inputs_create,outputs_get,routes_get,inputs_get}_test.go
└── cmd/sonora/main_test.go           # --include-routes elsewhere
```

**Structure Decision**: the existing single-module layout. One new internal package,
`internal/cli/joinmode`, is shared by three command packages. No other new packages.

## Complexity Tracking

None.
