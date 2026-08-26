# Feature Specification: Instant Playback (`play`)

**Feature Branch**: `006-play-command`

**Created**: 2026-08-26

**Status**: Draft

**Input**: User description: "`play` should be next. The pattern so far has been list/get for main resources (outputs, inputs, routes). Looking at the OpenAPI spec, /api/v2/play is a dedicated \"Playback\" feature that creates an ephemeral input and route in one call — it's core functionality mentioned in the API description."

## Clarifications

### Session 2026-08-26

- Q: Should the command return as soon as the hub accepts the playback request (route
  status `STARTING`), or wait and poll until the route reaches a terminal state (`ACTIVE`
  or `FAILED`) before reporting success or failure? → A: Return immediately with whatever
  status the hub's response contains; no polling.
- Q: Should there be an explicit option to force single-output targeting (e.g.
  `--output`), symmetric to the `--group` option, for the case where an output and a group
  share an identifier? → A: Yes — add a symmetric `--output` option; the ambiguous case can
  be resolved toward either type.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Play a URI to a single output (Priority: P1)

A person operating the audio system wants to play an audio source (a stream URL, a
YouTube/SoundCloud link, etc.) directly to one speaker without first creating an input and
a route as two separate steps.

**Why this priority**: This is the core value of the feature — instant playback in one
command — and is called out in the API description as core functionality. Without it,
users must fall back to separate input-creation and route-creation steps that don't yet
exist as CLI commands.

**Independent Test**: Can be fully tested by running the play command with a valid URI and
the identifier of a single existing output, then verifying the command immediately reports
a created input identifier and the route's identifier and status exactly as returned by
the hub, without waiting for the route to reach a further state.

**Acceptance Scenarios**:

1. **Given** a reachable audio URI and the identifier of an existing, enabled output,
   **When** the user runs the play command targeting that output, **Then** the command
   succeeds and displays the created input identifier, the route identifier, the route
   status, and a human-readable confirmation message.
2. **Given** the same successful playback request, **When** the user requests
   machine-readable output, **Then** the result is valid, strictly parseable structured
   data containing the same fields as the human-readable view.

---

### User Story 2 - Play a URI to an output group (Priority: P2)

