# Main Project Specification

> **Revision**: 2026-11-23 — Seeded from specs/001-list-outputs (first archival): user stories 1–3, edge cases, FR-001–FR-013, Output entity, SC-001–SC-005, AS-001–AS-005.
> **Revision**: 2026-11-23 — Archived specs/002-outputs-get: added User Stories 4–6, an edge case, FR-014–FR-017, SC-006–SC-007 and AS-006; folded its equivalent requirements, edge cases, outcomes, assumptions and the Output entity into existing entries (FR-004–FR-013, SC-002–SC-005, AS-001–AS-004), widening FR-004, FR-010, FR-011 and SC-005 to cover the get command and the not-found case.
> **Revision**: 2026-11-23 — Archived .specify/archive/003-inputs-list-get: added User Stories 7–11, FR-018–FR-027, SC-008–SC-011, AS-007 and the Input entity. Its requirements (003 FR-009, FR-010, FR-012–FR-015 and FR-017) were folded into FR-006–FR-011 and FR-013; its outcomes (003 SC-004–SC-006) into SC-002–SC-004; its edge cases into the existing edge cases; and its assumptions into AS-001–AS-006, widening those entries to cover the inputs commands as well as the outputs ones. No existing ID was renumbered. The feature's own FR and SC numbers collided with main-memory IDs, so its new items continue above FR-017 and SC-007.
> **Revision**: 2026-11-23 — Archived .specify/archive/004-routes-list-get (moved from specs/ in the same run): added User Stories 12–16, FR-028–FR-036, SC-012–SC-016, AS-008 and the Route entity. Its FR-008, FR-009, FR-011–FR-014 and FR-016 were folded into FR-006–FR-011 and FR-013, its SC-004–SC-006 into SC-002–SC-004, and its assumptions into AS-001–AS-006, widening those entries to cover the routes commands as well; its invalid-status and playback-not-started edge cases are new. No existing ID was renumbered; the feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-027, SC-011 and AS-007.
> **Revision**: 2026-11-23 — Archived .specify/archive/005-groups-list-get (moved from specs/ in the same run): added User Stories 17–21, FR-037–FR-046, SC-017–SC-021, AS-009 and the Output Group entity. Its FR-009, FR-010, FR-012–FR-015 and FR-017 were folded into FR-006–FR-011 and FR-013, its SC-004–SC-006 into SC-002–SC-004, and its assumptions into AS-001–AS-006, widening those entries to cover the groups commands as well; its no-identifier, unreachable, timeout, hub-error and malformed-response edge cases were folded into the existing edge cases, and its zero-member-outputs edge case is new. No existing ID was renumbered; the feature's own FR and SC numbers collided with main-memory IDs, so its new items continue above FR-036 and SC-016.

## User Scenarios & Testing

### User Story 1 - View currently active outputs (Priority: P1)

A person operating the audio system wants to see which speakers/zones are currently
available and controllable, along with their volume and mute state, so they know what
they can act on next (e.g. before adjusting volume or routing audio somewhere).

**Why this priority**: This is the baseline read operation. Without it, a user has no way
to discover output identifiers or current state before issuing any other command — it is
the entry point to every other interaction with outputs.

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled outputs and verifying each enabled output's identifier, display name,
volume, mute state, and availability are shown, and disabled outputs are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled outputs and one disabled output, **When** the user
   runs the list command with no flags, **Then** only the two enabled outputs are
   displayed, each showing its identifier, display name, volume, mute state, and
   availability.
2. **Given** the hub has zero enabled outputs, **When** the user runs the list command,
   **Then** the user sees a clear indication that no outputs are available rather than an
   empty or ambiguous response.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 1]

---

### User Story 2 - Include disabled outputs (Priority: P2)

A person managing the audio system (e.g. diagnosing why a speaker isn't selectable) wants
to see outputs that exist but are currently disabled, so they can distinguish "not
available for routing" from "doesn't exist."

**Why this priority**: Secondary to the default view, but necessary for troubleshooting
and administration; it reuses the same read path with one additional option.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled output and verifying it
appears in the results, distinguishable from enabled outputs.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled output, **When** the user runs the
   list command with the option to include disabled outputs, **Then** both outputs are
   displayed and each output's enabled/disabled state is visible in the result.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 2]

---

### User Story 3 - Consume output list from a script (Priority: P3)

An automation or script wants to retrieve the current outputs in a strict, parseable
format so it can act on the data programmatically (e.g. feed it into another tool).

**Why this priority**: Enables automation use cases but is not required for a human
operator to get value from the command; it is an alternate presentation of the same data
as User Story 1.

**Independent Test**: Can be fully tested by running the list command with the
machine-readable output option and verifying the result is strictly parseable and contains
the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one enabled output, **When** the user runs the list
   command requesting machine-readable output, **Then** the result is valid, strictly
   parseable structured data containing every output's identifier, display name, volume,
   mute state, availability, and enabled state.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 3]

