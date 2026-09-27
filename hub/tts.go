package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SpeakTimeout bounds a `speak` request. It is longer than the standard 5s
// client timeout so the hub's own provider-timeout error arrives before this
// client's request bound does.
const SpeakTimeout = 15 * time.Second

// maxTTSErrorBody bounds how much of a non-2xx TTS response body is read, so
// an oversized proxy error page can't flood memory.
const maxTTSErrorBody = 64 * 1024

// SpeakRequest mirrors #/components/schemas/SpeakRequest in api/openapi.json
// field-for-field. ProviderName/Voice/Language are sent only when supplied
// (omitempty).
type SpeakRequest struct {
	Text         string  `json:"text"`
	TargetName   string  `json:"targetName"`
	TargetType   string  `json:"targetType"`
	ProviderName *string `json:"providerName,omitempty"`
	Voice        *string `json:"voice,omitempty"`
	Language     *string `json:"language,omitempty"`
}

// SpeakAccepted mirrors #/components/schemas/SpeakAcceptedResponse in
// api/openapi.json, decoded from a 202.
type SpeakAccepted struct {
	AnnouncementID string `json:"announcementId"`
	CacheHit       bool   `json:"cacheHit"`
	QueueDepth     int32  `json:"queueDepth"`
}

// speakWire is the decode target for a 202 body: pointer fields distinguish
// "absent" from "zero value" so a missing or empty announcementId, or a
// missing cacheHit/queueDepth, is rejected as a *DecodeError.
type speakWire struct {
	AnnouncementID *string `json:"announcementId"`
	CacheHit       *bool   `json:"cacheHit"`
	QueueDepth     *int32  `json:"queueDepth"`
}

// ttsErrorShape mirrors #/components/schemas/TtsErrorResponse in
// api/openapi.json.
type ttsErrorShape struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// coreErrorShape is the core API's problem-details shape, used as a
// fallback message source when a TTS error body isn't TTS-shaped.
type coreErrorShape struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// SingleLine collapses any line breaks — "\r\n", bare "\n", and bare "\r"
// (old-Mac line endings, which would otherwise move the terminal cursor back
// to column 0 mid-message) — so a rendered error message never spans
// multiple terminal lines or overwrites itself. Every hub-supplied string
// that ends up embedded in a rendered error must be passed through this
// first.
func SingleLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// decodeTTSError reads a non-2xx, non-404 TTS response body (at most
// maxTTSErrorBody) and builds a *TTSError from it: the {error, message}
// shape when present, else the core problem-details detail/title as a
// message with an empty code.
func decodeTTSError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxTTSErrorBody))

	var shaped ttsErrorShape
	if err := json.Unmarshal(body, &shaped); err == nil && shaped.Error != "" {
		return &TTSError{StatusCode: resp.StatusCode, Code: shaped.Error, Message: SingleLine(shaped.Message)}
	}

	var core coreErrorShape
	msg := ""
	if err := json.Unmarshal(body, &core); err == nil {
		if core.Detail != "" {
			msg = core.Detail
		} else if core.Title != "" {
			msg = core.Title
		}
	}
	return &TTSError{StatusCode: resp.StatusCode, Message: SingleLine(msg)}
}

// Speak calls POST {baseURL}/api/tts/speak (operationId "speak"), triggering
// a TTS announcement. A 404 is returned as *TTSNotOfferedError; any other
// non-2xx goes through decodeTTSError. A 2xx decodes into SpeakAccepted,
// rejected as a *DecodeError if announcementId is missing/empty or
// cacheHit/queueDepth are missing.
func Speak(ctx context.Context, client *http.Client, baseURL string, req SpeakRequest) (*SpeakAccepted, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/tts/speak", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &TTSNotOfferedError{}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeTTSError(resp)
	}

	var wire speakWire
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if wire.AnnouncementID == nil || *wire.AnnouncementID == "" || wire.CacheHit == nil || wire.QueueDepth == nil {
		return nil, &DecodeError{Err: fmt.Errorf("speak response missing required announcementId/cacheHit/queueDepth")}
	}
	return &SpeakAccepted{AnnouncementID: *wire.AnnouncementID, CacheHit: *wire.CacheHit, QueueDepth: *wire.QueueDepth}, nil
}

