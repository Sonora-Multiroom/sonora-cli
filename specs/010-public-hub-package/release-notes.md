# Release notes: v0.1.0

The Go module path is now `github.com/Sonora-Multiroom/sonora-cli`. The `sonora`
binary and its CLI behavior are unchanged; this release makes the module
resolvable for other Sonora Go projects.

## New public packages

- **`hub`** — the Multiroom Audio Hub API client, moved out of `internal/hub`.
  It is protocol-only (request/response types, HTTP calls, error types and
  classification), standard-library-only, and documented for use outside this
  repo. See [contracts/public-packages.md](contracts/public-packages.md) and
  `go doc github.com/Sonora-Multiroom/sonora-cli/hub`.
- **`api`** — publishes the Multiroom Audio Hub OpenAPI specification as
  `api.Spec` (`[]byte`, via `go:embed`), so a consumer can verify it is
  testing against the exact spec version `hub` at this module version was
  built and tested against. `cmd/sonora` does not import `api`, so it never
  enters the CLI binary.

## Not part of the public API

- `hub.ErrorClass.ExitCode()` and `hub.ClassUsage` are removed. Exit-code
  mapping is a CLI concern; it now lives in this repo's internal
  `internal/cli/exitcode` package, which is not importable by other modules.
- `hub.ErrorClass`'s other constants keep their names, but their underlying
  integer values are not a stable contract — code must compare against the
  named constants (`hub.ClassNetwork`, `hub.ClassNotFound`, etc.), never
  against literal integers.

## Behavior changes in `hub`

- `hub.TTSUnavailableError`'s message no longer carries CLI-specific hints:
  the hub-address diagnosis drops its "set the correct address with
  --hub-url, MULTIROOM_URL, or the config file" suffix, and the
  version-mismatch diagnosis now says "this client version" instead of "this
  CLI version". The `sonora` CLI's own output is unaffected — it adds the
  same hints back in `internal/cli/tts`.
- `hub.SetOutputMuted` and `hub.SetGroupMuted` now return the new
  `*hub.OutputMute` / `*hub.GroupMute` types (`{outputId|groupId, muted,
  updatedAt}`), matching the spec's `OutputMuteResponse` /
  `GroupMuteResponse`. They used to decode the response as a full `Output` /
  `Group` and require `displayName`, which the hub does not send, so every
  successful call failed with a `*hub.DecodeError`.

## Requirements

- Go ≥ 1.27.

## CLI behavior

Commands and exit codes are unchanged. One output fix: `sonora mute` /
`sonora unmute` for `outputs/<id>` and `groups/<id>` now print the hub's mute
confirmation (`outputId`/`groupId`, `muted`, `updatedAt`) instead of failing
with a malformed-response error. `go install
github.com/Sonora-Multiroom/sonora-cli/cmd/sonora@v0.1.0` now works from
outside this repo.