---

### User Story 4 - Look up a specific output by identifier (Priority: P1)

A person operating the audio system already knows (e.g. from `outputs list`) the
identifier of a specific output and wants to see its full current state — display name,
volume, mute state, availability, and enabled state — without scanning the full list.

**Why this priority**: This is the sole purpose of the command; without it the feature
delivers no value. It is the direct single-item counterpart to `outputs list`.

**Independent Test**: Can be fully tested by running the get command with the identifier
of an output known to exist on the hub and verifying the identifier, display name,
volume, mute state, availability, and enabled state are all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has an enabled output with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command displays that output's
   identifier, display name, volume, mute state, availability, and enabled state.
2. **Given** the hub has a disabled output with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command still displays the output
   (unlike the list command's default filtering) with its disabled state clearly shown.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 1]

---

### User Story 5 - Handle a nonexistent output identifier (Priority: P2)

A person mistypes an output identifier, or refers to one that has since been removed,
and needs to know clearly that it doesn't exist rather than receiving a confusing error
or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path — the command must be usable correctly before edge cases
matter.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no output with the given identifier exists on the hub, **When** the user
   runs the get command with that identifier, **Then** the user sees a clear "output not
   found" message rather than an empty or ambiguous result.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 2]

---

### User Story 6 - Consume a single output's state from a script (Priority: P3)

An automation or script wants to retrieve one output's current state in a strict,
parseable format so it can act on the data programmatically (e.g. check volume before
deciding whether to adjust it).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from the command; it is an
alternate presentation of the same data as User Story 4.

**Independent Test**: Can be fully tested by running the get command with the
machine-readable output option against a known output identifier and verifying the
result is strictly parseable and contains the same fields available in the
human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has an output with a known identifier, **When** the user runs the
   get command with that identifier requesting machine-readable output, **Then** the
   result is valid, strictly parseable structured data containing the output's
   identifier, display name, volume, mute state, availability, and enabled state.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 3]

---

### User Story 7 - View currently active inputs (Priority: P1)

A person operating the audio system wants to see which audio inputs (e.g. streaming
sources, ephemeral playback sessions) are currently registered and usable, so they know
what they can route to an output next.

**Why this priority**: This is the baseline read operation for inputs — without it, a
user has no way to discover input identifiers before issuing any other command involving
an input (e.g. creating a route).

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled inputs and verifying each enabled input's identifier, display name,
source URI, enabled state, and source type are shown, and disabled inputs are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled inputs and one disabled input, **When** the user runs
   the list command with no flags, **Then** only the two enabled inputs are displayed,
   each showing its identifier, display name, source URI, source type, and enabled state.
2. **Given** the hub has zero enabled inputs, **When** the user runs the list command,
   **Then** the user sees a clear indication that no inputs are available rather than an
   empty or ambiguous response.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 1]

---

### User Story 8 - Look up a specific input by identifier (Priority: P1)

A person who already knows an input's identifier (e.g. from `inputs list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `inputs list`; both are equally core to enabling later
route-management commands.

**Independent Test**: Can be fully tested by running the get command with the identifier
of an input known to exist on the hub and verifying the identifier, display name, source
URI, enabled state, source type, and other attributes are all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has an enabled input with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command displays that input's
   identifier, display name, source URI, enabled state, source type, and remaining
   attributes.
2. **Given** the hub has a disabled input with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command still displays the input
   (unlike the list command's default filtering) with its disabled state clearly shown.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 2]

---

### User Story 9 - Include disabled inputs in the list (Priority: P2)

A person managing the audio system wants to see inputs that exist but are currently
disabled, so they can distinguish "not currently usable" from "doesn't exist."

**Why this priority**: Secondary to the default view, but necessary for troubleshooting
and administration; it reuses the same read path with one additional option.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled input and verifying it
appears in the results, distinguishable from enabled inputs.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled input, **When** the user runs the
   list command with the option to include disabled inputs, **Then** both inputs are
   displayed and each input's enabled/disabled state is visible in the result.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 3]

---

### User Story 10 - Handle a nonexistent input identifier (Priority: P2)

A person mistypes an input identifier, or refers to an ephemeral input that has since
been removed, and needs to know clearly that it doesn't exist rather than receiving a
confusing error or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no input with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "input not found"
   message rather than an empty or ambiguous result.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 4]

---

### User Story 11 - Consume input data from a script (Priority: P3)

