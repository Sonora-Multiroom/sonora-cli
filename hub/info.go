package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Info says which hub is running and since when. It mirrors the reply of
// GET /api/v2/info. Product, Version and APIVersion are passed through as the
// hub sent them: Version and APIVersion can be "unknown" when the build
// carries no version, so do not assume a semantic version. StartedAt is an
// RFC 3339 time that is equal on every call of one hub process, so a change
// means a restart; it can be wrong after a cold boot of the host. UptimeSeconds
// is nil when the hub did not send it.
type Info struct {
	Product       string `json:"product"`
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	StartedAt     string `json:"startedAt,omitempty"`
	UptimeSeconds *int64 `json:"uptimeSeconds,omitempty"`
}

// infoWire is the decode target: pointer fields distinguish an absent
// product, version or apiVersion from an empty one.
type infoWire struct {
	Product       *string `json:"product"`
	Version       *string `json:"version"`
	APIVersion    *string `json:"apiVersion"`
	StartedAt     string  `json:"startedAt"`
	UptimeSeconds *int64  `json:"uptimeSeconds"`
}

// GetInfo calls GET {baseURL}/api/v2/info (operationId "getHubInfo") and
// returns the decoded hub identity. A 404 is an *InfoNotOfferedError (a hub
// older than 0.1.33); any other non-2xx status is a *StatusError. A body that
// is not a JSON object, or that lacks a non-empty product, version or
// apiVersion, is a *DecodeError.
func GetInfo(ctx context.Context, client *http.Client, baseURL string) (*Info, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &InfoNotOfferedError{}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var wire infoWire
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if wire.Product == nil || *wire.Product == "" ||
		wire.Version == nil || *wire.Version == "" ||
		wire.APIVersion == nil || *wire.APIVersion == "" {
		return nil, &DecodeError{Err: fmt.Errorf("info response missing required product/version/apiVersion")}
	}
	return &Info{
		Product: *wire.Product, Version: *wire.Version, APIVersion: *wire.APIVersion,
		StartedAt: wire.StartedAt, UptimeSeconds: wire.UptimeSeconds,
	}, nil
}
