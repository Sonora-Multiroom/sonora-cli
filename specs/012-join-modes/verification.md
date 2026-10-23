# Verification: Join Modes in the CLI

## Baseline (2026-10-23, before any change)

- Branch `012-join-modes`; `api/openapi.json` reads `"version": "0.1.28"`.
- `go version`: go1.27.0 windows/amd64.
- `gofmt -l .`: exit 0, printed nothing.
- `go vet ./...`: exit 0.
- `go test -count=1 ./...`: exit 0. Packages with tests: `cmd/sonora`, `tests/contract`,
  `tests/integration`, `tests/unit`, all `ok` (4 packages); 1111 passing tests and subtests
  (`go test -count=1 -json ./...`, counting `"Action":"pass"` events that name a test).

## Implementation checks (2026-10-23)

Every new test was run and seen failing before its implementation. The contract and `joinmode`
tests failed to compile (missing `JoinMode*`, the request fields, `GetRoutesForOutput`,
`RouteOnOutput`, `internal/cli/joinmode`); the help, invalid-mode, dispatch, output-with-routes and
YAML tests failed on their assertions. `TestRenderRouteYAML_OlderHubOmitsBoth` and
`TestRoutesGet_OlderHubBodyOmitsFields` passed before the change, as intended: they guard the old
ten-line shape.

None of the existing `create`, `rename` or `enable|disable inputs/<id>` integration tests compares
the exact YAML (they check for fields with `strings.Contains`), so none needed the trailing
`defaultJoinMode` line. The new line is covered by `TestRenderInputYAML_*` and `TestInputsGet_*`.

The help texts put the flag's argument name in the text (`MODE is REPLACE, …`, `join MODE for …`)
because the `flag` package takes the argument name from a back-quoted word in the usage string;
`contracts/cli.md` was updated to match.

| Command | Result |
|---|---|
| `gofmt -l .` | exit 0, printed nothing |
| `go vet ./...` | exit 0 |
| `go test -count=1 ./...` | exit 0; `cmd/sonora`, `tests/contract`, `tests/integration`, `tests/unit` all `ok` |
| `go test -count=1 -json ./...`, passing tests and subtests | 1208 (baseline 1111, +97); 0 failures |
| `git diff main -- go.mod go.sum` | printed nothing: no new dependency |
| `git ls-files --eol` over the changed and new files | none shows `w/crlf` |

## Quickstart manual checks (2026-10-23)

- `sonora play --help`, `route --help`, `create inputs/x --help` and `get outputs/x --help`, from a
  built binary with `MULTIROOM_URL=http://127.0.0.1:1`: each exits 0, sends no request, and shows
  its flag text from `contracts/cli.md`.
- **Read-only, against `http://multiroom.lan:8080`** (`/api-docs` reports `0.1.28`). Only `GET`
  requests were sent; nothing was played, stopped or changed.
  - `get outputs/office --include-routes`: the record plus one route (`kissfmdeep`, `REPLACE`,
    `duckState: "FULL"`, `main: true`). With `--json` the `routes` array is identical, field for
    field, to `curl -s …/api/v2/outputs/office/routes`.
  - `get outputs/kitchen --include-routes`: `routes: []` (nothing playing there).
  - `get outputs/nosuch --include-routes`: `error: output not found: nosuch`, exit 5.
  - `get routes/route-1ea355f908ab`: the ten lines, then `joinMode: "REPLACE"` and
    `outputs: ["office"]`; the JSON is unchanged in shape.
  - `get inputs/kissfmdeep`: ends with `defaultJoinMode: null`; the JSON has
    `"defaultJoinMode":null`.

## Audible check on `office` (2026-10-23, with the user's permission)

State before: `office` volume 6, unmuted, enabled; master mute off; one route playing, the user's
`route-1ea355f908ab` (`kissfmdeep`, `REPLACE`, `FULL`), which was left alone throughout.

| Step | Result |
|---|---|
| `play https://online.kissfm.ua/KissFM outputs/office --join-mode duck_others --volume 6` | exit 0. The hub logged `joinMode=DUCK_OTHERS` and "lowering 1 route(s) under ducking route"; the stream then ended after 2 s ("completed unexpectedly (hardware failure or EOF)") and the hub stopped the route and removed its input. A source problem (same server as `kissfmdeep`), not the CLI |
| `route inputs/krainafm outputs/office --join-mode duck_others` | exit 0. `get outputs/office --include-routes`: `kissfmdeep` `duckState: "LOWERED"`, the new route `joinMode: "DUCK_OTHERS"`, `FULL`, `main: false`. `get routes/<id>` shows `joinMode: "DUCK_OTHERS"`, `outputs: ["office"]` |
| `delete routes/<that route>` | exit 0; `kissfmdeep` back to `FULL` |
| `play http://91.218.212.84:8000/radionv.mp3 outputs/office --join-mode mix --volume 6` | exit 0. Both routes `FULL`; the new one `joinMode: "MIX"`, `main: false` |
| `delete routes/<that route>` | exit 0; its ephemeral input was removed with it (404) |

State after: identical to before; `audio.output.routes{output=office}` reads 1 (the user's route).
