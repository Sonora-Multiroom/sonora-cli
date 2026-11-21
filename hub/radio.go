package hub

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// RADIO-002: the read-only calls of the hub's radio extension (station search
// and lookup, and the catalogue's countries, languages and tags). The types
// mirror the Radio* schemas in api/openapi.json.

// RadioTimeout bounds one radio request. It is longer than the hub's own
// catalogue budget (5 s by default), so the hub's "catalogue unavailable"
// answer arrives before this client gives up. Pass it to NewClientWithTimeout.
const RadioTimeout = 10 * time.Second

// maxRadioErrorBody bounds how much of a non-2xx radio response body is read,
// so an oversized proxy error page cannot flood memory.
const maxRadioErrorBody = 64 * 1024

// radioStationNotFoundType is the problem type of a lookup for a station the
// catalogue does not have.
const radioStationNotFoundType = radioProblemPrefix + "station-not-found"

// RadioNotOfferedError indicates a radio request was answered 404 and it was
// not "station not found": the hub is not serving the radio API at all.
// DiagnoseRadioUnavailable says why; ClassifyError gives it only a generic
// message.
type RadioNotOfferedError struct{}

func (e *RadioNotOfferedError) Error() string {
	return "the hub's radio extension is not offering this operation"
}

// RadioStation is one station of the radio catalogue. The text fields come
// from a community catalogue and are not trusted: a name or tag can hold any
// characters, including terminal control sequences. Every field is always
// marshalled, a missing optional one as null.
type RadioStation struct {
	// UUID identifies the station in the catalogue, in lower case.
	UUID string `json:"uuid"`
	// URI is the address to give to play: "radiobrowser://" followed by the UUID.
	URI string `json:"uri"`
	// Name is the station's display name (up to 256 characters), nil when unknown.
	Name *string `json:"name"`
	// Favicon is the address of the station's icon, which the hub has not fetched
	// (up to 2048 characters), nil when unknown.
	Favicon *string `json:"favicon"`
	// Homepage is the station's web page (up to 2048 characters), nil when unknown.
	Homepage *string `json:"homepage"`
	// CountryCode is the ISO 3166-1 alpha-2 code of the station's country, nil
	// when unknown.
	CountryCode *string `json:"countryCode"`
	// Languages names the languages the station broadcasts in (up to 10). It is
	// never nil after decoding.
	Languages []string `json:"languages"`
	// Tags are the station's catalogue tags (up to 30). It is never nil after
	// decoding.
	Tags []string `json:"tags"`
	// Codec is the catalogue's codec label, for example "MP3" or "AAC+" (up to 32
	// characters), nil when unknown.
	Codec *string `json:"codec"`
	// Bitrate is the stream's bitrate in kbit/s, nil when unknown.
	Bitrate *int32 `json:"bitrate"`
	// HLS is true when the stream is an HLS playlist.
	HLS bool `json:"hls"`
	// LastCheckOK is false when the catalogue's last check of the stream failed.
	LastCheckOK bool `json:"lastCheckOk"`
	// Votes is the number of votes the station has received.
	Votes int64 `json:"votes"`
	// ClickCount is the number of plays in the last 24 hours.
	ClickCount int64 `json:"clickCount"`
}

// radioStationWire is the decode target for a station: a pointer tells a
// missing required field apart from its zero value.
type radioStationWire struct {
	UUID        *string   `json:"uuid"`
	URI         *string   `json:"uri"`
	Name        *string   `json:"name"`
	Favicon     *string   `json:"favicon"`
	Homepage    *string   `json:"homepage"`
	CountryCode *string   `json:"countryCode"`
	Languages   *[]string `json:"languages"`
	Tags        *[]string `json:"tags"`
	Codec       *string   `json:"codec"`
	Bitrate     *int32    `json:"bitrate"`
	HLS         *bool     `json:"hls"`
	LastCheckOK *bool     `json:"lastCheckOk"`
	Votes       *int64    `json:"votes"`
	ClickCount  *int64    `json:"clickCount"`
}

