package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/hub"
)

// Response shapes here mirror #/components/schemas/GroupResponse and the
// getGroup operation in api/openapi.json (constitution Principle II).

func TestGetGroup_RequestAndDecodeContract_WithMemberOutputs(t *testing.T) {
	var gotPath string
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groupId": "living-room", "displayName": "Living Room Speakers",
			"outputIds": []string{"office-speaker", "bedroom-speaker"},
			"muted":     false, "enabled": true,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	group, err := hub.GetGroup(context.Background(), client, srv.URL, "living-room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/groups/living-room" {
		t.Errorf("got path %q, want /api/v2/groups/living-room", gotPath)
	}
	if group.GroupID != "living-room" || group.DisplayName != "Living Room Speakers" ||
		len(group.OutputIDs) != 2 || group.Muted || !group.Enabled {
		t.Errorf("unexpected decoded group: %+v", group)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-013)", got)
	}
}

func TestGetGroup_RequestAndDecodeContract_NoMemberOutputs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groupId": "empty-group", "displayName": "Empty Group",
			"outputIds": []string{},
			"muted":     false, "enabled": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	group, err := hub.GetGroup(context.Background(), client, srv.URL, "empty-group")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.OutputIDs == nil || len(group.OutputIDs) != 0 {
		t.Errorf("expected empty (non-nil) OutputIDs, got: %+v", group.OutputIDs)
	}
}

func TestGetGroup_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetGroup(context.Background(), client, srv.URL, "living-room")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestGetGroup_MalformedBodyRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"displayName":"Living Room Speakers","outputIds":["office-speaker"],"muted":false,"enabled":true}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetGroup(context.Background(), client, srv.URL, "living-room")
	if err == nil {
		t.Fatal("expected an error for a malformed body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetGroup(context.Background(), client, srv.URL, "missing-group")
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
