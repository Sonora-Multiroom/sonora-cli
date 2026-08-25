package hub

import (
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
