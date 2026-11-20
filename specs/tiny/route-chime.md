# TinySpec: a chime before a route (HUB-040, API 0.1.34)

**Branch**: feature/route-chime
**Date**: 2026-11-20
**Status**: implemented, tag pending
**Complexity**: small (top of the range: ~12 files, all additive)

## What

Hub 0.1.34 lets a route start with a short sound. Bring `api/openapi.json` and the `hub` package up to that
contract, add `--chime` to `sonora play` and `sonora route`, `sonora get chimes`, and show a set chime in
`get outputs`, `get groups` and `get inputs`. The `sonora-mcp` bump follows the tag and is not part of this.
Source: backlog `adopt-route-chime.md`; hub `specs/040-route-chime/contracts/chime-api.md`.

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Re-fetched with `scripts/update-openapi.sh` (0.1.33 → 0.1.35: the live hub is newer than 0.1.34) |
| `hub/play.go`, `hub/routes.go` | `PlaybackRequest` and `CreateRouteRequest` gain `Chime *string` |
| `hub/inputs.go` | `Input.DefaultChime`; `CreateInputRequest.DefaultChime` (hub field only, no CLI flag) |
| `hub/outputs.go`, `hub/groups.go` | `Output.Chime`, `Group.Chime` |
| `hub/chimes.go` | New: `Chime`, `ListChimes`, `ChimeNone` |
| `internal/cli/play/play.go`, `internal/cli/route/route.go` | `--chime` flag |
| `internal/cli/chimes/get.go` | New: `sonora get chimes`, modelled on `internal/cli/hubinfo/info.go` |
| `cmd/sonora/main.go` | `dispatchGetList`: route `chimes` like `info` (`list chimes` is a usage error) |
| `internal/render/{outputs,groups,inputs,chimes}.go` | Chime keys, only when set; new chimes renderer |
| `tests/contract/`, `tests/unit/` | New and extended tests |
| `README.md`, `docs/backlog/` | Document the flag and command; trim the backlog item |

## Requirements

1. `api/openapi.json` is the hub's 0.1.34 document, LF endings; `go test ./tests/unit/` (`api.Spec`) passes.
2. `Chime` is sent only when non-nil, on play, route and (hub package) input creation. Absent decodes to
   nil on `Output`, `Group` and `Input`; `none` is a value, never nil. JSON views omit a nil chime.
3. `ListChimes(ctx, client, baseURL) ([]Chime, error)` calls `GET /api/v2/chimes`; `Chime` has `ID` and
   `DurationMillis`; an empty list is `[]`, not an error; a malformed body is `*DecodeError`. Godoc stands
   without this repo and says `none` is reserved and never listed.
4. `--chime <id|none>` on play and route sends the value verbatim (`none` is matched exactly, lower case).
   A blank value is a usage error (exit 2). Without the flag nothing is sent.
5. **Open question 1, decided:** an unknown id is not checked client-side. The hub's 400
   (`Unknown chime: <id>`) is reported through the existing `APIError` path, so the flag costs no extra
   request and cannot disagree with the hub.
6. `sonora get chimes` prints `chimes:` with `id` and `durationMillis` per entry (YAML default, `--json`);
   an empty list prints `# no chimes found` and `chimes: []`, as inputs do.
7. `get outputs[/id]`, `get groups[/id]` print `chime:` and `get inputs[/id]` prints `defaultChime:` only
   when set, in YAML and JSON; output with no chime is byte-identical to today.

## Plan

1. `scripts/update-openapi.sh`; check the diff covers only HUB-040's additions.
2. Tests first (contract for the hub calls, unit for flags and renderers), then `hub/`, then the CLI.
3. Document in `README.md`; delete the CLI half of `adopt-route-chime.md`, leaving the `sonora-mcp` bump,
   and sync `docs/backlog/INDEX.md`.

## Tasks

- [x] Refresh `api/openapi.json` (the live hub serves 0.1.35)
- [x] Contract tests: `chime` on play and route bodies (set, `none`, omitted); `ListChimes` (200, empty, malformed, path and method)
- [x] Contract tests: `chime` / `defaultChime` decoded on output, group, input (set, `none`, absent) and sent on input creation
- [x] Unit tests: `--chime` (value, `none`, blank, absent) on play and route; 400 `Unknown chime` reaches stderr
- [x] Unit tests: renderers show the keys only when set (YAML and JSON); `get chimes` YAML, JSON, empty; `list chimes` rejected
- [x] Implement `hub/` changes and `hub/chimes.go`
- [x] Implement `--chime` in play and route; `internal/cli/chimes`; dispatch in `main.go`; renderers
- [x] README and backlog updates
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...`

## Done When

- [x] All tasks checked off; tests pass; no vet or gofmt findings; `git ls-files --eol` shows no `w/crlf`
- [x] Read-only live check on `multiroom.lan`: `sonora get chimes` lists the hub's chimes
- [ ] Tag released (additive, no breaking change); `sonora-mcp` bump is a separate item

## Verification (2026-11-20)

- `scripts/update-openapi.sh`: 0.1.33 → 0.1.35, because `multiroom.lan` already runs 0.1.35. The diff is
  wider than HUB-040: it also holds `PUT /api/v2/inputs/{inputId}/persistent`, `InputResponse`
  `origin`/`playable`/`unplayableReason`, `CreateInputRequest.persistent`, `SpeakRequest.chime` and the
  `radio` extension's `/api/radio/*` paths. None of that is adopted here; `hub.SpeakRequest` and `Input`
  lag the schema as before. Not `SpeakRequest.chime` either: `speak` is TTS, outside this item.
- New tests failed to compile before the `hub` changes (`undefined: hub.ChimeNone`, `unknown field Chime`).
- `gofmt -l .`: no output. `go vet ./...`: exit 0. `go test ./...`: exit 0. No `w/crlf` file.
- Live, `multiroom.lan` hub 0.1.35, read-only, built binary: `get chimes` (YAML and `--json`) lists
  `cabin`, `tritone`, `bingbong`, `alert` with the hub's durations; `get outputs|groups|inputs` print no
  chime key because none is set there. No route was started, so `--chime` was tested against mock hubs only.