// UnmarshalJSON decodes a station and rejects one that lacks a required
// field or has an empty uuid or uri.
func (s *RadioStation) UnmarshalJSON(data []byte) error {
	var w radioStationWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	switch {
	case w.UUID == nil || *w.UUID == "":
		return errors.New("radio station has no uuid")
	case w.URI == nil || *w.URI == "":
		return errors.New("radio station has no uri")
	case w.Languages == nil || w.Tags == nil:
		return errors.New("radio station is missing languages or tags")
	case w.HLS == nil || w.LastCheckOK == nil || w.Votes == nil || w.ClickCount == nil:
		return errors.New("radio station is missing hls, lastCheckOk, votes or clickCount")
	}
	*s = RadioStation{
		UUID: *w.UUID, URI: *w.URI, Name: w.Name, Favicon: w.Favicon, Homepage: w.Homepage,
		CountryCode: w.CountryCode, Languages: *w.Languages, Tags: *w.Tags, Codec: w.Codec, Bitrate: w.Bitrate,
		HLS: *w.HLS, LastCheckOK: *w.LastCheckOK, Votes: *w.Votes, ClickCount: *w.ClickCount,
	}
	return nil
}

// RadioStationPage is one page of a station search.
type RadioStationPage struct {
	// Stations are the page's stations, possibly none. It is never nil after
	// decoding.
	Stations []RadioStation `json:"stations"`
	// NextOffset is the offset of the next page, nil on the last page.
	NextOffset *int32 `json:"nextOffset"`
}

// radioStationPageWire is the decode target for a page: a pointer tells a
// missing stations list apart from an empty one.
type radioStationPageWire struct {
	Stations   *[]RadioStation `json:"stations"`
	NextOffset *int32          `json:"nextOffset"`
}

// Values for StationSearch.Order and StationSearch.Direction.
const (
	// RadioOrderName sorts by station name.
	RadioOrderName = "name"
	// RadioOrderVotes sorts by votes.
	RadioOrderVotes = "votes"
	// RadioOrderClickCount sorts by plays in the last 24 hours.
	RadioOrderClickCount = "clickCount"
	// RadioOrderBitrate sorts by bitrate.
	RadioOrderBitrate = "bitrate"
	// RadioDirectionAsc sorts from lowest to highest.
	RadioDirectionAsc = "asc"
	// RadioDirectionDesc sorts from highest to lowest.
	RadioDirectionDesc = "desc"
)

// StationSearch holds the optional filters, order and paging of a station
// search. A nil field is not sent, so the hub applies its default; with every
// field nil the hub returns the 20 most-voted stations. Values are sent as
// given: the hub checks them and answers 400 for one it refuses.
type StationSearch struct {
	// Name matches part of the station name, in any letter case.
	Name *string
	// Tag matches one whole catalogue tag.
	Tag *string
	// CountryCode is an ISO 3166-1 alpha-2 country code.
	CountryCode *string
	// Language matches one whole language name.
	Language *string
	// Codec matches the catalogue's codec label, for example "MP3".
	Codec *string
	// BitrateMin is the lowest bitrate in kbit/s.
	BitrateMin *int
	// Order is one of the RadioOrder constants.
	Order *string
	// Direction is RadioDirectionAsc or RadioDirectionDesc.
	Direction *string
	// Limit is the page size.
	Limit *int
	// Offset is where the page starts; use the previous page's NextOffset.
	Offset *int
}

// values returns the parameters that are set, each once.
func (q StationSearch) values() url.Values {
	v := url.Values{}
	setString(v, "name", q.Name)
	setString(v, "tag", q.Tag)
	setString(v, "countryCode", q.CountryCode)
	setString(v, "language", q.Language)
	setString(v, "codec", q.Codec)
	setInt(v, "bitrateMin", q.BitrateMin)
	setString(v, "order", q.Order)
	setString(v, "direction", q.Direction)
	setInt(v, "limit", q.Limit)
	setInt(v, "offset", q.Offset)
	return v
}

func setString(v url.Values, key string, value *string) {
	if value != nil {
		v.Set(key, *value)
	}
}

func setInt(v url.Values, key string, value *int) {
	if value != nil {
		v.Set(key, strconv.Itoa(*value))
	}
}

// SearchStations calls GET {baseURL}/api/radio/stations (operationId
// "searchRadioStations") and returns one page of stations. Use RadioTimeout
// for the client. An empty page is a result, not an error. A 404 that is not
// "station not found" is a *RadioNotOfferedError; a radio problem is an
// *APIError with its Type, Parameter and RetryAfter; a body that is not the
// declared shape is a *DecodeError.
func SearchStations(ctx context.Context, client *http.Client, baseURL string, q StationSearch) (*RadioStationPage, error) {
	var wire radioStationPageWire
	if err := radioGet(ctx, client, baseURL, "/api/radio/stations", q.values(), "", &wire); err != nil {
		return nil, err
	}
	if wire.Stations == nil {
		return nil, &DecodeError{Err: errors.New("station page missing required stations")}
	}
	return &RadioStationPage{Stations: *wire.Stations, NextOffset: wire.NextOffset}, nil
}

