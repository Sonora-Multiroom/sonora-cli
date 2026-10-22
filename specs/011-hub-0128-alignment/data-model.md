# Data Model: Align the CLI with Hub 0.1.25–0.1.28

## `hub.PlaybackResponse` (mirrors `#/components/schemas/PlaybackResponse`, 0.1.28)

| Field | JSON | Type | Change |
|---|---|---|---|
| `InputID` | `inputId` | string | unchanged |
| `Route` | `route` | `Route` | unchanged |
| `Message` | `message` | string | unchanged |
| `DisplayName` | `displayName` | string | **new**; the input's name once the route exists; `""` when the hub omits it (< 0.1.28) |

No validation: an empty name is not an error.

## Rendered play record (`sonora play`)

Field order, YAML and JSON:

| # | Field | Present when |
|---|---|---|
| 1 | `inputId` | always |
| 2 | `displayName` | **new**; only when non-empty |
| 3 | `routeId` | always |
| 4 | `status` | always |
| 5 | `message` | always |

With no display name the record is byte-identical to the current four-field record.

## Route refusal (`hub.APIError`, 409)

| Field | Change |
|---|---|
| `Reason` | still an open string; doc comment lists `ROUTE_CHANGING_OUTPUTS` as the seventh known value, transient |
| `OutputID` | doc comment: empty for `INPUT_DISABLED`, `GROUP_DISABLED` and `ROUTE_CHANGING_OUTPUTS` |

Classification of a 409 `*APIError`:

| Reason | Class | Message |
|---|---|---|
| `ROUTE_CHANGING_OUTPUTS` | `ClassConflict` (exit 14) | "the route is still adding an output; retry in a moment" |
| any other, or none | `ClassConflict` (exit 14) | hub `detail`, else `title`, else the generic conflict text (unchanged) |

## Transfer attempt sequence (`sonora transfer`)

```text
pre-check target ──► attempt 1 ──accepted──► print result, exit 0
                         │
                   transient 409
                         ▼
          wait 100 ms ► attempt 2 ──accepted──► print result, exit 0
                         │ transient 409
          wait 200 ms ► attempt 3 ── …
          wait 400 ms ► attempt 4 ── …
          wait 800 ms ► attempt 5 ──transient 409──► transient message, exit 14

Any attempt failing with anything else (other 409, 404, 400, 422, 5xx, network)
  ──► reported as today, no further attempts.
```
