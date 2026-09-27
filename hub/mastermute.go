package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// MasterMute mirrors #/components/schemas/MasterMuteResponse in
// api/openapi.json field-for-field (constitution Principle II).
type MasterMute struct {
	Muted bool `json:"muted"`
}

// GetMasterMute calls GET {baseURL}/api/v2/master-mute (operationId
// "getMasterMute") and returns the decoded, current system-wide mute state.
// master-mute is a singleton with no identifier, so unlike GetOutput/GetGroup
// there is no 404 case; any non-2xx status is a *StatusError, and a
// malformed 200 body is a *DecodeError.
func GetMasterMute(ctx context.Context, client *http.Client, baseURL string) (*MasterMute, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/master-mute"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

	var mm MasterMute
	if err := json.NewDecoder(resp.Body).Decode(&mm); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &mm, nil
}

// SetMasterMute calls PUT {baseURL}/api/v2/master-mute (operationId
// "setMasterMute") with {"muted": muted} and returns the decoded, updated
// system-wide mute state. Neither this call nor GetMasterMute has a 404 or
// documented 400 response in api/openapi.json, so any non-2xx status is a
// *StatusError, and a malformed 200 body is a *DecodeError.
func SetMasterMute(ctx context.Context, client *http.Client, baseURL string, muted bool) (*MasterMute, error) {
	body, err := json.Marshal(struct {
		Muted bool `json:"muted"`
	}{Muted: muted})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/master-mute"
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var mm MasterMute
	if err := json.NewDecoder(resp.Body).Decode(&mm); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &mm, nil
}
