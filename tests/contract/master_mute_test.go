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

// Response/request shapes here mirror #/components/schemas/MasterMuteResponse
// and the getMasterMute/setMasterMute operations in api/openapi.json
// (constitution Principle II).

func TestGetMasterMute_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"muted": true})
	}))
	defer srv.Close()

	client := hub.NewClient()
	mm, err := hub.GetMasterMute(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("got method %q, want GET", gotMethod)
	}
	if gotPath != "/api/v2/master-mute" {
		t.Errorf("got path %q, want /api/v2/master-mute", gotPath)
	}
	if !mm.Muted {
		t.Errorf("got muted=%t, want true", mm.Muted)
	}
}

func TestGetMasterMute_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetMasterMute(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestGetMasterMute_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetMasterMute(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestSetMasterMute_Mute_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"muted": true})
	}))
	defer srv.Close()

	client := hub.NewClient()
	mm, err := hub.SetMasterMute(context.Background(), client, srv.URL, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v2/master-mute" {
		t.Errorf("got path %q, want /api/v2/master-mute", gotPath)
	}
	if gotBody["muted"] != true {
		t.Errorf("got request body %+v, want muted=true", gotBody)
	}
	if !mm.Muted {
		t.Errorf("got muted=%t, want true", mm.Muted)
	}
}

func TestSetMasterMute_Unmute_RequestBody(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"muted": false})
	}))
	defer srv.Close()

	client := hub.NewClient()
	mm, err := hub.SetMasterMute(context.Background(), client, srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["muted"] != false {
		t.Errorf("got request body %+v, want muted=false", gotBody)
	}
	if mm.Muted {
		t.Errorf("expected decoded muted=false, got: %+v", mm)
	}
}

func TestSetMasterMute_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetMasterMute(context.Background(), client, srv.URL, true)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestSetMasterMute_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetMasterMute(context.Background(), client, srv.URL, true)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