// Station calls GET {baseURL}/api/radio/stations/{uuid} (operationId
// "getRadioStation") and returns that station. The uuid is path-escaped and
// otherwise sent as given; the hub checks it and answers 400 for one it
// refuses. Use RadioTimeout for the client. A 404 "station not found" is a
// *NotFoundError, any other 404 a *RadioNotOfferedError; other failures are as
// for SearchStations.
func Station(ctx context.Context, client *http.Client, baseURL, uuid string) (*RadioStation, error) {
	var st RadioStation
	if err := radioGet(ctx, client, baseURL, "/api/radio/stations/"+url.PathEscape(uuid), nil, uuid, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// RadioCountry is a country of the radio catalogue with its station count.
type RadioCountry struct {
	// Code is the ISO 3166-1 alpha-2 country code, usable as StationSearch.CountryCode.
	Code string `json:"code"`
	// Name is the country's name.
	Name string `json:"name"`
	// StationCount is the number of stations in the country.
	StationCount int64 `json:"stationCount"`
}

// UnmarshalJSON decodes a country and rejects one that lacks a required field.
func (c *RadioCountry) UnmarshalJSON(data []byte) error {
	var w struct {
		Code         *string `json:"code"`
		Name         *string `json:"name"`
		StationCount *int64  `json:"stationCount"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	if w.Code == nil || w.Name == nil || w.StationCount == nil {
		return errors.New("radio country is missing code, name or stationCount")
	}
	*c = RadioCountry{Code: *w.Code, Name: *w.Name, StationCount: *w.StationCount}
	return nil
}

// RadioLanguage is a language of the radio catalogue with its station count.
type RadioLanguage struct {
	// Name is the language's name, usable as StationSearch.Language.
	Name string `json:"name"`
	// Code is the language's ISO 639 code, nil when the catalogue has none.
	Code *string `json:"code"`
	// StationCount is the number of stations in the language.
	StationCount int64 `json:"stationCount"`
}

// UnmarshalJSON decodes a language and rejects one that lacks a required field.
func (l *RadioLanguage) UnmarshalJSON(data []byte) error {
	var w struct {
		Name         *string `json:"name"`
		Code         *string `json:"code"`
		StationCount *int64  `json:"stationCount"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	if w.Name == nil || w.StationCount == nil {
		return errors.New("radio language is missing name or stationCount")
	}
	*l = RadioLanguage{Name: *w.Name, Code: w.Code, StationCount: *w.StationCount}
	return nil
}

// RadioTag is a tag of the radio catalogue with its station count.
type RadioTag struct {
	// Name is the tag, usable as StationSearch.Tag.
	Name string `json:"name"`
	// StationCount is the number of stations with the tag.
	StationCount int64 `json:"stationCount"`
}

// UnmarshalJSON decodes a tag and rejects one that lacks a required field.
func (g *RadioTag) UnmarshalJSON(data []byte) error {
	var w struct {
		Name         *string `json:"name"`
		StationCount *int64  `json:"stationCount"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	if w.Name == nil || w.StationCount == nil {
		return errors.New("radio tag is missing name or stationCount")
	}
	*g = RadioTag{Name: *w.Name, StationCount: *w.StationCount}
	return nil
}

// RadioTagPage is one page of the catalogue's tags.
type RadioTagPage struct {
	// Tags are the page's tags, possibly none. It is never nil after decoding.
	Tags []RadioTag `json:"tags"`
	// NextOffset is the offset of the next page, nil on the last page.
	NextOffset *int32 `json:"nextOffset"`
}

// TagSearch holds the optional name filter and paging of a tag listing. A nil
// field is not sent, so the hub applies its default.
type TagSearch struct {
	// Name matches part of the tag, in any letter case.
	Name *string
	// Limit is the page size.
	Limit *int
	// Offset is where the page starts; use the previous page's NextOffset.
	Offset *int
}

// values returns the parameters that are set, each once.
func (q TagSearch) values() url.Values {
	v := url.Values{}
	setString(v, "name", q.Name)
	setInt(v, "limit", q.Limit)
	setInt(v, "offset", q.Offset)
	return v
}

// Countries calls GET {baseURL}/api/radio/countries (operationId
// "listRadioCountries") and returns every country of the catalogue in the
// hub's order, largest first. Use RadioTimeout for the client. Failures are as
// for SearchStations.
func Countries(ctx context.Context, client *http.Client, baseURL string) ([]RadioCountry, error) {
	var list *[]RadioCountry
	if err := radioGet(ctx, client, baseURL, "/api/radio/countries", nil, "", &list); err != nil {
		return nil, err
	}
	if list == nil {
		return nil, &DecodeError{Err: errors.New("country list is null")}
	}
	return *list, nil
}

// Languages calls GET {baseURL}/api/radio/languages (operationId
// "listRadioLanguages") and returns every language of the catalogue in the
// hub's order, largest first. Use RadioTimeout for the client. Failures are as
// for SearchStations.
func Languages(ctx context.Context, client *http.Client, baseURL string) ([]RadioLanguage, error) {
	var list *[]RadioLanguage
	if err := radioGet(ctx, client, baseURL, "/api/radio/languages", nil, "", &list); err != nil {
		return nil, err
	}
	if list == nil {
		return nil, &DecodeError{Err: errors.New("language list is null")}
	}
	return *list, nil
}

// radioTagPageWire is the decode target for a tag page: a pointer tells a
// missing tags list apart from an empty one.
type radioTagPageWire struct {
	Tags       *[]RadioTag `json:"tags"`
	NextOffset *int32      `json:"nextOffset"`
}

// Tags calls GET {baseURL}/api/radio/tags (operationId "listRadioTags") and
// returns one page of the catalogue's tags. Use RadioTimeout for the client.
// Failures are as for SearchStations.
func Tags(ctx context.Context, client *http.Client, baseURL string, q TagSearch) (*RadioTagPage, error) {
	var wire radioTagPageWire
	if err := radioGet(ctx, client, baseURL, "/api/radio/tags", q.values(), "", &wire); err != nil {
		return nil, err
	}
	if wire.Tags == nil {
		return nil, &DecodeError{Err: errors.New("tag page missing required tags")}
	}
	return &RadioTagPage{Tags: *wire.Tags, NextOffset: wire.NextOffset}, nil
}

// radioGet sends GET {baseURL}{path}?{query} and decodes a 2xx body into out.
// path must already be escaped. notFoundID names the station a "station not
// found" answer is about. Every radio call goes through it.
func radioGet(ctx context.Context, client *http.Client, baseURL, path string, query url.Values, notFoundID string, out any) error {
	reqURL := strings.TrimRight(baseURL, "/") + path
	if len(query) > 0 {
		// A space as %20 rather than "+", which every server reads the same way.
		reqURL += "?" + strings.ReplaceAll(query.Encode(), "+", "%20")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return radioError(resp, notFoundID)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return &DecodeError{Err: err}
	}
	return nil
}

// radioProblem mirrors #/components/schemas/RadioProblem in api/openapi.json.
type radioProblem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Parameter string `json:"parameter"`
}

// radioError turns a non-2xx radio response into an error: 404 is "station
// not found" or "radio not served", any other status with a problem body is an
// *APIError, and a body that is not JSON is a *StatusError. At most
// maxRadioErrorBody bytes are read.
func radioError(resp *http.Response, notFoundID string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxRadioErrorBody))

	var p radioProblem
	parsed := json.Unmarshal(body, &p) == nil

	if resp.StatusCode == http.StatusNotFound {
		if parsed && p.Type == radioStationNotFoundType {
			return &NotFoundError{Resource: "radio station", ID: notFoundID}
		}
		return &RadioNotOfferedError{}
	}
	if !parsed {
		return &StatusError{StatusCode: resp.StatusCode}
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Title:      SingleLine(p.Title),
		Detail:     SingleLine(p.Detail),
		Type:       p.Type,
		Parameter:  SingleLine(p.Parameter),
		RetryAfter: retryAfter(resp.Header.Get("Retry-After")),
	}
}

// retryAfter reads a Retry-After header given as whole seconds. It returns
// zero for an absent header, a date, a non-number or a non-positive number.
func retryAfter(value string) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
