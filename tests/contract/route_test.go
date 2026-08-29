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

// Request/response shapes here mirror #/components/schemas/CreateRouteRequest,
// RouteResponse, and ErrorResponse, and the createRoute operation, in
// api/openapi.json (constitution Principle II).

func TestCreateRoute_RequestBody_AlwaysIncludesAllFields(t *testing.T) {
	var gotBody map[string]any
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route_1", "inputId": "spotify-1", "targetId": "office-speaker",
			"targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
			"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["inputId"] != req.InputID || gotBody["targetId"] != req.TargetID || gotBody["targetType"] != req.TargetType {
		t.Errorf("expected inputId/targetId/targetType always present, got body: %+v", gotBody)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry, FR-009)", got)
	}
}

func TestCreateRoute_Success_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/routes" {
			t.Errorf("got path %q, want /api/v2/routes", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("got method %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route_abc123", "inputId": "spotify-1", "targetId": "office-speaker",
			"targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
			"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	resp, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RouteID != "route_abc123" || resp.InputID != "spotify-1" || resp.TargetID != "office-speaker" {
		t.Errorf("unexpected decoded route: %+v", resp)
	}
	if resp.TargetType != "SINGLE_OUTPUT" || resp.Status != "STARTING" {
		t.Errorf("unexpected decoded route: %+v", resp)
	}
}

func testCreateRouteErrorStatus(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:validation-error", "title": "Validation Error", "detail": "inputId must not be blank",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatalf("expected an error for status %d, got nil", status)
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != status || apiErr.Title != "Validation Error" || apiErr.Detail != "inputId must not be blank" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestCreateRoute_400_DecodesAsAPIError(t *testing.T) {
	testCreateRouteErrorStatus(t, http.StatusBadRequest)
}
func TestCreateRoute_422_DecodesAsAPIError(t *testing.T) {
	testCreateRouteErrorStatus(t, http.StatusUnprocessableEntity)
}

func TestCreateRoute_ErrorStatus_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
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

func TestCreateRoute_ErrorStatus_EmptyBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for an empty error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
}

func TestCreateRoute_404_NotFoundNamesTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "input, output, or group not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "missing-id", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
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

func testCreateRouteMalformedBody(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.CreateRouteRequest{InputID: "spotify-1", TargetID: "office-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.CreateRoute(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a malformed 201 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestCreateRoute_MalformedBody_MissingRouteID(t *testing.T) {
	testCreateRouteMalformedBody(t, `{"routeId":"","inputId":"spotify-1","targetId":"office-speaker","targetType":"SINGLE_OUTPUT","status":"STARTING"}`)
}

func TestCreateRoute_MalformedBody_MissingInputID(t *testing.T) {
	testCreateRouteMalformedBody(t, `{"routeId":"route_1","inputId":"","targetId":"office-speaker","targetType":"SINGLE_OUTPUT","status":"STARTING"}`)
}

func TestCreateRoute_MalformedBody_MissingTargetID(t *testing.T) {
	testCreateRouteMalformedBody(t, `{"routeId":"route_1","inputId":"spotify-1","targetId":"","targetType":"SINGLE_OUTPUT","status":"STARTING"}`)
}

func TestCreateRoute_MalformedBody_UnrecognizedTargetType(t *testing.T) {
	testCreateRouteMalformedBody(t, `{"routeId":"route_1","inputId":"spotify-1","targetId":"office-speaker","targetType":"BOGUS","status":"STARTING"}`)
}

func TestCreateRoute_MalformedBody_UnrecognizedStatus(t *testing.T) {
	testCreateRouteMalformedBody(t, `{"routeId":"route_1","inputId":"spotify-1","targetId":"office-speaker","targetType":"SINGLE_OUTPUT","status":"BOGUS"}`)
}

// Request/response shapes here mirror the deleteRoute operation in
// api/openapi.json (constitution Principle II): DELETE /api/v2/routes/{routeId}
// returns 204 on success, 404 if the route doesn't exist, 422 if the stop
// fails.

func TestDeleteRoute_Success_204(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteRoute(context.Background(), client, srv.URL, "route_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("got method %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v2/routes/route_abc123" {
		t.Errorf("got path %q, want /api/v2/routes/route_abc123", gotPath)
	}
}

func TestDeleteRoute_404_NotFoundNamesRoute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "route not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteRoute(context.Background(), client, srv.URL, "missing-route")
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "route" || notFoundErr.ID != "missing-route" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestDeleteRoute_422_DecodesAsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:route-stop-error", "title": "Route Stop Error", "detail": "route could not be stopped",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteRoute(context.Background(), client, srv.URL, "route_abc123")
	if err == nil {
		t.Fatal("expected an error for a 422 response, got nil")
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity || apiErr.Title != "Route Stop Error" || apiErr.Detail != "route could not be stopped" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestDeleteRoute_422_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteRoute(context.Background(), client, srv.URL, "route_abc123")
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("got StatusCode %d, want 422", statusErr.StatusCode)
	}
}

func TestDeleteRoute_OtherErrorStatus_IsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.DeleteRoute(context.Background(), client, srv.URL, "route_abc123")
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("got StatusCode %d, want 500", statusErr.StatusCode)
	}
}

// Request/response shapes here mirror #/components/schemas/TransferRequest,
// RouteResponse, and ErrorResponse, and the transferRoute operation, in
// api/openapi.json (constitution Principle II): POST
// /api/v2/routes/{routeId}/transfer returns 200 with the new route on
// success, 404 if the route doesn't exist, 400 if it isn't transferable,
// 422 if the transfer fails.