// TTSCacheStats mirrors #/components/schemas/CacheStats in api/openapi.json
// field-for-field. EntriesByProvider is never nil after decoding: an absent
// or null value is normalised to an empty map.
type TTSCacheStats struct {
	TotalEntries      int32
	TotalSizeBytes    int64
	MaxSizeBytes      int64
	EntriesByProvider map[string]int32
}

// cacheStatsWire is the decode target for a 200 body: pointer fields on the
// three required totals distinguish "absent" from "zero".
type cacheStatsWire struct {
	TotalEntries      *int32           `json:"totalEntries"`
	TotalSizeBytes    *int64           `json:"totalSizeBytes"`
	MaxSizeBytes      *int64           `json:"maxSizeBytes"`
	EntriesByProvider map[string]int32 `json:"entriesByProvider"`
}

// GetTTSCacheStats calls GET {baseURL}/api/tts/cache/stats (operationId
// "getCacheStats") and returns the decoded cache statistics. A 404 is
// returned as *TTSNotOfferedError; any other non-2xx goes through
// decodeTTSError. A 2xx missing totalEntries, totalSizeBytes, or
// maxSizeBytes is a *DecodeError.
func GetTTSCacheStats(ctx context.Context, client *http.Client, baseURL string) (*TTSCacheStats, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/tts/cache/stats"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &TTSNotOfferedError{}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeTTSError(resp)
	}

	var wire cacheStatsWire
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if wire.TotalEntries == nil || wire.TotalSizeBytes == nil || wire.MaxSizeBytes == nil {
		return nil, &DecodeError{Err: fmt.Errorf("cache stats response missing required totalEntries/totalSizeBytes/maxSizeBytes")}
	}
	entries := wire.EntriesByProvider
	if entries == nil {
		entries = map[string]int32{}
	}
	return &TTSCacheStats{
		TotalEntries: *wire.TotalEntries, TotalSizeBytes: *wire.TotalSizeBytes, MaxSizeBytes: *wire.MaxSizeBytes,
		EntriesByProvider: entries,
	}, nil
}

// ClearTTSCache calls DELETE {baseURL}/api/tts/cache (operationId
// "clearCache"), clearing all cache entries, or one provider's when
// provider is non-nil (added as a URL-encoded providerName query
// parameter). A 404 is returned as *TTSNotOfferedError; any other non-2xx
// goes through decodeTTSError. A 204 returns nil; clearing an already-empty
// cache still succeeds.
func ClearTTSCache(ctx context.Context, client *http.Client, baseURL string, provider *string) error {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/tts/cache"
	if provider != nil {
		q := url.Values{}
		q.Set("providerName", *provider)
		reqURL += "?" + q.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &TTSNotOfferedError{}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeTTSError(resp)
	}
	return nil
}

// TTSVoice mirrors #/components/schemas/TtsVoiceDescriptor in
// api/openapi.json field-for-field. Every field is optional there, and the
// hub omits null fields (a google-gemini voice has no language, and gender
// may be missing), so pointers keep "absent" distinct from "empty".
type TTSVoice struct {
	ShortName *string `json:"shortName"`
	FullName  *string `json:"fullName"`
	Engine    *string `json:"engine"`
	Language  *string `json:"language"`
	Gender    *string `json:"gender"`
}

// TTSVoiceList mirrors #/components/schemas/TtsVoiceListResponse in
// api/openapi.json field-for-field. Voices is never nil after decoding: an
// absent or null value is normalised to an empty slice.
type TTSVoiceList struct {
	ProviderName *string    `json:"providerName"`
	Voices       []TTSVoice `json:"voices"`
}

