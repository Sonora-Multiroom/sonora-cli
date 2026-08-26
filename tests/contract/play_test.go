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

// Request/response shapes here mirror #/components/schemas/PlaybackRequest,
// PlaybackResponse, and ErrorResponse, and the playback operation, in
// api/openapi.json (constitution Principle II).

func TestPlayback_RequestBody_OmitsOptionalFieldsWhenUnset(t *testing.T) {
	var gotBody map[string]any
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "playback_1",
			"route": map[string]any{
				"routeId": "route_1", "inputId": "playback_1", "targetId": "office-speaker",
				"targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
				"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
			},
			"message": "Playback started",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := gotBody["displayName"]; ok {
		t.Errorf("expected displayName omitted, got body: %+v", gotBody)
	}
	if _, ok := gotBody["volume"]; ok {
		t.Errorf("expected volume omitted, got body: %+v", gotBody)
	}
	if gotBody["uri"] != req.URI || gotBody["targetId"] != req.TargetID || gotBody["targetType"] != req.TargetType {
		t.Errorf("expected uri/targetId/targetType always present, got body: %+v", gotBody)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-010)", got)
	}
}

func TestPlayback_RequestBody_IncludesOptionalFieldsWhenSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "playback_1",
			"route": map[string]any{
				"routeId": "route_1", "status": "STARTING",
			},
			"message": "Playback started",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	name := "Kitchen Radio"
	volume := 50
	req := hub.PlaybackRequest{
		URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT",
		DisplayName: &name, Volume: &volume,
	}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["displayName"] != name {
		t.Errorf("expected displayName %q present, got body: %+v", name, gotBody)
	}
	if gotBody["volume"] != float64(volume) {
		t.Errorf("expected volume %d present, got body: %+v", volume, gotBody)
	}
}

func TestPlayback_Success_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/play" {
			t.Errorf("got path %q, want /api/v2/play", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("got method %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": "playback_1782345678",
			"route": map[string]any{
				"routeId": "route_abc123", "inputId": "playback_1782345678", "targetId": "office-speaker",
				"targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
				"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
			},
			"message": "Playback started: Radio Stream → office-speaker",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	resp, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.InputID != "playback_1782345678" {
		t.Errorf("got InputID %q, want playback_1782345678", resp.InputID)
	}
	if resp.Route.RouteID != "route_abc123" || resp.Route.Status != "STARTING" {
		t.Errorf("unexpected decoded route: %+v", resp.Route)
	}
	if resp.Message != "Playback started: Radio Stream → office-speaker" {
		t.Errorf("unexpected message: %q", resp.Message)
	}
}

func testPlaybackErrorStatus(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:validation-error", "title": "Validation Error", "detail": "uri must not be blank",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatalf("expected an error for status %d, got nil", status)
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != status || apiErr.Title != "Validation Error" || apiErr.Detail != "uri must not be blank" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestPlayback_400_DecodesAsAPIError(t *testing.T) {
	testPlaybackErrorStatus(t, http.StatusBadRequest)
}
func TestPlayback_422_DecodesAsAPIError(t *testing.T) {
	testPlaybackErrorStatus(t, http.StatusUnprocessableEntity)
}
func TestPlayback_502_DecodesAsAPIError(t *testing.T) {
	testPlaybackErrorStatus(t, http.StatusBadGateway)
}
func TestPlayback_503_DecodesAsAPIError(t *testing.T) {
	testPlaybackErrorStatus(t, http.StatusServiceUnavailable)
}

func TestPlayback_ErrorStatus_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusBadRequest {
		t.Errorf("got StatusCode %d, want 400", statusErr.StatusCode)
	}
}

func TestPlayback_ErrorStatus_EmptyBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for an empty error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
}

func TestPlayback_404_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "target not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "missing-id", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "target" || notFoundErr.ID != "missing-id" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func testPlaybackMalformedBody(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.PlaybackRequest{URI: "https://stream.example.com/live.mp3", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Playback(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a malformed 200 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestPlayback_MalformedBody_MissingInputID(t *testing.T) {
	testPlaybackMalformedBody(t, `{"route":{"routeId":"route_1","status":"STARTING"},"message":"ok"}`)
}

func TestPlayback_MalformedBody_MissingRouteID(t *testing.T) {
	testPlaybackMalformedBody(t, `{"inputId":"playback_1","route":{"routeId":"","status":"STARTING"},"message":"ok"}`)
}

func TestPlayback_MalformedBody_MissingRouteStatus(t *testing.T) {
	testPlaybackMalformedBody(t, `{"inputId":"playback_1","route":{"routeId":"route_1","status":""},"message":"ok"}`)
}
