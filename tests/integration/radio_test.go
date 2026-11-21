package integration

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

const radioErrPrefix = "urn:multiroom:radio:error:"

// mockRadioHub fakes the radio paths and /api/v2/extensions, and records every
// request so a test can assert exactly what the CLI sent and how often.
type mockRadioHub struct {
	mu sync.Mutex

	// answer is served for every /api/radio/ request.
	status  int
	headers map[string]string
	body    string

	extStatus int
	extBody   string

	requests []string // "path?rawquery" of each radio request
	queries  []url.Values
	extCount int
}

func newMockRadioHub(t *testing.T, status int, body string) (*httptest.Server, *mockRadioHub) {
	t.Helper()
	m := &mockRadioHub{status: status, body: body, extStatus: 200, extBody: `{"extensions":[]}`}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v2/extensions" {
			m.extCount++
			w.WriteHeader(m.extStatus)
			_, _ = w.Write([]byte(m.extBody))
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/radio/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.requests = append(m.requests, r.URL.Path+"?"+r.URL.RawQuery)
		m.queries = append(m.queries, r.URL.Query())
		for k, v := range m.headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(m.status)
		_, _ = w.Write([]byte(m.body))
	}))
	t.Cleanup(srv.Close)
	return srv, m
}

func radioProblemBody(typ string, status int, detail string) string {
	return `{"type":"` + radioErrPrefix + typ + `","title":"t","detail":"` + detail + `","status":` + strconv.Itoa(status) + `}`
}

const kissStationJSON = `{"uuid":"9617a958-0601-11e8-ae97-52543be04c81","uri":"radiobrowser://9617a958-0601-11e8-ae97-52543be04c81","name":"Kiss FM","favicon":null,"homepage":"https://www.kissfm.ua/","countryCode":"UA","languages":["ukrainian"],"tags":["dance","electronic"],"codec":"MP3","bitrate":128,"hls":false,"lastCheckOk":true,"votes":1520,"clickCount":37}`

func kissStation() hub.RadioStation {
	name, home, cc, codec := "Kiss FM", "https://www.kissfm.ua/", "UA", "MP3"
	br := int32(128)
	return hub.RadioStation{
		UUID: "9617a958-0601-11e8-ae97-52543be04c81", URI: "radiobrowser://9617a958-0601-11e8-ae97-52543be04c81",
		Name: &name, Homepage: &home, CountryCode: &cc, Languages: []string{"ukrainian"}, Tags: []string{"dance", "electronic"},
		Codec: &codec, Bitrate: &br, LastCheckOK: true, Votes: 1520, ClickCount: 37,
	}
}

const kissPageJSON = `{"stations":[` + kissStationJSON + `],"nextOffset":20}`

func kissPage() hub.RadioStationPage {
	next := int32(20)
	return hub.RadioStationPage{Stations: []hub.RadioStation{kissStation()}, NextOffset: &next}
}

func (m *mockRadioHub) radioRequests() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

func TestSearchRadio_AllFlagsSentOnce(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, kissPageJSON)

	res := runCLI(t, "search", "radio-stations", "kiss", "--tag", "pop", "--country", "UA", "--language", "ukrainian",
		"--codec", "MP3", "--min-bitrate", "128", "--order", "bitrate", "--desc", "--limit", "5", "--offset", "10", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	want := url.Values{
		"name": {"kiss"}, "tag": {"pop"}, "countryCode": {"UA"}, "language": {"ukrainian"}, "codec": {"MP3"},
		"bitrateMin": {"128"}, "order": {"bitrate"}, "direction": {"desc"}, "limit": {"5"}, "offset": {"10"},
	}
	if len(m.queries) != 1 || !reflect.DeepEqual(m.queries[0], want) {
		t.Errorf("queries = %v, want [%v]", m.queries, want)
	}
}