An automation or script wants to retrieve the current inputs, or a single input's state,
in a strict, parseable format so it can act on the data programmatically.

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 7 and 8.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and
contains the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one enabled input, **When** the user runs the list
   command requesting machine-readable output, **Then** the result is valid, strictly
   parseable structured data containing every input's identifier, display name, source
   URI, enabled state, source type, and remaining attributes.
2. **Given** the hub has an input with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that input's full attribute set.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 5]

---

### User Story 12 - View currently active routes (Priority: P1)

A person operating the audio system wants to see which audio routes currently exist —
each connecting a specific input to a specific output or output group — along with each
route's status, so they know what audio is currently flowing where.

**Why this priority**: This is the baseline read operation for routes — without it, a
user has no way to discover route identifiers or see the current routing state before
issuing any other command involving a route (e.g. transferring or stopping one).

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more routes and verifying each route's identifier, source input identifier, target
identifier, target type, and status are shown.

**Acceptance Scenarios**:

1. **Given** the hub has two active routes, **When** the user runs the list command with
   no flags, **Then** both routes are displayed, each showing its identifier, source
   input identifier, target identifier, target type, and status.
2. **Given** the hub has zero routes, **When** the user runs the list command, **Then**
   the user sees a clear indication that no routes are currently active rather than an
   empty or ambiguous response.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 1]

---

### User Story 13 - Look up a specific route by identifier (Priority: P1)

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

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 2]

---

### User Story 14 - Narrow the route list by status, input, or target (Priority: P2)

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

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 3]

---

### User Story 15 - Handle a nonexistent route identifier (Priority: P2)

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

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 4]

---

### User Story 16 - Consume route data from a script (Priority: P3)

An automation or script wants to retrieve the current routes, or a single route's state,
in a strict, parseable format so it can act on the data programmatically (e.g. detect
`FAILED` routes and alert on them).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 12 and 13.

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

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 5]

---

### User Story 17 - View currently enabled output groups (Priority: P1)

A person operating the audio system wants to see which output groups currently exist —
each a named collection of outputs that can be routed to as a single target — along with
each group's member outputs and mute state, so they know what groups are available for
routing.

**Why this priority**: This is the baseline read operation for groups — without it, a
user has no way to discover group identifiers or see which outputs belong to a group
before issuing any command that routes audio to a group.

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled groups and verifying each group's identifier, display name, member
output identifiers, and mute state are shown, and disabled groups are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled groups and one disabled group, **When** the user runs
   the list command with no flags, **Then** only the two enabled groups are displayed,
   each showing its identifier, display name, member output identifiers, and mute state.
2. **Given** the hub has zero enabled groups, **When** the user runs the list command,
   **Then** the user sees a clear indication that no groups are currently available rather
   than an empty or ambiguous response.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 1]

---

### User Story 18 - Include disabled groups (Priority: P2)

A person managing the audio system wants to see groups that exist but are currently
disabled, so they can distinguish "not available for new routes" from "does not exist"
when troubleshooting.

**Why this priority**: Secondary to the default view; needed for full visibility into the
system but not required for the common case of routing to an available group.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled group and verifying it
appears in the results, distinguishable from enabled groups.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled group, **When** the user runs the
   list command with the option to include disabled groups, **Then** both groups are
   displayed and each group's enabled/disabled state is visible in the result.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 2]

---

### User Story 19 - Look up a specific group by identifier (Priority: P1)

A person who already knows a group's identifier (e.g. from `groups list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `groups list`; both are equally core to giving a user
visibility into the groups available for routing.

**Independent Test**: Can be fully tested by running the get command with the identifier
of a group known to exist on the hub, whether enabled or disabled, and verifying the
identifier, display name, member output identifiers, mute state, and enabled state are
all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has a group with a known identifier, **When** the user runs the get
   command with that identifier, **Then** the command displays that group's identifier,
   display name, member output identifiers, mute state, and enabled state — regardless of
   whether the group is currently enabled or disabled.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 3]

---

### User Story 20 - Handle a nonexistent group identifier (Priority: P2)

A person mistypes a group identifier, or refers to a group that has since been removed,
and needs to know clearly that it doesn't exist rather than receiving a confusing error or
empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message and
a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no group with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "group not found"
   message rather than an empty or ambiguous result.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 4]

---

### User Story 21 - Consume group data from a script (Priority: P3)

An automation or script wants to retrieve the current groups, or a single group's state,
in a strict, parseable format so it can act on the data programmatically (e.g. pick a
group's member outputs before issuing a route command).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 17 and 19.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and contains
the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one group, **When** the user runs the list command
   requesting machine-readable output, **Then** the result is valid, strictly parseable
   structured data containing every group's full attribute set.
2. **Given** the hub has a group with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that group's full attribute set.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 5]

---

### Edge Cases

- What happens when the hub is unreachable (host down, wrong address, network partition)?
  The user must see a clear, actionable error rather than the command hanging indefinitely
  or crashing. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub is unreachable"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub is unreachable"]
