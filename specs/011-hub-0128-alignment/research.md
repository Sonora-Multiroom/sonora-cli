# Research: Align the CLI with Hub 0.1.25–0.1.28

All Technical Context items were known; these are the design decisions.

## 1. Where the enable/disable playback text goes in `--help`

- **Decision**: Append a blank line and one description line to the usage string passed to
  `clihelp.SetUsage`/`PrintUsage` in `internal/cli/{inputs,outputs,groups}/enabled.go`, chosen per verb.
  `PrintUsage` prints the usage string with `Fprintln`, so a multi-line string needs no helper change,
  and the first line stays the `usage: …` line that tests and users expect.
- **Rationale**: Smallest change; no other command has a description today, so no shared helper exists
  to reuse, and adding one for six strings is not worth it.
- **Alternatives**: A `description` parameter on `clihelp.SetUsage` (touches every caller); text only in
  the top-level `sonora --help` list (one line for three resources cannot say three different things).

## 2. Rendering `displayName` in `sonora play`

- **Decision**: In `internal/render/play.go`, `playbackPayload` gets
  ``DisplayName string `json:"displayName,omitempty"` `` placed after `InputID`; the YAML renderer writes
  `displayName: %q` after `inputId` only when non-empty.
- **Rationale**: Absent and empty are both "unknown" (spec edge case), so `omitempty` gives FR-006
  byte-identical output for older hubs. `%q` (`strconv.Quote`) keeps printable Unicode such as `—` and
  Cyrillic unescaped, as the existing `message` field already shows ("→"); `json.Marshal` does too.
- **Alternatives**: Always print the field (changes output for old hubs, violates FR-006); print the
  name inside `message` only (prose, not scriptable).

## 3. Recognising the transient refusal in `hub`

- **Decision**: Export `const ReasonRouteChangingOutputs = "ROUTE_CHANGING_OUTPUTS"` and
  `func IsTransientConflict(err error) bool` (true when `errors.As` finds an `*APIError` with status 409
  and that reason). `APIError.Reason` stays a `string`.
- **Rationale**: The CLI and sonora-mcp both need the same test; a predicate keeps the meaning of
  "transient" with the protocol (Principle VII) while leaving policy to callers. One constant, not a
  typed set, keeps the earlier `decode-route-refusals-409` decision: an unknown reason needs no client
  release.
- **Alternatives**: Constants for all seven reasons (not needed by any caller yet; can be added later
  without breaking anything); string comparison in the CLI (duplicated in sonora-mcp).

## 4. The transient message

- **Decision**: `ClassifyError` returns `ClassConflict` with the fixed message
  "the route is still adding an output; retry in a moment" for that reason, instead of the hub's
  `detail`. The hub's own text (today "Route is changing its outputs, try again: <routeId>") remains in
  the error and is shown with `--verbose`.
- **Rationale**: FR-010 and FR-012 need wording the CLI controls; the class (and so exit 14) is
  unchanged. Every command that reports via `ClassifyError` gets it, which covers FR-012.
- **Alternatives**: Append to the hub's detail (wording depends on the hub); a new error class and exit
  code (the spec keeps 14).

## 5. Retry schedule for `sonora transfer`

- **Decision**: In `internal/cli/route/transfer.go`, after `hub.TransferRoute` fails with
  `IsTransientConflict`, wait 100, 200, 400, then 800 ms before each retry (at most 4 retries, 1.5 s of
  waiting in total), then report the last error through the existing `reportError`. Any other error ends
  the loop at once and goes through the existing paths (404 → `reportNotFound`). The target pre-check is
  not repeated.
- **Rationale**: The hub measured a rejoin at 0.16 s, so the first or second retry normally succeeds
  (SC-003: < 1 s extra). The waits sum to 1.5 s, inside the spec's "about 2 seconds" and SC-004's 3 s
  even with a few fast round trips; each attempt keeps the client's 5 s timeout, so a hub that stops
  answering is a network error and is not retried. Doubling keeps load on a busy hub low.
- **Alternatives**: Fixed 250 ms interval (more requests, same window); retry inside
  `hub.TransferRoute` (breaks "one call, one request" for other consumers); a `--retry` flag (the spec
  rules out new flags).
- **Testability**: The exhaustion test waits the real 1.5 s; one such test is acceptable, so no clock
  seam is added.

## 6. Contract tests against 0.1.28 shapes

- **Decision**: Extend `tests/contract/play_test.go` with a 200 body carrying `displayName` and one
  without it; extend `tests/contract/route_refusal_test.go`'s table with `ROUTE_CHANGING_OUTPUTS` and no
  `outputId` on the transfer endpoint.
- **Rationale**: These are the response shapes the 0.1.28 contract adds; the existing tables already
  have the right structure.
