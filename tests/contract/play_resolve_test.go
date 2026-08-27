package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/hub"
)

// fakeResolveHub serves GET /api/v2/outputs/{id} and GET /api/v2/groups/{id},
// reusing the existing GetOutput/GetGroup response shapes, and counts
// requests per endpoint so tests can assert an un-needed endpoint was never
// called.
type fakeResolveHub struct {
	srv            *httptest.Server
	outputExists   bool
	groupExists    bool
	outputRequests int32
	groupRequests  int32
}

func newFakeResolveHub(t *testing.T, outputExists, groupExists bool) *fakeResolveHub {
	t.Helper()
	f := &fakeResolveHub{outputExists: outputExists, groupExists: groupExists}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v2/outputs/"):
			atomic.AddInt32(&f.outputRequests, 1)
			if !f.outputExists {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"outputId": "shared-id", "displayName": "Office Speaker", "volume": 50,
				"muted": false, "available": true, "enabled": true,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/groups/"):
			atomic.AddInt32(&f.groupRequests, 1)
			if !f.groupExists {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"groupId": "shared-id", "displayName": "Whole House", "outputIds": []string{"a", "b"},
				"muted": false, "enabled": true,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func TestResolveTarget_Output_Found(t *testing.T) {
	f := newFakeResolveHub(t, true, true)
	client := hub.NewClient()
	err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", "SINGLE_OUTPUT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&f.outputRequests); got != 1 {
		t.Errorf("expected exactly 1 outputs request, got %d", got)
	}
	if got := atomic.LoadInt32(&f.groupRequests); got != 0 {
		t.Errorf("expected groups endpoint never called, got %d requests", got)
	}
}

func TestResolveTarget_Output_NotFound(t *testing.T) {
	f := newFakeResolveHub(t, false, true)
	client := hub.NewClient()
	err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", "SINGLE_OUTPUT")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "output" {
		t.Errorf("got Resource %q, want output", notFoundErr.Resource)
	}
	if got := atomic.LoadInt32(&f.groupRequests); got != 0 {
		t.Errorf("expected groups endpoint never called, got %d requests", got)
	}
}

func TestResolveTarget_Group_Found(t *testing.T) {
	f := newFakeResolveHub(t, true, true)
	client := hub.NewClient()
	err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", "OUTPUT_GROUP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&f.groupRequests); got != 1 {
		t.Errorf("expected exactly 1 groups request, got %d", got)
	}
	if got := atomic.LoadInt32(&f.outputRequests); got != 0 {
		t.Errorf("expected outputs endpoint never called, got %d requests", got)
	}
}

func TestResolveTarget_Group_NotFound(t *testing.T) {
	f := newFakeResolveHub(t, true, false)
	client := hub.NewClient()
	err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", "OUTPUT_GROUP")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "group" {
		t.Errorf("got Resource %q, want group", notFoundErr.Resource)
	}
	if got := atomic.LoadInt32(&f.outputRequests); got != 0 {
		t.Errorf("expected outputs endpoint never called, got %d requests", got)
	}
}
