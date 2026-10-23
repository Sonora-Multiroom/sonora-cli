# Data Model: Join Modes in the CLI

## `hub` types (public package)

### Changed request types: one optional field each

| Type | New field | JSON key | Go type | When unset |
|---|---|---|---|---|
| `CreateRouteRequest` | `JoinMode` | `joinMode` | `*string`, `omitempty` | key left out; the hub uses the input's default, else `REPLACE` |
| `PlaybackRequest` | `JoinMode` | `joinMode` | `*string`, `omitempty` | key left out; the hub uses `REPLACE` |
| `CreateInputRequest` | `DefaultJoinMode` | `defaultJoinMode` | `*string`, `omitempty` | key left out; the input has none |

When it is set, the value is sent as given. `hub` does not check it, so the hub decides.

### Response types already in place (no change)

- `Route.JoinMode string`, `Route.Outputs []string` (nil when the hub leaves them out, before 0.1.21).
- `Input.DefaultJoinMode *string` (nil when the hub sends null or leaves it out).

### New: `RouteOnOutput`

Mirrors `#/components/schemas/RouteOnOutputResponse` field for field, in contract order.

| Field | JSON key | Go type | Notes |
|---|---|---|---|
| `RouteID` | `routeId` | `string` | required (checked) |
| `InputID` | `inputId` | `string` | required (checked) |
| `TargetID` | `targetId` | `string` | required (checked); for a group route, the group |
| `TargetType` | `targetType` | `string` | `SINGLE_OUTPUT` or `OUTPUT_GROUP` (checked) |
| `Status` | `status` | `string` | route status enum (checked) |
| `JoinMode` | `joinMode` | `string` | not checked; an unknown value passes through |
| `CreatedAt` | `createdAt` | `string` | |
| `StartedAt` | `startedAt` | `*string` | null before playback starts |
| `Transferable` | `transferable` | `bool` | |
| `Pauseable` | `pauseable` | `bool` | |
| `Paused` | `paused` | `bool` | |
| `DuckState` | `duckState` | `string` | `FULL` or `LOWERED`, open to more values; not checked |
| `Main` | `main` | `bool` | the output's main route |

Method: `Lowered() bool`, true only when `DuckState == "LOWERED"`.

### New constants

| Name | Value |
|---|---|
| `JoinModeReplace` | `"REPLACE"` |
| `JoinModeMix` | `"MIX"` |
| `JoinModeDuckOthers` | `"DUCK_OTHERS"` |
| `DuckStateFull` | `"FULL"` |
| `DuckStateLowered` | `"LOWERED"` |

### New operation: `GetRoutesForOutput`

`GET /api/v2/outputs/{outputId}/routes` (`getRoutesForOutput`). One request.

| Hub answer | Returned |
|---|---|
| 200, array | `[]RouteOnOutput` in the hub's order; null or `[]` becomes an empty slice |
| 200, an entry fails the checks, or the body is not an array | `*DecodeError` |
| 404 | `*NotFoundError{Resource: "output", ID: outputID}` |
| any other non-2xx | `*StatusError{StatusCode}` |
| transport failure | the transport error |

## CLI values

### Join mode flag value

- Accepted: `REPLACE`, `MIX`, `DUCK_OTHERS`, in any case. Sent upper-cased.
- Rejected with exit 2 and no request: anything else, including `""` and `duck-others`.
- The flag is left out: no field is sent.

### Rendered records

**`get outputs/<id> --include-routes`, YAML**:

```yaml
outputId: "kitchen"
displayName: "Kitchen"
volume: 30
muted: false
available: true
enabled: true
routes:
  - routeId: "route-1"
    inputId: "radio"
    targetId: "downstairs"
    targetType: "OUTPUT_GROUP"
    status: "ACTIVE"
    joinMode: "REPLACE"
    duckState: "LOWERED"
    main: true
    paused: false
```

With no routes, the last part is `routes: []`. In JSON it is one object: the output's six keys,
then `"routes": [ …every RouteOnOutput key… ]`.

**`get routes/<id>`, YAML**: the ten current lines, then:

```yaml
joinMode: "MIX"
outputs:
  - "kitchen"
  - "office"
```

`outputs: []` when it plays on no output. Each of the two keys is left out when the hub did not send
it.

**Input record YAML** (`get`, `create`, `rename`, `enable|disable inputs/<id>`): the eight current
lines, then `defaultJoinMode: "DUCK_OTHERS"` or `defaultJoinMode: null`.

### Exit codes (no new codes)

| Case | Code |
|---|---|
| Invalid or empty join mode; `--include-routes` outside `get outputs/<id>` | 2 |
| `--include-routes`: output found, routes request 404 or 405 | 3 |
| `--include-routes`: output 404 | 5, as today |
| `--include-routes`: routes request fails in any other way | that failure's usual code |
