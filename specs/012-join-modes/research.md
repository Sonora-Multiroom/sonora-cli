# Research: Join Modes in the CLI

The Technical Context left nothing open. These are the design decisions and two findings that change
the spec's wording.

## 1. What an older hub answers for the output's routes (finding)

- **Finding**: `DELETE /api/v2/outputs/{outputId}/routes` (`stopRoutesForOutput`) has been in the hub
  since 2026-09-11 (`sonora-multiroom` 231d8a6). `GET` on the same path arrived with join modes
  (0d5886a, 0.1.21). A hub that has the `DELETE` but not the `GET` matches the path but not the
  method. Its `V2ExceptionHandler` is scoped to `multiroom.rest.controllers.v2`, so it does not see a
  request that never reached a controller. Spring's default resolver then answers **405**, not 404.
  Only a hub from before 2026-09-11 answers 404.
- **Decision**: treat both 404 and 405 from the routes request, after the output request succeeded, as
  "this hub cannot list an output's routes". The spec's FR-010, scenario 5 of User Story 2 and its
  edge case now say "404 or 405".
- **Alternatives**: 404 only, as the spec first said. A 0.1.13–0.1.20 hub would then exit 3 with
  "hub reported an error (HTTP 405)", which does not tell the user what to do.

## 2. Which commands show `defaultJoinMode` in YAML (finding)

- **Finding**: `render.RenderInputYAML` prints the input record for `get inputs/<id>`, but also for
  `create inputs/<id>`, `rename inputs/<id>` and `enable|disable inputs/<id>`. Their JSON already
  includes `defaultJoinMode`, because it prints `hub.Input` unchanged.
- **Decision**: add `defaultJoinMode` to `RenderInputYAML`, so all four commands show it. One input
  record has one shape, as its JSON already does. `create inputs --default-join-mode` then shows the
  mode it just set. FR-013 now names every command that prints one input. SC-004's "byte for byte"
  excludes the fields FR-012 and FR-013 add on purpose.
- **Alternatives**: a separate renderer for `get` only. That gives two YAML shapes for the same record,
  and `create` would not echo the mode it sent.

## 3. Request fields in `hub`

- **Decision**: `CreateRouteRequest.JoinMode`, `PlaybackRequest.JoinMode` and
  `CreateInputRequest.DefaultJoinMode`, each `*string` with `omitempty`.
- **Rationale**: the contract types them `["string","null"]`. Optional request fields in `hub` are
  already pointers with `omitempty` (`PlaybackRequest.DisplayName`, `Volume`;
  `CreateInputRequest.Enabled`, `AutoRemove`). A nil pointer leaves the key out, so FR-004's "same
  request body as today" needs no extra code.
- **Alternatives**: `string` with `omitempty`. It is simpler, but breaks the idiom, and "" can then not
  mean anything else later.

## 4. Join-mode and duck-state values in `hub`

- **Decision**: export `JoinModeReplace`, `JoinModeMix`, `JoinModeDuckOthers`, `DuckStateFull` and
  `DuckStateLowered` as untyped string constants. The fields stay `string`.
- **Rationale**: these are protocol values (Principle VII). sonora-mcp needs them as well as the CLI.
  Untyped constants keep an unknown future value representable, as the contract asks for duck state.
  They follow `ReasonRouteChangingOutputs`.
- **Alternatives**: a named `JoinMode` type, which is a breaking change to `Route.JoinMode` and
  `Input.DefaultJoinMode`; or values only in the CLI, which sonora-mcp would duplicate.

## 5. The new operation

- **Decision**: `hub.GetRoutesForOutput(ctx, client, baseURL, outputID) ([]RouteOnOutput, error)`,
  named after the contract's operationId. The rest of its behaviour follows `ListRoutes` and
  `GetOutput`:
  - 404 is `*NotFoundError{Resource: "output"}`;
  - any other non-2xx status is `*StatusError`;
  - a malformed body is `*DecodeError`;
  - null or `[]` is an empty slice.

  Each entry is checked like a `Route`: it must have a route ID, an input and a target, and a known
  target type and status. Join mode and duck state are not checked, so that unknown values pass through.
