# CLI Contract: Join Modes

## New flags

| Command | Flag | `--help` text |
|---|---|---|
| `sonora play <uri> <outputs\|groups>/<id>` | `--join-mode MODE` | how the playback joins what already plays on the target: MODE is REPLACE, MIX or DUCK_OTHERS (any case); without it, REPLACE |
| `sonora route inputs/<id> <outputs\|groups>/<id>` | `--join-mode MODE` | how the route joins what already plays on the target: MODE is REPLACE, MIX or DUCK_OTHERS (any case); without it, the input's default join mode, else REPLACE |
| `sonora create inputs/<id> <uri> --display-name <name>` | `--default-join-mode MODE` | join MODE for routes from this input that name none: REPLACE, MIX or DUCK_OTHERS (any case); without it, the input declares none and its routes use REPLACE |
| `sonora get outputs/<id>` | `--include-routes` | also list the routes playing on this output, in join order |

The `flag` package's `--help` listing is alphabetical; nothing else about the help output changes.

## Errors (stderr; stdout empty)

| Case | stderr | Exit |
|---|---|---|
| `--join-mode fade`, `--join-mode ""`, `--join-mode duck-others` (same for `--default-join-mode`) | the command's one-line usage, then `error: unknown join mode "fade"; valid modes: REPLACE, MIX, DUCK_OTHERS` | 2 |
| `--include-routes` on `get outputs`, `list outputs`, `get inputs/<id>`, `get groups/<id>`, `get routes/<id>`, `get master-mute`, … | `usage: sonora get <resource>[/<id>] [flags]` (or the `list` usage), then `error: --include-routes applies only to 'sonora get outputs/<id>'` | 2 |
| `--include-routes`, output answers 404 | `error: output not found: <id> (hub URL: <url>)` (as today) | 5 |
| `--include-routes`, output found, routes answers 404 or 405 | `error: this hub cannot list an output's routes; it needs hub 0.1.21 or later (hub URL: <url>)` | 3 |
| `--include-routes`, routes request fails otherwise | that failure's usual message | its usual code |

No request is sent in the exit-2 cases.

## Requests

- With the flag, the body carries `"joinMode": "<UPPER>"` (or `"defaultJoinMode"`) along with every
  field it carries today, including `volume` and `displayName` on `play`.
- Without the flag, the body is byte-identical to today's.
- `--include-routes` sends `GET /api/v2/outputs/<id>`, then `GET /api/v2/outputs/<id>/routes`, on
  one client. Without it, only the first is sent.

## Output

The shapes are in [data-model.md](../data-model.md#rendered-records). In summary:

- `get outputs/<id> --include-routes`: the output record plus a `routes` list (YAML has 9 fields per
  entry; JSON has every field).
- `get routes/<id>`: `joinMode` and `outputs` after `paused` in YAML, each left out when absent.
  JSON is unchanged.
- Input record (YAML of `get`, `create`, `rename`, `enable|disable inputs/<id>`): `defaultJoinMode`
  last, null when the input has none.
- `play`, `route`, `transfer`, `list routes`, `list inputs`, `list outputs`: unchanged.

## README

- A short paragraph defines `REPLACE`, `MIX` and `DUCK_OTHERS`.
- `--join-mode` is documented under `play` and `route`, `--default-join-mode` under `create inputs`,
  and `--include-routes` under `get outputs/<id>`, with one example each.
- The exit-code table is unchanged.
