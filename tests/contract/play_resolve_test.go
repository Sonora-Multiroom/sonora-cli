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
// called in forced modes.
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

func TestResolveTarget_Default_OnlyOutputFound(t *testing.T) {
	f := newFakeResolveHub(t, true, false)
	client := hub.NewClient()
	targetType, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetType != "SINGLE_OUTPUT" {
		t.Errorf("got targetType %q, want SINGLE_OUTPUT", targetType)
	}
}

func TestResolveTarget_Default_OnlyGroupFound(t *testing.T) {
	f := newFakeResolveHub(t, false, true)
	client := hub.NewClient()
	targetType, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetType != "OUTPUT_GROUP" {
		t.Errorf("got targetType %q, want OUTPUT_GROUP", targetType)
	}
}

func TestResolveTarget_Default_BothFound_Ambiguous(t *testing.T) {
	f := newFakeResolveHub(t, true, true)
	client := hub.NewClient()
	_, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, false)
	if err == nil {
		t.Fatal("expected an error when both output and group match, got nil")
	}
	var ambiguousErr *hub.AmbiguousTargetError
	if !errors.As(err, &ambiguousErr) {
		t.Fatalf("expected a *hub.AmbiguousTargetError, got %T: %v", err, err)
	}
	if ambiguousErr.ID != "shared-id" {
		t.Errorf("got ID %q, want shared-id", ambiguousErr.ID)
	}
}

func TestResolveTarget_Default_NeitherFound_NotFound(t *testing.T) {
	f := newFakeResolveHub(t, false, false)
	client := hub.NewClient()
	_, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, false)
	if err == nil {
		t.Fatal("expected an error when neither matches, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "target" {
		t.Errorf("got Resource %q, want target", notFoundErr.Resource)
	}
}

func TestResolveTarget_ForceGroup_OnlyCallsGroupsEndpoint(t *testing.T) {
	f := newFakeResolveHub(t, true, true)
	client := hub.NewClient()
	targetType, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetType != "OUTPUT_GROUP" {
		t.Errorf("got targetType %q, want OUTPUT_GROUP", targetType)
	}
	if got := atomic.LoadInt32(&f.outputRequests); got != 0 {
		t.Errorf("expected outputs endpoint never called, got %d requests", got)
	}
	if got := atomic.LoadInt32(&f.groupRequests); got != 1 {
		t.Errorf("expected exactly 1 groups request, got %d", got)
	}
}

func TestResolveTarget_ForceGroup_MissOnGroupNotFoundEvenIfOutputExists(t *testing.T) {
	f := newFakeResolveHub(t, true, false)
	client := hub.NewClient()
	_, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", true, false)
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

func TestResolveTarget_ForceOutput_OnlyCallsOutputsEndpoint(t *testing.T) {
	f := newFakeResolveHub(t, true, true)
	client := hub.NewClient()
	targetType, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetType != "SINGLE_OUTPUT" {
		t.Errorf("got targetType %q, want SINGLE_OUTPUT", targetType)
	}
	if got := atomic.LoadInt32(&f.groupRequests); got != 0 {
		t.Errorf("expected groups endpoint never called, got %d requests", got)
	}
	if got := atomic.LoadInt32(&f.outputRequests); got != 1 {
		t.Errorf("expected exactly 1 outputs request, got %d", got)
	}
}

func TestResolveTarget_ForceOutput_MissOnOutputNotFoundEvenIfGroupExists(t *testing.T) {
	f := newFakeResolveHub(t, false, true)
	client := hub.NewClient()
	_, err := hub.ResolveTarget(context.Background(), client, f.srv.URL, "shared-id", false, true)
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
