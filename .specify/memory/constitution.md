<!--
Sync Impact Report
==================
Version change: 1.1.2 → 1.1.3
Modified principles:
  - V. CLI UX Consistency — the illustrative command-shape example was updated from
    `sonora <noun> <verb> [args]` to `sonora <verb> <resource>[/<id>] [args]`, and the
    "verb first, resource second" ordering is now stated explicitly. The principle's
    normative content is unchanged: commands MUST still follow one consistent, predictable
    structure across the tool. Only the example naming that structure was stale — it named
    the noun-first grammar that feature 007-refactor-cli-commands removes and that the
    already-specified `route` command (008-route-command) never used.
Added principles: none
Removed sections: none
Other changes:
  - Performance Standards, "No hidden I/O" bullet: the example invocation
    `sonora route list` was corrected to `sonora get routes`. Same staleness class as the
    Principle V example — an illustrative command in the removed grammar (and one that was
    never valid even under it, which used `sonora routes list`). No normative change.
Deferred / TODO items: none
Rationale for PATCH bump: both edits replace illustrative examples only. No principle was
  added, removed, or redefined, and no MUST/SHOULD obligation changed in scope or strength,
  which is the definition of PATCH under this document's own versioning policy.
Templates requiring follow-up: none. Downstream templates (plan/spec/tasks) read this file
  at runtime and need no edits here.
Follow-up for in-flight work: specs/007-refactor-cli-commands/plan.md's Constitution Check
  row for Principle V (and research.md §6) describe this example as stale and defer the
  amendment. That amendment is now made; those two notes may be simplified to a plain
  "Pass" on the next edit of those files, but they are not incorrect as written.
-->

# Sonora Multiroom CLI Constitution

## Core Principles

### I. Instant Startup & Low-Latency Execution (NON-NEGOTIABLE)
The CLI MUST feel instant: from process start to dispatching the first HTTP request, there
MUST be no perceptible lag for an interactive user. Concretely:
- No blocking work (network calls, disk scans, config discovery beyond the invoked command's
  needs, plugin/dynamic loading) MAY occur before argument parsing completes for the
  requested command.
- CLI framework and dependency choices MUST be justified against their startup cost.
  Reflection-heavy or DI-container-style libraries are disallowed unless a benchmark proves
  negligible overhead versus a stdlib-based alternative.
- Any change that adds work to the startup path MUST include a benchmark
  (`go test -bench`) or measured timing showing the added latency, and that latency MUST be
  justified in the PR description.
Rationale: this tool controls real-time audio (volume, mute, routing) interactively; a slow
launch or a delayed HTTP call directly degrades the user's live listening experience, which is
the tool's core value proposition.

### II. API Contract Fidelity (OpenAPI-Driven)
The Multiroom Audio Hub OpenAPI specification (`openapi.json`) is the single source of truth
for every REST interaction the CLI performs.
- Request and response types MUST be generated from, or explicitly validated against, the
  spec — hand-maintained structs that silently drift from the spec are a defect, not a
  shortcut.
- When the spec's API version changes, all affected endpoints in the CLI MUST be reconciled
  (regenerated/updated and re-tested) before merging further changes that touch them.
- The CLI MUST NOT invent behavior not described by the spec (undocumented endpoints,
  guessed fields) without first updating or confirming the spec with the service team.
Rationale: the hub API is under active development (currently v0.1.11); treating the spec as
authoritative is the only way to keep a fast-moving client correct without duplicating
service-side logic.

### III. Minimal, Justified Dependencies
Prefer the Go standard library (`net/http`, `encoding/json`, `flag`, etc.) over third-party
frameworks. A new dependency MAY be added only when it provides value that cannot be
reasonably achieved with the standard library, and its impact on binary size and startup time
MUST be noted in the PR that introduces it.
Rationale: every dependency is a tax on binary size, build time, and — per Principle I —
startup latency; keeping the dependency graph small keeps the tool fast and easy to audit.

### IV. Resilient, Transparent HTTP Client Behavior
All calls to the Multiroom Audio Hub API MUST have explicit connect and request timeouts —
unbounded waits are not permitted. Failures MUST be handled explicitly:
- API and network errors MUST be translated into clear, actionable messages for the user;
  raw Go errors or stack traces MUST NOT be the default output, but the underlying error
  MUST remain available via a `--verbose`/debug mode.