func TestSearchRadio_FlagsBeforeAndAfterName(t *testing.T) {
	for name, args := range map[string][]string{
		"flags before": {"search", "radio-stations", "--country", "UA", "kiss"},
		"flags after":  {"search", "radio-stations", "kiss", "--country", "UA"},
		"name between": {"search", "radio-stations", "--limit", "3", "kiss", "--country", "UA"},
	} {
		t.Run(name, func(t *testing.T) {
			srv, m := newMockRadioHub(t, 200, kissPageJSON)
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 0 {
				t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
			}
			if got := m.queries[0]; got.Get("name") != "kiss" || got.Get("countryCode") != "UA" {
				t.Errorf("query = %v", got)
			}
		})
	}
}

func TestSearchRadio_NoFlags_NoQueryString(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, kissPageJSON)

	res := runCLI(t, "search", "radio-stations", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
	}
	if !reflect.DeepEqual(m.requests, []string{"/api/radio/stations?"}) {
		t.Errorf("requests = %v, want one with no query string", m.requests)
	}
}

func TestSearchRadio_AscAndEmptyString(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, kissPageJSON)

	res := runCLI(t, "search", "radio-stations", "--asc", "--tag", "", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
	}
	want := url.Values{"direction": {"asc"}}
	if !reflect.DeepEqual(m.queries[0], want) {
		t.Errorf("query = %v, want %v (an empty --tag is not sent)", m.queries[0], want)
	}
}

func TestSearchRadio_UsageErrors_NoRequest(t *testing.T) {
	cases := map[string][]string{
		"two positionals": {"search", "radio-stations", "kiss", "fm"},
		"asc and desc":    {"search", "radio-stations", "--asc", "--desc"},
		"limit 0":         {"search", "radio-stations", "--limit", "0"},
		"limit 101":       {"search", "radio-stations", "--limit", "101"},
		"offset -1":       {"search", "radio-stations", "--offset", "-1"},
		"offset 10001":    {"search", "radio-stations", "--offset", "10001"},
		"limit not int":   {"search", "radio-stations", "--limit", "abc"},
		"unknown flag":    {"search", "radio-stations", "--nope"},
		"no resource":     {"search"},
		"radio-tags":      {"search", "radio-tags"},
		"other resource":  {"search", "outputs"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			srv, m := newMockRadioHub(t, 200, kissPageJSON)
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("stdout = %q, want empty", res.stdout)
			}
			if n := m.radioRequests(); n != 0 || m.extCount != 0 {
				t.Errorf("requests = %d radio, %d inventory; want none", n, m.extCount)
			}
		})
	}
}

func TestSearchRadio_UsageLineOnMissingOrWrongResource(t *testing.T) {
	for _, args := range [][]string{{"search"}, {"search", "radio-tags"}} {
		res := runCLI(t, args...)
		if res.exitCode != 2 || !strings.Contains(res.stderr, "usage: sonora search radio-stations [name] [flags]") {
			t.Errorf("%v: exit %d, stderr %q", args, res.exitCode, res.stderr)
		}
	}
}

func TestSearchRadio_FlagRangeErrorsNameTheFlag(t *testing.T) {
	for flagName, args := range map[string][]string{
		"--limit":  {"search", "radio-stations", "--limit", "101"},
		"--offset": {"search", "radio-stations", "--offset", "10001"},
		"--asc":    {"search", "radio-stations", "--asc", "--desc"},
	} {
		res := runCLI(t, args...)
		if !strings.Contains(res.stderr, flagName) {
			t.Errorf("%v: stderr %q does not name %s", args, res.stderr, flagName)
		}
	}
}

func TestSearchRadio_Help(t *testing.T) {
	for _, args := range [][]string{{"search", "--help"}, {"search", "radio-stations", "--help"}} {
		res := runCLI(t, args...)
		if res.exitCode != 0 || !strings.Contains(res.stdout, "usage: sonora search radio-stations [name]") {
			t.Errorf("%v: exit %d, stdout %q", args, res.exitCode, res.stdout)
		}
	}
	res := runCLI(t, "search", "radio-stations", "--help")
	for _, want := range []string{"--tag", "--country", "--language", "--codec", "--min-bitrate", "--order", "--asc", "--desc", "--limit", "--offset", "--json"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("help does not list %s:\n%s", want, res.stdout)
		}
	}
}

