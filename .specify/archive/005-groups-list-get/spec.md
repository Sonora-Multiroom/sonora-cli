# Feature Specification: List and Get Output Groups

**Feature Branch**: `005-groups-list-get`

**Created**: 2026-08-26

**Status**: Approved

**Input**: User description: "next feature, similar to already implemented `list` and `get` for outputs, inputs and routes, logically, would be `groups list` and `groups get`"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View currently enabled output groups (Priority: P1)

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

---

### User Story 2 - Include disabled groups (Priority: P2)

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

---

### User Story 3 - Look up a specific group by identifier (Priority: P1)

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

---

### User Story 4 - Handle a nonexistent group identifier (Priority: P2)

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

---

### User Story 5 - Consume group data from a script (Priority: P3)

An automation or script wants to retrieve the current groups, or a single group's state,
in a strict, parseable format so it can act on the data programmatically (e.g. pick a
group's member outputs before issuing a route command).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 1 and 3.

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

---

### Edge Cases

- What happens when no group identifier is provided to the get command? The user must see
  a clear usage error explaining that an identifier is required, not a confusing failure.
- What happens when a group has zero member outputs? Its member output list must be
  displayed clearly as empty rather than as an error or omitted field.
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

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio output groups.
- **FR-002**: By default, the list command MUST show only enabled groups.
- **FR-003**: The list command MUST provide an option to include disabled groups in the
  results.
- **FR-004**: For each group returned by the list command, the CLI MUST display: its
  identifier, its display name, its member output identifiers, its mute state, and its
  enabled state — in both the default human-readable view and the machine-readable
  (--json) view.
- **FR-005**: The CLI MUST provide a command that retrieves a single audio output group
  from the hub by its identifier.
- **FR-006**: The get command MUST require exactly one group identifier as input and MUST
  fail with a clear usage error if it is omitted.
- **FR-007**: The get command MUST return the requested group regardless of its
  enabled/disabled state.
- **FR-008**: For the group returned by the get command, the CLI MUST display: its
  identifier, its display name, its member output identifiers, its mute state, and its
  enabled state — in both the default human-readable view and the machine-readable
  (--json) view.
- **FR-009**: Both commands' default output MUST be human-readable and legible without any
  additional flags or tools.
- **FR-010**: Both commands MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-011**: The get command MUST fail with a clear, actionable "group not found"
  message, distinguishable from other failure types, when no group with the given
  identifier exists on the hub.
- **FR-012**: Both commands MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable.
- **FR-013**: Both commands MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; neither command MUST wait
  indefinitely. Each command MUST make a single attempt and MUST NOT automatically retry
  on network or timeout failure.
- **FR-014**: Both commands MUST fail with a clear, actionable error message when the hub
  returns an error response other than "not found" (for get) or a plain success (for
  list), distinguishing that case from a connectivity failure and, for get, from a "not
  found" result.
- **FR-015**: Both commands MUST exit with a status that is programmatically
  distinguishable between success, a usage error, "not found" (get only), a hub-reported
  error, and a connectivity/network error.
- **FR-016**: The list command MUST clearly communicate when there are zero groups to
  show, distinct from a failure.
- **FR-017**: Both commands MUST reject group data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Output Group**: A named collection of one or more audio outputs, managed by the hub,
  that can be used as a single routing target. Attributes relevant to this feature: unique
  identifier, human-readable display name, list of member output identifiers, whether all
  member outputs are currently muted, and whether the group is enabled for new route
  creation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently enabled groups in under 1 second
  under normal network conditions.
- **SC-002**: A user can retrieve a specific group's full state in under 1 second under
  normal network conditions.
- **SC-003**: When the requested group does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such cases.
- **SC-004**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases, for both
  commands.
- **SC-005**: 100% of machine-readable output produced by either command can be parsed by
  a standard structured-data parser without error.
- **SC-006**: A user can always tell, without ambiguity, whether "no groups are available"
  or "the command failed" occurred.
- **SC-007**: A user can identify a specific group's member outputs and mute state from a
  single command invocation, without needing to run any other command or cross-reference a
  list.
- **SC-008**: A user can view groups that exist but are currently disabled without needing
  to run a separate, differently-named command.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same
  mechanism established by the `outputs`/`inputs`/`routes` features; this feature consumes
  that mechanism but does not redefine it.
- The hub API requires no authentication for these read-only endpoints, consistent with
  the current API specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing, consistent with the
  outputs, inputs, and routes commands.
- The default request timeout is 5 seconds, consistent with the outputs, inputs, and
  routes commands.
- The group identifier is an opaque, case-sensitive string as returned by the hub; the
  command performs no client-side normalization or fuzzy matching.
- Groups are unordered by the hub API; the list command may present them in the order
  received without imposing additional sorting guarantees.
- Creating, deleting, and modifying groups (membership, volume, mute, enabled state) are
  out of scope for this feature; only read (list/get) operations are covered here,
  mirroring how `outputs list`/`outputs get`, `inputs list`/`inputs get`, and
  `routes list`/`routes get` preceded any corresponding mutation commands.
