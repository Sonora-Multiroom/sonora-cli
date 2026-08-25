# Feature Specification: List and Get Audio Routes

**Feature Branch**: `004-routes-list-get`

**Created**: 2026-08-25

**Status**: Draft

**Input**: User description: "next feature, similar to already implemented `outputs list`, `outputs get`, `inputs list` and `inputs get`, logically, would be `routes list` and `routes get`"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View currently active routes (Priority: P1)

A person operating the audio system wants to see which audio routes currently exist —
each connecting a specific input to a specific output or output group — along with each
route's status, so they know what audio is currently flowing where.

**Why this priority**: This is the baseline read operation for routes — without it, a
user has no way to discover route identifiers or see the current routing state before
issuing any other command involving a route (e.g. transferring or stopping one).

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more routes and verifying each route's identifier, source input identifier,
target identifier, target type, and status are shown.

**Acceptance Scenarios**:

1. **Given** the hub has two active routes, **When** the user runs the list command with
   no flags, **Then** both routes are displayed, each showing its identifier, source
   input identifier, target identifier, target type, and status.
2. **Given** the hub has zero routes, **When** the user runs the list command, **Then**
   the user sees a clear indication that no routes are currently active rather than an
   empty or ambiguous response.

---

### User Story 2 - Look up a specific route by identifier (Priority: P1)

A person who already knows a route's identifier (e.g. from `routes list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `routes list`; both are equally core to giving a user visibility
into the routing state of the system.

**Independent Test**: Can be fully tested by running the get command with the identifier
of a route known to exist on the hub and verifying the identifier, source input
identifier, target identifier, target type, status, and remaining attributes are all
shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has a route with a known identifier, **When** the user runs the get
   command with that identifier, **Then** the command displays that route's identifier,
   source input identifier, target identifier, target type, status, creation timestamp,
   playback-started timestamp (or an explicit indication that playback has not started),
   whether it can be transferred, whether it supports pause/resume, and whether it is
   currently paused.

---

### User Story 3 - Narrow the route list by status, input, or target (Priority: P2)

A person managing the audio system wants to see only the routes relevant to a specific
input, a specific output/group, or a specific lifecycle status (e.g. only `FAILED`
routes), so they can focus on the subset they care about without scanning the full list.

**Why this priority**: Secondary to the default view, but valuable for troubleshooting
and for scripts that only care about one input or target; it reuses the same read path
with additional optional filters.

**Independent Test**: Can be fully tested by running the list command with each filter
option in turn against a hub with routes in different states, sourced from different
inputs, and pointed at different targets, and verifying only the matching routes are
returned.

**Acceptance Scenarios**:

1. **Given** the hub has routes with different statuses, **When** the user runs the list
   command filtered to a specific status, **Then** only routes with that status are
   displayed.
2. **Given** the hub has routes from different source inputs, **When** the user runs the
   list command filtered to a specific input identifier, **Then** only routes originating
   from that input are displayed.
3. **Given** the hub has routes to different targets, **When** the user runs the list
   command filtered to a specific target identifier, **Then** only routes to that target
   are displayed.
4. **Given** the user supplies more than one filter at once, **When** the user runs the
   list command, **Then** only routes matching all supplied filters are displayed.

---

### User Story 4 - Handle a nonexistent route identifier (Priority: P2)

A person mistypes a route identifier, or refers to a route that has since stopped and
been removed, and needs to know clearly that it doesn't exist rather than receiving a
confusing error or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no route with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "route not found"
   message rather than an empty or ambiguous result.

---

### User Story 5 - Consume route data from a script (Priority: P3)

An automation or script wants to retrieve the current routes, or a single route's state,
in a strict, parseable format so it can act on the data programmatically (e.g. detect
`FAILED` routes and alert on them).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 1 and 2.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and
contains the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one route, **When** the user runs the list command
   requesting machine-readable output, **Then** the result is valid, strictly parseable
   structured data containing every route's full attribute set.
2. **Given** the hub has a route with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that route's full attribute set.

---

### Edge Cases

- What happens when no route identifier is provided to the get command? The user must
  see a clear usage error explaining that an identifier is required, not a confusing
  failure.
- What happens when an invalid status value is supplied to the list command's status
  filter? The user must see a clear usage or hub-reported error rather than a silently
  empty or incorrect result.
- What happens when the hub is unreachable (host down, wrong address, network partition)?
  The user must see a clear, actionable error rather than either command hanging
  indefinitely or crashing.
- What happens when the hub takes too long to respond? Both commands must fail with a
  clear timeout error rather than waiting indefinitely.
