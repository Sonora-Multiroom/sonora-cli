# Public `hub/` Package Contract: Additions

All additions are backward compatible; release as the next minor tag.

## New exported identifiers

```go
// PlaybackResponse
DisplayName string `json:"displayName"`

// ReasonRouteChangingOutputs is the APIError.Reason of a transfer refused while the
// route is still adding an output that was just enabled. It clears by itself within
// moments, so retrying the same request shortly is the right response.
const ReasonRouteChangingOutputs = "ROUTE_CHANGING_OUTPUTS"

// IsTransientConflict reports whether err is a 409 route refusal that clears without
// any change of state (today only ReasonRouteChangingOutputs). It is false for every
// other error, including other 409 refusals.
func IsTransientConflict(err error) bool
```

## Changed behaviour

- `ClassifyError` on a 409 `*APIError` with `ReasonRouteChangingOutputs`: class `ClassConflict`
  (unchanged), message "the route is still adding an output; retry in a moment" instead of the hub's
  detail. All other inputs classify exactly as before.

## Changed documentation only

- `APIError.Reason`: lists seven known reasons, notes `ROUTE_CHANGING_OUTPUTS` as transient, keeps
  "treat an unknown one as a generic conflict".
- `APIError.OutputID`: empty for `INPUT_DISABLED`, `GROUP_DISABLED` and `ROUTE_CHANGING_OUTPUTS`.

## Unchanged

- `TransferRoute` and every other call send exactly one request; no call retries.
- `APIError.Reason` remains `string`.
