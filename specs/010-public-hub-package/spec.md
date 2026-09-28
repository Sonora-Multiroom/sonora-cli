# Feature Specification: Public Hub Client Package

**Feature Branch**: `010-public-hub-package`

**Created**: 2026-09-27

**Status**: Draft

**Input**: User description: "Phase 2 of
sonora-mcp/docs/future/go-rewrite-shared-hub-client.md — promote the Sonora Multiroom CLI's
internal Multiroom Audio Hub API client to a public, versioned package inside this module so
that other Sonora projects (at minimum sonora-mcp) can import it instead of maintaining their
own client. Pure refactor, no CLI behavior change."

## Context

Sonora Multiroom has two clients for the same Multiroom Audio Hub API: the tested client
inside this CLI, and a separate hand-written client in sonora-mcp (a different language,
covering 24 of the 49 hub operations, with no timeouts or tests). Keeping both in step with
every `api/openapi.json` change is ongoing duplicated work. The decision (option 2a in the
source plan) is to make this CLI's client importable by other projects, then rewrite
sonora-mcp on top of it. This feature is only the sonora-cli side; the sonora-mcp rewrite is a
separate feature in that repository.

Because this is a developer-facing refactor, the "users" below are developers of other Sonora
projects, maintainers of this CLI, and existing CLI users (who must notice nothing).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Import the hub client from another project (Priority: P1)

A developer of another Sonora project (sonora-mcp) adds this module as a dependency by its
public import path, pinned to a released tag, and calls hub operations (list outputs, set
volume, create a route, speak, …) through the shared client, getting the same timeouts,
typed errors and error classification the CLI uses.

**Why this priority**: This is the whole point of the feature. Without an importable,
released client, sonora-mcp cannot drop its own client.

**Independent Test**: In a scratch module outside this repository, require this module at the
released tag, call one read operation and one write operation against a fake hub, and
classify a failure — with no access to anything under this repo's internal packages.

**Acceptance Scenarios**:

1. **Given** a released tag of this module, **When** an external project requires the
   module by its public path and imports the hub client package, **Then** it resolves and
   builds with no replace directives, workspace files or private-module settings.