- What happens when the hub takes too long to respond? The command must fail with a clear
  timeout error rather than waiting indefinitely. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub takes too long to respond?"]
- What happens when the hub returns an error response (e.g. 5xx)? The user must see a
  clear message distinguishing "the hub reported a problem" from a plain network problem
  (and, for the get command, from a "not found" result).
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns an error response"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub returns an error response other than"]
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? The command must fail clearly rather than displaying
  garbled or silently incomplete output.
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"]
- What happens when an output has a physically disconnected/unavailable device but is
  still enabled? Its availability state must be visibly distinguishable from an available
  output. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when an output has a physically disconnected/unavailable device"; .specify/archive/002-outputs-get/spec.md -> "What happens when the requested output exists but is enabled with unavailable hardware?"]
- What happens when no identifier is provided to the outputs, inputs, routes or groups get command? The user must see a
  clear usage error explaining that an identifier is required, not a confusing failure.
  [Source: .specify/archive/002-outputs-get/spec.md -> "What happens when no identifier is provided?"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when no input identifier is provided to the get command?"; .specify/archive/004-routes-list-get/spec.md -> "What happens when no route identifier is provided to the get command?"; .specify/archive/005-groups-list-get/spec.md -> "What happens when no group identifier is provided to the get command?"]
- What happens when a static (YAML-configured) input is returned? It has no creation
  timestamp; this must be displayed clearly (e.g. "n/a") rather than as an error or blank
  field. [Source: .specify/archive/003-inputs-list-get/spec.md -> "What happens when a static (YAML-configured) input is returned?"]
- What happens when an invalid status value is supplied to the routes list command's status filter? The user must see a clear usage or hub-reported error rather than a silently empty or incorrect result.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "What happens when an invalid status value is supplied to the list command's status filter?"]
- What happens when a route's playback has not yet started? Its playback-started timestamp must be displayed clearly (e.g. `null`) rather than as an error or blank field, consistent with how `inputs get` shows a static input's missing creation timestamp.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "What happens when a route's playback has not yet started?"]
- What happens when a group has zero member outputs? Its member output list must be displayed clearly as empty (`outputIds: []`) rather than as an error or omitted field.
  [Source: .specify/archive/005-groups-list-get/spec.md -> "What happens when a group has zero member outputs?"]

## Requirements

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio outputs. [Source: .specify/archive/001-list-outputs/spec.md -> FR-001]
- **FR-002**: By default, the command MUST show only enabled outputs. [Source: .specify/archive/001-list-outputs/spec.md -> FR-002]
- **FR-003**: The command MUST provide an option for the user to additionally include
  disabled outputs in the results. [Source: .specify/archive/001-list-outputs/spec.md -> FR-003]
- **FR-004**: For each output returned, the command MUST display: its identifier, its
  human-readable display name, its current volume level, its current mute state, whether
  it is enabled, and whether its underlying hardware is currently connected
  (availability) — in both the default human-readable view and the machine-readable
  (--json) view. The same applies to the single output returned by the get command.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-004; .specify/archive/002-outputs-get/spec.md -> FR-004]
- **FR-005**: An output that is enabled but whose hardware is currently unavailable MUST
  be visibly distinguishable from an available output in the default human-readable view.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-005; .specify/archive/002-outputs-get/spec.md -> FR-005]
- **FR-006**: Each outputs, inputs, routes and groups command's default output MUST be human-readable and legible without any
  additional flags or tools. [Source: .specify/archive/001-list-outputs/spec.md -> FR-006; .specify/archive/002-outputs-get/spec.md -> FR-006; .specify/archive/003-inputs-list-get/spec.md -> FR-009; .specify/archive/004-routes-list-get/spec.md -> FR-008; .specify/archive/005-groups-list-get/spec.md -> FR-009]
- **FR-007**: Each outputs, inputs, routes and groups command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-007; .specify/archive/002-outputs-get/spec.md -> FR-007; .specify/archive/003-inputs-list-get/spec.md -> FR-010; .specify/archive/004-routes-list-get/spec.md -> FR-009; .specify/archive/005-groups-list-get/spec.md -> FR-010]
- **FR-008**: Each outputs, inputs, routes and groups command MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable. [Source: .specify/archive/001-list-outputs/spec.md -> FR-008; .specify/archive/002-outputs-get/spec.md -> FR-009; .specify/archive/003-inputs-list-get/spec.md -> FR-012; .specify/archive/004-routes-list-get/spec.md -> FR-011; .specify/archive/005-groups-list-get/spec.md -> FR-012]
- **FR-009**: Each outputs, inputs, routes and groups command MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; it MUST NOT wait indefinitely. Each
  command MUST make a single attempt and MUST NOT automatically retry on network or
  timeout failure. [Source: .specify/archive/001-list-outputs/spec.md -> FR-009; .specify/archive/002-outputs-get/spec.md -> FR-010; .specify/archive/003-inputs-list-get/spec.md -> FR-013; .specify/archive/004-routes-list-get/spec.md -> FR-012; .specify/archive/005-groups-list-get/spec.md -> FR-013]
