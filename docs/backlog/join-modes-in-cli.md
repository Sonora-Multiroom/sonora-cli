# Support join modes: list the routes on an output and choose how a route joins

**Status:** Proposed. Written 2026-10-07, from comparing `api/openapi.json` at hub 0.1.18 and 0.1.22.

**Origin:** Hub 0.1.21 lets an output play several routes at once. A route joins its target with a join
mode: `REPLACE` stops the routes playing there (except ducking ones), `MIX` plays alongside them, and
`DUCK_OTHERS` plays alongside and lowers them while it plays. The contract now has an endpoint that lists
the routes on one output and request fields that pick the join mode. The `hub` client has none of them,
so the CLI can only create `REPLACE` routes (or the input's default) and cannot show what an output is
playing.

**Depends on:** Hub 0.1.21 (join modes, `getRoutesForOutput`). sonora-tts's `SpeakRequest.playbackMode`
is out of scope until sonora-tts documents it.

**Effort:** Small to medium. One new `hub` call and type, three request fields, `--include-routes` on
`get outputs/<id>`, three create flags, rendering and tests.

## Problem

1. **No way to see what an output plays.** `GET /api/v2/outputs/{outputId}/routes`
   (`getRoutesForOutput`) returns every route playing on the output, whether addressed to it or to a group
   it belongs to, in join order. Each entry (`RouteOnOutputResponse`) has the route's fields plus
   `joinMode`, `duckState` (`FULL` or `LOWERED`; treat an unknown value as `FULL`) and `main` (the
   output's main route). `sonora list routes --target-id <id>` is not the same thing: it filters on what a
   route was addressed to, so it misses group routes playing on the output and has no duck state.
2. **No way to choose the join mode.** `CreateRouteRequest.joinMode` and `PlaybackRequest.joinMode`, and
   `CreateInputRequest.defaultJoinMode`, are missing from the `hub` request types, which claim to mirror
   the contract field for field. When `joinMode` is omitted, `createRoute` uses the input's default join
   mode and `playback` uses `REPLACE`, so today `sonora route` and `sonora play` always replace what is
   playing. Announcements over music (`DUCK_OTHERS`) or a second stream alongside (`MIX`) are impossible.
3. **The join mode is not shown.** `hub.Route` decodes `joinMode` and `outputs` since #21, and
   `hub.Input` decodes `defaultJoinMode` since #22, but `get routes/<id>` and `get inputs/<id>` do not
   render them.

## Proposal

- `hub`: add `GetRoutesForOutput(ctx, client, baseURL, outputID) ([]RouteOnOutput, error)` with a
  `RouteOnOutput` type that mirrors `RouteOnOutputResponse`; a 404 is a `*NotFoundError` naming the
  output. Add `JoinMode *string` (`omitempty`) to `CreateRouteRequest` and `PlaybackRequest`, and
  `DefaultJoinMode *string` (`omitempty`) to `CreateInputRequest`, so that leaving them unset keeps the
  hub's defaults.
- `sonora get outputs/<id> --include-routes`: the output's record plus a `routes:` list, in join order
  with `joinMode`, `duckState` and `main` next to the usual route fields; see
  [Showing an output's routes](#showing-an-outputs-routes).
- `--join-mode <mode>` on `sonora route` and `sonora play`, and `--default-join-mode <mode>` on
  `sonora create inputs/<id>`. Check the value against the enum case-insensitively and upper-case it, as
  `--status` does; an unknown mode is a usage error (exit 2).
- Show `joinMode` and `outputs` in `get routes/<id>`, and `defaultJoinMode` in `get inputs/<id>`, in
  YAML and JSON.
- README: the new flags and command, and a short paragraph on join modes.
- Tests: contract tests for `getRoutesForOutput` (success, empty list, 404, unknown `duckState`) and for
  each request field being sent only when set; integration tests for the flags, an invalid mode, and
  `get outputs/<id> --include-routes` (including a hub older than 0.1.21).
- After release, sonora-mcp can offer the same (a route tool's join mode, and what an output plays).

## Showing an output's routes

**Decided (2026-10-22): `sonora get outputs/<id> --include-routes`.** An opt-in flag on the output's own
record. With it, the CLI makes a second call (`getRoutesForOutput`) and adds a `routes:` list to the
output; without it, `get outputs/<id>` keeps its current output and its single hub call. Help text:
"also list the routes playing on this output, in join order".

- It answers a question about the output ("what is playing on kitchen?"), so it stays separate from the
  route filters, and it mirrors `stop outputs/<id>`, which stops the same set of routes.
- The name follows `--include-disabled` on `list inputs`, `list outputs` and `list groups`, and says
  that it adds data rather than filters or replaces the output. Unlike `--include-disabled`, which adds
  items of the same shape, it adds a field; the help text makes that clear. It is the first flag that
  adds data to a `get` instead of changing its format.
- It maps 1:1 to the opt-in `?includeRoutes=true` proposed for `GET /api/v2/outputs/{outputId}` (open
  question 3 of `sonora-multiroom/docs/backlog/routes-per-output-in-outputs-list.md`). Once the hub has
  it, the flag becomes one call and one consistent snapshot, and an older hub shows itself by the missing
  `routes` field, so the CLI can fall back to the second call instead of reading a 404.

A hub older than 0.1.21 answers 404 for the routes call. The output call has already succeeded by then,
so that 404 must be reported as "this hub does not support listing an output's routes", not as "output
not found".

### Rejected

- **`sonora get outputs/<id> --routes`.** Same behaviour, shorter name, but it can read as "show the
  routes instead of the output" or as a filter, and no other boolean flag in the CLI is a bare noun.
- **`sonora list routes --playing-on <id>`.** Two similar filters side by side with different meanings
  (`--target-id` = addressed to, `--playing-on` = sounding on); unlike the other filters it calls a
  different endpoint and returns a different shape (`duckState`, `main`); it cannot be combined with
  `--status`, `--input-id` or `--target-id`, because the endpoint takes no query parameters.
- **Always include the routes in `sonora get outputs/<id>`.** Every `get outputs/<id>` would make two
  hub calls, and the YAML and JSON shape would change for every caller, including scripts that never
  needed the routes.

- **`sonora get outputs/<id>/routes`.** It mirrors the API path, but no command has a three-part path
  today. It would set a precedent for the whole grammar for one endpoint, and users would expect it to
  work for other kinds too: `get groups/<id>/routes` does not exist in the API (only `DELETE` does), so
  it would be a confusing gap.
- **`sonora list routes --output <id>`.** `list routes` already has `--target-id ID` ("only return
  routes pointed at this target ID"), and `--output kitchen` reads as the same filter while meaning
  something else: it would also return group routes playing on the output, in a different shape. In the
  kubectl/gh flag style this CLI follows, `-o`/`--output` is also the conventional name for the output
  format, so the name would mislead twice.

## Open questions

- **Route lists:** should every route list (`list routes`, `get routes/<id>`) also show `joinMode`, or
  only the view of an output's routes?