- The CLI MUST NOT panic on network failure, malformed API responses, or unreachable hosts;
  it MUST exit with a non-zero, distinguishable exit code instead.
- Where a session issues multiple requests, the HTTP client SHOULD reuse connections
  (keep-alive) rather than paying a new connection/TLS-handshake cost per call.
Rationale: a control tool that hangs or dies opaquely on a flaky local network is worse than
one that fails fast with a clear message — speed and clarity of failure matter as much as
speed of success.

### V. CLI UX Consistency
Commands MUST follow a consistent, predictable structure — verb first, resource second
(e.g. `sonora <verb> <resource>[/<id>] [args]`) — across the tool.
- Output MUST default to YAML for data-returning commands; a `--json` flag MUST be
  available to switch output to JSON for scripting and automation that requires it.
- Exit codes MUST distinguish between classes of failure at minimum: success (0), user/usage
  error, API error (e.g. hub returned 4xx/5xx), and network/connectivity error.
- Flag and argument names MUST be consistent across commands (no synonyms for the same
  concept in different subcommands).
Rationale: predictable CLI behavior is required both for the human operator adjusting audio
live and for scripts/automations that may drive the same tool.

### VI. Test-First Development (NON-NEGOTIABLE)
Tests MUST be written before implementation, for every command, API client method, and
bugfix. The cycle is strict: write the test → get it reviewed/approved → watch it fail →
implement the minimum code to pass → refactor (Red-Green-Refactor). Implementation code
written before its corresponding test exists MUST be rejected in review.
- Unit tests are required for command logic and the API client wrapper.
- Contract/integration tests (run against a mock server built from `openapi.json`) are
  required for endpoints critical to core flows (playback, routing, volume/mute) before the
  code that calls them is merged.
- A PR that adds behavior without a preceding failing test demonstrating that behavior was
  needed MUST NOT be merged.
Rationale: for a fast, always-on control tool, regressions in correctness are as damaging as
regressions in latency; TDD is the mechanism that keeps both the HTTP client's contract
handling (Principle II) and its failure handling (Principle IV) verifiably correct as the
tool evolves.

## Performance Standards

- **Startup budget**: cold start to first HTTP request dispatched MUST target well under
  50ms on typical developer/user hardware; regressions against this budget block merge
  unless explicitly justified.
- **Connection reuse**: within a single CLI invocation that issues multiple requests, the
  HTTP client MUST reuse the underlying transport/connection pool rather than creating a new
  client per call.
- **No hidden I/O**: config/credential loading MUST be lazy and scoped to what the invoked
  command actually needs — a `sonora --help` or `sonora get routes` MUST NOT pay the cost of
  unrelated subsystems initializing.
- Performance-sensitive changes (anything touching startup path, client construction, or
  request dispatch) SHOULD include a benchmark or measured before/after timing in the PR.

## Development Workflow

- Code MUST pass `gofmt`, `go vet`, and the project's configured linter (e.g.
  `golangci-lint`) before merge.
- All new command logic, API client methods, and bugfixes MUST follow the Test-First
  workflow in Principle VI — no exceptions. Endpoints critical to core flows (playback,
  routing, volume/mute) MUST have contract/integration tests against a mock server built
  from the OpenAPI spec before the calling code is merged.
- Every PR touching startup path, dependency list, or HTTP client construction MUST include
  a self-review against Principles I, III, IV, and VI before requesting merge.
- Changes to `openapi.json` (or its upstream source) MUST trigger a review of all CLI code
  paths that consume the changed endpoints, per Principle II.

## Governance

This constitution supersedes any conflicting informal practice. Amendments are made by
editing this file via the `/speckit-constitution` workflow and MUST include an updated Sync
Impact Report.

**Versioning policy** (semantic versioning applied to this document):
- **MAJOR**: backward-incompatible governance changes or removal/redefinition of a principle.
- **MINOR**: a new principle or materially expanded section is added.
- **PATCH**: clarifications, wording fixes, or non-semantic refinements.

**Compliance review**: every PR/spec/plan produced for this project MUST be checked against
these principles before merge; unresolved conflicts must be justified explicitly in the PR
description or resolved by amending this constitution first. Use this file as the runtime
guidance for planning and implementation commands (`/speckit-plan`, `/speckit-tasks`,
`/speckit-implement`).

**Version**: 1.1.3 | **Ratified**: 2026-08-24 | **Last Amended**: 2026-08-27
