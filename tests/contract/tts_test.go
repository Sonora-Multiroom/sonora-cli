package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/hub"
)

// Request/response shapes here mirror #/components/schemas/SpeakRequest,
// SpeakAcceptedResponse, and TtsErrorResponse, and the speak operation, in
// api/openapi.json (constitution Principle II).

func strPtr(s string) *string { return &s }

func TestSpeakTimeout_Is15Seconds(t *testing.T) {
	if hub.SpeakTimeout != 15*time.Second {
		t.Errorf("hub.SpeakTimeout = %v, want 15s", hub.SpeakTimeout)
	}
}

func TestSpeak_RequestBody_OmitsOptionalFieldsWhenUnset(t *testing.T) {
	var gotBody map[string]any
	var gotContentType, gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1", "cacheHit": false, "queueDepth": 1,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("got method %q, want POST", gotMethod)
	}
	if gotPath != "/api/tts/speak" {
		t.Errorf("got path %q, want /api/tts/speak", gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", gotContentType)
	}
	if len(gotBody) != 3 {
		t.Errorf("expected exactly 3 keys (text, targetName, targetType), got: %+v", gotBody)
	}
	for _, key := range []string{"providerName", "voice", "language"} {
		if _, ok := gotBody[key]; ok {
			t.Errorf("expected %q omitted, got body: %+v", key, gotBody)
		}
	}
	if gotBody["text"] != "Hi" || gotBody["targetName"] != "kitchen" || gotBody["targetType"] != "SINGLE_OUTPUT" {
		t.Errorf("unexpected required fields in body: %+v", gotBody)
	}
}

func TestSpeak_RequestBody_IncludesOptionalFieldsWhenSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1", "cacheHit": false, "queueDepth": 1,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{
		Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT",
		ProviderName: strPtr("openai"), Voice: strPtr("alloy"), Language: strPtr("en-US"),
	}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["providerName"] != "openai" || gotBody["voice"] != "alloy" || gotBody["language"] != "en-US" {
		t.Errorf("expected optional fields present, got: %+v", gotBody)
	}
}

// TestSpeak_RequestBody_OnlyVoiceSet covers 009-tts-commands US3 (T021):
// a request with only Voice set produces only the voice key.
func TestSpeak_RequestBody_OnlyVoiceSet(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1", "cacheHit": false, "queueDepth": 1,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT", Voice: strPtr("alloy")}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["voice"] != "alloy" {
		t.Errorf("expected voice=alloy, got: %+v", gotBody)
	}
	if _, ok := gotBody["providerName"]; ok {
		t.Errorf("expected providerName omitted, got: %+v", gotBody)
	}
	if _, ok := gotBody["language"]; ok {
		t.Errorf("expected language omitted, got: %+v", gotBody)
	}
}

func TestSpeak_202_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "cacheHit": true, "queueDepth": 2,
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	resp, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AnnouncementID != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" || !resp.CacheHit || resp.QueueDepth != 2 {
		t.Errorf("unexpected decoded response: %+v", resp)
	}
}

func testSpeakMalformedBody(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a malformed 202 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestSpeak_MalformedBody_MissingAnnouncementID(t *testing.T) {
	testSpeakMalformedBody(t, `{"cacheHit":true,"queueDepth":1}`)
}
func TestSpeak_MalformedBody_MissingCacheHit(t *testing.T) {
	testSpeakMalformedBody(t, `{"announcementId":"a1","queueDepth":1}`)
}
func TestSpeak_MalformedBody_MissingQueueDepth(t *testing.T) {
	testSpeakMalformedBody(t, `{"announcementId":"a1","cacheHit":true}`)
}
func TestSpeak_MalformedBody_EmptyAnnouncementID(t *testing.T) {
	testSpeakMalformedBody(t, `{"announcementId":"","cacheHit":true,"queueDepth":1}`)
}
func TestSpeak_MalformedBody_NonJSON(t *testing.T) {
	testSpeakMalformedBody(t, `not json`)
}

func testSpeakTTSErrorStatus(t *testing.T, status int, code string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": code, "message": "boom"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatalf("expected an error for status %d code %s, got nil", status, code)
	}
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.StatusCode != status || ttsErr.Code != code || ttsErr.Message != "boom" {
		t.Errorf("unexpected TTSError: %+v", ttsErr)
	}
}