- **FR-010**: Each outputs, inputs, routes and groups command MUST fail with a clear, actionable error message when the hub
  returns an error response, distinguishing that case from a connectivity failure and,
  for the get commands, from a "not found" result.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-010; .specify/archive/002-outputs-get/spec.md -> FR-011; .specify/archive/003-inputs-list-get/spec.md -> FR-014; .specify/archive/004-routes-list-get/spec.md -> FR-013; .specify/archive/005-groups-list-get/spec.md -> FR-014]
- **FR-011**: Each outputs, inputs, routes and groups command MUST exit with a status that is programmatically distinguishable
  between success, a usage error, a hub-reported error, and a connectivity/network error
  and, for the get commands, "not found".
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-011; .specify/archive/002-outputs-get/spec.md -> FR-012; .specify/archive/003-inputs-list-get/spec.md -> FR-015; .specify/archive/004-routes-list-get/spec.md -> FR-014; .specify/archive/005-groups-list-get/spec.md -> FR-015]
- **FR-012**: The command MUST clearly communicate when there are zero outputs to show,
  distinct from a failure. [Source: .specify/archive/001-list-outputs/spec.md -> FR-012]
- **FR-013**: Each outputs, inputs and routes command MUST reject any output, input, route or group data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results. [Source: .specify/archive/001-list-outputs/spec.md -> FR-013; .specify/archive/002-outputs-get/spec.md -> FR-013; .specify/archive/003-inputs-list-get/spec.md -> FR-017; .specify/archive/004-routes-list-get/spec.md -> FR-016; .specify/archive/005-groups-list-get/spec.md -> FR-017]
- **FR-014**: The CLI MUST provide a command that retrieves a single audio output from
  the hub by its identifier. [Source: .specify/archive/002-outputs-get/spec.md -> FR-001]
- **FR-015**: The get command MUST require exactly one output identifier as input and MUST
  fail with a clear usage error if it is omitted. [Source: .specify/archive/002-outputs-get/spec.md -> FR-002]