func TestTransferRoute_Success_Decodes(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route_new456", "inputId": "spotify-1", "targetId": "bedroom-speaker",
			"targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
			"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	resp, err := hub.TransferRoute(context.Background(), client, srv.URL, "route_abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("got method %q, want POST", gotMethod)
	}
	if gotPath != "/api/v2/routes/route_abc123/transfer" {
		t.Errorf("got path %q, want /api/v2/routes/route_abc123/transfer", gotPath)
	}
	if gotBody["targetId"] != req.TargetID || gotBody["targetType"] != req.TargetType {
		t.Errorf("expected targetId/targetType in request body, got: %+v", gotBody)
	}
	if resp.RouteID != "route_new456" || resp.TargetID != "bedroom-speaker" {
		t.Errorf("unexpected decoded route: %+v", resp)
	}
}

func TestTransferRoute_404_NotFoundNamesRoute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "route not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.TransferRoute(context.Background(), client, srv.URL, "missing-route", req)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "route" || notFoundErr.ID != "missing-route" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func testTransferRouteErrorStatus(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:transfer-error", "title": "Transfer Error", "detail": "route is not transferable",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.TransferRoute(context.Background(), client, srv.URL, "route_abc123", req)
	if err == nil {
		t.Fatalf("expected an error for status %d, got nil", status)
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != status || apiErr.Title != "Transfer Error" || apiErr.Detail != "route is not transferable" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestTransferRoute_400_DecodesAsAPIError(t *testing.T) {
	testTransferRouteErrorStatus(t, http.StatusBadRequest)
}
func TestTransferRoute_422_DecodesAsAPIError(t *testing.T) {
	testTransferRouteErrorStatus(t, http.StatusUnprocessableEntity)
}

func TestTransferRoute_ErrorStatus_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.TransferRoute(context.Background(), client, srv.URL, "route_abc123", req)
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError fallback, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("got StatusCode %d, want 422", statusErr.StatusCode)
	}
}

func TestTransferRoute_OtherErrorStatus_IsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.TransferRoute(context.Background(), client, srv.URL, "route_abc123", req)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("got StatusCode %d, want 500", statusErr.StatusCode)
	}
}

func testTransferRouteMalformedBody(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.TransferRequest{TargetID: "bedroom-speaker", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.TransferRoute(context.Background(), client, srv.URL, "route_abc123", req)
	if err == nil {
		t.Fatal("expected an error for a malformed 200 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestTransferRoute_MalformedBody_MissingRouteID(t *testing.T) {
	testTransferRouteMalformedBody(t, `{"routeId":"","inputId":"spotify-1","targetId":"bedroom-speaker","targetType":"SINGLE_OUTPUT","status":"STARTING"}`)
}

func TestTransferRoute_MalformedBody_UnrecognizedTargetType(t *testing.T) {
	testTransferRouteMalformedBody(t, `{"routeId":"route_new456","inputId":"spotify-1","targetId":"bedroom-speaker","targetType":"BOGUS","status":"STARTING"}`)
}

// Request/response shapes here mirror #/components/schemas/PauseRequest and
// RouteResponse, and the setPauseState operation, in api/openapi.json
// (constitution Principle II).

func testSetPauseState_Success(t *testing.T, paused bool) {
	t.Helper()
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"routeId": "route_abc123", "inputId": "spotify-1", "targetId": "office-speaker",
			"targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "createdAt": "2026-01-01T00:00:00Z",
			"startedAt": "2026-01-01T00:00:01Z", "transferable": true, "pauseable": true, "paused": paused,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	resp, err := hub.SetPauseState(context.Background(), client, srv.URL, "route_abc123", paused)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if gotPath != "/api/v2/routes/route_abc123/pause" {
		t.Errorf("got path %q, want /api/v2/routes/route_abc123/pause", gotPath)
	}
	if gotBody["paused"] != paused {
		t.Errorf("expected paused=%v in request body, got: %+v", paused, gotBody)
	}
	if resp.RouteID != "route_abc123" || resp.Paused != paused {
		t.Errorf("unexpected decoded route: %+v", resp)
	}
}

func TestSetPauseState_Pause_Success(t *testing.T) {
	testSetPauseState_Success(t, true)
}
func TestSetPauseState_Resume_Success(t *testing.T) {
	testSetPauseState_Success(t, false)
}

func TestSetPauseState_404_NotFoundNamesRoute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "route not found"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetPauseState(context.Background(), client, srv.URL, "missing-route", true)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notFoundErr *hub.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected a *hub.NotFoundError, got %T: %v", err, err)
	}
	if notFoundErr.Resource != "route" || notFoundErr.ID != "missing-route" {
		t.Errorf("unexpected NotFoundError: %+v", notFoundErr)
	}
}

func TestSetPauseState_400_DecodesAsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "urn:multiroom:error:validation-error", "title": "Validation Error", "detail": "route is not active",
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetPauseState(context.Background(), client, srv.URL, "route_abc123", true)
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	var apiErr *hub.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *hub.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Title != "Validation Error" || apiErr.Detail != "route is not active" {
		t.Errorf("unexpected APIError: %+v", apiErr)
	}
}

func TestSetPauseState_400_NonJSONBodyFallsBackToStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetPauseState(context.Background(), client, srv.URL, "route_abc123", true)
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

func TestSetPauseState_OtherErrorStatus_IsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetPauseState(context.Background(), client, srv.URL, "route_abc123", true)
	if err == nil {
		t.Fatal("expected an error for a 500 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("got StatusCode %d, want 500", statusErr.StatusCode)
	}
}

func TestSetPauseState_MalformedSuccessBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"routeId":"","inputId":"spotify-1","targetId":"office-speaker","targetType":"SINGLE_OUTPUT","status":"ACTIVE"}`))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.SetPauseState(context.Background(), client, srv.URL, "route_abc123", true)
	if err == nil {
		t.Fatal("expected an error for a malformed 200 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}
