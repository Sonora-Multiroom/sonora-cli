package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// PlaybackRequest mirrors #/components/schemas/PlaybackRequest in
// api/openapi.json field-for-field (constitution Principle II).
type PlaybackRequest struct {
	URI         string  `json:"uri"`
	TargetID    string  `json:"targetId"`
	TargetType  string  `json:"targetType"`
	DisplayName *string `json:"displayName,omitempty"`
	Volume      *int    `json:"volume,omitempty"`
}

// PlaybackResponse mirrors #/components/schemas/PlaybackResponse in
// api/openapi.json field-for-field (constitution Principle II). Route
// decodes into the existing hub.Route struct (internal/hub/routes.go).
type PlaybackResponse struct {
	InputID string `json:"inputId"`
	Route   Route  `json:"route"`
	Message string `json:"message"`
}

// errorResponse mirrors #/components/schemas/ErrorResponse in
// api/openapi.json (constitution Principle II) — only the fields Playback
// needs to construct an *APIError.
type errorResponse struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Playback calls POST {baseURL}/api/v2/play (operationId "playback"),
// creating an ephemeral input and route in one hub round trip. On 200, the
// decoded PlaybackResponse is returned, rejected as a *DecodeError if
// InputID, Route.RouteID, or Route.Status is empty (FR-012). A 404 is
// returned as a *NotFoundError naming the target; a 400/422/502/503 attempts
// to decode the body as an ErrorResponse into a *APIError, falling back to a
// *StatusError if that decode fails; any other non-2xx status is a
// *StatusError.
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusBadGateway, http.StatusServiceUnavailable:
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
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

// ResolveTarget determines whether targetID names a single output or an
// output group by calling the existing GetOutput/GetGroup (per
// data-model.md's Target Resolution State table). forceGroup calls only
// GetGroup; forceOutput calls only GetOutput; neither set calls GetOutput
// then GetGroup and returns *AmbiguousTargetError if both succeed.
func ResolveTarget(ctx context.Context, client *http.Client, baseURL, targetID string, forceGroup, forceOutput bool) (string, error) {
	if forceGroup {
		if _, err := GetGroup(ctx, client, baseURL, targetID); err != nil {
			return "", err
		}
		return "OUTPUT_GROUP", nil
	}
	if forceOutput {
		if _, err := GetOutput(ctx, client, baseURL, targetID); err != nil {
			return "", err
		}
		return "SINGLE_OUTPUT", nil
	}

	_, outputErr := GetOutput(ctx, client, baseURL, targetID)
	outputFound := outputErr == nil
	if outputErr != nil {
		var notFoundErr *NotFoundError
		if !errors.As(outputErr, &notFoundErr) {
			return "", outputErr
		}
	}

	_, groupErr := GetGroup(ctx, client, baseURL, targetID)
	groupFound := groupErr == nil
	if groupErr != nil {
		var notFoundErr *NotFoundError
		if !errors.As(groupErr, &notFoundErr) {
			return "", groupErr
		}
	}

	switch {
	case outputFound && groupFound:
		return "", &AmbiguousTargetError{ID: targetID}
	case outputFound:
		return "SINGLE_OUTPUT", nil
	case groupFound:
		return "OUTPUT_GROUP", nil
	default:
		return "", &NotFoundError{Resource: "target", ID: targetID}
	}
}
