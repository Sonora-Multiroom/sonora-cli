package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// PlaybackRequest mirrors #/components/schemas/PlaybackRequest in
// api/openapi.json field-for-field.
type PlaybackRequest struct {
	URI         string  `json:"uri"`
	TargetID    string  `json:"targetId"`
	TargetType  string  `json:"targetType"`
	DisplayName *string `json:"displayName,omitempty"`
	Volume      *int    `json:"volume,omitempty"`
}

// PlaybackResponse mirrors #/components/schemas/PlaybackResponse in
// api/openapi.json field-for-field. Route decodes into the Route struct
// defined in routes.go.
type PlaybackResponse struct {
	InputID string `json:"inputId"`
	Route   Route  `json:"route"`
	Message string `json:"message"`
}

// errorResponse mirrors #/components/schemas/ErrorResponse in
// api/openapi.json — only the fields needed to construct an *APIError.
type errorResponse struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Reason   string `json:"reason"`
	OutputID string `json:"outputId"`
}

// Playback calls POST {baseURL}/api/v2/play (operationId "playback"),
// creating an ephemeral input and route in one hub round trip. On 200, the
// decoded PlaybackResponse is returned, rejected as a *DecodeError if
// InputID, Route.RouteID, or Route.Status is empty. A 404 is returned as a
// *NotFoundError naming the target; a 400/409/422/502/503 attempts
// to decode the body as an ErrorResponse into a *APIError, falling back to a
// *StatusError if that decode fails; any other non-2xx status is a
// *StatusError. A 409 is a route the hub refused in its current state, with
// APIError.Reason saying why.
func Playback(ctx context.Context, client *http.Client, baseURL string, req PlaybackRequest) (*PlaybackResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v2/play", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Resource: "target", ID: req.TargetID}
	}
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, apiErrorFromBody(resp.StatusCode, resp.Body)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var playbackResp PlaybackResponse
	if err := json.NewDecoder(resp.Body).Decode(&playbackResp); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if playbackResp.InputID == "" || playbackResp.Route.RouteID == "" || playbackResp.Route.Status == "" {
		return nil, &DecodeError{Err: fmt.Errorf("playback response missing required inputId/route.routeId/route.status")}
	}
	return &playbackResp, nil
}

// ResolveTarget verifies that a target of the given, already-known type
// (targetType "SINGLE_OUTPUT" or "OUTPUT_GROUP") exists, by calling the
// matching GetOutput/GetGroup. The caller is expected to already know
// targetType — for example, from the kind of path or identifier the target
// was addressed by — so there is no auto-detect/ambiguity branch here:
// exactly one endpoint is called.
func ResolveTarget(ctx context.Context, client *http.Client, baseURL, targetID, targetType string) error {
	if targetType == "OUTPUT_GROUP" {
		_, err := GetGroup(ctx, client, baseURL, targetID)
		return err
	}
	_, err := GetOutput(ctx, client, baseURL, targetID)
	return err
}
