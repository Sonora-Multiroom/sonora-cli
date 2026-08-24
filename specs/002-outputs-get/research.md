# Research: Get Single Audio Output

**Feature**: `002-outputs-get` | **Date**: 2026-08-25

This document resolves every open technical question for implementing `sonora outputs get
<output-id>` before design (Phase 1) begins. The spec has no unresolved `NEEDS
CLARIFICATION` markers — this feature deliberately reuses every foundational decision
`001-list-outputs` already made (language, dependency policy, HTTP client construction, hub
URL resolution, YAML/JSON rendering approach, `--verbose`/error-translation shape, testing
strategy). Only the decisions genuinely new to this feature are recorded below; see
[../001-list-outputs/research.md](../001-list-outputs/research.md) for the rest, which apply
unchanged.

## 1. HTTP call: `GET /api/v2/outputs/{outputId}` (operationId `getOutput`)

- **Decision**: Add `hub.GetOutput(ctx, client, baseURL, outputID string) (*Output, error)`
  in `internal/hub/outputs.go`, alongside the existing `ListOutputs`. It builds
  `{baseURL}/api/v2/outputs/{outputId}` (path-escaped), issues a single GET through the
  same `*http.Client` returned by `hub.NewClient()`, and decodes the body into the existing
  `Output` struct (`#/components/schemas/OutputResponse` — identical schema to the list
  endpoint's items, confirmed in `api/openapi.json`). No new response type is needed.
- **Rationale**: `getOutput`'s 200 response is the same `OutputResponse` schema `ListOutputs`
  already decodes into (Principle II); reusing `Output` avoids a duplicate, drift-prone
  struct. Splitting the call into its own function (rather than parameterizing
  `ListOutputs`) matches the one-function-per-operation shape `001-list-outputs` already
  established.
- **Alternatives considered**: Generalizing `ListOutputs` to also fetch a single output —
  rejected: the two operations have different paths, different response shapes (array vs.
  single object), and different error semantics (404 is meaningful only for `getOutput`);
  conflating them would produce a confusing signature.

## 2. The 404 "not found" case needs its own exit-code class (FR-012)

- **Decision**: Add a `NotFoundError{OutputID string}` type in `internal/hub/errors.go`,
  distinct from the existing `StatusError`. `GetOutput` returns `&NotFoundError{OutputID:
  outputID}` specifically when the hub responds `404`, and `&StatusError{...}` for any other
  non-2xx status (consistent with `ListOutputs`'s existing handling of non-2xx as a hub
  error). Add a fourth `ErrorClass`, `ClassNotFound`, with **exit code `5`**.
  `ClassifyError` gains a branch: `errors.As(err, &notFoundErr)` → `ClassNotFound`, message
  `"output not found: <id>"`.
- **Rationale**: FR-012 requires success / usage error / not-found / hub error / network
  error to all be programmatically distinguishable — five classes, not four. `001-list-
  outputs`'s research.md §6 fixed `0/2/3/4` for its four classes; `outputs list` has no
  "not found" concept (an empty result there is still success, per its FR-012), so those
  meanings are unchanged and this feature only *appends* `5` rather than renumbering
  anything — `outputs list`'s existing exit-code contract and tests stay valid untouched.
- **Alternatives considered**:
  - Reusing exit code `3` (hub error) for not-found — rejected: FR-012 explicitly requires
    "not found" to be distinguishable from "hub reported a problem" (a 5xx, or a malformed
    body), which is exactly FR-011's distinction; collapsing them back together would fail
    both requirements.
  - Renumbering the whole scheme (e.g. `0/2/3/4/5` → some other ordering so "not found"
    sits next to "hub error") — rejected: `outputs list` already ships with `3`=hub,
    `4`=network; changing those meanings would silently break that command's documented
    contract and passing tests for no benefit — appending `5` is strictly additive.

## 3. CLI dispatch: two verbs under one noun package

- **Decision**: Rename the existing exported `outputs.Run` (in
  `internal/cli/outputs/list.go`) to `outputs.RunList`, and add `outputs.RunGet` in a new
  `internal/cli/outputs/get.go` in the same package. Update `cmd/sonora/main.go`'s `verb`
  switch to route `"list"` → `outputs.RunList` and `"get"` → `outputs.RunGet`. Update the
  two existing call sites that reference `outputs.Run` by name
  (`tests/unit/cli_outputs_test.go`, `tests/integration/outputs_list_test.go`) to
  `outputs.RunList` — a pure rename, no behavior change.
- **Rationale**: `main.go` already groups commands as `<noun> <verb>` (constitution
  Principle V) with one package per noun (`internal/cli/outputs`). A second verb in that
  package needs a distinct exported name; `Run` was only unambiguous while `list` was the
  sole verb. `RunList`/`RunGet` keeps both names self-describing at the call site in
  `main.go`, which is where `outputs.Run` used to read ambiguously once a second verb
  exists.
- **Alternatives considered**: A single dispatching `outputs.Run(verb string, args []string,
  ...)` — rejected: `flag.NewFlagSet` per verb already needs verb-specific flag sets (`get`
  takes a positional ID and no `--include-disabled`; `list` takes no positional and has
  `--include-disabled`), so a shared entry point would just re-dispatch internally with no
  real benefit over two top-level functions `main.go` calls directly.

## 4. Output identifier is a required positional argument, not a flag

- **Decision**: `sonora outputs get <output-id> [--json] [--verbose] [--hub-url URL]` — the
  identifier is `fs.Arg(0)` after `flag.Parse()`, not a named flag. Exactly one positional
  argument is required: zero args ("missing identifier") or more than one ("unexpected
  argument(s)") are both usage errors (exit `2`), mirroring the existing "unexpected
  argument(s)" usage-error path `outputs list` already has for its (zero-positional-arg)
  case.
- **Rationale**: FR-002 requires "exactly one output identifier as input" with a clear usage
  error if omitted (Edge Cases: "no identifier provided"). A positional argument is the
  conventional shape for a single required identifier in `<noun> <verb> <id>` CLIs (e.g.
  `git show <ref>`, `kubectl get pod <name>`), consistent with Principle V's "predictable,
  consistent structure."
- **Alternatives considered**: `--output-id` as a required flag — rejected: heavier syntax
  for a single mandatory value with no default, and inconsistent with the "one clear way per
  concept" reading of Principle V's flag-consistency rule; the flags this command *does*
  share with `outputs list` (`--json`, `--verbose`, `--hub-url`) keep their exact names.

## 5. Single-output rendering (FR-004, FR-005, FR-007)

- **Decision**: Add `render.RenderOutputYAML(o hub.Output) string` and
  `render.RenderOutputJSON(o hub.Output) string` in `internal/render/outputs.go`, alongside
  the existing list renderers. Both emit the same six fields as the list renderers'
  per-record shape, but as a single top-level record — not wrapped in an `outputs:`/
  `{"outputs": [...]}` list — since exactly one output is ever returned. Every field
  (including `available: false`) is always emitted explicitly, reusing the same
  never-omit-availability rule `RenderYAML` already established for FR-005.
- **Rationale**: FR-004 requires the same field set as the list command's per-output view;
  the spec Assumptions state the human-readable default "presents the same set of fields as
  the machine-readable output... consistent with `outputs list`." A bare single-record
  document (no list wrapper) most directly represents "one output," and keeps the
  human/`--json` views trivially parallel to each other, mirroring `RenderYAML`/`RenderJSON`'s
  existing relationship.
- **Alternatives considered**: Reusing `RenderYAML([]Output{o})`/`RenderJSON([]Output{o})`
  as-is (a one-element list) — rejected: it would present a single lookup result as a
  collection, which is a worse shape for both a human skimming one output's state and a
  script that (per SC-005) shouldn't need to index into an array to read one record.

## 6. Testing strategy — same three layers, extended for `get`

- **Decision**: Add `tests/unit/render_output_get_test.go` (single-output YAML/JSON
  rendering, including the unavailable-output distinguishability case), extend
  `tests/contract/` with `outputs_get_test.go` (mock server serving both the `200
  OutputResponse` and `404 ErrorResponse` shapes from `api/openapi.json`'s `getOutput`
  operation, asserting request path and both decode paths), and extend `tests/integration/`
  with `outputs_get_test.go` (full `sonora outputs get <id>` process invocation: found,
  not-found, missing-identifier usage error, `--json`, exit codes). All written before their
  corresponding implementation code, per Principle VI, exactly as `001-list-outputs`
  established.
- **Rationale**: Same rationale as `001-list-outputs` research.md §8 — this reuses that
  three-layer taxonomy rather than inventing a new one. `outputs get` is not itself in
  Principle VI's named "critical" flows (playback/routing/volume/mute), but it shares the
  exact HTTP-client/decode/render skeleton those flows depend on, so validating it here is
  equally cheap and high-leverage.
- **Alternatives considered**: None specific to this feature — the layer choice was already
  settled by `001-list-outputs`.

## Related specs (from memory)

- _none found_ (the `memsearch:memory-recall` tool was unavailable in this environment;
  per `fail_open`, no memory-tool search was performed. The directly relevant sibling spec,
  `001-list-outputs`, is already cited throughout this document.)