func TestSearchRadio_OutputMatchesRenderer(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, kissPageJSON)

	yaml := runCLI(t, "search", "radio-stations", "kiss", "--hub-url", srv.URL)
	js := runCLI(t, "search", "radio-stations", "kiss", "--json", "--hub-url", srv.URL)

	if yaml.exitCode != 0 || yaml.stdout != render.RenderStationPageYAML(kissPage()) {
		t.Errorf("YAML: exit %d, stdout\n%s", yaml.exitCode, yaml.stdout)
	}
	if js.exitCode != 0 || js.stdout != render.RenderStationPageJSON(kissPage()) {
		t.Errorf("JSON: exit %d, stdout\n%s", js.exitCode, js.stdout)
	}
	if m.radioRequests() != 2 || m.extCount != 0 {
		t.Errorf("a successful search is one request each: radio %d, inventory %d", m.radioRequests(), m.extCount)
	}
}

func TestSearchRadio_EmptyPage_ExitZero(t *testing.T) {
	srv, _ := newMockRadioHub(t, 200, `{"stations":[],"nextOffset":null}`)

	res := runCLI(t, "search", "radio-stations", "zzz", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
	}
	if want := "# no stations found\nstations: []\nnextOffset: null\n"; res.stdout != want {
		t.Errorf("stdout = %q, want %q", res.stdout, want)
	}
}

func TestSearchRadio_Failures(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		headers    map[string]string
		body       string
		wantExit   int
		wantStderr string
		wantExt    int
	}{
		{"400 invalid parameter", 400, nil, radioProblemBody("invalid-parameter", 400, "limit must be between 1 and 100"), 6, "error: limit must be between 1 and 100 (hub URL: ", 0},
		{"502 bad answer", 502, nil, radioProblemBody("catalogue-bad-answer", 502, "not a list"), 9, "error: the radio catalogue gave an unusable answer: not a list (hub URL: ", 0},
		{"503 with retry", 503, map[string]string{"Retry-After": "30"}, radioProblemBody("catalogue-unavailable", 503, "timed out"), 10, "error: the radio catalogue is unavailable; retry in 30 s (hub URL: ", 0},
		{"503 without retry", 503, nil, radioProblemBody("catalogue-unavailable", 503, "timed out"), 10, "error: the radio catalogue is unavailable; retry later (hub URL: ", 0},
		{"500 radio problem", 500, nil, radioProblemBody("internal-error", 500, "it broke"), 3, "error: it broke (hub URL: ", 0},
		{"405 radio problem", 405, nil, radioProblemBody("method-not-allowed", 405, "wrong method"), 3, "error: wrong method (hub URL: ", 0},
		{"malformed 200", 200, nil, `{"nope":true}`, 3, "error: hub returned an unexpected or malformed response (hub URL: ", 0},
		{"400 not json", 400, nil, `<html>`, 3, "error: hub reported an error (HTTP 400) (hub URL: ", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, m := newMockRadioHub(t, c.status, c.body)
			m.headers = c.headers
			res := runCLI(t, "search", "radio-stations", "kiss", "--hub-url", srv.URL)
			if res.exitCode != c.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", res.exitCode, c.wantExit, res.stderr)
			}
			if !strings.HasPrefix(res.stderr, c.wantStderr) {
				t.Errorf("stderr = %q, want prefix %q", res.stderr, c.wantStderr)
			}
			if res.stdout != "" {
				t.Errorf("stdout = %q, want empty", res.stdout)
			}
			if m.radioRequests() != 1 || m.extCount != c.wantExt {
				t.Errorf("requests: radio %d (want 1), inventory %d (want %d)", m.radioRequests(), m.extCount, c.wantExt)
			}
		})
	}
}

