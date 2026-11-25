# Agent Instructions

## Product naming

- The product is **Sonora Multiroom**. This CLI's full name is **Sonora Multiroom CLI**.
- "Sonora CLI" is informal shorthand — do not use it in specs, docs, or headings; use
  "Sonora Multiroom CLI" (or "Sonora Multiroom" when referring to the product/hub as a
  whole) instead.
- The invoked binary/command name is `sonora` (e.g. `sonora get outputs`) and is
  unaffected by this — only prose naming changes.
- The backend service this CLI talks to is the **Multiroom Audio Hub API**
  (`api/openapi.json`), a separate component from the CLI itself.

## CLI flag conventions

- Strict GNU/POSIX style (like `kubectl`/`gh`): long flags require `--` (e.g.
  `--version`, `--help`), short flags are single-letter with `-` (e.g. `-v`, `-h`).
  Single-dash multi-letter forms (`-version`, `-help`) are invalid and must not be
  added.

## Live testing

- Piper (`--provider piper-local`) on the Pi takes about 19 s to synthesize an uncached phrase, longer
  than the 15 s default bound. Pass `--timeout 60s` to `sonora speak` and `sonora warm tts-cache` when
  testing live; on a timeout the hub still stores the audio, so a retry is a cache hit.

## Git / PR conventions

- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
  — see [CONTRIBUTING.md](CONTRIBUTING.md) for the types this repo uses.
- Never add a `Co-Authored-By` trailer (or similar attribution) to git commit messages.
- Never add "Generated with Claude Code" (or similar attribution) to pull request
  descriptions.

## Project context

- Governance and non-negotiable engineering principles live in
  `.specify/memory/constitution.md` — read it before planning or implementing.
- API contract fidelity: all HTTP request/response handling MUST be validated against
  `api/openapi.json`, the single source of truth for the hub API.
- Feature work follows the Spec Kit workflow (`specs/<NNN-feature>/spec.md` →
  `plan.md` → `tasks.md`), driven by the `/speckit-*` skills.

## Specification Sources (Speckit)

### `specs/` — Active Feature Specs
Contains specs for features currently in development that have not yet been archived.
**When to use:**
- When working on a feature that is actively being built or changed.
- When no entry exists yet in `.specify/memory/`.
**When NOT to use:**
- Do not treat these as final — they may be drafts or works in progress.
**Not yet archived:** small changes in `specs/tiny/`.
**Archived tinyspecs** live in `.specify/archive/tiny/<date>-<name>/`, indexed by folder name; check there before concluding a small change has no spec.

### `.specify/memory/` — Consolidated Specs (primary reference)
Distilled, up-to-date specs summarising the current understanding of each feature or domain.
**When to use:**
- Before starting any task — check here first.
- As the authoritative source of truth for current requirements.
- When writing, reviewing, or refactoring code to verify alignment with defined behaviour.
- When resolving ambiguity about how something is supposed to work.
**Files:**
- `.specify/memory/spec.md` — main specification; `plan.md` — main implementation plan
- `.specify/memory/constitution.md` — principles and mandates
- `.specify/memory/changelog.md` — log of merged features

### `.specify/archive/` — Archived Feature Specs (historical reference)
Retired or superseded specs kept for historical context and decision archaeology.
**When to use:**
- When a consolidated spec in `.specify/memory/` lacks rationale or context.
- When investigating why a past decision was made.
- When a feature was deprecated and you need its original intent.
- As a fallback if no spec exists in `specs/` or `.specify/memory/`.
**Rules:**
- Treat as read-only historical context, not current requirements.
- If an archived spec contradicts `.specify/memory/`, the memory spec wins.
- If an archived spec should be promoted back to active, flag it to the team — do not act on it directly.
**Index** (one line per archived feature; open `.specify/archive/<folder>/spec.md` for the behaviour and the reasoning):

