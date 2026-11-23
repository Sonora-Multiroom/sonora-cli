# Verification: Align the CLI with Hub 0.1.25–0.1.28

## Baseline (2026-10-22, before any change)

- `api/openapi.json` reads `"version": "0.1.28"`.
- `go version`: go1.27.0 windows/amd64.
- `go test -count=1 ./...`: exit 0. Packages with tests: `cmd/sonora`, `tests/contract`,
  `tests/integration`, `tests/unit`, all `ok`; 1057 passing tests and subtests
  (`go test -count=1 -json ./...`, counting `"Action":"pass"` events that name a test).

## Implementation checks (2026-10-22)

Every new test was run and seen failing before its implementation (missing `DisplayName`,
`ReasonRouteChangingOutputs` and `IsTransientConflict` failed to compile; the help and transfer tests
failed on their assertions). Two guards were already passing before the change, as intended: the
`*_UsageErrorOmitsDescription` tests and the `ROUTE_CHANGING_OUTPUTS` decode in
`TestRouteRefusal_409_DecodesReasonAndOutputID`.

| Command | Result |
|---|---|
| `gofmt -l .` | printed nothing |
| `go vet ./...` | exit 0 |
| `go test -count=1 ./...` | exit 0; `cmd/sonora`, `tests/contract`, `tests/integration`, `tests/unit` all `ok` |
| `go test -count=1 -json ./...`, passing tests and subtests | 1095 (baseline 1057, +38) |
| `git ls-files --eol` | no changed file shows `w/crlf` |

## Quickstart manual checks (2026-10-22)

- `sonora <enable|disable> <inputs|outputs|groups>/x --help` for all six combinations, using a built
  binary with `MULTIROOM_URL=http://127.0.0.1:1`: each prints the `usage: …` line, a blank line, then
  its description line from `contracts/cli.md`, and exits 0 without a hub request.
- `sonora play --help` shows the new `--display-name` description.
- README review against `api/openapi.json` 0.1.28: the enable/disable rules match the descriptions
  of `setInputEnabled`, `setOutputEnabled` and `setGroupEnabled`; the `play` paragraph matches
  `PlaybackResponse.displayName`; the transfer sentence and exit-code row 14 match
  `ErrorResponse.reason` (`ROUTE_CHANGING_OUTPUTS`: transient, retry shortly). SC-001 met.
- No request was sent to `multiroom.lan`; the audible checks are out of scope.
