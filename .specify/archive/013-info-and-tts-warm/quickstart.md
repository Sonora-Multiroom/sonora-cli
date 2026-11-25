# Quickstart: validating hub info and TTS warm-up

## Automated

```sh
gofmt -l .
go vet ./...
go test ./...
```

Expected: no `gofmt` output, `go vet` exit 0, all suites pass. The new tests are in
`tests/contract/info_test.go`, `tests/contract/tts_test.go`, `tests/unit/render_info_test.go`,
`tests/unit/render_tts_warm_test.go` and `tests/integration/{info,tts_warm}_test.go`. See
[contracts/cli.md](contracts/cli.md) and [contracts/hub-package.md](contracts/hub-package.md) for the
behaviour they check.

## Live hub (`multiroom.lan`, hub 0.1.33)

1. **Hub info** (read-only, allowed without asking):
   `sonora get info` shows the five fields; `sonora get info --json` parses with `jq`.
2. **Older hub** is covered by the contract test (404 gives the "needs hub 0.1.33" message); no live
   check.
3. **Warm** (writes one entry to the production TTS cache, no sound, no speaker change; ask the user
   first):
   - `sonora warm tts-cache "cache warm test" --provider piper-local` gives `cacheHit: false`;
   - the same command again gives `cacheHit: true`;
   - `sonora speak` of that text is *not* run (it would play); the cache hit is shown by the second warm.
4. **Afterwards**: `sonora get tts-cache` shows the entry count went up by one. Do not run
   `clear tts-cache` (it would drop entries the user already has).
