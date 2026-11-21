package unit

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/render"
)

func TestSafeText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Kiss FM", "Kiss FM"},
		{"Київ", "Київ"},
		{"a\x1b[2Jb", `a\x1b[2Jb`},
		{"bell\a", `bell\a`},
		{"del\x7f", `del\x7f`},
		{"c1\u009bx", `c1\u009bx`},
		{"c1\u0080", `c1\u0080`},
		{"flip\u202eme", `flip\u202eme`},
		{"zero\u200bwidth", `zero\u200bwidth`},
		{"two\nlines", `two\nlines`},
		{"cr\rback", `cr\rback`},
		{"tab\there", `tab\there`},
		{"bad\xffbyte", `bad\xffbyte`},
	}
	for _, c := range cases {
		if got := render.SafeText(c.in); got != c.want {
			t.Errorf("SafeText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// hostile holds the characters a terminal acts on, written as byte escapes so this
// source file holds no invisible characters: ESC, BEL, a C1 control (U+009B) and a
// right-to-left override (U+202E).
const hostile = "x\x1b[2J\x07\xc2\x9b\xe2\x80\xaey"

// assertTerminalSafe fails if out holds a rune a terminal could act on.
func assertTerminalSafe(t *testing.T, label, out string) {
	t.Helper()
	for _, r := range out {
		switch {
		case r == '\x1b', r == '\x07', r == 0x7f, r >= 0x80 && r <= 0x9f, unicode.Is(unicode.Cf, r):
			t.Errorf("%s holds raw rune %U", label, r)
		case r < ' ' && r != '\n':
			t.Errorf("%s holds raw control rune %U", label, r)
		}
	}
}

func radioStr(s string) *string { return &s }
func radioInt32(i int32) *int32 { return &i }

func fullStation() hub.RadioStation {
	return hub.RadioStation{
		UUID: "u1", URI: "radiobrowser://u1", Name: radioStr("Kiss FM"), Homepage: radioStr("https://www.kissfm.ua/"),
		CountryCode: radioStr("UA"), Languages: []string{"ukrainian"}, Tags: []string{"dance", "electronic"},
		Codec: radioStr("MP3"), Bitrate: radioInt32(128), LastCheckOK: true, Votes: 1520, ClickCount: 37,
	}
}

func bareStation() hub.RadioStation {
	return hub.RadioStation{UUID: "u2", URI: "radiobrowser://u2", Languages: []string{}, Tags: []string{}, HLS: true}
}

const fullStationYAML = `    uuid: "u1"
    uri: "radiobrowser://u1"
    name: "Kiss FM"
    favicon: null
    homepage: "https://www.kissfm.ua/"
    countryCode: "UA"
    languages: ["ukrainian"]
    tags: ["dance", "electronic"]
    codec: "MP3"
    bitrate: 128
    hls: false
    lastCheckOk: true
    votes: 1520
    clickCount: 37
`

const bareStationYAML = `    uuid: "u2"
    uri: "radiobrowser://u2"
    name: null
    favicon: null
    homepage: null
    countryCode: null
    languages: []
    tags: []
    codec: null
    bitrate: null
    hls: true
    lastCheckOk: false
    votes: 0
    clickCount: 0
`

func TestRenderStationPageYAML(t *testing.T) {
	page := hub.RadioStationPage{Stations: []hub.RadioStation{fullStation(), bareStation()}, NextOffset: radioInt32(20)}

	got := render.RenderStationPageYAML(page)

	want := "stations:\n  - " + strings.TrimPrefix(fullStationYAML, "    ") + "  - " + strings.TrimPrefix(bareStationYAML, "    ") + "nextOffset: 20\n"
	if got != want {
		t.Errorf("YAML =\n%s\nwant\n%s", got, want)
	}
}

func TestRenderStationPageYAML_LastPageAndEmptyPage(t *testing.T) {
	last := render.RenderStationPageYAML(hub.RadioStationPage{Stations: []hub.RadioStation{bareStation()}})
	if !strings.HasSuffix(last, "nextOffset: null\n") {
		t.Errorf("last page must end with nextOffset: null, got:\n%s", last)
	}

	for name, page := range map[string]hub.RadioStationPage{"empty slice": {Stations: []hub.RadioStation{}}, "nil slice": {}} {
		if got, want := render.RenderStationPageYAML(page), "# no stations found\nstations: []\nnextOffset: null\n"; got != want {
			t.Errorf("%s: YAML = %q, want %q", name, got, want)
		}
	}
}

func TestRenderStationPageJSON(t *testing.T) {
	page := hub.RadioStationPage{Stations: []hub.RadioStation{fullStation(), bareStation()}, NextOffset: radioInt32(20)}

	got := render.RenderStationPageJSON(page)

	if !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 1 {
		t.Errorf("JSON must be one line plus a newline, got %q", got)
	}
	var decoded struct {
		Stations   []map[string]any `json:"stations"`
		NextOffset *int             `json:"nextOffset"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, got)
	}
	if len(decoded.Stations) != 2 || decoded.NextOffset == nil || *decoded.NextOffset != 20 {
		t.Fatalf("decoded = %+v", decoded)
	}
	wantKeys := []string{"bitrate", "clickCount", "codec", "countryCode", "favicon", "hls", "homepage", "languages", "lastCheckOk", "name", "tags", "uri", "uuid", "votes"}
	for i, st := range decoded.Stations {
		keys := make([]string, 0, len(st))
		for k := range st {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if !reflect.DeepEqual(keys, wantKeys) {
			t.Errorf("station %d keys = %v, want %v", i, keys, wantKeys)
		}
	}
	if v, ok := decoded.Stations[1]["name"]; !ok || v != nil {
		t.Errorf("a nil name must be null, got %v (present %v)", v, ok)
	}

	if got := render.RenderStationPageJSON(hub.RadioStationPage{}); got != `{"stations":[],"nextOffset":null}`+"\n" {
		t.Errorf("empty page JSON = %q", got)
	}
}

func TestRenderStationPage_HostileTextIsEscaped(t *testing.T) {
	st := fullStation()
	st.Name = radioStr("name" + hostile)
	st.Homepage = radioStr("http://h" + hostile)
	st.Tags = []string{"tag" + hostile}
	st.Languages = []string{"lang" + hostile}
	st.Codec = radioStr("codec" + hostile)
	page := hub.RadioStationPage{Stations: []hub.RadioStation{st}}

	yaml := render.RenderStationPageYAML(page)
	js := render.RenderStationPageJSON(page)

	assertTerminalSafe(t, "YAML", yaml)
	assertTerminalSafe(t, "JSON", js)

	var decoded struct {
		Stations []hub.RadioStation `json:"stations"`
	}
	if err := json.Unmarshal([]byte(js), &decoded); err != nil {
		t.Fatalf("escaped JSON does not parse: %v\n%s", err, js)
	}
	got := decoded.Stations[0]
	if got.Name == nil || *got.Name != "name"+hostile || got.Tags[0] != "tag"+hostile || *got.Homepage != "http://h"+hostile {
		t.Errorf("JSON values changed by escaping: %+v", got)
	}
}

func TestRenderStationYAML_BareRecord(t *testing.T) {
	want := strings.ReplaceAll(fullStationYAML, "\n    ", "\n")
	want = strings.TrimPrefix(want, "    ")

	if got := render.RenderStationYAML(fullStation()); got != want {
		t.Errorf("YAML =\n%s\nwant\n%s", got, want)
	}
	if got := render.RenderStationYAML(bareStation()); strings.Contains(got, "stations:") || !strings.HasPrefix(got, "uuid: \"u2\"\n") || !strings.Contains(got, "bitrate: null\n") {
		t.Errorf("bare station YAML =\n%s", got)
	}
}

func TestRenderStationJSON_OneObject(t *testing.T) {
	got := render.RenderStationJSON(fullStation())

	if !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 1 {
		t.Errorf("JSON must be one line plus a newline, got %q", got)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if len(m) != 14 || m["uuid"] != "u1" || m["bitrate"] != float64(128) {
		t.Errorf("decoded = %v", m)
	}
	if _, ok := m["stations"]; ok {
		t.Error("a bare record has no stations wrapper")
	}
	var bare map[string]any
	_ = json.Unmarshal([]byte(render.RenderStationJSON(hub.RadioStation{UUID: "u", URI: "radiobrowser://u"})), &bare)
	if v, ok := bare["languages"]; !ok || v == nil {
		t.Errorf("nil languages must marshal as [], got %v", v)
	}
}

func TestRenderStation_HostileTextIsEscaped(t *testing.T) {
	st := fullStation()
	st.Name = radioStr("name" + hostile)
	st.Tags = []string{"tag" + hostile}
	st.Homepage = radioStr("http://h" + hostile)

	assertTerminalSafe(t, "YAML", render.RenderStationYAML(st))
	js := render.RenderStationJSON(st)
	assertTerminalSafe(t, "JSON", js)

	var back hub.RadioStation
	if err := json.Unmarshal([]byte(js), &back); err != nil {
		t.Fatalf("escaped JSON does not parse: %v", err)
	}
	if *back.Name != "name"+hostile || back.Tags[0] != "tag"+hostile {
		t.Errorf("values changed by escaping: %+v", back)
	}
}

func TestRenderCountries(t *testing.T) {
	list := []hub.RadioCountry{{Code: "UA", Name: "Ukraine", StationCount: 120}, {Code: "PL", Name: "Poland", StationCount: 90}}

	wantYAML := "countries:\n  - code: \"UA\"\n    name: \"Ukraine\"\n    stationCount: 120\n  - code: \"PL\"\n    name: \"Poland\"\n    stationCount: 90\n"
	if got := render.RenderCountriesYAML(list); got != wantYAML {
		t.Errorf("YAML = %q, want %q", got, wantYAML)
	}
	wantJSON := `{"countries":[{"code":"UA","name":"Ukraine","stationCount":120},{"code":"PL","name":"Poland","stationCount":90}]}` + "\n"
	if got := render.RenderCountriesJSON(list); got != wantJSON {
		t.Errorf("JSON = %q, want %q", got, wantJSON)
	}
	for _, empty := range [][]hub.RadioCountry{nil, {}} {
		if got := render.RenderCountriesYAML(empty); got != "# no countries found\ncountries: []\n" {
			t.Errorf("empty YAML = %q", got)
		}
		if got := render.RenderCountriesJSON(empty); got != `{"countries":[]}`+"\n" {
			t.Errorf("empty JSON = %q", got)
		}
	}
}

func TestRenderLanguages(t *testing.T) {
	list := []hub.RadioLanguage{{Name: "ukrainian", Code: radioStr("uk"), StationCount: 80}, {Name: "klingon", StationCount: 1}}

	wantYAML := "languages:\n  - name: \"ukrainian\"\n    code: \"uk\"\n    stationCount: 80\n  - name: \"klingon\"\n    code: null\n    stationCount: 1\n"
	if got := render.RenderLanguagesYAML(list); got != wantYAML {
		t.Errorf("YAML = %q, want %q", got, wantYAML)
	}
	wantJSON := `{"languages":[{"name":"ukrainian","code":"uk","stationCount":80},{"name":"klingon","code":null,"stationCount":1}]}` + "\n"
	if got := render.RenderLanguagesJSON(list); got != wantJSON {
		t.Errorf("JSON = %q, want %q", got, wantJSON)
	}
	for _, empty := range [][]hub.RadioLanguage{nil, {}} {
		if got := render.RenderLanguagesYAML(empty); got != "# no languages found\nlanguages: []\n" {
			t.Errorf("empty YAML = %q", got)
		}
		if got := render.RenderLanguagesJSON(empty); got != `{"languages":[]}`+"\n" {
			t.Errorf("empty JSON = %q", got)
		}
	}
}

func TestRenderTagPage(t *testing.T) {
	page := hub.RadioTagPage{Tags: []hub.RadioTag{{Name: "jazz", StationCount: 500}, {Name: "smooth jazz", StationCount: 40}}, NextOffset: radioInt32(20)}

	wantYAML := "tags:\n  - name: \"jazz\"\n    stationCount: 500\n  - name: \"smooth jazz\"\n    stationCount: 40\nnextOffset: 20\n"
	if got := render.RenderTagPageYAML(page); got != wantYAML {
		t.Errorf("YAML = %q, want %q", got, wantYAML)
	}
	wantJSON := `{"tags":[{"name":"jazz","stationCount":500},{"name":"smooth jazz","stationCount":40}],"nextOffset":20}` + "\n"
	if got := render.RenderTagPageJSON(page); got != wantJSON {
		t.Errorf("JSON = %q, want %q", got, wantJSON)
	}
	page.NextOffset = nil
	if got := render.RenderTagPageYAML(page); !strings.HasSuffix(got, "nextOffset: null\n") {
		t.Errorf("last page YAML = %q", got)
	}
	for _, empty := range []hub.RadioTagPage{{}, {Tags: []hub.RadioTag{}}} {
		if got := render.RenderTagPageYAML(empty); got != "# no tags found\ntags: []\nnextOffset: null\n" {
			t.Errorf("empty YAML = %q", got)
		}
		if got := render.RenderTagPageJSON(empty); got != `{"tags":[],"nextOffset":null}`+"\n" {
			t.Errorf("empty JSON = %q", got)
		}
	}
}

func TestRenderCatalogueLists_HostileTextIsEscaped(t *testing.T) {
	countries := []hub.RadioCountry{{Code: "U" + hostile, Name: "n" + hostile, StationCount: 1}}
	languages := []hub.RadioLanguage{{Name: "l" + hostile, Code: radioStr("c" + hostile), StationCount: 1}}
	tags := hub.RadioTagPage{Tags: []hub.RadioTag{{Name: "t" + hostile, StationCount: 1}}}

	for label, out := range map[string]string{
		"countries YAML": render.RenderCountriesYAML(countries), "countries JSON": render.RenderCountriesJSON(countries),
		"languages YAML": render.RenderLanguagesYAML(languages), "languages JSON": render.RenderLanguagesJSON(languages),
		"tags YAML": render.RenderTagPageYAML(tags), "tags JSON": render.RenderTagPageJSON(tags),
	} {
		assertTerminalSafe(t, label, out)
	}
	var back struct {
		Tags []hub.RadioTag `json:"tags"`
	}
	if err := json.Unmarshal([]byte(render.RenderTagPageJSON(tags)), &back); err != nil || back.Tags[0].Name != "t"+hostile {
		t.Errorf("tag JSON does not round-trip: %v %+v", err, back)
	}
}
