# Feature Specification: List Audio Outputs

**Feature Branch**: `001-list-outputs`

**Created**: 2026-08-24

**Status**: Completed

**Input**: User description: "I'd start with sonora outputs list — listing the hub's audio outputs. It's the simplest read-only GET in the spec, but it forces you to build the full skeleton every other command will reuse: HTTP client construction with timeouts (Principle IV), JSON decoding validated against openapi.json (Principle II), and the dual human/--json output format (Principle V) — all without the complexity of a write/mutation flow yet."

## Clarifications

### Session 2026-08-24

- Q: When the hub is unreachable or times out, should the command retry automatically before giving up, or fail immediately on the first failure? → A: Fail fast, no retries — a single attempt; on network/timeout failure, report a clear error immediately.
- Q: What should the default request timeout be for the outputs list command, before it fails with a timeout error? → A: 5 seconds.
- Q: Should each output's hardware-availability state (connected vs. disconnected) be shown directly in the default human-readable list, or only included in the --json output? → A: Show in both — availability is a visible field in the default view as well as in --json.

## User Scenarios & Testing *(mandatory)*

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

---

### Edge Cases

- What happens when the hub is unreachable (host down, wrong address, network partition)?
  The user must see a clear, actionable error rather than the command hanging indefinitely
  or crashing.
- What happens when the hub takes too long to respond? The command must fail with a clear
  timeout error rather than waiting indefinitely.
- What happens when the hub returns an error response (e.g. 5xx)? The user must see a
  clear message distinguishing "the hub reported a problem" from a plain network problem.
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? The command must fail clearly rather than displaying
  garbled or silently incomplete output.
- What happens when an output has a physically disconnected/unavailable device but is
  still enabled? Its availability state must be visibly distinguishable from an available
  output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio outputs.
- **FR-002**: By default, the command MUST show only enabled outputs.
- **FR-003**: The command MUST provide an option for the user to additionally include
  disabled outputs in the results.
- **FR-004**: For each output returned, the command MUST display: its identifier, its
  human-readable display name, its current volume level, its current mute state, whether
  it is enabled, and whether its underlying hardware is currently connected
  (availability) — in both the default human-readable view and the machine-readable
  (--json) view.
- **FR-005**: An output that is enabled but whose hardware is currently unavailable MUST
  be visibly distinguishable from an available output in the default human-readable view.
- **FR-006**: The command's default output MUST be human-readable and legible without any
  additional flags or tools.
- **FR-007**: The command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-008**: The command MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable.
- **FR-009**: The command MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; it MUST NOT wait indefinitely. The
  command MUST make a single attempt and MUST NOT automatically retry on network or
  timeout failure.
- **FR-010**: The command MUST fail with a clear, actionable error message when the hub
  returns an error response, distinguishing that case from a connectivity failure.
- **FR-011**: The command MUST exit with a status that is programmatically distinguishable
  between success, a usage error, a hub-reported error, and a connectivity/network error.
- **FR-012**: The command MUST clearly communicate when there are zero outputs to show,
  distinct from a failure.
- **FR-013**: The command MUST reject any output data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Output**: An audio output device/zone managed by the hub. Attributes relevant to this
  feature: unique identifier, human-readable display name, current volume level (a
  bounded numeric range), current mute state (on/off), whether its hardware is currently
  connected (availability), and whether it is enabled for use.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently available outputs in under 1
  second under normal network conditions.
- **SC-002**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases.
- **SC-003**: 100% of machine-readable output produced by the command can be parsed by a
  standard structured-data parser without error.
- **SC-004**: A user can always tell, without ambiguity, whether "no outputs were found"
  or "the command failed" occurred.
- **SC-005**: A user can identify a specific output's current volume and mute state from
  the command's output alone, without needing to run any other command.

## Assumptions

- The hub's network location (host/port) is available to the CLI through a mechanism
  established once (e.g. a default matching the API's documented default server, overridable
  by the user); this feature consumes that mechanism but does not define it.
- The hub API requires no authentication for this read-only endpoint, consistent with the
  current API specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing.
- The default request timeout is 5 seconds.
- Outputs are unordered by the hub API; the command may present them in the order received
  without imposing additional sorting guarantees.
