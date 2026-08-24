# Feature Specification: Get Single Audio Output

**Feature Branch**: `002-outputs-get`

**Created**: 2026-08-24

**Status**: Approved

**Input**: User description: "Natural next step following the same list-then-mutate pattern: outputs get <id> — reuses existing hub client, render, and error-handling code almost as-is, and is the simplest way to validate those \"reusable patterns\" the plan.md mentions before tackling routes/groups/inputs."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Look up a specific output by identifier (Priority: P1)

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

---

### User Story 2 - Handle a nonexistent output identifier (Priority: P2)

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

---

### User Story 3 - Consume a single output's state from a script (Priority: P3)

An automation or script wants to retrieve one output's current state in a strict,
parseable format so it can act on the data programmatically (e.g. check volume before
deciding whether to adjust it).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from the command; it is an
alternate presentation of the same data as User Story 1.

**Independent Test**: Can be fully tested by running the get command with the
machine-readable output option against a known output identifier and verifying the
result is strictly parseable and contains the same fields available in the
human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has an output with a known identifier, **When** the user runs the
   get command with that identifier requesting machine-readable output, **Then** the
   result is valid, strictly parseable structured data containing the output's
   identifier, display name, volume, mute state, availability, and enabled state.

---

### Edge Cases

- What happens when no identifier is provided? The user must see a clear usage error
  explaining that an identifier is required, not a confusing failure.
- What happens when the hub is unreachable (host down, wrong address, network
  partition)? The user must see a clear, actionable error rather than the command
  hanging indefinitely or crashing.
- What happens when the hub takes too long to respond? The command must fail with a
  clear timeout error rather than waiting indefinitely.
- What happens when the hub returns an error response other than "not found" (e.g.
  5xx)? The user must see a clear message distinguishing "the hub reported a problem"
  from a plain network problem or a "not found" result.
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? The command must fail clearly rather than
  displaying garbled or silently incomplete output.
- What happens when the requested output exists but is enabled with unavailable
  hardware? Its availability state must be visibly distinguishable from an available
  output, consistent with the list command.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that retrieves a single audio output from
  the hub by its identifier.
- **FR-002**: The command MUST require exactly one output identifier as input and MUST
  fail with a clear usage error if it is omitted.
- **FR-003**: The command MUST return the requested output regardless of its
  enabled/disabled state (unlike the list command's default filtering).
- **FR-004**: For the output returned, the command MUST display: its identifier, its
  human-readable display name, its current volume level, its current mute state,
  whether it is enabled, and whether its underlying hardware is currently connected
  (availability) — in both the default human-readable view and the machine-readable
  (--json) view.
- **FR-005**: An output that is enabled but whose hardware is currently unavailable MUST
  be visibly distinguishable from an available output in the default human-readable
  view.
- **FR-006**: The command's default output MUST be human-readable and legible without
  any additional flags or tools.
- **FR-007**: The command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-008**: The command MUST fail with a clear, actionable "output not found" message,
  distinguishable from other failure types, when no output with the given identifier
  exists on the hub.
- **FR-009**: The command MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable.
- **FR-010**: The command MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; it MUST NOT wait indefinitely. The
  command MUST make a single attempt and MUST NOT automatically retry on network or
  timeout failure.
- **FR-011**: The command MUST fail with a clear, actionable error message when the hub
  returns an error response other than "not found," distinguishing that case from a
  connectivity failure and from a "not found" result.
- **FR-012**: The command MUST exit with a status that is programmatically
  distinguishable between success, a usage error, "not found," a hub-reported error, and
  a connectivity/network error.
- **FR-013**: The command MUST reject output data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Output**: An audio output device/zone managed by the hub. Attributes relevant to
  this feature: unique identifier, human-readable display name, current volume level (a
  bounded numeric range), current mute state (on/off), whether its hardware is currently
  connected (availability), and whether it is enabled for use.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can retrieve a specific output's full state in under 1 second under
  normal network conditions.
- **SC-002**: When the requested output does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such
  cases.
- **SC-003**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases.
- **SC-004**: 100% of machine-readable output produced by the command can be parsed by a
  standard structured-data parser without error.
- **SC-005**: A user can identify a specific output's current volume, mute state, and
  availability from a single command invocation, without needing to run any other
  command or cross-reference a list.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same
  mechanism established by the `outputs list` feature; this feature consumes that
  mechanism but does not redefine it.
- The hub API requires no authentication for this read-only endpoint, consistent with
  the current API specification.
- "Human-readable" default output presents the same set of fields as the
  machine-readable output, formatted for readability rather than strict parsing,
  consistent with `outputs list`.
- The default request timeout is 5 seconds, consistent with `outputs list`.
- The output identifier is an opaque, case-sensitive string as returned by the hub; the
  command performs no client-side normalization or fuzzy matching.
