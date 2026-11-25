# Public `hub/` Package Contract: Additions

All additions are backward compatible. The one exception is a consumer that builds `CreateRouteRequest`,
`PlaybackRequest` or `CreateInputRequest` with unkeyed struct literals, which `go vet` already flags.
Release as the next minor tag, `v0.3.0`; the last tag is `v0.2.1`.

## New exported identifiers

```go
// Join modes: how a route joins the outputs it targets.
const (
	// JoinModeReplace stops the routes already playing on the target, except ducking ones.
	JoinModeReplace = "REPLACE"
	// JoinModeMix plays alongside the routes already there.
	JoinModeMix = "MIX"
	// JoinModeDuckOthers plays alongside them and lowers them while it plays.
	JoinModeDuckOthers = "DUCK_OTHERS"
)

// Duck states: how a route sounds on one output. The hub may add values; treat an
// unknown one as DuckStateFull.
const (
	DuckStateFull    = "FULL"
	DuckStateLowered = "LOWERED"
)

// RouteOnOutput is one route as seen from one output: the route's fields plus its
// duck state there and whether it is the output's main route. A group route appears
// under each member output it plays on.
type RouteOnOutput struct { /* fields in data-model.md */ }

// Lowered reports whether the route is lowered under a ducking route on this output.
// Any duck state other than DuckStateLowered counts as not lowered.
func (r RouteOnOutput) Lowered() bool

// GetRoutesForOutput lists the routes playing on one output, addressed to it or to a
// group it belongs to, in the order they joined it. A 404 is a *NotFoundError naming
// the output.
func GetRoutesForOutput(ctx context.Context, client *http.Client, baseURL, outputID string) ([]RouteOnOutput, error)
```

## Changed types: new optional fields

```go
// CreateRouteRequest
// JoinMode is how the route joins its target (a JoinMode* value). Nil leaves it to
// the hub: the input's default join mode, else REPLACE.
JoinMode *string `json:"joinMode,omitempty"`

// PlaybackRequest
// JoinMode is how the playback joins its target (a JoinMode* value). Nil means REPLACE.
JoinMode *string `json:"joinMode,omitempty"`

// CreateInputRequest
// DefaultJoinMode is the join mode for routes from this input whose request names
// none (a JoinMode* value). Nil declares none, and such routes use REPLACE.
DefaultJoinMode *string `json:"defaultJoinMode,omitempty"`
```

## Unchanged

- Every call still sends exactly one request. `CreateRoute`, `Playback` and `CreateInput` behave as
  before when the new field is nil, and send identical bytes.
- `Route`, `Input` and `ClassifyError` do not change.
- `hub` checks no join mode value; the hub does.
