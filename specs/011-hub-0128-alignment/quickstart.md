# Quickstart: Validate Hub 0.1.25–0.1.28 Alignment

## Prerequisites

- Go 1.27; run from the repository root in Git Bash. There is no `make` on this machine.

## Automated checks

```sh
gofmt -l . ; go vet ./... ; go test -count=1 ./... > "$TEMP/011-test.log" 2>&1; echo $?
```

Expected: `gofmt` prints nothing, exit 0, and these new tests pass (names are indicative):

| Scenario | Where | Expected |
|---|---|---|
| Play response with `displayName` decodes | `tests/contract/play_test.go` | `DisplayName` set |
| Play response without it decodes | same | `DisplayName == ""`, no error |
| Play render with name (YAML, JSON), incl. `—` and Cyrillic | `tests/unit/render_play_test.go` | field second, unescaped |
| Play render without name | same | byte-identical to the 0.1.24-era output |
| Transfer refusal `ROUTE_CHANGING_OUTPUTS` decodes, no `outputId` | `tests/contract/route_refusal_test.go` | `Reason` set, `OutputID == ""` |
| `IsTransientConflict` / `ClassifyError` message | `tests/unit/hub_client_test.go` | true only for that 409; message fixed; class conflict |
| Transfer: 1 transient refusal then success | `tests/integration/route_test.go` | exit 0, 2 transfer requests, stderr empty |
| Transfer: always transient | same | exit 14, 5 requests, transient message, under 3 s |
| Transfer: `OUTPUT_DISABLED` | same | exit 14, 1 request, hub detail shown |
| Transfer: transient, then `OUTPUT_DISABLED`, 404 or a dropped connection | same | that error as today, 2 requests |
| Enable/disable help | `tests/unit/cli_*_enabled_test.go` | first line `usage: …`, then the description line from [contracts/cli.md](contracts/cli.md); usage errors unchanged |

## Manual checks (read-only against production)

The production hub is the user's real audio system: these steps only read help or use a fake hub.

1. `go run ./cmd/sonora disable inputs/x --help` (and the five other combinations): description lines
   match [contracts/cli.md](contracts/cli.md); no hub request is made for `--help`.
2. README review: enable/disable, `play` and exit-code-14 text agree with `api/openapi.json` 0.1.28
   descriptions of `setInputEnabled`, `setOutputEnabled`, `setGroupEnabled`, `playback` and
   `ErrorResponse.reason` (SC-001).
3. An audible check of `sonora play` naming or a real rejoin-then-transfer is **not** part of this
   validation; if wanted, ask the user first and follow the shared rules for audible tests.
