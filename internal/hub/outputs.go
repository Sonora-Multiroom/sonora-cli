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
// field-for-field (constitution Principle II).
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

// OutputVolume mirrors #/components/schemas/OutputVolumeResponse in
// api/openapi.json field-for-field (constitution Principle II).
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
