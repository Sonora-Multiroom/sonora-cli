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

// Response/request shapes here mirror #/components/schemas/EnabledRequest,
// #/components/schemas/OutputResponse, and the setOutputEnabled operation in
// api/openapi.json (constitution Principle II).

func TestSetOutputEnabled_Enable_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"outputId": "office-speaker", "displayName": "Office Speaker",
			"volume": 40, "muted": false, "available": true, "enabled": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	output, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "office-speaker", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v2/outputs/office-speaker/enabled" {
		t.Errorf("got path %q, want /api/v2/outputs/office-speaker/enabled", gotPath)
	}
	if gotBody["enabled"] != true {
		t.Errorf("got request body %+v, want enabled=true", gotBody)
	}
	if output.OutputID != "office-speaker" || !output.Enabled {
		t.Errorf("unexpected decoded output: %+v", output)
	}
}

func TestSetOutputEnabled_Disable_RequestBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"outputId": "office-speaker", "displayName": "Office Speaker",
			"volume": 40, "muted": false, "available": true, "enabled": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	output, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "office-speaker", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["enabled"] != false {
		t.Errorf("got request body %+v, want enabled=false", gotBody)
	}
	if output.Enabled {
		t.Errorf("expected decoded output to have enabled=false, got: %+v", output)
	}
}

func TestSetOutputEnabled_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "missing-output", true)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "output" || notFoundErr.ID != "missing-output" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestSetOutputEnabled_ValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Validation error", "detail": "enabled must be a boolean"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "office-speaker", true)
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Detail != "enabled must be a boolean" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestSetOutputEnabled_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "office-speaker", true)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetOutputEnabled_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"displayName":"Office Speaker","volume":40,"muted":false,"available":true,"enabled":true}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetOutputEnabled(context.Background(), client, srv.URL, "office-speaker", true)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