- `001-list-outputs` — `sonora outputs list`: enabled outputs, `--include-disabled`, `--json`, exit codes, 5s timeout
- `002-outputs-get` — `sonora outputs get <output-id>`: one output, `--json`, not-found exit code 5
- `003-inputs-list-get` — `sonora inputs list` and `sonora inputs get <input-id>`: inputs, `--include-disabled`, `--json`, not-found exit code 5
- `004-routes-list-get` — `sonora routes list` and `sonora routes get <route-id>`: routes, `--status`/`--input-id`/`--target-id`, `--json`, not-found exit code 5
- `005-groups-list-get` — `sonora groups list` and `sonora groups get <group-id>`: output groups, `--include-disabled`, `--json`, not-found exit code 5
- `006-play-command` — `sonora play <uri> <target-id>`: one playback request, `--group`/`--output`, `--volume`, `--name`, `--json`; exit codes 6–10 for the play failure classes
- `007-refactor-cli-commands` — verb-first grammar: `sonora get <resource>[/<id>]`, `sonora list <resource>`, aliases `in`/`out`/`gr`/`rt`; `play <uri> <outputs|groups>/<id>` with `--display-name`; old noun-verb forms and `--group`/`--output`/`--name` removed; exit code 7 retired
- `008-route-command` — `sonora route inputs/<input-id> <outputs|groups>/<target-id>`: connects an existing input to one output or group; exit codes 11 (input not found) and 12 (target not found)
- `009-tts-commands` — `sonora speak <text|-> <outputs|groups>/<id>`, `sonora get tts-cache`, `sonora clear tts-cache`: TTS announcements and cache, `--provider`/`--voice`/`--language`/`--timeout`, exit code 13 (TTS not available)
- `010-public-hub-package` — public Go client package `hub/` (importable by other Sonora projects) and `api.Spec` (the embedded `api/openapi.json`); exit-code mapping and the CLI-only usage class moved to `internal/cli/exitcode`; no CLI behavior change
- `011-hub-0128-alignment` — hub 0.1.25–0.1.28 alignment: `enable`/`disable` help and README describe the effect on running playback; `sonora play` shows the hub's `displayName`; `sonora transfer` retries the transient `ROUTE_CHANGING_OUTPUTS` 409 (`hub.IsTransientConflict`), then exits 14
- `012-join-modes` — `sonora play|route --join-mode`, `sonora create inputs --default-join-mode`, `sonora get outputs/<id> --include-routes` (a second request; an older hub exits 3); join modes and outputs shown in the YAML of `get routes/<id>` and `get inputs/<id>`
- `013-info-and-tts-warm` — `sonora get info` (hub info: product, version, apiVersion, startedAt, uptimeSeconds; a 404 means a hub older than 0.1.33) and `sonora warm tts-cache <text|->` (`--provider`/`--voice`/`--language`/`--timeout`; prints `cacheHit`, plays nothing)
- `014-radio-station-search` — `sonora search radio-stations [name]` (`--tag`, `--country`, `--language`, `--codec`, `--min-bitrate`, `--order`, `--asc`/`--desc`, `--limit`, `--offset`), `sonora get radio-stations/<uuid>`, `sonora get|list radio-countries|radio-languages|radio-tags`: radio station search through the hub's radio extension; exit code `15` means radio is not available

## Active Technologies

- Go 1.27.0, standard library only (`net/http`, `encoding/json`, `flag`, `net/url`)

## Project Structure

- `cmd/sonora` entrypoint; public `hub/` (the hub client, 010) and `api/` (`api.Spec`); `internal/{config,cli,render}`; `internal/cli/respath` parses resource paths (`outputs/<id>`, aliases); `internal/cli/route` (`sonora route`) creates a route from an existing input; `tests/{unit,contract,integration}`
- `internal/cli/tts` (009): `speak`, `get tts-cache`, `clear tts-cache`; `hub/tts.go` and `hub/extensions.go` (`listExtensions`, used only for the 404 diagnosis); `internal/render/tts.go`. The TTS commands share one failure path (`report.go`) because the extension's error shape differs from the core API's.
- `internal/cli/joinmode` (012): the `--join-mode` and `--default-join-mode` check shared by `play`, `route` and `create inputs`; `hub/outputs.go` holds `RouteOnOutput` and `GetRoutesForOutput`
- `internal/cli/hubinfo` (013): `get info` (`RunGet`), with `hub/info.go` holding `GetInfo`; `internal/cli/tts/warm.go` (`RunWarm`) shares the TTS failure path with `speak`
- `internal/cli/radio` (014): `RunSearch`, `RunGetStation`, `RunCountries`, `RunLanguages`, `RunTags`, `ReportError`; `hub/radio.go` holds the five radio calls and `internal/render/radio.go` the output

