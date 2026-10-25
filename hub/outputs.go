package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Output mirrors #/components/schemas/OutputResponse in api/openapi.json
// field-for-field.
type Output struct {
	OutputID    string `json:"outputId"`
	DisplayName string `json:"displayName"`
	Volume      int    `json:"volume"`
	Muted       bool   `json:"muted"`
	Available   bool   `json:"available"`
	Enabled     bool   `json:"enabled"`
}

// ListOutputs calls GET {baseURL}/api/v2/outputs, optionally including
// disabled outputs, and returns the decoded output list. Any transport,
// non-2xx, or shape-mismatch failure is returned as an error suitable for
// hub.ClassifyError.
func ListOutputs(ctx context.Context, client *http.Client, baseURL string, includeDisabled bool) ([]Output, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/outputs")
	if err != nil {
		return nil, fmt.Errorf("invalid hub URL %q: %w", baseURL, err)
	}
	q := reqURL.Query()
	q.Set("includeDisabled", strconv.FormatBool(includeDisabled))
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var outputs []Output
	if err := json.NewDecoder(resp.Body).Decode(&outputs); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if outputs == nil {
		outputs = []Output{}
	}
	for _, o := range outputs {
		if o.OutputID == "" || o.DisplayName == "" {
			return nil, &DecodeError{Err: fmt.Errorf("output missing required outputId/displayName")}
		}
	}
	return outputs, nil
}

