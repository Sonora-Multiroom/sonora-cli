package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Shapes here mirror the radio operations and the RadioStation, RadioStationPage,
// RadioCountry, RadioLanguage, RadioTag, RadioTagPage and RadioProblem schemas
// in api/openapi.json (constitution Principle II).

const radioPrefix = "urn:multiroom:radio:error:"

// radioRequest is what a test server saw.
type radioRequest struct {
	method   string
	path     string
	rawPath  string
	rawQuery string
	query    url.Values
}

// radioHub serves every request with one fixed answer and records the requests.
type radioHub struct {
	srv      *httptest.Server
	requests []radioRequest
}

func newRadioHub(t *testing.T, status int, headers map[string]string, body string) *radioHub {
	t.Helper()
	h := &radioHub{}
	h.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.requests = append(h.requests, radioRequest{
			method: r.Method, path: r.URL.Path, rawPath: r.URL.EscapedPath(), rawQuery: r.URL.RawQuery, query: r.URL.Query(),
		})
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(h.srv.Close)
	return h
}

func radioProblem(typ string, status int, detail, parameter string) string {
	m := map[string]any{"type": radioPrefix + typ, "title": "problem", "detail": detail, "status": status}
	if parameter != "" {
		m["parameter"] = parameter
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func sp(s string) *string { return &s }
func ip(i int) *int       { return &i }

// stationFixture holds every RadioStation property, with the characters a URL
// and a terminal care about in the free-text fields.
const stationFixture = `{
  "uuid": "9617a958-0601-11e8-ae97-52543be04c81",
  "uri": "radiobrowser://9617a958-0601-11e8-ae97-52543be04c81",
  "name": "Kiss FM",
  "favicon": "https://example.com/icon.png",
  "homepage": "https://www.kissfm.ua/",
  "countryCode": "UA",
  "languages": ["ukrainian"],
  "tags": ["dance", "electronic"],
  "codec": "MP3",
  "bitrate": 128,
  "hls": false,
  "lastCheckOk": true,
  "votes": 1520,
  "clickCount": 37
}`

const bareStationFixture = `{
  "uuid": "00000000-0000-0000-0000-000000000001",
  "uri": "radiobrowser://00000000-0000-0000-0000-000000000001",
  "name": null, "favicon": null, "homepage": null, "countryCode": null, "codec": null, "bitrate": null,
  "languages": [], "tags": [],
  "hls": true, "lastCheckOk": false, "votes": 0, "clickCount": 0
}`

func stationPageFixture(next string) string {
	return `{"stations":[` + stationFixture + `,` + bareStationFixture + `],"nextOffset":` + next + `}`
}

func TestSearchStations_NoFilters_NoQueryString(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("20"))

	if _, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(h.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(h.requests))
	}
	r := h.requests[0]
	if r.method != http.MethodGet || r.path != "/api/radio/stations" || r.rawQuery != "" {
		t.Errorf("request = %+v, want GET /api/radio/stations with no query", r)
	}
}

