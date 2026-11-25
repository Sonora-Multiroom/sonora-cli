# Phase 0 Research: Instant Playback (`play`)

## 1. Target resolution: reuse `GetOutput`/`GetGroup` sequentially, not `list`, not concurrent

**Decision**: Resolve `<target-id>`'s type by calling the existing `hub.GetOutput(ctx, ...)`
and/or `hub.GetGroup(ctx, ...)` (by ID, not by listing and scanning), sequentially, and only
the ones actually needed for the requested mode:

- `--group` given: call `GetGroup` only. Not found → `NotFoundError{Resource: "group", ID}`.
- `--output` given: call `GetOutput` only. Not found → `NotFoundError{Resource: "output", ID}`.
- Neither given (default): call `GetOutput` then `GetGroup`. Both found →
  `AmbiguousTargetError{ID}`. Exactly one found → use that type. Neither found →
  `NotFoundError{Resource: "target", ID}`.

**Rationale**: `GetOutput`/`GetGroup` already exist, already return a typed `NotFoundError` on
404, and a lookup by ID is the minimal request shape for "does this identifier exist as this
type" — no new hub endpoint or query parameter is needed. Sequential (not concurrent) calls
were chosen over goroutine fan-out for the default path: the added concurrency-safety surface
(shared error handling, cancellation on first error, `errgroup`-style coordination) is not
justified by the win, since SC-003's 5-second budget is about *normal* network conditions,
under which two extra sub-100ms lookups plus the `play` call comfortably fit within budget.
Simplicity here also keeps this feature's error handling identical in shape to every prior
command (one call in flight at a time).

**Alternatives considered**:
- *Call `ListOutputs`/`ListGroups` and scan for a matching ID*: rejected — pulls the entire
  collection over the wire (and pays list-endpoint semantics like `includeDisabled`) to answer
  a single-ID existence question `GetOutput`/`GetGroup` already answer directly.
- *Concurrent lookups via goroutines*: rejected for this feature — no measured latency problem
  to justify it; revisit only if a future benchmark shows the sequential path missing SC-003
  under realistic conditions.
- *Let the hub's `POST /api/v2/play` decide the type itself (send targetId only, no
  targetType)*: not possible — `PlaybackRequest.targetType` is a required enum field
  (`SINGLE_OUTPUT`/`OUTPUT_GROUP`) per `api/openapi.json`; the hub does not offer
  auto-detection, so the CLI must resolve it client-side per the spec's clarification.

## 2. `play` is verb-less: `sonora play <uri> <target-id>`, not `sonora play start ...`

**Decision**: `play` is a top-level noun with no verb — `cmd/sonora/main.go` gains a special
case that hands its arguments to a single `play.Run(args, stdout, stderr) int`, mirroring
`RunList`/`RunGet`'s signature shape exactly but without the verb dispatch layer.

This cannot be added as an ordinary `case "play":` inside the existing noun `switch`, because
`run()` computes `noun, verb, rest := args[0], args[1], args[2:]` (behind a `len(args) < 2`
usage-error gate) *before* that switch ever runs, on the assumption that every command has a
verb token. For `sonora play <uri> <target-id>`, that assumption would bind `<uri>` (`args[1]`)
to the discarded `verb` variable and exclude it from `rest`, silently dropping it before
`play.Run` ever sees it. The fix is to special-case `play` *earlier*, immediately after the
existing `--version`/`-v` check and before the `len(args) < 2` gate: `if len(args) >= 1 &&
args[0] == "play" { return play.Run(args[1:], stdout, stderr) }`, so every token after `play`
(zero, one, or two positionals, plus flags) reaches `play.Run` unchanged, and `play.Run`'s own
parsing is solely responsible for validating that count.

**Rationale**: The OpenAPI spec exposes exactly one operation (`playback`) under this feature;
every prior noun (`outputs`/`inputs`/`routes`/`groups`) has a verb because each wraps two
distinct read operations (`list`/`get`). Inventing a verb (e.g. `play start`) for a single
operation would violate Principle V's UX-consistency intent in spirit (predictable structure)
more than it would serve it — the existing `<noun> <verb> [args]` pattern documented in the
constitution allows `[args]` to stand alone when there is exactly one verb, and `sonora play
<uri> <target-id>` matches how the feature is described in both the spec and the OpenAPI
`summary` ("Play audio from a URI").

**Alternatives considered**:
- *`sonora play start <uri> <target-id>`*: rejected — adds a verb with no second verb to
  contrast against, pure ceremony.
- *`sonora playback ...`*: rejected — the spec, spec title, and user-facing language all use
  "play"; matching operationId (`playback`) over user vocabulary was rejected since Principle V
  is about the CLI's UX, not the API's internal naming.

## 3. Exit code scheme: five new classes appended, nothing renumbered

**Decision**: Extend `hub.ErrorClass` with five new values, appended after the existing four
(`ClassUsage=2`, `ClassHub=3`, `ClassNetwork=4`, `ClassNotFound=5`):