- **FR-016**: The get command MUST return the requested output regardless of its
  enabled/disabled state (unlike the list command's default filtering).
  [Source: .specify/archive/002-outputs-get/spec.md -> FR-003]
- **FR-017**: The get command MUST fail with a clear, actionable "output not found"
  message, distinguishable from other failure types, when no output with the given
  identifier exists on the hub. [Source: .specify/archive/002-outputs-get/spec.md -> FR-008]
- **FR-018**: The CLI MUST provide a command that lists the hub's audio inputs. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-001]
- **FR-019**: By default, the inputs list command MUST show only enabled inputs. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-002]
- **FR-020**: The inputs list command MUST provide an option for the user to additionally include disabled inputs in the results. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-003]
- **FR-021**: For each input returned by the inputs list command, the CLI MUST display: its identifier, its human-readable display name, its source URI, whether it is enabled, its source type (static or ephemeral), whether it auto-removes when its route stops, whether it supports pause/resume, and its creation timestamp (or an explicit indication that none exists) — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-004]
- **FR-022**: The CLI MUST provide a command that retrieves a single audio input from the hub by its identifier. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-005]
- **FR-023**: The inputs get command MUST require exactly one input identifier as input and MUST fail with a clear usage error if it is omitted. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-006]
- **FR-024**: The inputs get command MUST return the requested input regardless of its enabled/disabled state (unlike the inputs list command's default filtering). [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-007]
- **FR-025**: For the input returned by the inputs get command, the CLI MUST display the same set of attributes specified in FR-021, in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-008]
- **FR-026**: The inputs get command MUST fail with a clear, actionable "input not found" message, distinguishable from other failure types, when no input with the given identifier exists on the hub. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-011]
- **FR-027**: The inputs list command MUST clearly communicate when there are zero inputs to show, distinct from a failure. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-016]
- **FR-028**: The CLI MUST provide a command that lists the hub's audio routes. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-001]
- **FR-029**: By default, the routes list command MUST show all routes regardless of status, since routes (unlike inputs and outputs) have no enabled/disabled concept to filter on. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-002]
- **FR-030**: The routes list command MUST provide optional filters to narrow results by status, by source input identifier, and by target identifier; when more than one filter is supplied, only routes matching all supplied filters MUST be returned. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-003]
- **FR-031**: For each route returned by the routes list command, the CLI MUST display: its identifier, its source input identifier, its target identifier, its target type (single output or output group), and its current status — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-004]
- **FR-032**: The CLI MUST provide a command that retrieves a single audio route from the hub by its identifier. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-005]
- **FR-033**: The routes get command MUST require exactly one route identifier as input and MUST fail with a clear usage error if it is omitted. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-006]
- **FR-034**: For the route returned by the routes get command, the CLI MUST display: its identifier, its source input identifier, its target identifier, its target type, its current status, its creation timestamp, its playback-started timestamp (or an explicit indication that playback has not started), whether it is transferable, whether it supports pause/resume, and whether it is currently paused — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-007]
- **FR-035**: The routes get command MUST fail with a clear, actionable "route not found" message, distinguishable from other failure types, when no route with the given identifier exists on the hub. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-010]
- **FR-036**: The routes list command MUST clearly communicate when there are zero routes to show, distinct from a failure. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-015]
- **FR-037**: The CLI MUST provide a command that lists the hub's audio output groups. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-001]
- **FR-038**: By default, the groups list command MUST show only enabled groups. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-002]
- **FR-039**: The groups list command MUST provide an option for the user to additionally include disabled groups in the results. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-003]
- **FR-040**: For each group returned by the groups list command, the CLI MUST display: its identifier, its display name, its member output identifiers, its mute state, and its enabled state — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-004]
- **FR-041**: The CLI MUST provide a command that retrieves a single audio output group from the hub by its identifier. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-005]
- **FR-042**: The groups get command MUST require exactly one group identifier as input and MUST fail with a clear usage error if it is omitted. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-006]
- **FR-043**: The groups get command MUST return the requested group regardless of its enabled/disabled state. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-007]
- **FR-044**: For the group returned by the groups get command, the CLI MUST display: its identifier, its display name, its member output identifiers, its mute state, and its enabled state — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-008]
- **FR-045**: The groups get command MUST fail with a clear, actionable "group not found" message, distinguishable from other failure types, when no group with the given identifier exists on the hub. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-011]
- **FR-046**: The groups list command MUST clearly communicate when there are zero groups to show, distinct from a failure. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-016]

### Key Entities

- **Output**: An audio output device/zone managed by the hub. Attributes relevant to this
  feature: unique identifier, human-readable display name, current volume level (a
  bounded numeric range), current mute state (on/off), whether its hardware is currently
  connected (availability), and whether it is enabled for use.
  [Source: .specify/archive/001-list-outputs/spec.md -> Output; .specify/archive/002-outputs-get/spec.md -> Output]
  The get command reuses this entity unchanged; its response is one decoded object, not an array.
  Field mapping to `OutputResponse` in `api/openapi.json` (required fields; wrong JSON type or empty `outputId`/`displayName` is malformed per FR-013; volume is passed through without re-validating bounds):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `OutputID` | string | `outputId` |
  | `DisplayName` | string | `displayName` |
  | `Volume` | int (0–100 domain range) | `volume` |
  | `Muted` | bool | `muted` |
  | `Available` | bool | `available` |
  | `Enabled` | bool | `enabled` |

  [Source: .specify/archive/001-list-outputs/data-model.md -> Output; .specify/archive/002-outputs-get/data-model.md -> Output]

- **Input**: An audio source registered with the hub, either statically configured or created ephemerally via the API. Attributes relevant to this feature: unique identifier, human-readable display name, source URI, whether it is enabled, source type (STATIC or EPHEMERAL), whether it auto-removes when its route stops, whether it supports pause/resume, and its creation timestamp (absent for static inputs). The inputs get command returns one decoded object, regardless of enabled state. Field mapping to `InputResponse` in `api/openapi.json` (required fields; wrong JSON type, empty `inputId` or `displayName`, or a `source` other than STATIC or EPHEMERAL is malformed per FR-013; `uri` is passed through unvalidated):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `InputID` | string | `inputId` |
  | `DisplayName` | string | `displayName` |
  | `URI` | string | `uri` |
  | `Enabled` | bool | `enabled` |
  | `AutoRemove` | bool | `autoRemove` |
  | `Source` | string (STATIC \| EPHEMERAL) | `source` |
  | `CreatedAt` | string, nullable (RFC 3339; `nil` for static inputs) | `createdAt` |
  | `Pauseable` | bool | `pauseable` |

  [Source: .specify/archive/003-inputs-list-get/spec.md -> "Input"; .specify/archive/003-inputs-list-get/data-model.md -> "Input"]

