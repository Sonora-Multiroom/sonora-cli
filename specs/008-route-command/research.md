# Phase 0 Research: Route an Existing Input (`route`)

## 1. Target resolution: reuse `hub.ResolveTarget` unchanged — no new resolution code

**Decision**: Call the existing `hub.ResolveTarget(ctx, client, baseURL, targetID, targetType)`
(`internal/hub/play.go`) as-is, with `targetType` already fixed by the target path's prefix
(`outputs`/`out` → `SINGLE_OUTPUT`, `groups`/`gr` → `OUTPUT_GROUP`) via `internal/cli/respath`.

**Rationale**: `hub.ResolveTarget`'s own doc comment already states its contract: "The CLI's
resource-path parsing (`internal/cli/respath`) determines targetType before this is ever
called ... so there is no auto-detect/ambiguity branch: exactly one endpoint is called." That
contract was established for `play` after `007-refactor-cli-commands` moved both commands to
path-style addressing, and is exactly what `route` needs — no `--group`/`--output` flags, no
auto-detect, one lookup (`GetOutput` or `GetGroup`) per invocation. Adding a second,
route-specific resolution function would duplicate logic already proven correct and tested.

**Alternatives considered**:
- *Write a route-local resolution helper*: rejected — `hub.ResolveTarget` already has the exact
  signature and behavior needed; duplicating it would violate Principle III (minimal, justified
  dependencies/code) for no behavioral gain.

## 2. Input existence: a new `GetInput`-based pre-check, ordered before target resolution

**Decision**: Before calling the hub's `createRoute` endpoint, call the existing
`hub.GetInput(ctx, client, baseURL, inputID)` (`internal/hub/inputs.go`, unchanged). If it
returns a `*hub.NotFoundError`, the command fails with "input not found" (FR-004) — checked
*before* `hub.ResolveTarget` is called, so when both the input and the target are missing, the
user sees "input not found" first, matching the order the spec's Edge Cases list them in.

**Rationale**: `play` never needed this check — its input doesn't exist yet, it's created by
the same call. `route`'s whole premise (FR-001) is that the input *already* exists, so its
existence has to be confirmed the same way the target's is: a direct by-ID lookup via the
existing `GetInput`, not a new endpoint or a `list`-and-scan.

**Alternatives considered**:
- *Skip the pre-check and rely solely on `createRoute`'s 404*: rejected — the hub's 404 for
  `createRoute` is documented as "Input, output, or group not found" with no field
  distinguishing which one, which cannot satisfy FR-004's requirement that "input not found" be
  distinguishable from "target not found." A typed pre-check with the existing per-resource
  `NotFoundError.Resource` values (`"input"` vs `"output"`/`"group"`) is the only way to get
  that distinction without inventing hub-side behavior the spec doesn't describe (Principle II).

## 3. Two new exit-code classes, applied locally in `route.Run` — `hub.ClassifyError` unchanged

**Decision**: Add two new `hub.ErrorClass` values in `internal/hub/errors.go`:

| Class | Exit code | Trigger (route command only) |
|---|---|---|
| `ClassInputNotFound` | 11 | `GetInput` (or `CreateRoute`'s 404 fallback) returns a `*NotFoundError` with `Resource == "input"` |
| `ClassTargetNotFound` | 12 | `ResolveTarget` (or `CreateRoute`'s 404 fallback) returns a `*NotFoundError` with `Resource != "input"` (i.e. `"output"` or `"group"`) |

These are resolved by a small check inside `route.Run`'s own error-reporting helper — `errors.As`
on `*hub.NotFoundError`, branching on `.Resource` — not inside `hub.ClassifyError`, which
continues to map every `*NotFoundError` to the existing generic `ClassNotFound` (exit 5)
exactly as it does today for `inputs get`, `outputs get`, `groups get`, `routes get`, and
`play`'s own target resolution.

**Rationale**: FR-010 requires "input not found" and "target not found" to be
*programmatically* distinguishable exit codes for this command specifically — a stronger
requirement than the message-level distinction (SC-003) that `ClassNotFound` already provides
for every other command. Splitting the classification inside `hub.ClassifyError` itself would
change the exit code of every existing `<resource> get <bad-id>` invocation from 5 to something
resource-dependent, silently breaking scripts written against those four already-shipped
commands — well outside this feature's scope. Keeping the split local to `route.Run` gets the
new requirement without touching any other command's contract, the same way `play` added its
five new classes without renumbering or reinterpreting the four it inherited
(`006-play-command/research.md` §3).

**Alternatives considered**:
- *Extend `hub.ClassifyError`'s `NotFoundError` branch to switch on `.Resource` globally*:
  rejected for the regression reason above.
- *Reuse the existing `ClassNotFound` (5) for both and rely on the message text alone*: rejected
  — fails FR-010's explicit "exit with a status that is programmatically distinguishable"
  wording, which play's own FR-011 precedent already establishes as meaning the exit code, not
  just the message (`006-play-command/research.md` §3).

## 4. The confirmation message is CLI-constructed, not decoded from the hub

**Decision**: `#/components/schemas/RouteResponse` (`createRoute`'s 201 body) has no `message`
field — unlike `PlaybackResponse`, which carries a hub-generated one. `internal/hub/routes.go`'s
new `CreateRoute` function decodes exactly `RouteResponse`'s fields into the existing `hub.Route`
struct (Principle II: no invented fields on the hub-facing type). The human-readable confirmation
string required by FR-005 is built in `internal/cli/route`, from the input/target paths and the
returned route, and handed to `internal/render/route.go` alongside the route for rendering.

**Rationale**: Inventing a `message` field on `hub.Route`/`hub.CreateRoute`'s return type to
satisfy FR-005 would misrepresent what the hub actually returns, violating Principle II ("The
CLI MUST NOT invent behavior not described by the spec"). Constructing the message in the CLI
layer — the same layer that already knows the human-typed `inputs/<id>` and
`outputs|groups/<id>` arguments — keeps the hub client a faithful mirror of the API and keeps
message construction next to the only place that has both the request and the response in hand.

**Alternatives considered**:
- *Add a synthetic `Message` field to `hub.Route`, populated by `CreateRoute` itself*: rejected —
  blurs the hub package's contract-fidelity guarantee; every other field on `hub.Route` is a
  direct mirror of `RouteResponse`, and a client-manufactured field breaks that invariant for
  callers that also use `hub.Route` for `routes get`/`routes list` (which never see a message).

## 5. `route` is a verb-less top-level noun, dispatched exactly like `play`

**Decision**: `cmd/sonora/main.go` gains one more early special case,
`if args[0] == "route" { return route.Run(args[1:], stdout, stderr) }`, placed alongside the
existing `play` check, before the `get`/`list` verb switch.

**Rationale**: The spec's own clarification settles this — "a new top-level command, `sonora
route <input-id> <target-id>`, singular and un-nested, mirroring the existing top-level `sonora
play <uri> <target-id>` command exactly." `route`, like `play`, wraps exactly one hub operation
(`createRoute`) with two required positional resource paths — the same shape `play` already
established a dispatch precedent for.

**Alternatives considered**:
- *`sonora routes create ...`*: rejected explicitly by the spec's clarification in favor of the
  singular, top-level, path-addressed form.

## 6. Testing strategy: contract test for `createRoute` alone; integration test for the full chain

**Decision**: `tests/contract/route_test.go` stands up an `httptest.Server` serving only
`POST /api/v2/routes`, shaped from `createRoute`'s documented 201/400/404/422 bodies —
verifying `hub.CreateRoute`'s request body and response/error decoding in isolation, matching
`tests/contract/play_test.go`'s single-endpoint pattern. `tests/integration/route_test.go`
stands up a fake hub serving `GET /api/v2/inputs/{id}`, `GET /api/v2/outputs/{id}`,
`GET /api/v2/groups/{id}`, and `POST /api/v2/routes` together, exercising the full CLI binary
path end-to-end for every user story and every exit-code class — the existing per-resource `get`
contract tests already cover `GetInput`/`GetOutput`/`GetGroup` in isolation, so they are reused,
not duplicated, at the contract layer.

**Rationale**: Same layering `play` established (`006-play-command/research.md` §5), extended to
this feature's three-endpoint interaction (one more than `play`'s two, since `route` also
pre-checks the input). `route` is named by the constitution as a core flow (routing) requiring
contract/integration coverage before merge (Principle VI).

**Alternatives considered**:
- *Duplicate `GetInput`/`GetOutput`/`GetGroup` contract coverage inside `route_test.go`*:
  rejected — those functions are unchanged and already covered by
  `inputs_get_test.go`/`outputs_get_test.go`/`groups_get_test.go`; re-testing them here would be
  redundant coverage of code this feature does not modify.