- What happens when the hub returns an error response other than "not found" (e.g. 5xx)?
  The user must see a clear message distinguishing "the hub reported a problem" from a
  plain network problem or (for get) a "not found" result.
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? Both commands must fail clearly rather than
  displaying garbled or silently incomplete output.
- What happens when a route's playback has not yet started? Its playback-started
  timestamp must be displayed clearly (e.g. `null`) rather than as an error or blank
  field, consistent with how `inputs get` handles a static input's missing creation
  timestamp (`createdAt: null`).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio routes.
- **FR-002**: By default, the list command MUST show all routes regardless of status,
  since routes (unlike inputs/outputs) have no enabled/disabled concept to filter on.
- **FR-003**: The list command MUST provide optional filters to narrow results by status,
  by source input identifier, and by target identifier; when more than one filter is
  supplied, only routes matching all supplied filters MUST be returned.
- **FR-004**: For each route returned by the list command, the CLI MUST display: its
  identifier, its source input identifier, its target identifier, its target type (single
  output or output group), and its current status — in both the default human-readable
  view and the machine-readable (--json) view.
- **FR-005**: The CLI MUST provide a command that retrieves a single audio route from the
  hub by its identifier.
- **FR-006**: The get command MUST require exactly one route identifier as input and MUST
  fail with a clear usage error if it is omitted.
- **FR-007**: For the route returned by the get command, the CLI MUST display: its
  identifier, its source input identifier, its target identifier, its target type, its
  current status, its creation timestamp, its playback-started timestamp (or an explicit
  indication that playback has not started), whether it is transferable, whether it
  supports pause/resume, and whether it is currently paused — in both the default
  human-readable view and the machine-readable (--json) view.
- **FR-008**: Both commands' default output MUST be human-readable and legible without
  any additional flags or tools.
- **FR-009**: Both commands MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-010**: The get command MUST fail with a clear, actionable "route not found"
  message, distinguishable from other failure types, when no route with the given
  identifier exists on the hub.
- **FR-011**: Both commands MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable.
- **FR-012**: Both commands MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; neither command MUST wait
  indefinitely. Each command MUST make a single attempt and MUST NOT automatically retry
  on network or timeout failure.
- **FR-013**: Both commands MUST fail with a clear, actionable error message when the hub
  returns an error response other than "not found" (for get) or a plain success (for
  list), distinguishing that case from a connectivity failure and, for get, from a "not
  found" result.
- **FR-014**: Both commands MUST exit with a status that is programmatically
  distinguishable between success, a usage error, "not found" (get only), a hub-reported
  error, and a connectivity/network error.
- **FR-015**: The list command MUST clearly communicate when there are zero routes to
  show, distinct from a failure.
- **FR-016**: Both commands MUST reject route data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Route**: A live connection routing audio from one input to one output or output
  group, managed by the hub. Attributes relevant to this feature: unique identifier,
  source input identifier, target identifier, target type (single output or output
  group), current lifecycle status (starting, active, stopping, stopped, or failed),
  creation timestamp, playback-started timestamp (absent until playback begins), whether
  it can be transferred to another target, whether it supports pause/resume, and whether
  it is currently paused.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently active routes in under 1 second
  under normal network conditions.
- **SC-002**: A user can retrieve a specific route's full state in under 1 second under
  normal network conditions.
- **SC-003**: When the requested route does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such cases.
- **SC-004**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases, for
  both commands.
- **SC-005**: 100% of machine-readable output produced by either command can be parsed by
  a standard structured-data parser without error.
- **SC-006**: A user can always tell, without ambiguity, whether "no routes are active"
  or "the command failed" occurred.
- **SC-007**: A user can identify a specific route's source input, target, and current
  status from a single command invocation, without needing to run any other command or
  cross-reference a list.
- **SC-008**: A user can narrow a route listing to a single status, input, or target
  without needing to filter the output themselves after the fact.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same
  mechanism established by the `outputs`/`inputs` features; this feature consumes that
  mechanism but does not redefine it.
- The hub API requires no authentication for these read-only endpoints, consistent with
  the current API specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing, consistent with the
  outputs and inputs commands.
- The default request timeout is 5 seconds, consistent with the outputs and inputs
  commands.
- The route identifier is an opaque, case-sensitive string as returned by the hub; the
  command performs no client-side normalization or fuzzy matching.
- Routes are unordered by the hub API; the list command may present them in the order
  received without imposing additional sorting guarantees.
- Creating, stopping, transferring, and pausing/resuming routes are out of scope for this
  feature; only read (list/get) operations are covered here, mirroring how
  `outputs list`/`outputs get` and `inputs list`/`inputs get` preceded any
  outputs/inputs mutation commands.
</content>