func TestSpeak_400_TargetNotFound(t *testing.T) { testSpeakTTSErrorStatus(t, 400, "TARGET_NOT_FOUND") }
func TestSpeak_400_InvalidRequest(t *testing.T) { testSpeakTTSErrorStatus(t, 400, "INVALID_REQUEST") }
func TestSpeak_400_ProviderNotFound(t *testing.T) {
	testSpeakTTSErrorStatus(t, 400, "PROVIDER_NOT_FOUND")
}
func TestSpeak_503_ProviderTimeout(t *testing.T) { testSpeakTTSErrorStatus(t, 503, "PROVIDER_TIMEOUT") }
func TestSpeak_503_ProviderRateLimited(t *testing.T) {
	testSpeakTTSErrorStatus(t, 503, "PROVIDER_RATE_LIMITED")
}
func TestSpeak_503_ProviderError(t *testing.T) { testSpeakTTSErrorStatus(t, 503, "PROVIDER_ERROR") }
func TestSpeak_503_FormatNormalizationFailed(t *testing.T) {
	testSpeakTTSErrorStatus(t, 503, "FORMAT_NORMALIZATION_FAILED")
}
func TestSpeak_400_UnknownCode_KeptVerbatim(t *testing.T) {
	testSpeakTTSErrorStatus(t, 400, "SOMETHING_NEW")
}

func TestSpeak_ErrorBody_CoreShapedFallsBackToDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Error", "detail": "something went wrong"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "" || ttsErr.Message != "something went wrong" {
		t.Errorf("unexpected TTSError: %+v", ttsErr)
	}
}

func TestSpeak_ErrorBody_CoreShapedFallsBackToTitleWhenNoDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Error Title"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "" || ttsErr.Message != "Error Title" {
		t.Errorf("unexpected TTSError: %+v", ttsErr)
	}
}

func TestSpeak_ErrorBody_HTMLBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("<html><body>502 Bad Gateway</body></html>"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "" || ttsErr.Message != "" {
		t.Errorf("expected an empty code and message for an HTML body, got: %+v", ttsErr)
	}
}

func TestSpeak_ErrorBody_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "" || ttsErr.Message != "" {
		t.Errorf("expected an empty code and message for an empty body, got: %+v", ttsErr)
	}
}

func TestSpeak_ErrorBody_MessageCollapsedToSingleLine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"title": "Error", "detail": "line one\nline two"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if strings.Contains(ttsErr.Message, "\n") {
		t.Errorf("expected a single-line message, got: %q", ttsErr.Message)
	}
}

// TestSpeak_ErrorBody_BareCarriageReturn_CollapsedToSingleLine guards against
// an old-Mac-style bare "\r" (with no following "\n") surviving into a
// rendered error, where it would move the terminal cursor back to column 0
// mid-message and visually overwrite the text that follows it.
func TestSpeak_ErrorBody_BareCarriageReturn_CollapsedToSingleLine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "PROVIDER_ERROR", "message": "line1\rline2"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if strings.ContainsAny(ttsErr.Message, "\r\n") {
		t.Errorf("expected a single-line message with no bare \\r, got: %q", ttsErr.Message)
	}
}

func TestSpeak_ErrorBody_1MiB_HandledSafely(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("x", 1024*1024)))
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	done := make(chan error, 1)
	go func() {
		_, err := hub.Speak(context.Background(), client, srv.URL, req)
		done <- err
	}()
	select {
	case err := <-done:
		var ttsErr *hub.TTSError
		if !errors.As(err, &ttsErr) {
			t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
		}
		if strings.Contains(ttsErr.Message, "\n") {
			t.Errorf("expected a single-line message, got: %q", ttsErr.Message)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Speak did not return promptly for a 1 MiB error body")
	}
}

// GetTTSCacheStats contract tests (009-tts-commands US4, T025). Request/
// response shapes mirror #/components/schemas/CacheStats and the
// getCacheStats operation in api/openapi.json.

func TestGetTTSCacheStats_Decodes(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"totalEntries": 42, "totalSizeBytes": 15728640, "maxSizeBytes": 524288000,
			"entriesByProvider": map[string]any{"openai": 35, "piper-local": 7},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	stats, err := hub.GetTTSCacheStats(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/tts/cache/stats" {
		t.Errorf("got %s %s, want GET /api/tts/cache/stats", gotMethod, gotPath)
	}
	if stats.TotalEntries != 42 || stats.TotalSizeBytes != 15728640 || stats.MaxSizeBytes != 524288000 {
		t.Errorf("unexpected totals: %+v", stats)
	}
	if stats.EntriesByProvider["openai"] != 35 || stats.EntriesByProvider["piper-local"] != 7 {
		t.Errorf("unexpected EntriesByProvider: %+v", stats.EntriesByProvider)
	}
}

