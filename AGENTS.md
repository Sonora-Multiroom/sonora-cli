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
**Not yet archived:** `010-public-hub-package` through `014-radio-station-search`; small changes in `specs/tiny/`.
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

## Active Technologies

- Go 1.27.0, standard library only (`net/http`, `encoding/json`, `flag`, `net/url`) (.specify/archive/001-list-outputs, .specify/archive/002-outputs-get)

## Project Structure

- `cmd/sonora` entrypoint; `internal/{hub,config,cli,render}`; `internal/cli/respath` parses resource paths (`outputs/<id>`, aliases); `internal/cli/route` (`sonora route`) creates a route from an existing input; `tests/{unit,contract,integration}` (.specify/archive/001-list-outputs, .specify/archive/007-refactor-cli-commands, .specify/archive/008-route-command)
- `internal/cli/tts` (009): `speak`, `get tts-cache`, `clear tts-cache`; `internal/hub/tts.go` and `internal/hub/extensions.go` (`listExtensions`, used only for the 404 diagnosis); `internal/render/tts.go`. The TTS commands share one failure path (`report.go`) because the extension's error shape differs from the core API's.

## Commands

- `gofmt -l .` and `go vet ./...` baseline checks; `go test ./...` (.specify/archive/001-list-outputs)

## Recent Changes

- .specify/archive/009-tts-commands: `sonora speak <text|-> <outputs|groups>/<id>` (`--provider`, `--voice`, `--language`, `--timeout`, `--json`, `--verbose`, `--hub-url`) announces text through the hub's TTS extension (`POST /api/tts/speak`), returning on hub acceptance; `sonora get tts-cache` (`GET /api/tts/cache/stats`) and `sonora clear tts-cache [--provider]` (`DELETE /api/tts/cache`); exit code `13` for TTS not available, with a single `listExtensions` lookup on a 404
- .specify/archive/008-route-command: `sonora route inputs/<input-id> <outputs|groups>/<target-id>` routes an existing input to one output or group via one `POST /api/v2/routes` call, creating no input; input and target are checked first (`GET /api/v2/inputs/{id}`, then the target type's own endpoint); exit codes `11` (input not found) and `12` (target not found) added, route creation failed reuses `8`
- .specify/archive/007-refactor-cli-commands: verb-first grammar — `sonora get <resource>[/<id>]` and `sonora list <resource>` (inputs, outputs, groups, routes; aliases `in`, `out`, `gr`, `rt`) replace `sonora <resource> list|get`; `sonora play <uri> <outputs|groups>/<id>` with `--display-name` replaces `<target-id>` plus `--group`/`--output`/`--name`; old forms fail with a usage error (exit `2`); exit code `7` (ambiguous target) retired
- .specify/archive/006-play-command: `sonora play <uri> <target-id>` (`--group`, `--output`, `--volume`, `--name`, `--json`, `--verbose`, `--hub-url`); one `POST /api/v2/play` call, returns on hub acceptance without polling; exit codes 6–10 for validation, ambiguous target, route creation failed, source unreachable and service unavailable; target resolved via `GET /api/v2/outputs/{id}` and `GET /api/v2/groups/{id}`
- .specify/archive/005-groups-list-get: `sonora groups list` (`--include-disabled`, `--json`, `--verbose`, `--hub-url`) and `sonora groups get <group-id>` (`--json`, `--verbose`, `--hub-url`); 404 maps to exit code 5, shared with `outputs get`, `inputs get` and `routes get`
- .specify/archive/004-routes-list-get: `sonora routes list` (`--status`, `--input-id`, `--target-id`, `--json`, `--verbose`, `--hub-url`) and `sonora routes get <route-id>` (`--json`, `--verbose`, `--hub-url`); 404 maps to exit code 5, shared with `outputs get` and `inputs get`
- .specify/archive/003-inputs-list-get: `sonora inputs list` (`--include-disabled`, `--json`, `--verbose`, `--hub-url`) and `sonora inputs get <input-id>` (`--json`, `--verbose`, `--hub-url`); 404 maps to exit code 5, shared with `outputs get`
- .specify/archive/002-outputs-get: `sonora outputs get <output-id>` (`--json`, `--verbose`, `--hub-url`); 404 maps to exit code 5; `outputs.Run` renamed `RunList`
- .specify/archive/001-list-outputs: `sonora outputs list` (`--include-disabled`, `--json`, `--verbose`, `--hub-url`); exit codes 0/2/3/4; 5s single-attempt timeout

## Known Issues & Gotchas

### ⚠️ Flag re-parse keeps reading flags after `--`
**Issue:** `play`'s re-parse loop calls `fs.Parse` again on the arguments after a `--` terminator, so free text beginning with `-` is parsed as a flag. `speak` (009) stops parsing after `--`; `play` still has the quirk.
**Root Cause:** The loop re-enters `fs.Parse` for every remaining argument, and `flag` honours `--` only within the call that consumed it.
**Prevention Rule:** Once `--` has been consumed, treat every remaining argument as positional and stop calling `fs.Parse`, detected by checking whether the slice passed to `fs.Parse` contained `--` before its first returned positional. Apply it to any command that takes free text.

### ⚠️ Verb-less noun swallows its first argument
**Issue:** `sonora play <uri> <outputs|groups>/<id>` has no verb token after `play`, so a generic `verb, resource, rest` split would bind `<uri>` to the resource and drop it before dispatch; `run()` special-cases `play` first.
**Root Cause:** The `len(args) < 2` gate and the noun/verb split assume every command has a verb; the one-operation `play` noun does not.
**Prevention Rule:** Route a verb-less noun with an early special case in `cmd/sonora/main.go`, before the `len(args) < 2` gate and the noun/verb split, and let its own `Run` validate the positional count.
