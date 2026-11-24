# Public `hub` package additions

All additive; ships as a minor tag. Godoc stays self-contained (no spec or principle references).

```go
// Info is the identity and age of a running hub. Mirrors the reply of GET /api/v2/info.
type Info struct {
    Product       string `json:"product"`
    Version       string `json:"version"`
    APIVersion    string `json:"apiVersion"`
    StartedAt     string `json:"startedAt,omitempty"`
    UptimeSeconds *int64 `json:"uptimeSeconds,omitempty"`
}

// GetInfo calls GET /api/v2/info. A 404 is an *InfoNotOfferedError; another non-2xx status is a
// *StatusError; a malformed body, or one without product, version or apiVersion, is a *DecodeError.
func GetInfo(ctx context.Context, client *http.Client, baseURL string) (*Info, error)

// InfoNotOfferedError: the hub answered 404 to GET /api/v2/info (a hub before 0.1.33).
type InfoNotOfferedError struct{}

// WarmRequest mirrors #/components/schemas/WarmRequest. Optional fields are sent only when non-nil.
type WarmRequest struct {
    Text         string   `json:"text"`
    ProviderName *string  `json:"providerName,omitempty"`
    Voice        *string  `json:"voice,omitempty"`
    Language     *string  `json:"language,omitempty"`
    Engine       *string  `json:"engine,omitempty"`
    Pitch        *float64 `json:"pitch,omitempty"`
    SpeakingRate *float64 `json:"speakingRate,omitempty"`
    StylePrompt  *string  `json:"stylePrompt,omitempty"`
}

// WarmResult mirrors #/components/schemas/WarmResponse.
type WarmResult struct {
    CacheHit bool `json:"cacheHit"`
}

// WarmTTSCache calls POST /api/tts/cache/warm. A 404 is a *TTSNotOfferedError; any other non-2xx status
// is a *TTSError; a 2xx body without cacheHit is a *DecodeError.
func WarmTTSCache(ctx context.Context, client *http.Client, baseURL string, req WarmRequest) (*WarmResult, error)
```

`ClassifyError`: `*InfoNotOfferedError` gives `ClassHub` and the message in `contracts/cli.md`;
`CACHE_WRITE_FAILED` gives `ClassServiceUnavailable`.
