# Feature Specification: Adopt Verb-First Command Landscape

**Feature Branch**: `007-refactor-cli-commands`

**Created**: 2026-08-27

**Status**: Draft

**Input**: User description: "need refactoring - adopt already implemented commands as
specified in docs/cli-command-landscape.md"

## Clarifications

### Session 2026-08-27

- Q: When an old-style command (`sonora inputs list`, `sonora outputs get <id>`, or `play`
  with `--group`/`--output`/`--name`) is run after this refactor ships, should it be removed
  outright with no transition period, or kept working (with a deprecation warning) for some
  time before removal? → A: Hard cutover — old forms are removed immediately and fail with a
  usage error; no deprecation/warning period.
- Q: Can a resource identifier (input/output/group/route id) itself contain a `/`
  character, or are ids guaranteed slash-free? → A: Ids are unique identifiers matching
  `^[a-zA-Z0-9_-]{1,255}$` — slash-free by construction, so a resource path is parsed
  unambiguously by splitting on the first `/`.
- Q: When an old-style invocation fails, must the error message specifically explain what
  changed (e.g. name the new equivalent command), or is a standard "unrecognized command"
  usage error sufficient? → A: A standard usage error is sufficient — the CLI is not required
  to specifically detect or explain old-style syntax.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read commands follow the verb-first, path-style shape (Priority: P1)

An operator who has learned the new `sonora route inputs/<id> outputs/<id>` command (from
the separately specified `route` feature) expects every other resource — inputs, outputs,
groups, and routes — to read the same way: verb first, resource path second. Today, reading
those same resources instead uses the older `sonora <resource> <verb>` shape
(`sonora inputs list`, `sonora outputs get <id>`), which is inconsistent with `route` and
with the documented command landscape.

**Why this priority**: This is the actual inconsistency prompting the refactor — the CLI
currently speaks two different command grammars depending on which resource you're touching.
Unifying the already-shipped read commands onto one grammar is the highest-value, lowest-risk
piece of the refactor, since these commands don't mutate anything.