A person wants to play one audio source to a group of outputs at once (e.g. "whole
house") rather than a single speaker.

**Why this priority**: Groups are a first-class routing target already supported by
`groups list`/`groups get`; playback should work the same way against a group as against a
single output, but this is secondary to the single-output happy path.

**Independent Test**: Can be fully tested by running the play command with a valid URI and
the identifier of an existing output group, and verifying the same success output as User
Story 1, with the route's target reflecting the group.

**Acceptance Scenarios**:

1. **Given** a reachable audio URI and the identifier of an existing output group,
   **When** the user runs the play command targeting that group, **Then** the command
   succeeds and displays the created input identifier, the route identifier and status,
   and confirms the target is a group.
2. **Given** an output and a group that share the same identifier, **When** the user runs
   the play command with the explicit group option, **Then** the command targets the group
   rather than the output, regardless of which one auto-detection would otherwise have
   picked.
3. **Given** an output and a group that share the same identifier, **When** the user runs
   the play command with the explicit output option, **Then** the command targets the
   output rather than the group.

---

### User Story 3 - Set the starting volume (Priority: P3)

A person wants playback to start at a specific volume rather than whatever level the
target output or group was last left at.

**Why this priority**: A convenience on top of the core playback flow; playback is fully
usable without it since the target's existing volume applies by default.

**Independent Test**: Can be fully tested by running the play command with a volume option
set to a valid level and verifying the response confirms the requested volume was applied
before playback started.

**Acceptance Scenarios**:

1. **Given** a valid playback request with a volume option set to a value between 0 and
   100, **When** the user runs the play command, **Then** the target's volume is set to
   that value before playback begins.
2. **Given** a volume option set to a value outside 0-100, **When** the user runs the play
   command, **Then** the command fails with a clear usage/validation error before any
   request reaches the hub.

---

### User Story 4 - Name the playback session (Priority: P3)

A person wants the ephemeral input created for this playback to have a recognizable name
(e.g. "Kitchen Radio") instead of an auto-generated identifier, so it's easy to recognize
later in `inputs list`.

**Why this priority**: Cosmetic convenience; playback functions correctly with an
auto-generated name if this is omitted.

**Independent Test**: Can be fully tested by running the play command with a display name
option and verifying the created input's name matches what was supplied.

**Acceptance Scenarios**:

1. **Given** a valid playback request with a display name option, **When** the user runs
   the play command, **Then** the confirmation output reflects that name.

---

### Edge Cases

- What happens when the URI is malformed or of an unsupported scheme? The command must
  fail with a clear validation error rather than an ambiguous or generic failure.
- What happens when the target output or group identifier does not exist? The command
  must fail with a clear "not found" message distinct from other failure types.
- What happens when the hub accepts the request but cannot create the underlying route
  (e.g. target already at capacity)? The command must fail with a clear "route creation
  failed" message distinct from "not found" and from validation errors.
- What happens when the URI is well-formed but unreachable (e.g. dead stream link, DNS
  failure on the hub's side)? The command must fail with a clear error distinguishing "the
  source could not be reached" from a local network problem or a hub bug.
- What happens when the hub's upstream playback service is temporarily unavailable? The
  command must fail with a clear, distinguishable "service unavailable" error.
- What happens when neither `uri` nor `targetId` is supplied? The user must see a clear
  usage error explaining what's required, not a confusing failure.
- What happens when the target identifier matches neither an existing output nor an
  existing group? The command must fail with a clear "target not found" error.
- What happens when the target identifier matches both an existing output and an existing
  group, and the user did not use the explicit group or output option? The command must
  fail with a clear "ambiguous target" error rather than silently guessing.
- What happens when the user uses the explicit group (or output) option but the identifier
  does not exist as that type — even if it exists as the other type? The command must fail
  with a clear "target not found" error rather than silently falling back to the other
  type.
- What happens when the user supplies both the explicit group option and the explicit
  output option in the same invocation? The command must fail with a clear usage error
  rather than silently honoring one of them.
- What happens when the hub is unreachable or does not respond within a bounded time? The
  command must fail with a clear, actionable error rather than hanging indefinitely.
- What happens when the hub returns a response that doesn't match the expected shape? The
  command must fail clearly rather than displaying garbled or silently incomplete output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that starts instant playback of an audio URI
  to a target output or output group, creating the underlying ephemeral input and route in
  a single invocation.
- **FR-002**: The command MUST require a URI and a target identifier as input, and MUST
  fail with a clear usage error if either is omitted.
- **FR-002a**: The explicit group option and the explicit output option are mutually
  exclusive; the command MUST fail with a clear usage error if both are supplied in the
  same invocation.
- **FR-003**: By default, the command MUST determine on its own whether the target
  identifier refers to a single output or a group, by checking the identifier against
  both existing outputs and existing groups, without requiring the user to state which
  kind it is. If it matches exactly one of the two, the command MUST use that match's
  type. The command MUST also provide two explicit, mutually exclusive options — one to
  force the identifier to be treated as a group (e.g. `--group`), one to force it to be
  treated as a single output (e.g. `--output`) — for cases where an output and a group
  share the same identifier and auto-detection alone cannot tell which one the user
  means.
- **FR-003a**: When the target identifier matches both an existing output and an existing
  group and neither the explicit group option nor the explicit output option was used, the
  command MUST fail with a clear "ambiguous target" error naming both possibilities,
  rather than silently picking one.
- **FR-003b**: When an explicit group or output option is used, the command MUST validate
  the identifier only against that type; if the identifier does not exist as that type,
  the command MUST fail with a clear "target not found" error, even if the identifier
  exists as the other type.
- **FR-004**: The command MUST provide an optional way to set the starting volume (0-100)
  for the target, applied before playback begins, and MUST reject an out-of-range value
  with a clear usage/validation error before contacting the hub.
- **FR-005**: The command MUST provide an optional way to set a display name for the
  ephemeral input created for this playback session.
- **FR-006**: On success, the command MUST return as soon as the hub accepts the request,
  without waiting or polling for the route to reach any further state, and MUST display
  the created input's identifier, the resulting route's identifier and status exactly as
  returned by the hub, and a human-readable confirmation message — in both the default
  human-readable view and the machine-readable (--json) view.
- **FR-007**: The command's default output MUST be human-readable and legible without any
  additional flags or tools.
- **FR-008**: The command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-009**: The command MUST fail with a clear, actionable error, distinguishable from
  all other failure types, for each of: a validation error (bad URI or bad request
  fields), a "target not found" error, an "ambiguous target" error, a "route creation
  failed" error, a "source unreachable" error, and an "upstream service unavailable"
  error.
- **FR-010**: The command MUST fail with a clear, actionable error message when the hub is
  unreachable, and MUST fail with a clear timeout error when the hub does not respond
  within a bounded amount of time; the command MUST make a single attempt and MUST NOT
  automatically retry on network or timeout failure.
- **FR-011**: The command MUST exit with a status that is programmatically
  distinguishable across: success, a usage error, a validation error, "target not found",
  "ambiguous target", "route creation failed", "source unreachable", "upstream service
  unavailable", and a connectivity/network error.
- **FR-012**: The command MUST reject a playback result from the hub that does not match
  the expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Playback Request**: A one-shot instruction to play a given audio source URI to a given
  target (a single output or an output group), with an optional starting volume and an
  optional display name for the resulting input.
- **Playback Result**: The outcome of a playback request — the identifier of the
  ephemeral input that was created, the identifier and status of the route created to
  carry audio from that input to the target, and a confirmation message.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can start playback to a single output or a group in one command
  invocation, without separately creating an input or a route.
- **SC-002**: A user can start playback with a chosen starting volume and/or a chosen
  display name in the same single command invocation.
- **SC-003**: A user receives a response confirming the hub has accepted the playback
  request (or a clear failure) in under 5 seconds under normal network conditions; this
  confirms the request was accepted, not that audio has been confirmed playing.
- **SC-004**: 100% of failure cases return a message that lets the user tell apart: bad
  input, target not found, ambiguous target, route creation failure, unreachable source,
  unavailable service, and a connectivity/timeout problem.
- **SC-005**: 100% of machine-readable output produced by the command can be parsed by a
  standard structured-data parser without error.
- **SC-006**: A user can always tell, without ambiguity, whether playback started
  successfully or the command failed.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same
  mechanism established by the `outputs`/`inputs`/`routes`/`groups` features; this feature
  consumes that mechanism but does not redefine it.
- The hub API requires no authentication for this endpoint, consistent with the current
  API specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing, consistent with the
  outputs, inputs, routes, and groups commands.
- The default request timeout is 5 seconds, consistent with the outputs, inputs, routes,
  and groups commands.
- Stopping, pausing, or otherwise managing the route and input created by playback is out
  of scope for this feature; it covers only starting playback. Managing the resulting
  route/input is expected to be handled by future route/input management commands, not by
  this command.
- The set of supported URI schemes (e.g. YouTube, SoundCloud, direct stream URLs) is
  determined entirely by the hub; the CLI performs no client-side allow-listing of schemes
  and passes the URI through as given.
- Output identifiers and group identifiers may occasionally collide (the hub does not
  guarantee a shared namespace between the two), so auto-detection alone cannot always
  resolve target type; the explicit group and output options exist specifically to break
  that tie in either direction.
- The hub's playback response reflects only the route's initial state (typically
  `STARTING`); it is the caller's responsibility to poll `routes get` afterward if
  confirmation that audio has actually begun playing is needed — that polling is out of
  scope for this command.