## Commands

- `gofmt -l .` and `go vet ./...` baseline checks; `go test ./...`

## Recent Changes

Nothing is in progress: `specs/` holds no unfinished feature or tinyspec. Shipped work is folded into
`.specify/memory/` and listed per feature in `.specify/memory/changelog.md`; this section keeps only
work that is not archived yet.

## Known Issues & Gotchas

### ⚠️ Flag re-parse keeps reading flags after `--`
**Issue:** `play`'s re-parse loop calls `fs.Parse` again on the arguments after a `--` terminator, so free text beginning with `-` is parsed as a flag. `speak` (009) stops parsing after `--`; `play` still has the quirk.
**Root Cause:** The loop re-enters `fs.Parse` for every remaining argument, and `flag` honours `--` only within the call that consumed it.
**Prevention Rule:** Once `--` has been consumed, treat every remaining argument as positional and stop calling `fs.Parse`, detected by checking whether the slice passed to `fs.Parse` contained `--` before its first returned positional. Apply it to any command that takes free text.

### ⚠️ Verb-less noun swallows its first argument
**Issue:** `sonora play <uri> <outputs|groups>/<id>` has no verb token after `play`, so a generic `verb, resource, rest` split would bind `<uri>` to the resource and drop it before dispatch; `run()` special-cases `play` first.
**Root Cause:** The `len(args) < 2` gate and the noun/verb split assume every command has a verb; the one-operation `play` noun does not.
**Prevention Rule:** Route a verb-less noun with an early special case in `cmd/sonora/main.go`, before the `len(args) < 2` gate and the noun/verb split, and let its own `Run` validate the positional count.

### ⚠️ Older hub answers 405, not 404, for an output's routes
**Issue:** `GET /api/v2/outputs/{id}/routes` answers 405 on a hub that has the `DELETE` on the same path but not the `GET` (hubs from 2026-09-11 to 0.1.20), so `--include-routes` must read 405 as "older hub", not only 404.
**Root Cause:** The hub's `V2ExceptionHandler` covers only `multiroom.rest.controllers.v2`. A request that matches the path but not the method never reaches a controller, so Spring's default resolver answers 405. Only a hub from before 2026-09-11 answers 404.
**Prevention Rule:** Read 404 and 405 from an optional list request sent after its parent request succeeded as "this hub cannot list it", and exit with the unexpected-response code, not the not-found code.

### ⚠️ `RenderInputYAML` is shared by four commands
**Issue:** `get`, `create`, `rename` and `enable|disable inputs/<id>` all print one input record through `RenderInputYAML`, so a field added there appears in all four.
**Root Cause:** One renderer prints the input record for every command that returns it; the JSON output already prints `hub.Input` unchanged.
**Prevention Rule:** Change an input record's YAML in `internal/render/inputs.go` only, and check the golden output of all four commands.

### ⚠️ Radio requests need a longer client timeout than the standard 5 s
**Issue:** a radio call bounded at the standard 5 s races the hub's own 5 s catalogue budget, so the user sees "hub did not respond in time" instead of the hub's 503 with its retry delay.
**Root Cause:** the hub bounds one catalogue request at 5 s by default and answers 503 `catalogue-unavailable` at that point; a client bound of the same length usually fires first.
**Prevention Rule:** every radio call uses `hub.RadioTimeout` (10 s), never the standard client bound; a new radio call gets the same timeout.
