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

// Response shapes here mirror #/components/schemas/BulkStopResponse and the
// stopAllRoutes operation in api/openapi.json (constitution Principle II).

func TestStopAllRoutes_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stoppedCount": 2,
			"stoppedRoutes": []map[string]any{
				{"routeId": "route-1", "targetType": "SINGLE_OUTPUT", "targetId": "kitchen-speaker", "stopReason": nil},
				{"routeId": "route-2", "targetType": "OUTPUT_GROUP", "targetId": "living-room", "stopReason": nil},
			},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	result, err := hub.StopAllRoutes(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("got method %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v2/routes" {
		t.Errorf("got path %q, want /api/v2/routes", gotPath)
	}
	if result.StoppedCount != 2 || len(result.StoppedRoutes) != 2 {
		t.Errorf("unexpected decoded result: %+v", result)
	}
	if result.StoppedRoutes[0].StopReason != nil {
		t.Errorf("expected nil StopReason, got: %+v", result.StoppedRoutes[0].StopReason)
	}
}

func TestStopAllRoutes_ZeroStopped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"stoppedCount": 0, "stoppedRoutes": []any{}})
	}))
	defer srv.Close()

	client := hub.NewClient()
	result, err := hub.StopAllRoutes(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StoppedCount != 0 || result.StoppedRoutes == nil || len(result.StoppedRoutes) != 0 {
		t.Errorf("unexpected decoded result: %+v", result)
	}
}

func TestStopAllRoutes_NullStoppedRoutesNormalizedToEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stoppedCount":0,"stoppedRoutes":null}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	result, err := hub.StopAllRoutes(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StoppedRoutes == nil {
		t.Error("expected StoppedRoutes to be normalized to an empty (non-nil) slice")
	}
}

func TestStopAllRoutes_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.StopAllRoutes(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestStopAllRoutes_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.StopAllRoutes(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