- **Route**: A live connection routing audio from one input to one output or output group, managed by the hub. Attributes relevant to this feature: unique identifier, source input identifier, target identifier, target type (single output or output group), current lifecycle status (starting, active, stopping, stopped, or failed), creation timestamp, playback-started timestamp (absent until playback begins), whether it can be transferred to another target, whether it supports pause/resume, and whether it is currently paused. The routes list command shows the first five of these; the routes get command returns one decoded object with all ten. Field mapping to `RouteResponse` in `api/openapi.json` (required fields; wrong JSON type, empty `routeId`, `inputId` or `targetId`, a `targetType` other than SINGLE_OUTPUT or OUTPUT_GROUP, or a `status` outside the five values is malformed per FR-013; timestamps are passed through unparsed):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `RouteID` | string | `routeId` |
  | `InputID` | string | `inputId` |
  | `TargetID` | string | `targetId` |
  | `TargetType` | string (SINGLE_OUTPUT \| OUTPUT_GROUP) | `targetType` |
  | `Status` | string (STARTING \| ACTIVE \| STOPPING \| STOPPED \| FAILED) | `status` |
  | `CreatedAt` | string (RFC 3339, required, non-nullable) | `createdAt` |
  | `StartedAt` | string, nullable (RFC 3339; `nil` before playback starts) | `startedAt` |
  | `Transferable` | bool | `transferable` |
  | `Pauseable` | bool | `pauseable` |
  | `Paused` | bool | `paused` |

  [Source: .specify/archive/004-routes-list-get/spec.md -> "Route"; .specify/archive/004-routes-list-get/data-model.md -> "Route"]

- **Output Group**: A named collection of audio outputs managed by the hub, which can be used as a single routing target. Attributes relevant to this feature: unique identifier, human-readable display name, member output identifiers (a group may have none; see Edge Cases), whether all member outputs are currently muted, and whether the group is enabled for new route creation. The groups list command shows enabled groups by default; the groups get command returns one decoded object regardless of enabled state. Both commands render the same five fields. Field mapping to `GroupResponse` in `api/openapi.json` (required fields; wrong JSON type, or an empty `groupId` or `displayName`, is malformed per FR-013; an empty `outputIds` array is valid and displayed as `[]`):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `GroupID` | string | `groupId` |
  | `DisplayName` | string | `displayName` |
  | `OutputIDs` | string list (may be empty, never `nil` on decode) | `outputIds` |
  | `Muted` | bool | `muted` |
  | `Enabled` | bool | `enabled` |

  [Source: .specify/archive/005-groups-list-get/spec.md -> "Output Group"; .specify/archive/005-groups-list-get/data-model.md -> "Group"]

## Success Criteria

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently available outputs in under 1
  second under normal network conditions. [Source: .specify/archive/001-list-outputs/spec.md -> SC-001]
- **SC-002**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases, for the outputs, inputs, routes and groups commands.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-002; .specify/archive/002-outputs-get/spec.md -> SC-003; .specify/archive/003-inputs-list-get/spec.md -> SC-004; .specify/archive/004-routes-list-get/spec.md -> SC-004; .specify/archive/005-groups-list-get/spec.md -> SC-004]
- **SC-003**: 100% of machine-readable output produced by the outputs, inputs, routes or groups commands can be parsed by a
  standard structured-data parser without error. [Source: .specify/archive/001-list-outputs/spec.md -> SC-003; .specify/archive/002-outputs-get/spec.md -> SC-004; .specify/archive/003-inputs-list-get/spec.md -> SC-005; .specify/archive/004-routes-list-get/spec.md -> SC-005; .specify/archive/005-groups-list-get/spec.md -> SC-005]
- **SC-004**: A user can always tell, without ambiguity, whether "no outputs, inputs, routes or groups were found"
  or "the command failed" occurred. [Source: .specify/archive/001-list-outputs/spec.md -> SC-004; .specify/archive/003-inputs-list-get/spec.md -> SC-006; .specify/archive/004-routes-list-get/spec.md -> SC-006; .specify/archive/005-groups-list-get/spec.md -> SC-006]
- **SC-005**: A user can identify a specific output's current volume, mute state, and
  availability from the command's output alone (a single invocation), without needing to
  run any other command or cross-reference a list.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-005; .specify/archive/002-outputs-get/spec.md -> SC-005]
- **SC-006**: A user can retrieve a specific output's full state in under 1 second under
  normal network conditions. [Source: .specify/archive/002-outputs-get/spec.md -> SC-001]
- **SC-007**: When the requested output does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such
  cases. [Source: .specify/archive/002-outputs-get/spec.md -> SC-002]