func TestSearchRadio_RadioNotServed_ExitFifteen(t *testing.T) {
	srv, m := newMockRadioHub(t, 404, `{"title":"Not Found","status":404}`)
	m.extBody = `{"extensions":[{"id":"radio","status":"DISABLED"}]}`

	res := runCLI(t, "search", "radio-stations", "kiss", "--hub-url", srv.URL)

	if res.exitCode != 15 {
		t.Fatalf("exit code = %d, want 15; stderr: %s", res.exitCode, res.stderr)
	}
	want := "error: radio is not available on this hub: the radio extension is installed but disabled in the hub's configuration (hub URL: " + srv.URL + ")\n"
	if res.stderr != want {
		t.Errorf("stderr = %q, want %q", res.stderr, want)
	}
	if m.radioRequests() != 1 || m.extCount != 1 {
		t.Errorf("requests: radio %d, inventory %d; want 1 and 1", m.radioRequests(), m.extCount)
	}
}

func TestSearchRadio_RadioNotServed_InventoryAlso404_ExitFour(t *testing.T) {
	srv, m := newMockRadioHub(t, 404, ``)
	m.extStatus = 404

	res := runCLI(t, "search", "radio-stations", "--hub-url", srv.URL)

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	for _, want := range []string{srv.URL + " is not serving the Multiroom Audio Hub API", "--hub-url", "MULTIROOM_URL"} {
		if !strings.Contains(res.stderr, want) {
			t.Errorf("stderr %q lacks %q", res.stderr, want)
		}
	}
}

func TestSearchRadio_HubUnreachable_ExitFour(t *testing.T) {
	srv, _ := newMockRadioHub(t, 200, kissPageJSON)
	url := srv.URL
	srv.Close()

	res := runCLI(t, "search", "radio-stations", "--hub-url", url)

	if res.exitCode != 4 || !strings.Contains(res.stderr, "could not reach the hub") {
		t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
	}
}

func TestSearchRadio_VerboseKeepsTheCause(t *testing.T) {
	srv, _ := newMockRadioHub(t, 400, radioProblemBody("invalid-parameter", 400, "bad limit"))

	res := runCLI(t, "search", "radio-stations", "--verbose", "--hub-url", srv.URL)

	if res.exitCode != 6 || !strings.Contains(res.stderr, "detail: bad limit") {
		t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
	}
}

const kissUUID = "9617a958-0601-11e8-ae97-52543be04c81"

func TestGetRadioStation_PrintsTheStation(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, kissStationJSON)

	yaml := runCLI(t, "get", "radio-stations/"+kissUUID, "--hub-url", srv.URL)
	js := runCLI(t, "get", "radio-stations/"+kissUUID, "--json", "--hub-url", srv.URL)

	if yaml.exitCode != 0 || yaml.stdout != render.RenderStationYAML(kissStation()) {
		t.Errorf("YAML: exit %d, stdout\n%s\nstderr %s", yaml.exitCode, yaml.stdout, yaml.stderr)
	}
	if js.exitCode != 0 || js.stdout != render.RenderStationJSON(kissStation()) {
		t.Errorf("JSON: exit %d, stdout\n%s", js.exitCode, js.stdout)
	}
	if want := []string{"/api/radio/stations/" + kissUUID + "?", "/api/radio/stations/" + kissUUID + "?"}; !reflect.DeepEqual(m.requests, want) || m.extCount != 0 {
		t.Errorf("requests = %v, inventory %d", m.requests, m.extCount)
	}
}

func TestGetRadioStation_UnknownUUID_ExitFive(t *testing.T) {
	srv, _ := newMockRadioHub(t, 404, radioProblemBody("station-not-found", 404, "no such station"))

	res := runCLI(t, "get", "radio-stations/"+kissUUID, "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr %s", res.exitCode, res.stderr)
	}
	if want := "error: radio station not found: " + kissUUID + " (hub URL: " + srv.URL + ")\n"; res.stderr != want {
		t.Errorf("stderr = %q, want %q", res.stderr, want)
	}
}

func TestGetRadioStation_HubRejectsTheUUID_ExitSix(t *testing.T) {
	srv, _ := newMockRadioHub(t, 400, radioProblemBody("invalid-parameter", 400, "uuid is not a valid UUID"))

	res := runCLI(t, "get", "radio-stations/nope", "--hub-url", srv.URL)

	if res.exitCode != 6 || !strings.HasPrefix(res.stderr, "error: uuid is not a valid UUID (hub URL: ") {
		t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
	}
}

