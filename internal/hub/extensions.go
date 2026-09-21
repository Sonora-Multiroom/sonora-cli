package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Extension mirrors #/components/schemas/Extension in api/openapi.json
// field-for-field (constitution Principle II), decoding only the fields the
// TTS "not available" diagnosis needs (research.md §5): name, version,
// requiredApiVersion, and connectionState are not decoded.
type Extension struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejectionReason"`
}

// ExtensionInventory mirrors #/components/schemas/ExtensionInventory in
// api/openapi.json field-for-field (constitution Principle II).
// extensionsDirectory is not decoded, because nothing uses it.
type ExtensionInventory struct {
	LoadingEnabled *bool       `json:"loadingEnabled"`
	Extensions     []Extension `json:"extensions"`
}

// ListExtensions calls GET {baseURL}/api/v2/extensions (operationId
// "listExtensions") and returns the decoded inventory. It is used
// internally only, by the TTS "not available" diagnosis in
// internal/cli/tts/report.go, never surfaced as its own command
// (research.md §5). Any non-2xx status is a *StatusError, and an
// undecodable body is a *DecodeError. A missing extensions array decodes as
// a nil (empty) slice, which range treats identically to an empty one.
func ListExtensions(ctx context.Context, client *http.Client, baseURL string) (*ExtensionInventory, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/v2/extensions"
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

	var inv ExtensionInventory
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return nil, &DecodeError{Err: err}
	}
	return &inv, nil
}