func TestSearchStations_SendsEverySetFieldOnce(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("null"))
	order, dir := hub.RadioOrderClickCount, hub.RadioDirectionDesc

	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{
		Name: sp("kiss fm"), Tag: sp("pop"), CountryCode: sp("UA"), Language: sp("ukrainian"), Codec: sp("MP3"),
		BitrateMin: ip(128), Order: &order, Direction: &dir, Limit: ip(5), Offset: ip(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := url.Values{
		"name": {"kiss fm"}, "tag": {"pop"}, "countryCode": {"UA"}, "language": {"ukrainian"}, "codec": {"MP3"},
		"bitrateMin": {"128"}, "order": {"clickCount"}, "direction": {"desc"}, "limit": {"5"}, "offset": {"10"},
	}
	if got := h.requests[0].query; !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestSearchStations_OnlySetFieldsAreSent(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("null"))

	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{Tag: sp("jazz"), Offset: ip(0)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := url.Values{"tag": {"jazz"}, "offset": {"0"}}
	if got := h.requests[0].query; !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestSearchStations_ConstantsSendTheHubsWords(t *testing.T) {
	for constant, want := range map[string]string{
		hub.RadioOrderName: "name", hub.RadioOrderVotes: "votes", hub.RadioOrderClickCount: "clickCount", hub.RadioOrderBitrate: "bitrate",
		hub.RadioDirectionAsc: "asc", hub.RadioDirectionDesc: "desc",
	} {
		if constant != want {
			t.Errorf("constant = %q, want %q", constant, want)
		}
	}
}

func TestSearchStations_EncodesValuesAsUTF8(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("null"))
	// "Kyiv" in Ukrainian, written with byte escapes so the source holds no invisible or exotic characters.
	kyiv := "\xd0\x9a\xd0\xb8\xd1\x97\xd0\xb2"
	name := "a b&c+d=e " + kyiv

	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := h.requests[0]
	if got := r.query.Get("name"); got != name {
		t.Errorf("decoded name = %q, want %q", got, name)
	}
	for _, bad := range []string{" ", "&c", "+d", "=e", kyiv} {
		if strings.Contains(r.rawQuery, bad) {
			t.Errorf("raw query %q holds unencoded %q", r.rawQuery, bad)
		}
	}
	if !strings.Contains(r.rawQuery, "%D0%9A") {
		t.Errorf("raw query %q is not percent-encoded UTF-8", r.rawQuery)
	}
}

func TestSearchStations_DecodesEveryField(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("20"))

	page, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Stations) != 2 || page.NextOffset == nil || *page.NextOffset != 20 {
		t.Fatalf("page = %+v", page)
	}
	s := page.Stations[0]
	if s.UUID != "9617a958-0601-11e8-ae97-52543be04c81" || s.URI != "radiobrowser://9617a958-0601-11e8-ae97-52543be04c81" {
		t.Errorf("identity = %q %q", s.UUID, s.URI)
	}
	if s.Name == nil || *s.Name != "Kiss FM" || s.Favicon == nil || *s.Favicon != "https://example.com/icon.png" ||
		s.Homepage == nil || *s.Homepage != "https://www.kissfm.ua/" || s.CountryCode == nil || *s.CountryCode != "UA" ||
		s.Codec == nil || *s.Codec != "MP3" || s.Bitrate == nil || *s.Bitrate != 128 {
		t.Errorf("optional fields = %+v", s)
	}
	if !reflect.DeepEqual(s.Languages, []string{"ukrainian"}) || !reflect.DeepEqual(s.Tags, []string{"dance", "electronic"}) {
		t.Errorf("lists = %v %v", s.Languages, s.Tags)
	}
	if s.HLS || !s.LastCheckOK || s.Votes != 1520 || s.ClickCount != 37 {
		t.Errorf("flags and counts = %+v", s)
	}

	bare := page.Stations[1]
	if bare.Name != nil || bare.Favicon != nil || bare.Homepage != nil || bare.CountryCode != nil || bare.Codec != nil || bare.Bitrate != nil {
		t.Errorf("null fields must decode to nil: %+v", bare)
	}
	if bare.Languages == nil || len(bare.Languages) != 0 || bare.Tags == nil || len(bare.Tags) != 0 {
		t.Errorf("empty lists must decode to empty slices: %#v %#v", bare.Languages, bare.Tags)
	}
	if !bare.HLS || bare.LastCheckOK {
		t.Errorf("bare flags = %+v", bare)
	}
}

func TestSearchStations_FixtureAndMarshalKeysMatchSpec(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationPageFixture("20"))

	var raw struct {
		Stations []json.RawMessage `json:"stations"`
	}
	if err := json.Unmarshal([]byte(stationPageFixture("20")), &raw); err != nil {
		t.Fatal(err)
	}
	wantStation := specSchemaProperties(t, "RadioStation")
	for i, st := range raw.Stations {
		if got := keysOf(t, st); !reflect.DeepEqual(got, wantStation) {
			t.Errorf("fixture station %d keys = %v, spec has %v", i, got, wantStation)
		}
	}
	if got, want := keysOf(t, []byte(stationPageFixture("20"))), specSchemaProperties(t, "RadioStationPage"); !reflect.DeepEqual(got, want) {
		t.Errorf("fixture page keys = %v, spec has %v", got, want)
	}

	page, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, st := range page.Stations {
		data, err := json.Marshal(st)
		if err != nil {
			t.Fatal(err)
		}
		if got := keysOf(t, data); !reflect.DeepEqual(got, wantStation) {
			t.Errorf("marshalled station %d keys = %v, spec has %v", i, got, wantStation)
		}
	}
	data, _ := json.Marshal(page.Stations[1])
	var asMap map[string]any
	_ = json.Unmarshal(data, &asMap)
	for _, k := range []string{"name", "favicon", "homepage", "countryCode", "codec", "bitrate"} {
		if v, ok := asMap[k]; !ok || v != nil {
			t.Errorf("marshalled nil %s = %v (present %v), want null", k, v, ok)
		}
	}
}

