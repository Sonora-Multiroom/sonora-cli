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
// stopRoutesForGroup operation in api/openapi.json (constitution Principle
// II).

func TestStopRoutesForGroup_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		reason := "DIRECT"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stoppedCount": 1,
			"stoppedRoutes": []map[string]any{
				{"routeId": "route-1", "targetType": "OUTPUT_GROUP", "targetId": "living-room", "stopReason": reason},
			},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	result, err := hub.StopRoutesForGroup(context.Background(), client, srv.URL, "living-room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("got method %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v2/groups/living-room/routes" {
		t.Errorf("got path %q, want /api/v2/groups/living-room/routes", gotPath)
	}
	if result.StoppedCount != 1 || len(result.StoppedRoutes) != 1 {
		t.Fatalf("unexpected decoded result: %+v", result)
	}
	if result.StoppedRoutes[0].StopReason == nil || *result.StoppedRoutes[0].StopReason != "DIRECT" {
		t.Errorf("expected stopReason DIRECT, got: %+v", result.StoppedRoutes[0].StopReason)
	}
}

func TestStopRoutesForGroup_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.StopRoutesForGroup(context.Background(), client, srv.URL, "missing-group")
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "group" || notFoundErr.ID != "missing-group" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestStopRoutesForGroup_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.StopRoutesForGroup(context.Background(), client, srv.URL, "living-room")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestStopRoutesForGroup_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.StopRoutesForGroup(context.Background(), client, srv.URL, "living-room")
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
