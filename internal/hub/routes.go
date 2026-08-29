package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Route mirrors #/components/schemas/RouteResponse in api/openapi.json
// field-for-field (constitution Principle II). `routes list` renders only
// RouteID/InputID/TargetID/TargetType/Status; `routes get` renders all ten
// fields — see data-model.md.
type Route struct {
	RouteID      string  `json:"routeId"`
	InputID      string  `json:"inputId"`
	TargetID     string  `json:"targetId"`
	TargetType   string  `json:"targetType"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
	StartedAt    *string `json:"startedAt"`
	Transferable bool    `json:"transferable"`
	Pauseable    bool    `json:"pauseable"`
	Paused       bool    `json:"paused"`
}

// CreateRouteRequest mirrors #/components/schemas/CreateRouteRequest in
// api/openapi.json field-for-field (constitution Principle II).
type CreateRouteRequest struct {
	InputID    string `json:"inputId"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
}

// TransferRequest mirrors #/components/schemas/TransferRequest in
// api/openapi.json field-for-field (constitution Principle II).
type TransferRequest struct {
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
}

// PauseRequest mirrors #/components/schemas/PauseRequest in
// api/openapi.json field-for-field (constitution Principle II).
type PauseRequest struct {
	Paused bool `json:"paused"`
}

func validateRoute(r Route) error {
	if r.RouteID == "" || r.InputID == "" || r.TargetID == "" {
		return fmt.Errorf("route missing required routeId/inputId/targetId")
	}
	if r.TargetType != "SINGLE_OUTPUT" && r.TargetType != "OUTPUT_GROUP" {
		return fmt.Errorf("route %q has unrecognized targetType %q", r.RouteID, r.TargetType)
	}
	switch r.Status {
	case "STARTING", "ACTIVE", "STOPPING", "STOPPED", "FAILED":
	default:
		return fmt.Errorf("route %q has unrecognized status %q", r.RouteID, r.Status)
	}
	return nil
}

// ListRoutes calls GET {baseURL}/api/v2/routes, optionally filtered by
// status/inputID/targetID (each sent as a query parameter only when
// non-empty), and returns the decoded route list. Any transport, non-2xx, or
// shape-mismatch failure is returned as an error suitable for
// hub.ClassifyError.
func ListRoutes(ctx context.Context, client *http.Client, baseURL, status, inputID, targetID string) ([]Route, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/routes")
	if err != nil {
		return nil, fmt.Errorf("invalid hub URL %q: %w", baseURL, err)
	}
	q := reqURL.Query()
	if status != "" {
		q.Set("status", status)
	}
	if inputID != "" {
		q.Set("inputId", inputID)
	}
	if targetID != "" {
		q.Set("targetId", targetID)
	}
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

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if routes == nil {
		routes = []Route{}
	}
	for _, r := range routes {
		if err := validateRoute(r); err != nil {
			return nil, &DecodeError{Err: err}
		}
	}
	return routes, nil
}

// GetRoute calls GET {baseURL}/api/v2/routes/{routeId} and returns the
// decoded route. A 404 response is returned as a *NotFoundError; any other
// non-2xx status is returned as a *StatusError; a malformed 200 body is
// returned as a *DecodeError.
func GetRoute(ctx context.Context, client *http.Client, baseURL, routeID string) (*Route, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/routes/" + url.PathEscape(routeID))
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
		return nil, &NotFoundError{Resource: "route", ID: routeID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateRoute(route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &route, nil
}

// DeleteRoute calls DELETE {baseURL}/api/v2/routes/{routeId} (operationId
// "deleteRoute"), stopping playback and removing the route. On success (204)
// nil is returned. A 404 is returned as a *NotFoundError naming the route. A
// 422 attempts to decode the body as an errorResponse into an *APIError,
// falling back to a *StatusError if that decode fails (mirroring
// CreateRoute's 400/422 handling); any other non-2xx status is a
// *StatusError.
func DeleteRoute(ctx context.Context, client *http.Client, baseURL, routeID string) error {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/routes/" + url.PathEscape(routeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Resource: "route", ID: routeID}
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return &StatusError{StatusCode: resp.StatusCode}
		}
		return &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &StatusError{StatusCode: resp.StatusCode}
	}
	return nil
}

// CreateRoute calls POST {baseURL}/api/v2/routes (operationId "createRoute"),
// connecting an existing input to an existing output/group. On 201, the
// decoded Route is returned, validated via the existing validateRoute
// helper (malformed body → *DecodeError, FR-011). A 404 is returned as a
// *NotFoundError naming the target (data-model.md: the pre-checks in
// route.Run give the input/target distinction its real meaning; this is a
// same-shape backstop for the rare race where a resource vanishes between
// the pre-checks and this call, mirroring Playback's own 404 handling). A
// 400/422 attempts to decode the body as an errorResponse into an
// *APIError, falling back to a *StatusError if that decode fails; any other
// non-2xx status is a *StatusError.
func CreateRoute(ctx context.Context, client *http.Client, baseURL string, req CreateRouteRequest) (*Route, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v2/routes", bytes.NewReader(body))
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
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateRoute(route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &route, nil
}

// SetPauseState calls PUT {baseURL}/api/v2/routes/{routeId}/pause
// (operationId "setPauseState") to pause or resume playback on an active
// route. The call is idempotent — setting the same state twice still
// returns 200 with the current state. On success (200) the decoded and
// validated Route is returned, mirroring TransferRoute's success handling.
// A 404 is returned as a *NotFoundError naming the route. A 400 (route not
// active, input not pauseable, or other validation failure) attempts to
// decode the body as an errorResponse into an *APIError, falling back to a
// *StatusError if that decode fails; any other non-2xx status is a
// *StatusError.
func SetPauseState(ctx context.Context, client *http.Client, baseURL, routeID string, paused bool) (*Route, error) {
	body, err := json.Marshal(PauseRequest{Paused: paused})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/routes/" + url.PathEscape(routeID) + "/pause"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
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
		return nil, &NotFoundError{Resource: "route", ID: routeID}
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

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateRoute(route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &route, nil
}

// TransferRoute calls POST {baseURL}/api/v2/routes/{routeId}/transfer
// (operationId "transferRoute"), seamlessly moving an active route's
// playback to a new target. The hub replaces the old route with a new one,
// so on success (200) the decoded and validated *new* Route is returned,
// mirroring CreateRoute's success handling. A 404 is returned as a
// *NotFoundError naming the route. A 400/422 attempts to decode the body as
// an errorResponse into an *APIError, falling back to a *StatusError if
// that decode fails (mirroring CreateRoute's 400/422 handling); any other
// non-2xx status is a *StatusError.
func TransferRoute(ctx context.Context, client *http.Client, baseURL, routeID string, req TransferRequest) (*Route, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/routes/" + url.PathEscape(routeID) + "/transfer"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
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
		return nil, &NotFoundError{Resource: "route", ID: routeID}
	}
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateRoute(route); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &route, nil
}