func TestGetRadioStation_RadioNotServed_ExitFifteen(t *testing.T) {
	srv, m := newMockRadioHub(t, 404, `{"title":"Not Found","status":404}`)

	res := runCLI(t, "get", "radio-stations/"+kissUUID, "--hub-url", srv.URL)

	if res.exitCode != 15 || !strings.HasPrefix(res.stderr, "error: radio is not available on this hub") {
		t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
	}
	if m.extCount != 1 {
		t.Errorf("inventory requests = %d, want 1", m.extCount)
	}
}

func TestGetRadioStation_UsageErrors_NoRequest(t *testing.T) {
	cases := map[string][]string{
		"empty id":          {"get", "radio-stations/"},
		"extra segment":     {"get", "radio-stations/a/b"},
		"list with an id":   {"list", "radio-stations/" + kissUUID},
		"bare get":          {"get", "radio-stations"},
		"bare list":         {"list", "radio-stations"},
		"unexpected arg":    {"get", "radio-stations/" + kissUUID, "extra"},
		"search-only flag":  {"get", "radio-stations/" + kissUUID, "--tag", "pop"},
		"get with a search": {"get", "radio-stations", "kiss"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			srv, m := newMockRadioHub(t, 200, kissStationJSON)
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
			}
			if m.radioRequests() != 0 || m.extCount != 0 {
				t.Errorf("requests: radio %d, inventory %d; want none", m.radioRequests(), m.extCount)
			}
		})
	}
	for _, args := range [][]string{{"get", "radio-stations"}, {"list", "radio-stations"}} {
		res := runCLI(t, args...)
		if !strings.Contains(res.stderr, "error: use 'sonora search radio-stations' to find stations") {
			t.Errorf("%v: stderr %q does not point to search", args, res.stderr)
		}
	}
}

func radioListCountries() []hub.RadioCountry {
	return []hub.RadioCountry{{Code: "UA", Name: "Ukraine", StationCount: 120}}
}

const countriesJSON = `[{"code":"UA","name":"Ukraine","stationCount":120}]`
const languagesJSON = `[{"name":"ukrainian","code":"uk","stationCount":80}]`
const tagsJSON = `{"tags":[{"name":"jazz","stationCount":500}],"nextOffset":20}`

func TestRadioLists_GetAndListBothWork(t *testing.T) {
	next := int32(20)
	cases := []struct {
		noun, body, wantPath, wantYAML, wantJSON string
	}{
		{"radio-countries", countriesJSON, "/api/radio/countries", render.RenderCountriesYAML(radioListCountries()), render.RenderCountriesJSON(radioListCountries())},
		{"radio-languages", languagesJSON, "/api/radio/languages",
			render.RenderLanguagesYAML([]hub.RadioLanguage{{Name: "ukrainian", Code: strPtr("uk"), StationCount: 80}}),
			render.RenderLanguagesJSON([]hub.RadioLanguage{{Name: "ukrainian", Code: strPtr("uk"), StationCount: 80}})},
		{"radio-tags", tagsJSON, "/api/radio/tags",
			render.RenderTagPageYAML(hub.RadioTagPage{Tags: []hub.RadioTag{{Name: "jazz", StationCount: 500}}, NextOffset: &next}),
			render.RenderTagPageJSON(hub.RadioTagPage{Tags: []hub.RadioTag{{Name: "jazz", StationCount: 500}}, NextOffset: &next})},
	}
	for _, c := range cases {
		for _, verb := range []string{"get", "list"} {
			t.Run(verb+" "+c.noun, func(t *testing.T) {
				srv, m := newMockRadioHub(t, 200, c.body)
				yaml := runCLI(t, verb, c.noun, "--hub-url", srv.URL)
				js := runCLI(t, verb, c.noun, "--json", "--hub-url", srv.URL)
				if yaml.exitCode != 0 || yaml.stdout != c.wantYAML {
					t.Errorf("YAML: exit %d, stdout\n%s\nstderr %s", yaml.exitCode, yaml.stdout, yaml.stderr)
				}
				if js.exitCode != 0 || js.stdout != c.wantJSON {
					t.Errorf("JSON: exit %d, stdout\n%s", js.exitCode, js.stdout)
				}
				if want := []string{c.wantPath + "?", c.wantPath + "?"}; !reflect.DeepEqual(m.requests, want) || m.extCount != 0 {
					t.Errorf("requests = %v, inventory %d", m.requests, m.extCount)
				}
			})
		}
	}
}

