# Quickstart: Validating `sonora play`

## Prerequisites

- Go 1.27.0 toolchain (same as prior features).
- A reachable Multiroom Audio Hub, or the fake `httptest`-based hub used by the integration
  tests (see [research.md §5](research.md#5-testing-strategy-a-fake-hub-serving-three-endpoints-per-scenario)).
- Build the CLI: `go build -o sonora ./cmd/sonora` from the repo root.

## Scenario 1 — Play to a single output (User Story 1, P1)

```sh
./sonora play "https://stream.example.com/live.mp3" office-speaker
```

Expected: exit code 0; stdout (YAML) shows `inputId`, `routeId`, `status` (typically
`STARTING`), and `message`. Re-run with `--json` and confirm the same four fields parse as
strict JSON (`jq .` or equivalent).

## Scenario 2 — Play to a group (User Story 2, P2)

```sh
./sonora play "https://stream.example.com/live.mp3" whole-house
```

Expected: same four-field success output, with `routeId` belonging to a route whose
`targetType` is `OUTPUT_GROUP` (verifiable via a future `routes get <routeId>`, out of scope
for this command's own output per FR-006).

### Disambiguating a shared identifier

```sh
./sonora play "https://stream.example.com/live.mp3" shared-id --group
./sonora play "https://stream.example.com/live.mp3" shared-id --output
```

Expected: each targets the stated type regardless of what auto-detection would have picked;
running without either flag when `shared-id` matches both an output and a group exits 7 with
an "ambiguous target" message naming both.

## Scenario 3 — Starting volume (User Story 3, P3)

```sh
./sonora play "https://stream.example.com/live.mp3" office-speaker --volume 40
./sonora play "https://stream.example.com/live.mp3" office-speaker --volume 150
```

Expected: the first succeeds (exit 0); the second fails immediately (exit 6, no hub request
made — confirm via `--verbose` or a request log on the fake hub showing zero calls).

## Scenario 4 — Named playback session (User Story 4, P3)

```sh
./sonora play "https://stream.example.com/live.mp3" office-speaker --name "Kitchen Radio"
```

Expected: exit 0; the confirmation `message` field reflects the supplied name (per the hub's
own message construction), and a subsequent `sonora inputs list` (existing command) shows an
input named "Kitchen Radio".

## Edge cases to spot-check

| Invocation | Expected exit code |
|---|---|
| `sonora play` (no args) | 2 |
| `sonora play <uri>` (missing target) | 2 |
| `sonora play <uri> <id> --group --output` | 2 |
| `sonora play "not a uri" <id>` | 6 (hub-detected) |
| `sonora play <uri> nonexistent-id` | 5 |
| `sonora play <uri> id --group` where `id` is only an output | 5 |
| target at capacity (hub 422) | 8 |
| dead stream URL (hub 502) | 9 |
| upstream service down (hub 503) | 10 |
| hub unreachable | 4 |

Each row's exit code must be distinct per SC-004/FR-011 — verify with `echo $?` (or
`$LASTEXITCODE` on PowerShell) after each invocation.