**Independent Test**: Can be fully tested by running `sonora get inputs`, `sonora get
outputs`, `sonora get groups`, and `sonora get routes` (each with and without a trailing
`/<id>`), and verifying each returns exactly the same data as today's `sonora <resource>
list` / `sonora <resource> get <id>` commands.

**Acceptance Scenarios**:

1. **Given** existing inputs, outputs, groups, or routes, **When** the user runs `sonora get
   <resource>` for any of the four resources, **Then** the command returns the same
   collection data as today's `sonora <resource> list`, including support for the existing
   collection filter flags (`--include-disabled` for inputs/outputs/groups; `--status`,
   `--input-id`, `--target-id` for routes).
2. **Given** an existing resource of a known identifier, **When** the user runs `sonora get
   <resource>/<id>`, **Then** the command returns the same single-item data as today's
   `sonora <resource> get <id>`.
3. **Given** any of the four resources, **When** the user runs `sonora list <resource>`
   (no id), **Then** the result is identical to `sonora get <resource>` — `list` behaves as
   an exact synonym for the collection form only.
4. **Given** any of the four resources, **When** the user runs `sonora list <resource>/<id>`
   (an id after `list`), **Then** the command fails with a clear usage error, since `list` is
   valid only for the collection form.
5. **Given** the old command shape, **When** the user runs `sonora inputs list`, `sonora
   outputs get <id>`, or any other pre-refactor invocation, **Then** the command fails with a
   clear usage error rather than being silently reinterpreted or producing incorrect output.

---

### User Story 2 - Resource aliases work everywhere a resource path appears (Priority: P2)

An operator who frequently types commands wants to use the short aliases (`in`, `out`, `gr`,
`rt`) documented alongside the full resource names, so `sonora get out/<id>` works exactly
like `sonora get outputs/<id>`.

**Why this priority**: Aliases are a documented part of the same landscape and a natural
companion to User Story 1, but they're a convenience layer on top of the core grammar change,
not a prerequisite for it.

**Independent Test**: Can be fully tested by running the collection and single-item forms of
`get`/`list` for each resource using its alias instead of its full name, and verifying
identical results to using the full name.

**Acceptance Scenarios**:

1. **Given** an existing input, output, group, or route, **When** the user substitutes the
   resource's short alias (`in`, `out`, `gr`, `rt` respectively) for its full name in any
   `get`/`list` command, **Then** the command behaves identically to using the full resource
   name.
2. **Given** an unrecognized resource name or alias, **When** the user runs `sonora get
   <unrecognized>`, **Then** the command fails with a clear usage error naming the
   unrecognized resource.

---

### User Story 3 - `play`'s target and flags match the new landscape (Priority: P3)

An operator using `sonora play <uri> <target-id> --output` today expects `play` to address
its target the same way `route` does — a path like `outputs/<id>` or `groups/<id>` — instead
of a bare id plus a separate `--group`/`--output` flag, and expects the display-name flag to
be named consistently with the rest of the CLI.

**Why this priority**: `play` is the one already-implemented command that mutates state, so
its change carries the most user-facing risk (existing scripts calling it will break). It's
sequenced last so the lower-risk read-command grammar lands and stabilizes first.

**Independent Test**: Can be fully tested by running `sonora play <uri> outputs/<id>` and
`sonora play <uri> groups/<id>` (including with `--display-name` and `--volume`), and
verifying the same playback result as today's `sonora play <uri> <id> --output` /
`sonora play <uri> <id> --group`.

**Acceptance Scenarios**:

1. **Given** a playable URI and an existing output, **When** the user runs `sonora play <uri>
   outputs/<output-id>`, **Then** playback starts on that output exactly as today's
   `sonora play <uri> <output-id> --output` does.
2. **Given** a playable URI and an existing group, **When** the user runs `sonora play <uri>
   groups/<group-id>`, **Then** playback starts on that group exactly as today's
   `sonora play <uri> <group-id> --group` does.
3. **Given** a playable URI and a target, **When** the user supplies `--display-name`,
   **Then** it sets the new input's display name exactly as today's `--name` flag does.
4. **Given** the old invocation, **When** the user runs `sonora play <uri> <id> --group` or
   `sonora play <uri> <id> --output` or `sonora play <uri> <id> --name <name>`, **Then** the
   command fails with a clear usage error rather than silently accepting the old flags.

---

### Edge Cases

- What happens when a resource path's prefix names a real resource type but the id after the
  slash doesn't exist (e.g. `sonora get outputs/does-not-exist`)? The command must fail with
  the same "not found" error it produces today for the equivalent `sonora outputs get
  does-not-exist`.
- What happens when a resource path has no slash and no id at all except the bare resource
  name (e.g. `sonora get outputs`)? This is the valid collection form, not an error.
- What happens when the user runs a bare verb with nothing after it (e.g. `sonora get`)? The
  command must fail with a clear usage error listing the valid resources.
- What happens when the user mixes an alias and a full name incorrectly, e.g. adds a stray
  slash or extra segment (`sonora get out/foo/bar`)? Since ids match
  `^[a-zA-Z0-9_-]{1,255}$` and never contain `/`, an extra slash unambiguously means the
  argument is malformed; the command must fail with a clear usage error rather than guessing
  which segment is the id.
- What happens when existing collection filter flags (`--include-disabled`, `--status`,
  `--input-id`, `--target-id`) are combined with the new `get`/`list` verbs? They must
  continue to filter exactly as they do today — only the verb/resource-path shape changes.
- What happens to `--json`, `--hub-url`, and `--verbose`? They must keep working, unchanged,
  on every refactored command.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide `sonora get <resource>` for each of inputs, outputs,
  groups, and routes, returning the same collection data as today's `sonora <resource> list`,
  including today's collection filter flags.
- **FR-002**: The CLI MUST provide `sonora get <resource>/<id>` for each of inputs, outputs,
  groups, and routes, returning the same single-item data as today's `sonora <resource> get
  <id>`.
- **FR-003**: The CLI MUST accept `sonora list <resource>` as an exact synonym of `sonora get
  <resource>` (collection form only), and MUST fail with a clear usage error if `list` is
  given a resource path that includes an id.
- **FR-004**: The CLI MUST accept the short aliases `in`, `out`, `gr`, and `rt` interchangeably
  with `inputs`, `outputs`, `groups`, and `routes` respectively, everywhere a resource name or
  resource path is accepted.
- **FR-004a**: A resource path MUST be parsed by splitting on the first `/` into a
  resource-name segment and an identifier segment; the CLI MUST treat identifiers as matching
  `^[a-zA-Z0-9_-]{1,255}$` and MUST fail with a clear usage error if the identifier segment
  contains an additional `/` or otherwise doesn't match that pattern.
- **FR-005**: The CLI MUST remove the old `sonora <resource> list` and `sonora <resource> get
  <id>` command forms for inputs, outputs, groups, and routes, rather than keeping them as
  hidden or deprecated aliases.
- **FR-006**: The CLI MUST fail with a clear, standard usage error — not a silent
  misinterpretation or incorrect result — when given an old-style invocation, an unrecognized
  resource name or alias, or a malformed resource path. The CLI is NOT required to
  specifically detect old-style syntax in order to name it or point to its new equivalent; a
  generic "unrecognized command"-style error is sufficient.
- **FR-006a**: When `get` or `list` is given no resource argument at all (`sonora get`,
  `sonora list`), the CLI MUST fail with a usage error whose message enumerates the valid
  resource names. This is a deliberate exception to FR-006's "generic error is sufficient"
  allowance: FR-006 concerns invocations where the user typed *something* wrong and the CLI
  need not diagnose it, whereas here the user has typed a valid verb and supplied nothing,
  so the set of valid completions is both unambiguous and the only useful thing to say.
- **FR-007**: The CLI MUST change `play`'s target argument to a resource path
  (`outputs/<id>`, `groups/<id>`, or their aliases), removing the separate `--group` and
  `--output` disambiguation flags.
- **FR-008**: The CLI MUST rename `play`'s display-name flag from `--name` to
  `--display-name`, and MUST fail with a clear usage error if the old `--name`, `--group`, or
  `--output` flags are supplied.
- **FR-009**: The CLI MUST preserve `--json`, `--hub-url`, and `--verbose` unchanged, in
  meaning and behavior, on every refactored command.
- **FR-010**: The CLI MUST preserve the exact displayed data (fields and values, in both the
  default and `--json` views) of every refactored read command — only the invocation shape
  changes, not what's shown.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the currently implemented inputs/outputs/groups/routes read commands
  are reachable only through the verb-first, path-style shape described in
  `docs/cli-command-landscape.md` — none remain reachable through the old
  `sonora <resource> <verb>` shape.
- **SC-002**: A user substituting any documented resource alias (`in`/`out`/`gr`/`rt`) for
  its full resource name gets identical output, for every refactored command.
- **SC-003**: A user running any pre-refactor command form (old noun-verb commands, or
  `play` with `--group`/`--output`/`--name`) receives a clear usage error, in 100% of cases,
  rather than an incorrect or silently-reinterpreted result.
- **SC-004**: For every refactored command, the data displayed (in both default and `--json`
  output) is byte-for-byte unchanged from before the refactor, given the same underlying hub
  state.

## Assumptions

- Only commands that are already implemented today are in scope: `inputs`/`outputs`/`groups`/
  `routes` `list`/`get`, and `play`. The `route` command itself is out of scope — it's
  already specified separately (`specs/008-route-command`) and not yet implemented.
- Commands documented in `docs/cli-command-landscape.md` that aren't implemented yet
  (`create`, `delete`, `enable`/`disable`, `mute`/`unmute`, `set ... volume`, `transfer`,
  `stop`/`pause`/`resume`, `master-mute`) are out of scope here; this feature only brings
  already-shipped commands in line with the documented grammar.
- `--json`, `--hub-url`, and `--verbose` are unaffected by this refactor; it changes only
  command and argument parsing, not request behavior or output rendering.
- The underlying hub API and response shapes are unchanged; this is a CLI-surface-only
  refactor.
