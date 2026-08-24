package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sonora-cli/internal/hub"
)

// Response shapes here mirror #/components/schemas/OutputResponse and the
// listOutputs operation in api/openapi.json (constitution Principle II).

func TestListOutputs_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
			{"outputId": "kitchen-speaker", "displayName": "Kitchen Speaker", "volume": 50, "muted": true, "available": true, "enabled": true},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	outputs, err := hub.ListOutputs(context.Background(), client, srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/outputs" {
		t.Errorf("got path %q, want /api/v2/outputs", gotPath)
	}
	if gotQuery != "false" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "false")
	}
	if len(outputs) != 2 {
		t.Fatalf("got %d outputs, want 2", len(outputs))
	}
	if outputs[0].OutputID != "office-speaker" || outputs[0].DisplayName != "Office Speaker" ||
		outputs[0].Volume != 75 || outputs[0].Muted || !outputs[0].Available || !outputs[0].Enabled {
		t.Errorf("unexpected decoded output: %+v", outputs[0])
	}
}

func TestListOutputs_IncludeDisabledTrueSetsQueryParam(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer srv.Close()

	client := hub.NewClient()
	if _, err := hub.ListOutputs(context.Background(), client, srv.URL, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "true" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "true")
	}
}

func TestListOutputs_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// volume has the wrong JSON type (string instead of int) per FR-013.
		_, _ = w.Write([]byte(`[{"outputId":"office-speaker","displayName":"Office Speaker","volume":"loud","muted":false,"available":true,"enabled":true}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListOutputs(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
