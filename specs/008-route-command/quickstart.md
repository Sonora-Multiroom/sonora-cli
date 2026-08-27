# Quickstart: Validating `sonora route`

## Prerequisites

- Go 1.27.0 toolchain (same as prior features).
- A reachable Multiroom Audio Hub, or the fake `httptest`-based hub used by the integration
  tests (see [research.md §6](research.md#6-testing-strategy-contract-test-for-createroute-alone-integration-test-for-the-full-chain)).
- Build the CLI: `go build -o sonora ./cmd/sonora` from the repo root.
- An existing, enabled input already on the hub (e.g. one created by a prior
  `sonora play ... --display-name "Kitchen Radio"`, or a static input the hub already exposes).

## Scenario 1 — Route an existing input to a single output (User Story 1, P1)

```sh
./sonora route inputs/spotify-1 outputs/office-speaker
```

Expected: exit code 0; stdout (YAML) shows `routeId`, `status` (typically `STARTING`), and
`message`. Re-run with `--json` and confirm the same three fields parse as strict JSON (`jq .`
or equivalent). Follow up with `sonora get inputs/spotify-1` (existing command) and confirm the
input is unchanged — no new input was created.

## Scenario 2 — Route an existing input to a group (User Story 2, P2)

```sh
./sonora route inputs/spotify-1 groups/whole-house
```

Expected: same three-field success output. Follow up with `sonora get routes/<routeId>`
(existing command) and confirm `targetType` is `OUTPUT_GROUP`.

### Colliding identifiers

```sh
./sonora route inputs/spotify-1 groups/shared-id
./sonora route inputs/spotify-1 outputs/shared-id
```

Expected: each targets the type named by the prefix, regardless of a same-named resource
existing as the other type — no ambiguity error, no disambiguation flag needed.

## Edge cases to spot-check

| Invocation | Expected exit code |
|---|---|
| `sonora route` (no args) | 2 |
| `sonora route inputs/spotify-1` (missing target) | 2 |
| `sonora route outputs/x inputs/y` (arguments in the wrong order) | 2 |
| `sonora route routes/x outputs/y` (input path uses a non-`inputs` prefix) | 2 |
| `sonora route inputs/spotify-1 inputs/y` (target path uses a non-target prefix) | 2 |
| `sonora route inputs/nonexistent outputs/office-speaker` | 11 |
| `sonora route inputs/spotify-1 outputs/nonexistent` | 12 |
| `sonora route inputs/spotify-1 groups/nonexistent` | 12 |
| disabled input routed to a valid target (hub 422) | 8 |
| target already at capacity (hub 422) | 8 |
| hub unreachable | 4 |

Each row's exit code must be distinct per FR-010/SC-003 — verify with `echo $?` (or
`$LASTEXITCODE` on PowerShell) after each invocation.
