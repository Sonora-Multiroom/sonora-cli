# Verification: Hub Info and TTS Cache Warm-up

## Baseline (2026-11-09, branch 013-info-and-tts-warm, before any change)

- `gofmt -l .`: no output
- `go vet ./...`: exit 0
- `go test ./...`: exit 0 (cmd/sonora, tests/contract, tests/integration, tests/unit all `ok`)

## Spec check (T002)

`api/openapi.json` declares `HubInfoResponse` (five optional properties, no `required`), `WarmRequest`
(`text` required plus seven optional fields) and `WarmResponse` (`cacheHit`, no `required`). No difference
from `data-model.md`.

## Implementation checks (2026-11-09)

- Tests were written first and seen to fail (build errors for the missing `hub.InfoNotOfferedError`,
  `hub.GetInfo`, `hub.WarmTTSCache`, `render.RenderInfo*`, `render.RenderWarm*`) before each implementation.
- `gofmt -l .`: no output
- `go vet ./...`: exit 0
- `go test -count=1 ./...`: exit 0 (cmd/sonora, tests/contract, tests/integration, tests/unit all `ok`)
- `git ls-files --eol | grep -c w/crlf`: 0; untracked new files are LF too.
- `tests/unit/hub_godoc_test.go` enumerates exported identifiers of `hub/` and passes with the new ones
  (`Info`, `GetInfo`, `InfoNotOfferedError`, `WarmRequest`, `WarmResult`, `WarmTTSCache`); `api_spec_test.go` passes.
- `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./hub` lists only
  `github.com/Sonora-Multiroom/sonora-cli/hub`; no new `hub/` file imports `internal/`.

## Startup (Principle I)

The new `get info` and `warm` arms are two more branches in `dispatchGetList` and `run`'s `switch`, reached only
after the verb is known. They run no work before argument parsing, and the startup path is unchanged, so no
benchmark is needed (read `run` in `cmd/sonora/main.go`).

## Live check, read-only (2026-11-09, `http://multiroom.lan:8080`)

`sonora get info` and `get info --json`: exit 0, `product "multiroom"`, `version`/`apiVersion` `0.1.33`,
`startedAt 2026-11-09T12:07:32.320Z`, `uptimeSeconds 6962`.

## Live check, warm and cache round trip (2026-11-09, `multiroom.lan`, user permission given)

- Before: `get tts-cache` 94 entries (`piper-local` 34). Office speaker: volume 30, unmuted, enabled; master mute off; no routes.
- `warm tts-cache "Sonora warm-up check 013" --provider piper-local` with the 15 s default: exit 4 twice, "hub did not respond in time"
  (Piper on the Pi takes about 19 s; the hub still stored the entry). With `--timeout 60s`: `cacheHit: true`, 0.14 s.
- New phrase "... 013 b" with `--timeout 60s`: `cacheHit: false`, 18.9 s. After: 96 entries (`piper-local` 36). Left in the cache; `clear tts-cache` not run.
- SC-002: `speak "Sonora warm-up check 013" outputs/office --provider piper-local --route-volume 20` answered `cacheHit: true`.
- After: office volume 30, unmuted, enabled (unchanged, route volume used); no routes; ALSA `status` grep printed nothing;
  `audio.output.routes` for `output:office` = 0.
