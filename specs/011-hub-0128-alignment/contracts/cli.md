# CLI Contract: Align the CLI with Hub 0.1.25–0.1.28

No command, flag or exit code is added or removed. Changes below are the only user-visible ones.

## `enable` / `disable` help

The help keeps its `usage: …` first line, then a blank line, then one description line, then the
existing `Flags:` block.

| Command | Description line |
|---|---|
| `disable inputs/<id>` | Disabling stops every route playing the input, paused ones too, and removes an input registered with auto-remove. |
| `enable inputs/<id>` | Enabling lets new routes use the input again; it starts nothing. |
| `disable outputs/<id>` | Disabling takes the output out of every route playing on it; a route addressed only to it stops. |
| `enable outputs/<id>` | Enabling puts the output back into the group routes still playing on its groups; a route stopped by the disable does not come back. |
| `disable groups/<id>` | Disabling stops every route addressed to the group; routes addressed to a member output keep playing. |
| `enable groups/<id>` | Enabling lets new routes use the group again; it starts nothing. |

Wording may be tightened in review; the facts may not change.

## `sonora play` output

YAML, hub 0.1.28 answering with a name:

```yaml
inputId: "playback_1782345678"
displayName: "Flickermood — Forss"
routeId: "route-3f9a1c07b2e4"
status: "STARTING"
message: "Playback started: Flickermood — Forss → office-speaker"
```

JSON: `{"inputId":"…","displayName":"…","routeId":"…","status":"…","message":"…"}`.

With no name from the hub, the `displayName` line/key is absent and the output is identical to today's.

`--display-name` flag description: "display `NAME` for the input; without it the hub names a SoundCloud
or YouTube link after its track and anything else with a generated name".

## `sonora transfer`

| Hub behaviour | stdout | stderr | Exit |
|---|---|---|---|
| Accepts at once | route record (unchanged) | — | 0 |
| Transient 409 (`ROUTE_CHANGING_OUTPUTS`) 1–4 times, then accepts | route record, identical to an immediate success | — | 0 |
| Transient 409 on all 5 attempts | — | `error: the route is still adding an output; retry in a moment (hub URL: …)` | 14 |
| Same, with `--verbose` | — | the line above plus `detail: <hub's detail>` | 14 |
| Any other 409, 404, 4xx/5xx, network error | — | as today, after exactly one transfer request | as today |

## README

- Enable/disable paragraph: the five rules of FR-001.
- `play` paragraph: `--display-name` is optional; what the hub names the input without it; the output's
  `displayName`.
- Exit-code row 14: append "; a refusal because a route is still adding an output clears by itself, and
  `transfer` retries it for a moment".