// GetOutput calls GET {baseURL}/api/v2/outputs/{outputId} and returns the
// decoded output. A 404 response is returned as a *NotFoundError; any other
// non-2xx status is returned as a *StatusError; a malformed 200 body is
// returned as a *DecodeError.
func GetOutput(ctx context.Context, client *http.Client, baseURL, outputID string) (*Output, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID))
	if err != nil {
		return nil, fmt.Errorf("invalid hub URL %q: %w", baseURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var output Output
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if output.OutputID == "" || output.DisplayName == "" {
		return nil, &DecodeError{Err: fmt.Errorf("output missing required outputId/displayName")}
	}
	return &output, nil
}

// Duck states: how a route sounds on one output. The hub may add values;
// treat an unknown one as DuckStateFull.
const (
	DuckStateFull    = "FULL"
	DuckStateLowered = "LOWERED"
)

// RouteOnOutput mirrors #/components/schemas/RouteOnOutputResponse in
// api/openapi.json field-for-field: one route as seen from one output, with
// its duck state there and whether it is the output's main route. A group
// route appears under each member output it plays on.
type RouteOnOutput struct {
	RouteID    string `json:"routeId"`
	InputID    string `json:"inputId"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
	Status     string `json:"status"`
	// JoinMode is how the route joined its target (a JoinMode* value). The
	// hub may add values; an unknown one is kept as sent.
	JoinMode     string  `json:"joinMode"`
	CreatedAt    string  `json:"createdAt"`
	StartedAt    *string `json:"startedAt"`
	Transferable bool    `json:"transferable"`
	Pauseable    bool    `json:"pauseable"`
	Paused       bool    `json:"paused"`
	// DuckState is how the route sounds on this output (a DuckState* value),
	// kept as sent; use Lowered to test it.
	DuckState string `json:"duckState"`
	// Main reports whether this is the output's main route: the first
	// non-ducking route to join it, else the first route.
	Main bool `json:"main"`
	// RouteVolume is the route's own volume (0-100); nil when it follows
	// the output's volume. See Route.RouteVolume.
	RouteVolume *int `json:"routeVolume"`
	// DuckingLevelDb is how far a DUCK_OTHERS route lowers the others, in
	// dB; nil for other join modes.
	DuckingLevelDb *float64 `json:"duckingLevelDb"`
}

// Lowered reports whether the route is lowered under a ducking route on this
// output. Any duck state other than DuckStateLowered counts as not lowered.
func (r RouteOnOutput) Lowered() bool {
	return r.DuckState == DuckStateLowered
}

// GetRoutesForOutput calls GET {baseURL}/api/v2/outputs/{outputId}/routes
// (operationId "getRoutesForOutput") and lists every route playing on the
// output, addressed to it or to a group it belongs to, in the order they
// joined it. A null or empty body is an empty slice. A 404 is returned as a
// *NotFoundError naming the output; any other non-2xx status is a
// *StatusError; a malformed body, or an entry without a route ID, input and
// target or with an unknown target type or status, is a *DecodeError. The
// join mode and duck state are not checked, so unknown values pass through.
func GetRoutesForOutput(ctx context.Context, client *http.Client, baseURL, outputID string) ([]RouteOnOutput, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID) + "/routes")
	if err != nil {
		return nil, fmt.Errorf("invalid hub URL %q: %w", baseURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var routes []RouteOnOutput
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if routes == nil {
		routes = []RouteOnOutput{}
	}
	for _, r := range routes {
		if err := validateRouteFields(r.RouteID, r.InputID, r.TargetID, r.TargetType, r.Status); err != nil {
			return nil, &DecodeError{Err: err}
		}
	}
	return routes, nil
}

// SetOutputEnabled calls PUT {baseURL}/api/v2/outputs/{outputId}/enabled
// (operationId "setOutputEnabled") with {"enabled": enabled}. On success
// (200), the decoded, updated Output is returned (malformed body →
// *DecodeError). A 404 is returned as a *NotFoundError naming the output. A
// 400 attempts to decode the body as an errorResponse into an *APIError,
// falling back to a *StatusError if that decode fails (mirroring
// SetInputEnabled's 400 handling); any other non-2xx status is a
// *StatusError.
func SetOutputEnabled(ctx context.Context, client *http.Client, baseURL, outputID string, enabled bool) (*Output, error) {
	body, err := json.Marshal(struct {
		Enabled bool `json:"enabled"`
	}{Enabled: enabled})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID) + "/enabled"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var output Output
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if output.OutputID == "" || output.DisplayName == "" {
		return nil, &DecodeError{Err: fmt.Errorf("output missing required outputId/displayName")}
	}
	return &output, nil
}

// OutputMute mirrors #/components/schemas/OutputMuteResponse in
// api/openapi.json field-for-field.
type OutputMute struct {
	OutputID  string `json:"outputId"`
	Muted     bool   `json:"muted"`
	UpdatedAt string `json:"updatedAt"`
}

// SetOutputMuted calls PUT {baseURL}/api/v2/outputs/{outputId}/mute
// (operationId "setOutputMute") with {"muted": muted}. On success (200),
// the decoded OutputMute confirmation is returned (malformed body → *DecodeError). A
// 404 is returned as a *NotFoundError naming the output. A 400 attempts to
// decode the body as an errorResponse into an *APIError, falling back to a
// *StatusError if that decode fails (mirroring SetOutputEnabled's 400
// handling); any other non-2xx status is a *StatusError.
func SetOutputMuted(ctx context.Context, client *http.Client, baseURL, outputID string, muted bool) (*OutputMute, error) {
	body, err := json.Marshal(struct {
		Muted bool `json:"muted"`
	}{Muted: muted})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID) + "/mute"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var m OutputMute
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if m.OutputID == "" {
		return nil, &DecodeError{Err: fmt.Errorf("output mute response missing required outputId")}
	}
	return &m, nil
}

// OutputVolume mirrors #/components/schemas/OutputVolumeResponse in
// api/openapi.json field-for-field.
type OutputVolume struct {
	OutputID  string `json:"outputId"`
	Volume    int    `json:"volume"`
	UpdatedAt string `json:"updatedAt"`
}

// SetOutputVolume calls PUT {baseURL}/api/v2/outputs/{outputId}/volume
// (operationId "setOutputVolume") with {"volume": volume}. On success (200),
// the decoded OutputVolume is returned (malformed body → *DecodeError). A
// 404 is returned as a *NotFoundError naming the output. A 400 attempts to
// decode the body as an errorResponse into an *APIError, falling back to a
// *StatusError if that decode fails (mirroring SetInputEnabled's 400
// handling); any other non-2xx status is a *StatusError.
func SetOutputVolume(ctx context.Context, client *http.Client, baseURL, outputID string, volume int) (*OutputVolume, error) {
	body, err := json.Marshal(struct {
		Volume int `json:"volume"`
	}{Volume: volume})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID) + "/volume"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var ov OutputVolume
	if err := json.NewDecoder(resp.Body).Decode(&ov); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if ov.OutputID == "" {
		return nil, &DecodeError{Err: fmt.Errorf("output volume response missing required outputId")}
	}
	return &ov, nil
}

// StopRoutesForOutput calls DELETE {baseURL}/api/v2/outputs/{outputId}/routes
// (operationId "stopRoutesForOutput"), stopping every route addressed
// directly to this output plus any group route whose group currently
// includes it as a member. Idempotent — zero matching routes still returns
// 200 with stoppedCount 0. A 404 is returned as a *NotFoundError naming the
// output; any other non-2xx status is a *StatusError; a malformed 200 body
// is a *DecodeError.
func StopRoutesForOutput(ctx context.Context, client *http.Client, baseURL, outputID string) (*BulkStopResponse, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/outputs/" + url.PathEscape(outputID) + "/routes"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "output", ID: outputID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}
	return decodeBulkStopResponse(resp.Body)
}