// ListTTSVoices calls GET {baseURL}/api/tts/providers/{provider}/voices
// (operationId "listVoices"), with the path-escaped provider name and the
// language/engine query parameters added only when non-nil. A 404 is
// returned as *TTSNotOfferedError; any other non-2xx goes through
// decodeTTSError. The 200 is published as */*, so the body is decoded as
// JSON whatever its Content-Type; an undecodable body is a *DecodeError.
func ListTTSVoices(ctx context.Context, client *http.Client, baseURL, provider string, language, engine *string) (*TTSVoiceList, error) {
	reqURL := strings.TrimRight(baseURL, "/") + "/api/tts/providers/" + url.PathEscape(provider) + "/voices"
	q := url.Values{}
	if language != nil {
		q.Set("language", *language)
	}
	if engine != nil {
		q.Set("engine", *engine)
	}
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &TTSNotOfferedError{}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeTTSError(resp)
	}

	var list TTSVoiceList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, &DecodeError{Err: err}
	}
	if list.Voices == nil {
		list.Voices = []TTSVoice{}
	}
	return &list, nil
}

// TTSError indicates the hub's TTS extension rejected a request (400) or
// hit a provider failure (503), decoded from #/components/schemas/
// TtsErrorResponse in api/openapi.json. Code may be empty when the body
// wasn't TTS-shaped.
type TTSError struct {
	StatusCode int
	Code       string
	Message    string
}

// Error renders the failure message: "<code>: <message>" when Code is set,
// otherwise a message built from the HTTP status and any Message present.
func (e *TTSError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("hub rejected the request (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("hub reported an error (HTTP %d)", e.StatusCode)
}

// TTSNotOfferedError indicates a TTS operation answered 404: the hub isn't
// serving that endpoint at all. Callers that want a user-facing reason
// should follow up with one ListExtensions call and build a
// *TTSUnavailableError from the result, rather than surfacing this error
// directly.
type TTSNotOfferedError struct{}

func (e *TTSNotOfferedError) Error() string {
	return "the hub's TTS extension is not offering this operation"
}

// TTSDiagnosis explains why the TTS extension is unavailable, as determined
// by inspecting the hub's extension inventory (ListExtensions) after a
// *TTSNotOfferedError.
type TTSDiagnosis int

const (
	DiagnosisUnknown TTSDiagnosis = iota
	DiagnosisNotInstalled
	DiagnosisDisabled
	DiagnosisRejected
	DiagnosisInert
	DiagnosisVersionMismatch
	DiagnosisHubAddress
)

// TTSUnavailableError is the user-facing "TTS not available" failure. It
// carries the diagnosed reason (Diagnosis) and is meant to be built by a
// caller that has already run the ListExtensions diagnosis, not returned
// directly by a hub API call — the success path never triggers that
// diagnosis.
type TTSUnavailableError struct {
	Diagnosis       TTSDiagnosis
	Reason          string
	LoadingDisabled bool
	BaseURL         string
	Cause           error
}

// Error renders a message describing why TTS is unavailable, specific to
// the diagnosis.
func (e *TTSUnavailableError) Error() string {
	if e.Diagnosis == DiagnosisHubAddress {
		return fmt.Sprintf("%s is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded", e.BaseURL)
	}

	const head = "text-to-speech is not available on this hub"
	switch e.Diagnosis {
	case DiagnosisNotInstalled:
		if e.LoadingDisabled {
			return head + ": the TTS extension is not installed on this hub (extension loading is switched off in the hub's configuration)"
		}
		return head + ": the TTS extension is not installed on this hub"
	case DiagnosisDisabled:
		return head + ": the TTS extension is installed but disabled in the hub's configuration"
	case DiagnosisRejected:
		if e.Reason != "" {
			return head + ": the TTS extension failed to load: " + e.Reason
		}
		return head + ": the TTS extension failed to load"
	case DiagnosisInert:
		return head + ": the TTS extension is loaded but inactive"
	case DiagnosisVersionMismatch:
		return head + ": the hub's TTS API does not match this client version"
	default: // DiagnosisUnknown
		return head
	}
}

func (e *TTSUnavailableError) Unwrap() error { return e.Cause }