- **SC-008**: A user can view the full list of currently available inputs in under 1 second under normal network conditions. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-001]
- **SC-009**: A user can retrieve a specific input's full state in under 1 second under normal network conditions. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-002]
- **SC-010**: When the requested input does not exist, the user receives a clear "input not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-003]
- **SC-011**: A user can identify a specific input's source URI, enabled state, and source type from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-007]
- **SC-012**: A user can view the full list of currently active routes in under 1 second under normal network conditions. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-001]
- **SC-013**: A user can retrieve a specific route's full state in under 1 second under normal network conditions. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-002]
- **SC-014**: When the requested route does not exist, the user receives a clear "route not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-003]
- **SC-015**: A user can identify a specific route's source input, target, and current status from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-007]
- **SC-016**: A user can narrow a route listing to a single status, input, or target without needing to filter the output themselves after the fact. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-008]
- **SC-017**: A user can view the full list of currently enabled groups in under 1 second under normal network conditions. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-001]
- **SC-018**: A user can retrieve a specific group's full state in under 1 second under normal network conditions. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-002]
- **SC-019**: When the requested group does not exist, the user receives a clear "group not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-003]
- **SC-020**: A user can identify a specific group's member outputs and mute state from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-007]
- **SC-021**: A user can view groups that exist but are currently disabled without needing to run a separate, differently-named command. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-008]

## Assumptions

- **AS-001**: The hub's network location (host/port) is available to the CLI through a mechanism
  established once (e.g. a default matching the API's documented default server, overridable
  by the user); this feature consumes that mechanism but does not define it.
  [Source: .specify/archive/001-list-outputs/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/002-outputs-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/003-inputs-list-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/004-routes-list-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/005-groups-list-get/spec.md -> "The hub's network location (host/port) is available"]
- **AS-002**: The hub API requires no authentication for this read-only endpoint, consistent with the
  current API specification. [Source: .specify/archive/001-list-outputs/spec.md -> "The hub API requires no authentication"; .specify/archive/002-outputs-get/spec.md -> "The hub API requires no authentication"; .specify/archive/003-inputs-list-get/spec.md -> "The hub API requires no authentication"; .specify/archive/004-routes-list-get/spec.md -> "The hub API requires no authentication"; .specify/archive/005-groups-list-get/spec.md -> "The hub API requires no authentication"]
- **AS-003**: "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing.
  [Source: .specify/archive/001-list-outputs/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/002-outputs-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/003-inputs-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/004-routes-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/005-groups-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"]
- **AS-004**: The default request timeout is 5 seconds. [Source: .specify/archive/001-list-outputs/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/002-outputs-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/003-inputs-list-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/004-routes-list-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/005-groups-list-get/spec.md -> "The default request timeout is 5 seconds"]
- **AS-005**: Outputs, inputs, routes and groups are unordered by the hub API; the commands may present them in the order received
  without imposing additional sorting guarantees.
  [Source: .specify/archive/001-list-outputs/spec.md -> "Outputs are unordered by the hub API"; .specify/archive/003-inputs-list-get/spec.md -> "Inputs are unordered by the hub API"; .specify/archive/004-routes-list-get/spec.md -> "Routes are unordered by the hub API"; .specify/archive/005-groups-list-get/spec.md -> "Groups are unordered by the hub API"]
- **AS-006**: The output, input, route and group identifier is an opaque, case-sensitive string as returned by the hub; the
  get commands perform no client-side normalization or fuzzy matching.
  [Source: .specify/archive/002-outputs-get/spec.md -> "The output identifier is an opaque, case-sensitive string"; .specify/archive/003-inputs-list-get/spec.md -> "The input identifier is an opaque, case-sensitive string"; .specify/archive/004-routes-list-get/spec.md -> "The route identifier is an opaque, case-sensitive string"; .specify/archive/005-groups-list-get/spec.md -> "The group identifier is an opaque, case-sensitive string"]
- **AS-007**: Inputs have no volume, mute, or hardware-availability concept in the hub API; the inputs commands display the attributes the API defines for inputs (source URI, source type, auto-remove, pauseable, creation timestamp) rather than an outputs-shaped display.
  [Source: .specify/archive/003-inputs-list-get/spec.md -> "Unlike outputs, inputs have no volume, mute, or hardware-availability concept"]
- **AS-008**: The routes commands are read-only: creating, stopping, transferring, and pausing or resuming routes are out of scope for this feature, mirroring how the outputs and inputs read commands preceded their mutation commands.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "Creating, stopping, transferring, and pausing/resuming routes are out of scope for this feature"]
- **AS-009**: The groups commands are read-only: creating, deleting, and modifying groups (membership, volume, mute, enabled state) are out of scope for this feature, mirroring AS-008.
  [Source: .specify/archive/005-groups-list-get/spec.md -> "Creating, deleting, and modifying groups (membership, volume, mute, enabled state) are out of scope for this feature"]