2. **Given** an external project using the client, **When** a hub call fails (hub
   unreachable, timeout, 404, 4xx/5xx, malformed response), **Then** the project can obtain
   the error class and friendly message the client's classifier produces — the same
   classification the CLI starts from — without depending on any CLI-specific concept
   such as exit codes. (A caller may refine a generic not-found class with its own context,
   as the CLI's route commands do.)
3. **Given** the client package, **When** a developer reads its package documentation,
   **Then** every exported identifier is documented in terms that make sense without access
   to this repo's specs, research notes or constitution.
4. **Given** the client package's dependency list, **When** it is imported, **Then** it adds
   no third-party dependencies to the consumer beyond the standard library.

---

### User Story 2 - Existing CLI users notice no change (Priority: P1)

A person using `sonora` on Windows or Linux upgrades to the build that contains this refactor
and keeps using every command exactly as before.

**Why this priority**: The refactor is only acceptable if it is invisible to CLI users; a
behavior regression would break live audio control and scripts that rely on exit codes.

**Independent Test**: Run the full existing unit and contract test suite unchanged in intent
(only import paths adjusted) on Windows and Linux, and compare `sonora --help`, command
output and exit codes for representative commands before and after.

**Acceptance Scenarios**:

1. **Given** the refactored build, **When** any existing command is run, **Then** its
   output (YAML and `--json`), error messages and exit codes are identical to the previous
   build.
2. **Given** a build produced by the project's build scripts with a version supplied,
   **When** `sonora --version` is run, **Then** it prints the supplied version, not the
   default placeholder.
3. **Given** a hub error of each class, **When** a command hits it, **Then** the CLI exits
   with the same code as before, even though the error-class-to-exit-code mapping now lives
   inside the CLI rather than in the shared client.

---

### User Story 3 - Consumers verify against the same hub spec (Priority: P2)

A developer of a consuming project runs a conformance test against the exact Multiroom Audio
Hub API specification that the imported client version was built and tested against, without
keeping their own copy of `openapi.json`.

**Why this priority**: Removes the last duplicated artifact between repos, but the client is
usable without it.

**Independent Test**: From an external module, import the published spec and confirm it is
byte-identical to `api/openapi.json` at the same tag.

**Acceptance Scenarios**:

1. **Given** a released tag, **When** a consumer imports the published spec, **Then** it
   receives the full contents of `api/openapi.json` from that tag.
2. **Given** `api/openapi.json` is updated in this repo, **When** the module is rebuilt,
   **Then** the published spec reflects the update with no separate copy to keep in sync.

---

### User Story 4 - Maintainers release the shared client safely (Priority: P2)

A maintainer tags a release of this module that contains the public client so consumers can
pin it, following rules that keep already-published versions immutable.

**Why this priority**: Consumers can only depend on a tag; a mistaken or moved tag in a public
module is permanent damage.

**Independent Test**: Tag a release, fetch it from a clean environment through the public Go
module proxy, and build a consumer against it.

**Acceptance Scenarios**:

1. **Given** the refactor is merged, **When** a new tag is published, **Then** consumers can
   resolve that tag through the public module proxy.
2. **Given** a published tag has a problem, **When** the maintainer fixes it, **Then** a new
   tag is published and the existing tag is never moved, deleted or re-created.

### Edge Cases

- A build or release script still injects the version under the old module path: the build
  succeeds silently but the version stays at the default — this MUST be caught by a test or
  release check (US2 scenario 2).
- A stray reference to the old module path or the old internal client location remains in
  code, tests, scripts or docs: an automated test MUST fail on it.
- A CLI-only concept (exit codes, flags, rendering, config discovery) is still reachable from
  the public client: this is a boundary violation and MUST be removed before release.
- A consumer tries to import CLI internals (config discovery, renderers): this MUST remain
  impossible; only the client package and published spec are public.
- The error-classification messages must stay free of CLI-specific wording (flags, exit
  codes, environment variables, config files, "this CLI") so they read correctly when
  surfaced by a non-CLI consumer such as an AI assistant. Today the TTS "hub address" and
  "version mismatch" messages break this rule; the CLI must add that wording itself instead
  (FR-015).
- Contract tests move with the client: they MUST keep running in this repo as the
  conformance gate, not be dropped during the move.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The module MUST be addressable by its public repository import path
  (`github.com/Sonora-Multiroom/sonora-cli`), and every internal reference to the old
  module path MUST be updated, including build-time version injection in `Makefile` and
  `.goreleaser.yaml` (`release.sh` and `scripts/` must be checked too). The maintainer's
  local, untracked `build.sh` helper is updated too but is not part of the repository.
- **FR-002**: The Multiroom Audio Hub API client MUST be moved from its internal location to
  a public package at the module root (`hub/`), preserving its file history.
- **FR-003**: The public client MUST contain only hub protocol concerns: request/response
  types, HTTP calls, timeouts, error types, error classification and friendly error messages.
- **FR-004**: The mapping from error class to CLI exit code MUST move out of the public
  client into a CLI-internal package, together with any error class that only the CLI uses
  (the usage-error class); all CLI call sites MUST use the new mapping and produce the same
  exit codes as before.
- **FR-005**: Error classification and its friendly messages MUST remain in the public
  client so non-CLI consumers can present the same classification.
- **FR-006**: Helpers the client itself uses to build error messages (e.g. single-line
  normalization) MUST remain available from the public client.
- **FR-007**: Every exported identifier of the public client MUST have documentation that is
  self-contained — no references to this repo's constitution principles, `research.md`,
  `data-model.md` or other spec artifacts.
- **FR-008**: The public client MUST keep a standard-library-only dependency footprint.
- **FR-009**: Existing contract and unit tests for the client MUST remain in this repo and
  continue to run as the conformance gate against `api/openapi.json`.
- **FR-010**: The CLI MUST have no observable behavior change: same commands, flags, output
  (YAML and JSON), error messages and exit codes.
- **FR-011**: There MUST be an automated check that a build produced with a supplied version
  reports that version from `sonora --version`.
- **FR-012**: The hub API specification (`api/openapi.json`) SHOULD be published from the
  module as importable content, sourced from the single existing file (no second copy).
- **FR-013**: Hub URL configuration discovery and output rendering MUST stay internal to the
  CLI and MUST NOT be made public by this feature.
- **FR-014**: After merge, a release tag containing the public client (and published spec)
  MUST be created; published tags MUST never be moved or re-created.
- **FR-015**: Messages produced by the public client (`ClassifyError` and the `Error()`
  methods of its error types) MUST NOT mention CLI flags, environment variables, config
  files, exit codes or "the CLI". Where the CLI shows such hints today, the CLI MUST add
  them itself so its output stays identical (FR-010).

### Key Entities

- **Public hub client**: the importable Multiroom Audio Hub API client — operations,
  request/response types, error types and error classes. Its exported surface is a
  semver-governed contract.
- **Error class**: the category of a hub failure (e.g. network, timeout, not found, API
  error) with a friendly message; shared by all consumers.
- **Exit-code mapping**: CLI-only translation of an error class into a process exit code.
- **Published hub spec**: the contents of `api/openapi.json`, importable at the same version
  as the client.
- **Release tag**: an immutable module version consumers pin to.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A project outside this repository can depend on a released tag and complete
  one read call, one write call and one error classification against a fake hub, with zero
  workarounds (no replace directives, workspace files or private-module settings).
- **SC-002**: 100% of existing tests pass on both Windows and Linux after the refactor, with
  no test weakened. Tests that assert exit codes through the client move to the CLI-internal
  mapping with the same assertions; tests of hub message text change only where FR-015
  removes CLI wording, and the CLI tests that check that wording stay unchanged.
- **SC-003**: For every existing command, output and exit code are identical before and
  after the refactor across the success path and each error class. Evidence: the existing
  unit, integration and contract suites pass unchanged in intent (they cover every command
  and error class), supplemented by a manual before/after comparison of representative
  commands.
- **SC-004**: 0 references to the old module path or the old internal client location remain
  in code, tests, build/release scripts or current docs (README, CONTRIBUTING, AGENTS.md).
  Historical records — earlier feature specs under `specs/`, `docs/reviews/`, and the
  constitution's Sync Impact Report in `.specify/` — are excluded.
- **SC-005**: 100% of the public client's exported identifiers have documentation, and 0 of
  those docs reference internal spec artifacts or constitution principles.
- **SC-006**: The public client pulls in 0 third-party dependencies.
- **SC-007**: A build with a supplied version reports that exact version in 100% of release
  builds.

## Assumptions

- **Dependency**: The sonora-cli constitution is amended first (Phase 1 of the source plan,
  1.1.3 → 1.2.0) to add the "Public Hub Client Package" rules and to name `hub/` in
  Principle II as the single place the spec becomes Go types. That amendment is done via
  `/speckit-constitution`, not by this feature, and must land before or with this feature's
  merge.
- The repository at `github.com/Sonora-Multiroom/sonora-cli` is public, so consumers need no
  private-module configuration.
- While the module is `v0.x`, breaking changes to the public client are allowed but must be
  called out in release notes; the first release with the public client is `v0.1.0` unless
  an existing `v0.x` tag sequence dictates the next patch/minor number.
- Publishing the spec (FR-012 / US3) is recommended by the source plan; it is in scope but
  may be dropped at planning time without blocking US1/US2.
- Counts from the source plan were verified against the repo on 2026-09-27 and bound the
  scope of mechanical changes: 111 files import the module path and 77 import the internal
  client; 30 CLI files use the exit-code mapping (the source plan said 34).
- Out of scope: the sonora-mcp rewrite, new hub operations, new CLI commands, making
  config discovery public, and splitting the client into its own repository (option 2b).