| Class | Exit code | Trigger |
|---|---|---|
| `ClassValidation` | 6 | Hub `400` (bad URI / bad request field) **or** client-side out-of-range `--volume` |
| `ClassAmbiguous` | 7 | Target ID matches both an output and a group, no `--group`/`--output` given |
| `ClassRouteFailed` | 8 | Hub `422` (route creation failed, e.g. target at capacity) |
| `ClassSourceUnreachable` | 9 | Hub `502` (URI unreachable) |
| `ClassServiceUnavailable` | 10 | Hub `503` (upstream playback service unavailable) |

Existing classes/codes (`ClassNone=0`, `ClassUsage=2`, `ClassHub=3`, `ClassNetwork=4`,
`ClassNotFound=5`) are unchanged; `ClassNotFound` is reused as-is for "target not found"
(client-side, from `GetOutput`/`GetGroup`, or the hub's own `404` on `POST /api/v2/play` in the
unlikely race where the target vanishes between resolution and the play call). `ClassUsage` is
reused as-is for missing `<uri>`/`<target-id>`, both `--group` and `--output` given together,
and unparseable flags.

**Rationale**: FR-011 requires nine programmatically distinguishable outcomes (success + eight
failure classes); the four classes established by `001`-`005` only cover four of the eight
failure modes this feature introduces (`ClassUsage`, `ClassNotFound`, `ClassNetwork` are
reused; `ClassHub` remains the fallback for any *other* unexpected non-2xx status, keeping the
mapping total and forward-compatible with a hub response the CLI doesn't yet special-case).
Appending rather than renumbering preserves every existing script's exit-code contract for
`outputs`/`inputs`/`routes`/`groups`.

**Alternatives considered**:
- *Reuse `ClassHub` for all four hub-side failures (400/422/502/503) and rely on the message
  text alone to distinguish them*: rejected — FR-011 explicitly requires the *exit code* itself
  to be distinguishable, not just the message.
- *A single `ClassPlaybackFailed` covering 400/422/502/503*: rejected — collapses four
  spec-mandated distinct outcomes (FR-009) into one exit code, failing FR-011.

## 4. Client-side vs. hub-side validation: volume range checked before any request

**Decision**: `--volume` is parsed as an `int` flag and checked against `[0, 100]` in
`play.Run` itself, before `ResolveTarget` or `Playback` is ever called. An out-of-range value
returns `ClassValidation` (not `ClassUsage`) immediately, with no HTTP request made.

**Rationale**: FR-004 requires rejecting an out-of-range volume "before any request reaches
the hub" — this is a pure client-side short-circuit, matching the existing precedent of
`config.ResolveHubURL` errors being checked before any hub call. It is classified as
`ClassValidation` rather than `ClassUsage` because it is a *value* problem (a well-formed flag
whose content violates the API's own `minimum`/`maximum` constraint on `PlaybackRequest.volume`
in `api/openapi.json`), the same category as the hub's own 400 response for a bad request
field — `ClassUsage` is reserved for structural/invocation problems (missing positional args,
mutually exclusive flags used together, unparseable flags), consistent with how `routes
get`/`groups get` already use `ClassUsage` only for argument-count/parse problems, never value
problems.

**Alternatives considered**:
- *Let the hub reject an out-of-range volume via its own 400*: rejected — FR-004 is explicit
  that this must be caught before contacting the hub; it also saves a wasted round trip for a
  guaranteed-invalid request.

## 5. Testing strategy: a fake hub serving three endpoints per scenario

**Decision**: Contract tests (`tests/contract/play_test.go`) stand up an `httptest.Server`
that serves only `POST /api/v2/play`, shaped from `playback`'s documented 200/400/404/422/502/
503 bodies in `api/openapi.json` — verifying the CLI's request body and response/error
decoding in isolation, same pattern as every prior contract test. Integration tests
(`tests/integration/play_test.go`) stand up a fake hub serving all three endpoints the command
can call in one invocation (`/api/v2/play`, `/api/v2/outputs/{id}`, `/api/v2/groups/{id}`), so
target-resolution scenarios (single match, ambiguous match, no match, forced type mismatch) can
be exercised end-to-end through the real CLI binary path, matching the full-invocation style of
`tests/integration/routes_get_test.go` etc.

**Rationale**: This is the constitution's explicitly named core-flow requiring
contract/integration coverage before merge (Principle VI, Development Workflow); splitting
contract (single-endpoint, request/response shape) from integration (multi-endpoint,
end-to-end exit codes) mirrors the layering already established by every prior feature, just
extended to the three-endpoint interaction this feature is the first to require.

**Alternatives considered**:
- *Mock `hub.GetOutput`/`hub.GetGroup`/`hub.Playback` at the Go function level instead of over
  HTTP*: rejected — no prior feature introduces interface seams for mocking; `internal/hub`
  functions take `*http.Client` directly and are tested against real HTTP servers throughout
  the codebase, and introducing a mocking layer for this one feature would break that
  established, simpler pattern for no proven benefit.