- **Rationale**: one request per call, like every other `hub` function. The "older hub" reading of a
  404 is CLI policy, so it stays in `internal/cli/outputs`.
- **Shared check**: `validateRoute` becomes a thin wrapper over
  `validateRouteFields(routeID, inputID, targetID, targetType, status)`, which the new type uses too.

## 6. Telling whether a route is lowered

- **Decision**: `func (r RouteOnOutput) Lowered() bool`, true only when `DuckState == DuckStateLowered`.
- **Rationale**: the contract says "treat an unknown one as FULL". A method states that once, for
  every consumer.

## 7. Checking `--join-mode` / `--default-join-mode`

- **Decision**: a new package `internal/cli/joinmode` with:
  - `Values` (`REPLACE, MIX, DUCK_OTHERS`, built from the `hub` constants);
  - `Parse(s string) (string, bool)`, which upper-cases `s` and reports whether it is one of `Values`.

  Each command declares a `fs.String` flag and detects that it was given with `fs.Visit`, as `play`
  does for `--volume`. The check runs after the positional arguments are parsed and before the hub URL
  is resolved, so a bad value sends nothing. On failure the command prints the one-line usage, then
  `error: unknown join mode "<v>"; valid modes: REPLACE, MIX, DUCK_OTHERS`, and exits 2. An empty value
  fails the same way.
- **Rationale**: three commands in three packages need the same check. `routes.routeStatuses` is the
  model, but it is private to one package.
- **Alternatives**: put it in `clihelp`, which is about usage and parsing, not hub values; or copy it
  three times.

## 8. Rejecting `--include-routes` outside `get outputs/<id>`

- **Decision**: `dispatchGetList` in `cmd/sonora/main.go` scans its arguments up to a `--` for
  `--include-routes` or `--include-routes=…`, before any special case. It accepts the flag only when
  the verb is `get` and the path is `outputs/<id>`. Anywhere else it prints that verb's usage and
  `error: --include-routes applies only to 'sonora get outputs/<id>'`, then exits 2.
- **Rationale**: the dispatcher is the only place that knows the resource and the verb. Left to each
  command's `flag` set, the user would see "flag provided but not defined", which does not say where
  the flag belongs.

## 9. Output shape with `--include-routes`

- **YAML**: `RenderOutputYAML(o)` unchanged, then `routes: []`, or `routes:` with one entry per route.
  Each entry has `routeId`, `inputId`, `targetId`, `targetType`, `status`, `joinMode`, `duckState`,
  `main` and `paused`. Strings are written with `%q`, booleans with `%t`. There is no `# no routes`
  comment, which in the middle of a record would read as a separate document.
- **JSON**: one object, `struct{ hub.Output; Routes []hub.RouteOnOutput "routes" }`. The embedded
  fields are promoted, so the output's keys come first, then `routes`. A nil list becomes `[]`.
- **Rationale**: it matches "the output's record followed by a routes list". A script reads
  `.routes` from the same object it reads `.volume` from.

## 10. Route YAML for older hubs

- **Decision**: in `RenderRouteYAML`, after `paused`, write `joinMode: %q` when it is not empty. Write
  `outputs:` when the slice is not nil: `outputs: []` when it is empty, otherwise one `  - %q` line per
  output. `encoding/json` gives nil for an absent key and an empty slice for `[]`, so "older hub" and
  "playing on no members" stay apart.
- **JSON**: unchanged (`RenderRouteJSON` marshals `hub.Route`), as FR-014 asks.

## 11. Exit code for the older-hub case

- **Decision**: `exitcode.For(hub.ClassHub)` (3), with the message
  `error: this hub cannot list an output's routes; it needs hub 0.1.21 or later (hub URL: <url>)`.
  `--verbose` adds `detail: <the underlying error>`.
- **Rationale**: the README defines 3 as "hub returned an unexpected or malformed response". No new
  exit code is needed.
