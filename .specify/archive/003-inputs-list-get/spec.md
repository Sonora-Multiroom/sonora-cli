# Feature Specification: List and Get Audio Inputs

**Feature Branch**: `003-inputs-list-get`

**Created**: 2026-08-25

**Status**: Completed

**Input**: User description: "next feature - inputs list (mirrors outputs list) and inputs get (similar to outputs get)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View currently active inputs (Priority: P1)

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

---

### User Story 2 - Look up a specific input by identifier (Priority: P1)

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

---

### User Story 3 - Include disabled inputs in the list (Priority: P2)

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

---

### User Story 4 - Handle a nonexistent input identifier (Priority: P2)

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

---

### User Story 5 - Consume input data from a script (Priority: P3)

An automation or script wants to retrieve the current inputs, or a single input's state,
in a strict, parseable format so it can act on the data programmatically.

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 1 and 2.

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

---

### Edge Cases

- What happens when no input identifier is provided to the get command? The user must see
  a clear usage error explaining that an identifier is required, not a confusing failure.
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
- What happens when a static (YAML-configured) input is returned? It has no creation
  timestamp; this must be displayed clearly (e.g. "n/a") rather than as an error or blank
  field.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio inputs.
- **FR-002**: By default, the list command MUST show only enabled inputs.
- **FR-003**: The list command MUST provide an option for the user to additionally
  include disabled inputs in the results.
- **FR-004**: For each input returned by the list command, the CLI MUST display: its
  identifier, its human-readable display name, its source URI, whether it is enabled, its
  source type (static or ephemeral), whether it auto-removes when its route stops,
  whether it supports pause/resume, and its creation timestamp (or an explicit indication
  that none exists) — in both the default human-readable view and the machine-readable
  (--json) view.
- **FR-005**: The CLI MUST provide a command that retrieves a single audio input from the
  hub by its identifier.
- **FR-006**: The get command MUST require exactly one input identifier as input and MUST
  fail with a clear usage error if it is omitted.
- **FR-007**: The get command MUST return the requested input regardless of its
  enabled/disabled state (unlike the list command's default filtering).
- **FR-008**: For the input returned by the get command, the CLI MUST display the same
  set of attributes specified in FR-004, in both the default human-readable view and the
  machine-readable (--json) view.
- **FR-009**: Both commands' default output MUST be human-readable and legible without
  any additional flags or tools.
- **FR-010**: Both commands MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-011**: The get command MUST fail with a clear, actionable "input not found"
  message, distinguishable from other failure types, when no input with the given
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
- **FR-016**: The list command MUST clearly communicate when there are zero inputs to
  show, distinct from a failure.
- **FR-017**: Both commands MUST reject input data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Input**: An audio source registered with the hub, either statically configured or
  created ephemerally via the API. Attributes relevant to this feature: unique
  identifier, human-readable display name, source URI, whether it is enabled, source type
  (STATIC or EPHEMERAL), whether it auto-removes when its route stops, whether it
  supports pause/resume, and its creation timestamp (absent for static inputs).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently available inputs in under 1
  second under normal network conditions.
- **SC-002**: A user can retrieve a specific input's full state in under 1 second under
  normal network conditions.
- **SC-003**: When the requested input does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such cases.
- **SC-004**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases, for
  both commands.
- **SC-005**: 100% of machine-readable output produced by either command can be parsed by
  a standard structured-data parser without error.
- **SC-006**: A user can always tell, without ambiguity, whether "no inputs were found" or
  "the command failed" occurred.
- **SC-007**: A user can identify a specific input's source URI, enabled state, and
  source type from a single command invocation, without needing to run any other command
  or cross-reference a list.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same
  mechanism established by the `outputs list`/`outputs get` features; this feature
  consumes that mechanism but does not redefine it.
- The hub API requires no authentication for these read-only endpoints, consistent with
  the current API specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing, consistent with the
  outputs commands.
- The default request timeout is 5 seconds, consistent with the outputs commands.
- The input identifier is an opaque, case-sensitive string as returned by the hub; the
  command performs no client-side normalization or fuzzy matching.
- Inputs are unordered by the hub API; the list command may present them in the order
  received without imposing additional sorting guarantees.
- Unlike outputs, inputs have no volume, mute, or hardware-availability concept in the
  API; this feature displays the attributes the hub API actually defines for inputs
  (source URI, source type, auto-remove, pauseable, creation timestamp) rather than
  forcing an outputs-shaped display.
