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

// Group mirrors #/components/schemas/GroupResponse in api/openapi.json
// field-for-field (constitution Principle II). `groups list` and `groups
// get` render the same five fields — see data-model.md.
type Group struct {
	GroupID     string   `json:"groupId"`
	DisplayName string   `json:"displayName"`
	OutputIDs   []string `json:"outputIds"`
	Muted       bool     `json:"muted"`
	Enabled     bool     `json:"enabled"`
}

// ListGroups calls GET {baseURL}/api/v2/groups, optionally including
// disabled groups, and returns the decoded group list. Any transport,
// non-2xx, or shape-mismatch failure is returned as an error suitable for
// hub.ClassifyError.
func ListGroups(ctx context.Context, client *http.Client, baseURL string, includeDisabled bool) ([]Group, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/groups")
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

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if groups == nil {
		groups = []Group{}
	}
	for _, g := range groups {
		if g.GroupID == "" || g.DisplayName == "" {
			return nil, &DecodeError{Err: fmt.Errorf("group missing required groupId/displayName")}
		}
	}
	return groups, nil
}

// GetGroup calls GET {baseURL}/api/v2/groups/{groupId} and returns the
// decoded group. A 404 response is returned as a *NotFoundError; any other
// non-2xx status is returned as a *StatusError; a malformed 200 body is
// returned as a *DecodeError.
func GetGroup(ctx context.Context, client *http.Client, baseURL, groupID string) (*Group, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/groups/" + url.PathEscape(groupID))
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
		return nil, &NotFoundError{Resource: "group", ID: groupID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var group Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if group.GroupID == "" || group.DisplayName == "" {
		return nil, &DecodeError{Err: fmt.Errorf("group missing required groupId/displayName")}
	}
	return &group, nil
}

// SetGroupEnabled calls PUT {baseURL}/api/v2/groups/{groupId}/enabled
// (operationId "setGroupEnabled") with {"enabled": enabled}. On success
// (200), the decoded, updated Group is returned (malformed body →
// *DecodeError). A 404 is returned as a *NotFoundError naming the group. A
// 400 attempts to decode the body as an errorResponse into an *APIError,
// falling back to a *StatusError if that decode fails (mirroring
// SetOutputEnabled's 400 handling); any other non-2xx status is a
// *StatusError.
func SetGroupEnabled(ctx context.Context, client *http.Client, baseURL, groupID string, enabled bool) (*Group, error) {
	body, err := json.Marshal(struct {
		Enabled bool `json:"enabled"`
	}{Enabled: enabled})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/groups/" + url.PathEscape(groupID) + "/enabled"
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
		return nil, &NotFoundError{Resource: "group", ID: groupID}
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

	var group Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if group.GroupID == "" || group.DisplayName == "" {
		return nil, &DecodeError{Err: fmt.Errorf("group missing required groupId/displayName")}
	}
	return &group, nil
}

// GroupVolume mirrors #/components/schemas/GroupVolumeResponse in
// api/openapi.json field-for-field (constitution Principle II).
type GroupVolume struct {
	GroupID   string `json:"groupId"`
	Volume    int    `json:"volume"`
	UpdatedAt string `json:"updatedAt"`
}

// SetGroupVolume calls PUT {baseURL}/api/v2/groups/{groupId}/volume
// (operationId "setGroupVolume") with {"volume": volume}. On success (200),
// the decoded GroupVolume is returned (malformed body → *DecodeError). A
// 404 is returned as a *NotFoundError naming the group. A 400 attempts to
// decode the body as an errorResponse into an *APIError, falling back to a
// *StatusError if that decode fails (mirroring SetOutputVolume's 400
// handling); any other non-2xx status is a *StatusError.
func SetGroupVolume(ctx context.Context, client *http.Client, baseURL, groupID string, volume int) (*GroupVolume, error) {
	body, err := json.Marshal(struct {
		Volume int `json:"volume"`
	}{Volume: volume})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/groups/" + url.PathEscape(groupID) + "/volume"
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
		return nil, &NotFoundError{Resource: "group", ID: groupID}
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

	var gv GroupVolume
	if err := json.NewDecoder(resp.Body).Decode(&gv); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if gv.GroupID == "" {
		return nil, &DecodeError{Err: fmt.Errorf("group volume response missing required groupId")}
	}
	return &gv, nil
}
