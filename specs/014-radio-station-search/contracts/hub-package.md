# Public `hub` package additions

All additive; ships as a minor tag. Godoc stays self-contained (no spec, principle or feature-number
references). Field tables are in [data-model.md](../data-model.md).

```go
// RadioTimeout bounds one radio request. It is longer than the hub's own catalogue budget (5 s by
// default), so the hub's "catalogue unavailable" answer arrives before this client gives up.
const RadioTimeout = 10 * time.Second

type RadioStation struct { /* data-model.md */ }
type RadioStationPage struct {
    Stations   []RadioStation `json:"stations"`
    NextOffset *int32         `json:"nextOffset"`
}
type RadioCountry struct  { Code string `json:"code"`; Name string `json:"name"`; StationCount int64 `json:"stationCount"` }
type RadioLanguage struct { Name string `json:"name"`; Code *string `json:"code"`; StationCount int64 `json:"stationCount"` }
type RadioTag struct      { Name string `json:"name"`; StationCount int64 `json:"stationCount"` }
type RadioTagPage struct  { Tags []RadioTag `json:"tags"`; NextOffset *int32 `json:"nextOffset"` }

// StationSearch holds the optional filters, order and paging of a station search. A nil field is not
// sent, so the hub applies its default; with every field nil the hub returns the 20 most-voted stations.
type StationSearch struct {
    Name, Tag, CountryCode, Language, Codec *string
    BitrateMin                              *int
    Order, Direction                        *string
    Limit, Offset                           *int
}
const (
    RadioOrderName = "name"; RadioOrderVotes = "votes"; RadioOrderClickCount = "clickCount"; RadioOrderBitrate = "bitrate"
    RadioDirectionAsc = "asc"; RadioDirectionDesc = "desc"
)

// TagSearch holds the optional name filter and paging of a tag listing.
type TagSearch struct {
    Name          *string
    Limit, Offset *int
}

// GET /api/radio/stations
func SearchStations(ctx context.Context, client *http.Client, baseURL string, q StationSearch) (*RadioStationPage, error)
// GET /api/radio/stations/{uuid}; uuid is path-escaped and otherwise sent as given.
func Station(ctx context.Context, client *http.Client, baseURL, uuid string) (*RadioStation, error)
// GET /api/radio/countries
func Countries(ctx context.Context, client *http.Client, baseURL string) ([]RadioCountry, error)
// GET /api/radio/languages
func Languages(ctx context.Context, client *http.Client, baseURL string) ([]RadioLanguage, error)
// GET /api/radio/tags
func Tags(ctx context.Context, client *http.Client, baseURL string, q TagSearch) (*RadioTagPage, error)

// RadioNotOfferedError: a radio request answered 404 and it was not "station not found", so the hub is
// not serving the radio API. DiagnoseRadioUnavailable says why.
type RadioNotOfferedError struct{}

// RadioUnavailableError is the explained form of RadioNotOfferedError.
type RadioUnavailableError struct {
    Diagnosis       ExtensionDiagnosis
    Reason          string // the hub's rejection reason, single line
    LoadingDisabled bool
    Version         string // the installed extension's version, when the inventory gives one
    BaseURL         string
    Cause           error
}

// DiagnoseRadioUnavailable makes one ListExtensions call and maps the "radio" extension's status onto a
// diagnosis. It never returns nil.
func DiagnoseRadioUnavailable(ctx context.Context, client *http.Client, baseURL string) *RadioUnavailableError
```

## Changed, backward compatible

```go
// ExtensionDiagnosis explains why an extension's API is not served.
type ExtensionDiagnosis int

// TTSDiagnosis is the former name of ExtensionDiagnosis.
//
// Deprecated: use ExtensionDiagnosis. This alias will be removed in a later release.
type TTSDiagnosis = ExtensionDiagnosis

type APIError struct {
    // … existing fields …
    Type       string        // problem "type", e.g. urn:multiroom:radio:error:invalid-parameter; empty when absent
    Parameter  string        // the offending query parameter of an invalid-parameter problem
    RetryAfter time.Duration // from a Retry-After header in seconds; zero when absent or unreadable
}

type Extension struct {
    // … existing fields …
    Version *string `json:"version"`
}

const ClassRadioUnavailable ErrorClass // appended after ClassConflict, so existing values keep their numbers
```

`ClassifyError` gains, before its generic `*APIError` branch:

- `*RadioUnavailableError` → `ClassRadioUnavailable` with its message (`ClassNetwork` for the hub-address
  diagnosis);
- `*RadioNotOfferedError` → `ClassRadioUnavailable`, "radio is not available on this hub";
- `*APIError` whose `Type` starts with `urn:multiroom:radio:error:` → the table in research R5.

Every other branch and message is unchanged.

## Behaviour common to the five calls

- Non-2xx bodies are read up to 64 KiB.
- 2xx: a body that is not the declared shape, or lacks a required field, is a `*DecodeError`.
- 404: `station-not-found` → `*NotFoundError{Resource: "radio station", ID: uuid}` (only `Station` can get
  it); anything else → `*RadioNotOfferedError`.
- Other non-2xx with a problem body → `*APIError` (`Title`, `Detail`, `Type`, `Parameter`, `RetryAfter`);
  a body that is not JSON → `*StatusError`.
- Transport errors and deadlines are returned as they are, as for every other call.
