# Adopt HUB-040 in sonora-mcp: a chime before a route

**Status:** Proposed, not started. The CLI half shipped (tinyspec `specs/tiny/route-chime.md`: the `hub`
package, `--chime`, `get chimes`); what remains is the `sonora-mcp` bump.

**Origin:** HUB-040 (API 0.1.34) lets a route start with a short sound. The hub's backlog item
`route-chime.md` lists the clients as the next step.

**Depends on:** a `sonora-cli` tag that contains the `hub` chime support.

**Effort:** Small.

## Problem

`sonora-mcp` cannot ask for a chime or list the chimes a hub has, so an assistant cannot use the feature.

## Proposal

- In `sonora-mcp`, run `go get github.com/Sonora-Multiroom/sonora-cli@vX.Y.Z` and add a `chime` argument
  to the play and route tools, and a tool that lists the chimes (`hub.ListChimes`). Absent means no
  setting, and `none` means no chime.
- Leave an unknown id to the hub's 400 (`Unknown chime: <id>`), as the CLI does.

## Open questions

1. Should the MCP tool descriptions list the loaded chime ids, or only point at the list tool?
