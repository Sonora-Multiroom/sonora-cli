# Main Project Specification

> **Revision**: 2026-11-23 — Seeded from specs/001-list-outputs (first archival): user stories 1–3, edge cases, FR-001–FR-013, Output entity, SC-001–SC-005, AS-001–AS-005.

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

### Edge Cases

- What happens when the hub is unreachable (host down, wrong address, network partition)?
  The user must see a clear, actionable error rather than the command hanging indefinitely
  or crashing. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub is unreachable"]
- What happens when the hub takes too long to respond? The command must fail with a clear
  timeout error rather than waiting indefinitely. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub takes too long to respond?"]
- What happens when the hub returns an error response (e.g. 5xx)? The user must see a
  clear message distinguishing "the hub reported a problem" from a plain network problem.
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns an error response"]
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? The command must fail clearly rather than displaying
  garbled or silently incomplete output.
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"]
- What happens when an output has a physically disconnected/unavailable device but is
  still enabled? Its availability state must be visibly distinguishable from an available
  output. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when an output has a physically disconnected/unavailable device"]

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
  (--json) view. [Source: .specify/archive/001-list-outputs/spec.md -> FR-004]
- **FR-005**: An output that is enabled but whose hardware is currently unavailable MUST
  be visibly distinguishable from an available output in the default human-readable view.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-005]
- **FR-006**: The command's default output MUST be human-readable and legible without any
  additional flags or tools. [Source: .specify/archive/001-list-outputs/spec.md -> FR-006]
- **FR-007**: The command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-007]
- **FR-008**: The command MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable. [Source: .specify/archive/001-list-outputs/spec.md -> FR-008]
- **FR-009**: The command MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; it MUST NOT wait indefinitely. The
  command MUST make a single attempt and MUST NOT automatically retry on network or
  timeout failure. [Source: .specify/archive/001-list-outputs/spec.md -> FR-009]
- **FR-010**: The command MUST fail with a clear, actionable error message when the hub
  returns an error response, distinguishing that case from a connectivity failure.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-010]
- **FR-011**: The command MUST exit with a status that is programmatically distinguishable
  between success, a usage error, a hub-reported error, and a connectivity/network error.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-011]
- **FR-012**: The command MUST clearly communicate when there are zero outputs to show,
  distinct from a failure. [Source: .specify/archive/001-list-outputs/spec.md -> FR-012]
- **FR-013**: The command MUST reject any output data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results. [Source: .specify/archive/001-list-outputs/spec.md -> FR-013]

### Key Entities

- **Output**: An audio output device/zone managed by the hub. Attributes relevant to this
  feature: unique identifier, human-readable display name, current volume level (a
  bounded numeric range), current mute state (on/off), whether its hardware is currently
  connected (availability), and whether it is enabled for use.
  [Source: .specify/archive/001-list-outputs/spec.md -> Output]
  Field mapping to `OutputResponse` in `api/openapi.json` (required fields; wrong JSON type or empty `outputId`/`displayName` is malformed per FR-013; volume is passed through without re-validating bounds):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `OutputID` | string | `outputId` |
  | `DisplayName` | string | `displayName` |
  | `Volume` | int (0–100 domain range) | `volume` |
  | `Muted` | bool | `muted` |
  | `Available` | bool | `available` |
  | `Enabled` | bool | `enabled` |

  [Source: .specify/archive/001-list-outputs/data-model.md -> Output]

## Success Criteria

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently available outputs in under 1
  second under normal network conditions. [Source: .specify/archive/001-list-outputs/spec.md -> SC-001]
- **SC-002**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-002]
- **SC-003**: 100% of machine-readable output produced by the command can be parsed by a
  standard structured-data parser without error. [Source: .specify/archive/001-list-outputs/spec.md -> SC-003]
- **SC-004**: A user can always tell, without ambiguity, whether "no outputs were found"
  or "the command failed" occurred. [Source: .specify/archive/001-list-outputs/spec.md -> SC-004]
- **SC-005**: A user can identify a specific output's current volume and mute state from
  the command's output alone, without needing to run any other command.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-005]

## Assumptions

- **AS-001**: The hub's network location (host/port) is available to the CLI through a mechanism
  established once (e.g. a default matching the API's documented default server, overridable
  by the user); this feature consumes that mechanism but does not define it.
  [Source: .specify/archive/001-list-outputs/spec.md -> "The hub's network location (host/port) is available"]
- **AS-002**: The hub API requires no authentication for this read-only endpoint, consistent with the
  current API specification. [Source: .specify/archive/001-list-outputs/spec.md -> "The hub API requires no authentication"]
- **AS-003**: "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing.
  [Source: .specify/archive/001-list-outputs/spec.md -> "\"Human-readable\" default output presents the same set of fields"]
- **AS-004**: The default request timeout is 5 seconds. [Source: .specify/archive/001-list-outputs/spec.md -> "The default request timeout is 5 seconds"]
- **AS-005**: Outputs are unordered by the hub API; the command may present them in the order received
  without imposing additional sorting guarantees.
  [Source: .specify/archive/001-list-outputs/spec.md -> "Outputs are unordered by the hub API"]