func strPtr(s string) *string { return &s }

func TestRadioTags_SendsNameLimitOffset(t *testing.T) {
	srv, m := newMockRadioHub(t, 200, tagsJSON)

	res := runCLI(t, "get", "radio-tags", "--name", "jazz", "--limit", "10", "--offset", "20", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr %s", res.exitCode, res.stderr)
	}
	want := url.Values{"name": {"jazz"}, "limit": {"10"}, "offset": {"20"}}
	if len(m.queries) != 1 || !reflect.DeepEqual(m.queries[0], want) {
		t.Errorf("queries = %v, want [%v]", m.queries, want)
	}
}

func TestRadioLists_UsageErrors_NoRequest(t *testing.T) {
	cases := map[string][]string{
		"tags limit 0":            {"get", "radio-tags", "--limit", "0"},
		"tags limit 101":          {"get", "radio-tags", "--limit", "101"},
		"tags offset -1":          {"list", "radio-tags", "--offset", "-1"},
		"tags offset 10001":       {"list", "radio-tags", "--offset", "10001"},
		"countries with an id":    {"get", "radio-countries/UA"},
		"languages with an id":    {"get", "radio-languages/uk"},
		"tags with an id":         {"get", "radio-tags/jazz"},
		"list countries with id":  {"list", "radio-countries/UA"},
		"countries positional":    {"get", "radio-countries", "UA"},
		"tags positional":         {"get", "radio-tags", "jazz"},
		"countries --name":        {"get", "radio-countries", "--name", "x"},
		"countries --limit":       {"get", "radio-countries", "--limit", "5"},
		"countries --offset":      {"get", "radio-countries", "--offset", "5"},
		"languages --name":        {"list", "radio-languages", "--name", "x"},
		"languages --limit":       {"list", "radio-languages", "--limit", "5"},
		"languages --offset":      {"list", "radio-languages", "--offset", "5"},
		"tags with a search flag": {"get", "radio-tags", "--tag", "x"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			srv, m := newMockRadioHub(t, 200, tagsJSON)
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
			}
			if m.radioRequests() != 0 || m.extCount != 0 {
				t.Errorf("requests: radio %d, inventory %d; want none", m.radioRequests(), m.extCount)
			}
		})
	}
}

func TestRadioLists_Failures(t *testing.T) {
	for _, noun := range []string{"radio-countries", "radio-languages", "radio-tags"} {
		t.Run(noun+" not served", func(t *testing.T) {
			srv, m := newMockRadioHub(t, 404, `{"title":"Not Found","status":404}`)
			res := runCLI(t, "get", noun, "--hub-url", srv.URL)
			if res.exitCode != 15 || !strings.HasPrefix(res.stderr, "error: radio is not available on this hub") {
				t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
			}
			if m.extCount != 1 {
				t.Errorf("inventory requests = %d, want 1", m.extCount)
			}
		})
		t.Run(noun+" 503", func(t *testing.T) {
			srv, m := newMockRadioHub(t, 503, radioProblemBody("catalogue-unavailable", 503, "slow"))
			m.headers = map[string]string{"Retry-After": "12"}
			res := runCLI(t, "get", noun, "--hub-url", srv.URL)
			if res.exitCode != 10 || !strings.HasPrefix(res.stderr, "error: the radio catalogue is unavailable; retry in 12 s (hub URL: ") {
				t.Errorf("exit %d, stderr %q", res.exitCode, res.stderr)
			}
		})
	}
}
