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

// Input mirrors #/components/schemas/InputResponse in api/openapi.json
// field-for-field (constitution Principle II).
type Input struct {
	InputID     string  `json:"inputId"`
	DisplayName string  `json:"displayName"`
	URI         string  `json:"uri"`
	Enabled     bool    `json:"enabled"`
	AutoRemove  bool    `json:"autoRemove"`
	Source      string  `json:"source"`
	CreatedAt   *string `json:"createdAt"`
	Pauseable   bool    `json:"pauseable"`
}

// CreateInputRequest mirrors #/components/schemas/CreateInputRequest in
// api/openapi.json field-for-field (constitution Principle II).
type CreateInputRequest struct {
	InputID     string `json:"inputId"`
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
	Enabled     bool   `json:"enabled"`
	AutoRemove  bool   `json:"autoRemove"`
}

func validateInput(i Input) error {
	if i.InputID == "" || i.DisplayName == "" {
		return fmt.Errorf("input missing required inputId/displayName")
	}
	if i.Source != "STATIC" && i.Source != "EPHEMERAL" {
		return fmt.Errorf("input %q has unrecognized source %q", i.InputID, i.Source)
	}
	return nil
}

// ListInputs calls GET {baseURL}/api/v2/inputs, optionally including
// disabled inputs, and returns the decoded input list. Any transport,
// non-2xx, or shape-mismatch failure is returned as an error suitable for
// hub.ClassifyError.
func ListInputs(ctx context.Context, client *http.Client, baseURL string, includeDisabled bool) ([]Input, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/inputs")
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

	var inputs []Input
	if err := json.NewDecoder(resp.Body).Decode(&inputs); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if inputs == nil {
		inputs = []Input{}
	}
	for _, i := range inputs {
		if err := validateInput(i); err != nil {
			return nil, &DecodeError{Err: err}
		}
	}
	return inputs, nil
}

// GetInput calls GET {baseURL}/api/v2/inputs/{inputId} and returns the
// decoded input. A 404 response is returned as a *NotFoundError; any other
// non-2xx status is returned as a *StatusError; a malformed 200 body is
// returned as a *DecodeError.
func GetInput(ctx context.Context, client *http.Client, baseURL, inputID string) (*Input, error) {
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + "/api/v2/inputs/" + url.PathEscape(inputID))
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
		return nil, &NotFoundError{Resource: "input", ID: inputID}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var input Input
	if err := json.NewDecoder(resp.Body).Decode(&input); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateInput(input); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &input, nil
}

// SetInputEnabled calls PUT {baseURL}/api/v2/inputs/{inputId}/enabled
// (operationId "setInputEnabled") with {"enabled": enabled}. On success
// (200), the decoded, updated Input is returned, validated via the existing
// validateInput helper (malformed body → *DecodeError). A 404 is returned
// as a *NotFoundError naming the input. A 400 attempts to decode the body
// as an errorResponse into an *APIError, falling back to a *StatusError if
// that decode fails (mirroring CreateRoute's 400/422 handling); any other
// non-2xx status is a *StatusError.
func SetInputEnabled(ctx context.Context, client *http.Client, baseURL, inputID string, enabled bool) (*Input, error) {
	body, err := json.Marshal(struct {
		Enabled bool `json:"enabled"`
	}{Enabled: enabled})
	if err != nil {
		return nil, err
	}

	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/inputs/" + url.PathEscape(inputID) + "/enabled"
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
		return nil, &NotFoundError{Resource: "input", ID: inputID}
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

	var input Input
	if err := json.NewDecoder(resp.Body).Decode(&input); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateInput(input); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &input, nil
}

// CreateInput calls POST {baseURL}/api/v2/inputs (operationId
// "createInput"), registering a new ephemeral audio input. On success (201),
// the decoded Input is returned, validated via the existing validateInput
// helper (malformed body → *DecodeError, mirroring CreateRoute's success
// handling). A 400/409 attempts to decode the body as an errorResponse into
// an *APIError, falling back to a *StatusError if that decode fails
// (mirroring CreateRoute's 400/422 handling); any other non-2xx status is a
// *StatusError.
func CreateInput(ctx context.Context, client *http.Client, baseURL string, req CreateInputRequest) (*Input, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v2/inputs", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusConflict:
		var errBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, &StatusError{StatusCode: resp.StatusCode}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Title: errBody.Title, Detail: errBody.Detail}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var created Input
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if err := validateInput(created); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &created, nil
}

// DeleteInput calls DELETE {baseURL}/api/v2/inputs/{inputId} (operationId
// "deleteInput"), removing a previously registered ephemeral input. On
// success (204) nil is returned. A 404 is returned as a *NotFoundError
// naming the input. A 400 (the input is a static, YAML-configured input and
// cannot be deleted) attempts to decode the body as an errorResponse into an
// *APIError, falling back to a *StatusError if that decode fails; any other
// non-2xx status is a *StatusError.
func DeleteInput(ctx context.Context, client *http.Client, baseURL, inputID string) error {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/inputs/" + url.PathEscape(inputID)

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
		return &NotFoundError{Resource: "input", ID: inputID}
	}
	if resp.StatusCode == http.StatusBadRequest {
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