func TestSearchStations_NextOffset_NullAndAbsentAreNil(t *testing.T) {
	for name, body := range map[string]string{
		"null":   `{"stations":[],"nextOffset":null}`,
		"absent": `{"stations":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			h := newRadioHub(t, 200, nil, body)
			page, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
			if err != nil {
				t.Fatalf("an empty page must not be an error: %v", err)
			}
			if page.NextOffset != nil || len(page.Stations) != 0 {
				t.Errorf("page = %+v", page)
			}
		})
	}
}

func TestSearchStations_ShapeViolations_DecodeError(t *testing.T) {
	without := func(key string) string {
		var m map[string]json.RawMessage
		_ = json.Unmarshal([]byte(stationFixture), &m)
		delete(m, key)
		b, _ := json.Marshal(m)
		return `{"stations":[` + string(b) + `]}`
	}
	withValue := func(key, value string) string {
		var m map[string]json.RawMessage
		_ = json.Unmarshal([]byte(stationFixture), &m)
		m[key] = json.RawMessage(value)
		b, _ := json.Marshal(m)
		return `{"stations":[` + string(b) + `]}`
	}
	cases := map[string]string{
		"missing stations": `{"nextOffset":null}`,
		"not an object":    `[]`,
		"not json":         `not json`,
		"empty uuid":       withValue("uuid", `""`),
		"empty uri":        withValue("uri", `""`),
		"wrong type":       withValue("votes", `"many"`),
		"null languages":   withValue("languages", `null`),
	}
	for _, key := range []string{"clickCount", "hls", "languages", "lastCheckOk", "tags", "uri", "uuid", "votes"} {
		cases["missing "+key] = without(key)
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h := newRadioHub(t, 200, nil, body)
			_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
			var de *hub.DecodeError
			if !errors.As(err, &de) {
				t.Fatalf("err = %T %v, want *hub.DecodeError", err, err)
			}
		})
	}
}

func TestSearchStations_Errors(t *testing.T) {
	t.Run("400 invalid parameter", func(t *testing.T) {
		h := newRadioHub(t, 400, nil, radioProblem("invalid-parameter", 400, "limit must be between 1 and 100", "limit"))
		_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{Limit: ip(0)})
		var ae *hub.APIError
		if !errors.As(err, &ae) {
			t.Fatalf("err = %T %v, want *hub.APIError", err, err)
		}
		if ae.StatusCode != 400 || ae.Type != radioPrefix+"invalid-parameter" || ae.Title != "problem" ||
			ae.Detail != "limit must be between 1 and 100" || ae.Parameter != "limit" || ae.RetryAfter != 0 {
			t.Errorf("APIError = %+v", ae)
		}
	})

	t.Run("503 with Retry-After", func(t *testing.T) {
		h := newRadioHub(t, 503, map[string]string{"Retry-After": "30"}, radioProblem("catalogue-unavailable", 503, "timed out", ""))
		_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
		var ae *hub.APIError
		if !errors.As(err, &ae) || ae.RetryAfter != 30*time.Second || ae.Type != radioPrefix+"catalogue-unavailable" {
			t.Fatalf("err = %#v", err)
		}
	})

	for name, headers := range map[string]map[string]string{
		"503 without Retry-After":      nil,
		"503 with a date Retry-After":  {"Retry-After": "Wed, 21 Oct 2026 07:28:00 GMT"},
		"503 with a negative delay":    {"Retry-After": "-5"},
		"503 with a non-numeric delay": {"Retry-After": "soon"},
	} {
		t.Run(name, func(t *testing.T) {
			h := newRadioHub(t, 503, headers, radioProblem("catalogue-unavailable", 503, "timed out", ""))
			_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
			var ae *hub.APIError
			if !errors.As(err, &ae) || ae.RetryAfter != 0 {
				t.Fatalf("err = %#v, want *hub.APIError with RetryAfter 0", err)
			}
		})
	}

	for status, typ := range map[int]string{502: "catalogue-bad-answer", 500: "internal-error", 405: "method-not-allowed"} {
		t.Run(typ, func(t *testing.T) {
			h := newRadioHub(t, status, nil, radioProblem(typ, status, "detail text", ""))
			_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
			var ae *hub.APIError
			if !errors.As(err, &ae) || ae.StatusCode != status || ae.Type != radioPrefix+typ {
				t.Fatalf("err = %#v", err)
			}
		})
	}

	t.Run("detail is put on one line", func(t *testing.T) {
		h := newRadioHub(t, 400, nil, radioProblem("invalid-parameter", 400, "line one\nline two", "name"))
		_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
		var ae *hub.APIError
		if !errors.As(err, &ae) || ae.Detail != "line one line two" {
			t.Fatalf("err = %#v", err)
		}
	})
}

func TestSearchStations_404_NotOffered(t *testing.T) {
	bodies := map[string]string{
		"core problem body":    `{"title":"Not Found","detail":"No static resource api/radio/stations.","status":404}`,
		"radio not-found body": radioProblem("not-found", 404, "no such path", ""),
		"empty body":           ``,
		"not json":             `<html>nope</html>`,
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			h := newRadioHub(t, 404, nil, body)
			_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
			var ne *hub.RadioNotOfferedError
			if !errors.As(err, &ne) {
				t.Fatalf("err = %T %v, want *hub.RadioNotOfferedError", err, err)
			}
		})
	}
}

func TestSearchStations_StationNotFound404_IsNotFoundError(t *testing.T) {
	h := newRadioHub(t, 404, nil, radioProblem("station-not-found", 404, "no station with that uuid", ""))
	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
	var nf *hub.NotFoundError
	if !errors.As(err, &nf) || nf.Resource != "radio station" {
		t.Fatalf("err = %#v, want *hub.NotFoundError for a radio station", err)
	}
}

func TestSearchStations_NonJSONErrorBody_StatusError(t *testing.T) {
	h := newRadioHub(t, 400, nil, `<html>bad</html>`)
	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})
	var se *hub.StatusError
	if !errors.As(err, &se) || se.StatusCode != 400 {
		t.Fatalf("err = %#v, want *hub.StatusError 400", err)
	}
}

func TestSearchStations_ErrorBodyIsReadOnlyUpTo64KiB(t *testing.T) {
	// A valid problem whose detail pushes the end of the body past 64 KiB: only the
	// first 64 KiB are read, so the JSON is cut short and cannot be decoded.
	body := `{"type":"` + radioPrefix + `invalid-parameter","title":"t","status":400,"detail":"` + strings.Repeat("a", 100*1024) + `"}`
	h := newRadioHub(t, 400, nil, body)

	_, err := hub.SearchStations(context.Background(), hub.NewClient(), h.srv.URL, hub.StationSearch{})

	var se *hub.StatusError
	if !errors.As(err, &se) {
		t.Fatalf("err = %T %v, want *hub.StatusError for a body cut at 64 KiB", err, err)
	}
}

func TestSearchStations_TransportError_ReturnedAsIs(t *testing.T) {
	h := newRadioHub(t, 200, nil, `{}`)
	url := h.srv.URL
	h.srv.Close()

	_, err := hub.SearchStations(context.Background(), hub.NewClient(), url, hub.StationSearch{})
	if err == nil {
		t.Fatal("expected a transport error")
	}
	var de *hub.DecodeError
	var se *hub.StatusError
	if errors.As(err, &de) || errors.As(err, &se) {
		t.Errorf("a transport error must not be wrapped as %T", err)
	}
}

func TestRadioTimeout_IsTenSeconds(t *testing.T) {
	if hub.RadioTimeout != 10*time.Second {
		t.Errorf("RadioTimeout = %v, want 10s", hub.RadioTimeout)
	}
}

func TestStation_PathIsEscapedAndNoQuery(t *testing.T) {
	h := newRadioHub(t, 200, nil, stationFixture)

	st, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "9617a958-0601-11e8-ae97-52543be04c81")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.UUID != "9617a958-0601-11e8-ae97-52543be04c81" || st.Name == nil || *st.Name != "Kiss FM" {
		t.Errorf("station = %+v", st)
	}
	if r := h.requests[0]; r.method != http.MethodGet || r.path != "/api/radio/stations/9617a958-0601-11e8-ae97-52543be04c81" || r.rawQuery != "" {
		t.Errorf("request = %+v", r)
	}

	for _, id := range []string{"a b", "what?x=1", "a#b", "a/b"} {
		h := newRadioHub(t, 200, nil, stationFixture)
		if _, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, id); err != nil {
			t.Fatalf("id %q: %v", id, err)
		}
		r := h.requests[0]
		if want := "/api/radio/stations/" + url.PathEscape(id); r.rawPath != want || r.rawQuery != "" {
			t.Errorf("id %q: raw path %q query %q, want path %q and no query", id, r.rawPath, r.rawQuery, want)
		}
	}
}

func TestStation_BrokenStationDecodes(t *testing.T) {
	h := newRadioHub(t, 200, nil, bareStationFixture)

	st, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.LastCheckOK {
		t.Error("LastCheckOK = true, want false")
	}
	if st.Name != nil || len(st.Tags) != 0 {
		t.Errorf("station = %+v", st)
	}
}

func TestStation_Errors(t *testing.T) {
	t.Run("station-not-found", func(t *testing.T) {
		h := newRadioHub(t, 404, nil, radioProblem("station-not-found", 404, "no such station", ""))
		_, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "abc")
		var nf *hub.NotFoundError
		if !errors.As(err, &nf) || nf.Resource != "radio station" || nf.ID != "abc" {
			t.Fatalf("err = %#v, want *hub.NotFoundError{radio station, abc}", err)
		}
	})
	t.Run("the id is reported as given, not escaped", func(t *testing.T) {
		h := newRadioHub(t, 404, nil, radioProblem("station-not-found", 404, "no such station", ""))
		_, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "a b")
		var nf *hub.NotFoundError
		if !errors.As(err, &nf) || nf.ID != "a b" {
			t.Fatalf("err = %#v", err)
		}
	})
	t.Run("core 404 body", func(t *testing.T) {
		h := newRadioHub(t, 404, nil, `{"title":"Not Found","status":404}`)
		_, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "abc")
		var ne *hub.RadioNotOfferedError
		if !errors.As(err, &ne) {
			t.Fatalf("err = %#v, want *hub.RadioNotOfferedError", err)
		}
	})
	t.Run("400 invalid parameter names uuid", func(t *testing.T) {
		h := newRadioHub(t, 400, nil, radioProblem("invalid-parameter", 400, "uuid is not a UUID", "uuid"))
		_, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "nope")
		var ae *hub.APIError
		if !errors.As(err, &ae) || ae.Parameter != "uuid" || ae.StatusCode != 400 {
			t.Fatalf("err = %#v", err)
		}
	})
	t.Run("missing required field", func(t *testing.T) {
		h := newRadioHub(t, 200, nil, `{"uuid":"u","uri":"radiobrowser://u"}`)
		_, err := hub.Station(context.Background(), hub.NewClient(), h.srv.URL, "u")
		var de *hub.DecodeError
		if !errors.As(err, &de) {
			t.Fatalf("err = %#v, want *hub.DecodeError", err)
		}
	})
}

const countriesFixture = `[{"code":"UA","name":"Ukraine","stationCount":120},{"code":"PL","name":"Poland","stationCount":90}]`
const languagesFixture = `[{"name":"ukrainian","code":"uk","stationCount":80},{"name":"klingon","code":null,"stationCount":1}]`
const tagPageFixture = `{"tags":[{"name":"jazz","stationCount":500},{"name":"smooth jazz","stationCount":40}],"nextOffset":20}`

func firstKeys(t *testing.T, listJSON string) []string {
	t.Helper()
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(listJSON), &items); err != nil {
		t.Fatal(err)
	}
	return keysOf(t, items[0])
}

func TestCountries_DecodesInTheHubsOrder(t *testing.T) {
	h := newRadioHub(t, 200, nil, countriesFixture)

	list, err := hub.Countries(context.Background(), hub.NewClient(), h.srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []hub.RadioCountry{{Code: "UA", Name: "Ukraine", StationCount: 120}, {Code: "PL", Name: "Poland", StationCount: 90}}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("countries = %+v, want %+v", list, want)
	}
	if r := h.requests[0]; r.method != http.MethodGet || r.path != "/api/radio/countries" || r.rawQuery != "" {
		t.Errorf("request = %+v", r)
	}
	if got, want := firstKeys(t, countriesFixture), specSchemaProperties(t, "RadioCountry"); !reflect.DeepEqual(got, want) {
		t.Errorf("fixture keys = %v, spec has %v", got, want)
	}
	data, _ := json.Marshal(list[0])
	if got, want := keysOf(t, data), specSchemaProperties(t, "RadioCountry"); !reflect.DeepEqual(got, want) {
		t.Errorf("marshalled keys = %v, spec has %v", got, want)
	}
}

func TestLanguages_DecodesNullableCode(t *testing.T) {
	h := newRadioHub(t, 200, nil, languagesFixture)

	list, err := hub.Languages(context.Background(), hub.NewClient(), h.srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 || list[0].Name != "ukrainian" || list[0].Code == nil || *list[0].Code != "uk" || list[0].StationCount != 80 {
		t.Errorf("first language = %+v", list[0])
	}
	if list[1].Code != nil {
		t.Errorf("a null code must decode to nil, got %q", *list[1].Code)
	}
	if r := h.requests[0]; r.path != "/api/radio/languages" || r.rawQuery != "" {
		t.Errorf("request = %+v", r)
	}
	if got, want := firstKeys(t, languagesFixture), specSchemaProperties(t, "RadioLanguage"); !reflect.DeepEqual(got, want) {
		t.Errorf("fixture keys = %v, spec has %v", got, want)
	}
	data, _ := json.Marshal(list[1])
	if got, want := keysOf(t, data), specSchemaProperties(t, "RadioLanguage"); !reflect.DeepEqual(got, want) {
		t.Errorf("marshalled keys = %v, spec has %v", got, want)
	}
}

func TestTags_SendsOnlySetFieldsOnce(t *testing.T) {
	h := newRadioHub(t, 200, nil, tagPageFixture)
	if _, err := hub.Tags(context.Background(), hub.NewClient(), h.srv.URL, hub.TagSearch{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r := h.requests[0]; r.path != "/api/radio/tags" || r.rawQuery != "" {
		t.Errorf("no filters: request = %+v", r)
	}

	h = newRadioHub(t, 200, nil, tagPageFixture)
	if _, err := hub.Tags(context.Background(), hub.NewClient(), h.srv.URL, hub.TagSearch{Name: sp("smooth jazz"), Limit: ip(10), Offset: ip(20)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := url.Values{"name": {"smooth jazz"}, "limit": {"10"}, "offset": {"20"}}
	if got := h.requests[0].query; !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestTags_DecodesPage(t *testing.T) {
	h := newRadioHub(t, 200, nil, tagPageFixture)

	page, err := hub.Tags(context.Background(), hub.NewClient(), h.srv.URL, hub.TagSearch{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Tags) != 2 || page.Tags[0] != (hub.RadioTag{Name: "jazz", StationCount: 500}) || page.NextOffset == nil || *page.NextOffset != 20 {
		t.Errorf("page = %+v", page)
	}
	if got, want := keysOf(t, []byte(tagPageFixture)), specSchemaProperties(t, "RadioTagPage"); !reflect.DeepEqual(got, want) {
		t.Errorf("fixture page keys = %v, spec has %v", got, want)
	}
	var raw struct {
		Tags []json.RawMessage `json:"tags"`
	}
	_ = json.Unmarshal([]byte(tagPageFixture), &raw)
	if got, want := keysOf(t, raw.Tags[0]), specSchemaProperties(t, "RadioTag"); !reflect.DeepEqual(got, want) {
		t.Errorf("fixture tag keys = %v, spec has %v", got, want)
	}
	data, _ := json.Marshal(page)
	if got, want := keysOf(t, data), specSchemaProperties(t, "RadioTagPage"); !reflect.DeepEqual(got, want) {
		t.Errorf("marshalled page keys = %v, spec has %v", got, want)
	}

	last := newRadioHub(t, 200, nil, `{"tags":[]}`)
	page, err = hub.Tags(context.Background(), hub.NewClient(), last.srv.URL, hub.TagSearch{})
	if err != nil || page.NextOffset != nil || page.Tags == nil || len(page.Tags) != 0 {
		t.Errorf("an empty last page = %+v, %v", page, err)
	}
}

func TestRadioLists_ShapeViolations_DecodeError(t *testing.T) {
	call := map[string]func(url string) error{
		"countries": func(u string) error { _, err := hub.Countries(context.Background(), hub.NewClient(), u); return err },
		"languages": func(u string) error { _, err := hub.Languages(context.Background(), hub.NewClient(), u); return err },
		"tags": func(u string) error {
			_, err := hub.Tags(context.Background(), hub.NewClient(), u, hub.TagSearch{})
			return err
		},
	}
	bodies := map[string]map[string]string{
		"countries": {
			"object instead of an array": `{"countries":[]}`,
			"missing code":               `[{"name":"Ukraine","stationCount":1}]`,
			"missing name":               `[{"code":"UA","stationCount":1}]`,
			"missing stationCount":       `[{"code":"UA","name":"Ukraine"}]`,
			"not json":                   `nope`,
		},
		"languages": {
			"object instead of an array": `{}`,
			"missing name":               `[{"code":"uk","stationCount":1}]`,
			"missing stationCount":       `[{"name":"ukrainian","code":"uk"}]`,
		},
		"tags": {
			"missing tags":          `{"nextOffset":null}`,
			"array instead of page": `[]`,
			"missing name":          `{"tags":[{"stationCount":1}]}`,
			"missing stationCount":  `{"tags":[{"name":"jazz"}]}`,
		},
	}
	for kind, cases := range bodies {
		for name, body := range cases {
			t.Run(kind+"/"+name, func(t *testing.T) {
				h := newRadioHub(t, 200, nil, body)
				var de *hub.DecodeError
				if err := call[kind](h.srv.URL); !errors.As(err, &de) {
					t.Fatalf("err = %#v, want *hub.DecodeError", err)
				}
			})
		}
	}
}

func TestRadioLists_Errors(t *testing.T) {
	call := map[string]func(url string) error{
		"countries": func(u string) error { _, err := hub.Countries(context.Background(), hub.NewClient(), u); return err },
		"languages": func(u string) error { _, err := hub.Languages(context.Background(), hub.NewClient(), u); return err },
		"tags": func(u string) error {
			_, err := hub.Tags(context.Background(), hub.NewClient(), u, hub.TagSearch{})
			return err
		},
	}
	for kind, do := range call {
		t.Run(kind+" 404 core body", func(t *testing.T) {
			h := newRadioHub(t, 404, nil, `{"title":"Not Found","status":404}`)
			var ne *hub.RadioNotOfferedError
			if err := do(h.srv.URL); !errors.As(err, &ne) {
				t.Fatalf("err = %#v, want *hub.RadioNotOfferedError", err)
			}
		})
		t.Run(kind+" 503", func(t *testing.T) {
			h := newRadioHub(t, 503, map[string]string{"Retry-After": "7"}, radioProblem("catalogue-unavailable", 503, "slow", ""))
			var ae *hub.APIError
			if err := do(h.srv.URL); !errors.As(err, &ae) || ae.RetryAfter != 7*time.Second || ae.StatusCode != 503 {
				t.Fatalf("err = %#v", err)
			}
		})
	}
}