func testGetTTSCacheStats_EmptyMap(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	stats, err := hub.GetTTSCacheStats(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.EntriesByProvider == nil {
		t.Error("expected a non-nil empty map")
	}
	if len(stats.EntriesByProvider) != 0 {
		t.Errorf("expected an empty map, got: %+v", stats.EntriesByProvider)
	}
}

func TestGetTTSCacheStats_AbsentEntriesByProvider_YieldsEmptyMap(t *testing.T) {
	testGetTTSCacheStats_EmptyMap(t, `{"totalEntries":0,"totalSizeBytes":0,"maxSizeBytes":100}`)
}
func TestGetTTSCacheStats_NullEntriesByProvider_YieldsEmptyMap(t *testing.T) {
	testGetTTSCacheStats_EmptyMap(t, `{"totalEntries":0,"totalSizeBytes":0,"maxSizeBytes":100,"entriesByProvider":null}`)
}

func testGetTTSCacheStats_MissingRequiredField(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetTTSCacheStats(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a missing required field, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestGetTTSCacheStats_MissingTotalEntries(t *testing.T) {
	testGetTTSCacheStats_MissingRequiredField(t, `{"totalSizeBytes":0,"maxSizeBytes":100}`)
}
func TestGetTTSCacheStats_MissingTotalSizeBytes(t *testing.T) {
	testGetTTSCacheStats_MissingRequiredField(t, `{"totalEntries":0,"maxSizeBytes":100}`)
}
func TestGetTTSCacheStats_MissingMaxSizeBytes(t *testing.T) {
	testGetTTSCacheStats_MissingRequiredField(t, `{"totalEntries":0,"totalSizeBytes":0}`)
}

func TestGetTTSCacheStats_404_NotOffered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetTTSCacheStats(context.Background(), client, srv.URL)
	var notOffered *hub.TTSNotOfferedError
	if !errors.As(err, &notOffered) {
		t.Fatalf("expected a *hub.TTSNotOfferedError, got %T: %v", err, err)
	}
}

func TestGetTTSCacheStats_500_TTSErrorShaped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "SOMETHING", "message": "boom"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.GetTTSCacheStats(context.Background(), client, srv.URL)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "SOMETHING" || ttsErr.Message != "boom" {
		t.Errorf("unexpected TTSError: %+v", ttsErr)
	}
}

// ClearTTSCache contract tests (009-tts-commands US5, T034). Request/
// response shapes mirror the clearCache operation in api/openapi.json.

func TestClearTTSCache_NoProvider_NoQueryString(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.ClearTTSCache(context.Background(), client, srv.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/tts/cache" {
		t.Errorf("got %s %s, want DELETE /api/tts/cache", gotMethod, gotPath)
	}
	if gotQuery != "" {
		t.Errorf("expected no query string, got: %q", gotQuery)
	}
}

func TestClearTTSCache_WithProvider_SetsQueryParam(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	provider := "openai"
	err := hub.ClearTTSCache(context.Background(), client, srv.URL, &provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("providerName") != "openai" {
		t.Errorf("got providerName=%q, want openai", gotQuery.Get("providerName"))
	}
}

func TestClearTTSCache_ProviderWithSpecialChars_URLEncoded(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	provider := "my provider & co"
	err := hub.ClearTTSCache(context.Background(), client, srv.URL, &provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("providerName") != "my provider & co" {
		t.Errorf("got providerName=%q, want %q (URL decode roundtrip)", gotQuery.Get("providerName"), provider)
	}
}

func TestClearTTSCache_204_ReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := hub.NewClient()
	if err := hub.ClearTTSCache(context.Background(), client, srv.URL, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClearTTSCache_400_ProviderNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "PROVIDER_NOT_FOUND", "message": "no such provider"})
	}))
	defer srv.Close()

	client := hub.NewClient()
	provider := "no-such"
	err := hub.ClearTTSCache(context.Background(), client, srv.URL, &provider)
	var ttsErr *hub.TTSError
	if !errors.As(err, &ttsErr) {
		t.Fatalf("expected a *hub.TTSError, got %T: %v", err, err)
	}
	if ttsErr.Code != "PROVIDER_NOT_FOUND" {
		t.Errorf("got Code %q, want PROVIDER_NOT_FOUND", ttsErr.Code)
	}
}

func TestClearTTSCache_404_NotOffered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := hub.NewClient()
	err := hub.ClearTTSCache(context.Background(), client, srv.URL, nil)
	var notOffered *hub.TTSNotOfferedError
	if !errors.As(err, &notOffered) {
		t.Fatalf("expected a *hub.TTSNotOfferedError, got %T: %v", err, err)
	}
}

func TestSpeak_404_NotOffered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req := hub.SpeakRequest{Text: "Hi", TargetName: "kitchen", TargetType: "SINGLE_OUTPUT"}
	_, err := hub.Speak(context.Background(), client, srv.URL, req)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var notOffered *hub.TTSNotOfferedError
	if !errors.As(err, &notOffered) {
		t.Fatalf("expected a *hub.TTSNotOfferedError, got %T: %v", err, err)
	}
}
