# Adopt HUB-040: a chime before a route

**Status:** Proposed, not started.

**Origin:** HUB-040 (API 0.1.34) lets a route start with a short sound. The hub's backlog item
`route-chime.md` lists the clients as the next step.

**Depends on:** hub 0.1.34 or later. `sonora-mcp` follows this item (it uses only the `hub` package).

**Effort:** Small.

## Problem

The CLI cannot ask for a chime, list the chimes a hub has, or show which output, group or input has one,
so the feature can only be used by raw REST calls.

## Proposal

- Re-fetch `api/openapi.json` (published as `api.Spec`) from a 0.1.34 hub and check the version it was
  fetched at.
- `hub` package: `chime` on the play and route calls; `chime` on outputs and groups and `defaultChime` on
  inputs in the response types; `GET /api/v2/chimes` as a `Chimes` call. Absent means no setting, and
  `none` means no chime.
- Commands: `--chime <id|none>` on `sonora play` and `sonora route`; `sonora get chimes`; show the chime
  columns in `get outputs`, `get groups` and `get inputs` only when set.
- Release a tag, then bump `sonora-mcp` (`go get github.com/Sonora-Multiroom/sonora-cli@vX.Y.Z`) and add
  a `chime` argument to its play tool.

## Open questions

1. Should an unknown chime id be checked client-side against `GET /api/v2/chimes`, or left to the hub's
   400 (`Unknown chime: <id>`)?
