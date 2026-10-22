# Quickstart: Validate Join Modes

## Prerequisites

- Go 1.27. Run from the repository root in Git Bash; this machine has no `make`.

## Automated checks

```sh
gofmt -l . ; go vet ./... ; go test -count=1 ./... > "$TEMP/012-test.log" 2>&1; echo $?
```

Expected: `gofmt` prints nothing, the exit code is 0, and these tests pass. The names are indicative.

| Scenario | Where | Expected |
|---|---|---|
| `GetRoutesForOutput`: two routes, an empty list, `null`, 404, unknown duck state, a bad status | `tests/contract/outputs_routes_test.go` | order kept; empty slice; empty slice; `*NotFoundError` output; value kept and `Lowered()` false; `*DecodeError` |
| `CreateRoute`, `Playback`, `CreateInput` with the field nil and set | `tests/contract/route_test.go`, `play_test.go`, `inputs_create_test.go` | key absent when nil (body identical to before); present when set |
| `joinmode.Parse` | `tests/unit/joinmode_test.go` | `mix`, `Duck_Others`, `REPLACE` accepted, upper-cased; `""`, `fade`, `duck-others` rejected |
| `play`, `route`, `create inputs` with a valid mode | `tests/integration/play_test.go`, `route_test.go`, `inputs_create_test.go` | the body has the upper-cased mode; stdout and exit as without the flag |
| The same three with `fade` and `""` | the same files | exit 2, the usage line and valid modes on stderr, **zero** requests to the fake hub |
| `play --join-mode mix --volume 20 --display-name X` | `tests/integration/play_test.go` | all three fields in the body |
| `--help` on the three commands | `tests/unit/cli_play_test.go`, `cli_route_test.go`, `cli_inputs_create_test.go` | flag text as in [contracts/cli.md](contracts/cli.md) |
| `get outputs/<id> --include-routes`: two routes, an empty list, an unknown duck state | `tests/integration/outputs_get_test.go` | YAML and JSON as in [data-model.md](data-model.md); exit 0 |
| The same with the output 404 / routes 404 / routes 405 / routes 500 / malformed routes | the same | exit 5 / 3 / 3 / 3 / 3; stdout empty each time; the older-hub message only for 404 and 405 |
| `get outputs/<id>` without the flag | the same | exactly one request; output byte-identical |
| `--include-routes` on `get outputs`, `list outputs`, `get inputs/x`, `get groups/x`, `get routes/x`, `get master-mute` | `cmd/sonora/main_test.go` | exit 2, the applies-only message, no request |
| Route YAML with `joinMode` and two outputs, with an empty outputs list, and with both absent | `tests/unit/render_route_test.go` | lines after `paused`; `outputs: []`; the ten old lines only |
| Input YAML with a default and with null | `tests/unit/render_inputs_test.go` | `defaultJoinMode: "DUCK_OTHERS"` / `defaultJoinMode: null`, last |
| `list routes`, `list inputs` | the existing tests | unchanged, still passing |
| The godoc on new `hub` identifiers | `tests/unit/hub_godoc_test.go` | passes; no spec IDs in the comments |

## Manual checks

1. `go run ./cmd/sonora play --help`, `route --help`, `create inputs/x --help` and
   `get outputs/x --help` show the flag texts. `--help` sends no request.
2. **Read-only against production (allowed):** `go run ./cmd/sonora get outputs/<id> --include-routes`
   and `--json` against `http://multiroom.lan:8080`, which runs 0.1.28. The routes listed match
   `curl -s http://multiroom.lan:8080/api/v2/outputs/<id>/routes`. Do the same for
   `get routes/<id>` and `get inputs/<id>` with IDs taken from `list`.
3. **An audible check** (`play … --join-mode duck_others` over running music) is **not** part of this
   validation. If it is wanted, ask the user first and follow the shared rules for audible tests:
   record the state, keep the volume low, and stop only your own routes.
