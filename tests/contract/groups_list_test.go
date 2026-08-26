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
// listGroups operation in api/openapi.json (constitution Principle II).

func TestListGroups_RequestAndDecodeContract(t *testing.T) {
	var gotPath, gotQuery string
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"groupId": "living-room", "displayName": "Living Room Speakers", "outputIds": []string{"office-speaker", "bedroom-speaker"}, "muted": false, "enabled": true},
			{"groupId": "empty-group", "displayName": "Empty Group", "outputIds": []string{}, "muted": false, "enabled": true},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	groups, err := hub.ListGroups(context.Background(), client, srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v2/groups" {
		t.Errorf("got path %q, want /api/v2/groups", gotPath)
	}
	if gotQuery != "false" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "false")
	}
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[0].GroupID != "living-room" || groups[0].DisplayName != "Living Room Speakers" ||
		len(groups[0].OutputIDs) != 2 || groups[0].Muted || !groups[0].Enabled {
		t.Errorf("unexpected decoded group: %+v", groups[0])
	}
	if groups[1].GroupID != "empty-group" || groups[1].OutputIDs == nil || len(groups[1].OutputIDs) != 0 {
		t.Errorf("unexpected decoded group with empty outputIds: %+v", groups[1])
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-013)", got)
	}
}

func TestListGroups_IncludeDisabledTrueSetsQueryParam(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("includeDisabled")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer srv.Close()

	client := hub.NewClient()
	if _, err := hub.ListGroups(context.Background(), client, srv.URL, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "true" {
		t.Errorf("got includeDisabled=%q, want %q", gotQuery, "true")
	}
}

func TestListGroups_DecodesBothEnabledAndDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"groupId": "living-room", "displayName": "Living Room Speakers", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": true},
			{"groupId": "unused-group", "displayName": "Unused Group", "outputIds": []string{}, "muted": true, "enabled": false},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	groups, err := hub.ListGroups(context.Background(), client, srv.URL, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[1].GroupID != "unused-group" || groups[1].Enabled {
		t.Errorf("unexpected decoded disabled group: %+v", groups[1])
	}
}

func TestListGroups_HubErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListGroups(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected a *hub.StatusError, got %T: %v", err, err)
	}
}

func TestListGroups_MalformedBodyRejected_MissingGroupID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"displayName":"Living Room Speakers","outputIds":["office-speaker"],"muted":false,"enabled":true}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListGroups(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a malformed body (missing groupId), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestListGroups_MalformedBodyRejected_MissingDisplayName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"groupId":"living-room","outputIds":["office-speaker"],"muted":false,"enabled":true}]`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListGroups(context.Background(), client, srv.URL, false)
	if err == nil {
		t.Fatal("expected an error for a malformed body (missing displayName), got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Errorf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
