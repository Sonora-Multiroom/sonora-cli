# TinySpec: JSON field names on `hub.TTSCacheStats`

**Branch**: fix/tts-cache-stats-json-tags
**Date**: 2026-10-28
**Status**: implemented, tag pending
**Complexity**: small

## What

`hub.TTSCacheStats` has no `json` tags, so marshalling it gives `TotalEntries`, `TotalSizeBytes`,
`MaxSizeBytes` and `EntriesByProvider` instead of the property names of
`#/components/schemas/CacheStats`. The CLI never shows this, because it renders the stats through its
own view type (`internal/render/tts.go`). sonora-mcp's planned `getTtsCacheStats` tool returns the
type as is, and its conformance test would reject the Go names (sonora-mcp
`specs/002-tts-tools/research.md` R2). Every other exported `hub` response type is already tagged.

## Context

| File | Role |
|------|------|
| `hub/tts.go` | Modified: tags on `TTSCacheStats` |
| `tests/contract/tts_test.go` | Modified: marshal test |
| `internal/render/tts.go` | Unchanged: the CLI's view type keeps its own tags |

## Requirements

1. `json.Marshal(hub.TTSCacheStats{...})` gives `totalEntries`, `totalSizeBytes`, `maxSizeBytes` and
   `entriesByProvider`.
2. Decoding is unchanged: `GetTTSCacheStats` still decodes through its private wire type, so the
   missing-field and null-map tests pass unmodified.
3. CLI output is unchanged.

## Tasks

- [x] Failing marshal test in `tests/contract/tts_test.go`
- [x] Add the four tags
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...`

## Done When

- [x] All tasks checked off; tests pass; no vet or gofmt findings
- [ ] Tag `v0.3.4` released

## Verification (2026-10-28)

- Red first: `go test ./tests/contract/ -run TestTTSCacheStats_MarshalsWithSpecFieldNames` failed
  with `{"TotalEntries":42,...}`.
- With the tags: `gofmt -l .` no output; `go vet ./...` exit 0; `go test ./...` exit 0.
