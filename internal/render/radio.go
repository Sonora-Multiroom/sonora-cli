package render

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// SafeText returns s with every character a terminal could act on written as
// an escape sequence: C0 and C1 controls, DEL, line breaks, Unicode format
// characters (for example U+202E, which reverses text) and invalid UTF-8.
// Printable text, including non-ASCII letters, is returned unchanged. Use it
// for any text that came from outside this program, such as a hub's message.
func SafeText(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case strconv.IsPrint(r):
			b.WriteRune(r)
		default:
			q := strconv.Quote(string(r))
			b.WriteString(q[1 : len(q)-1])
		}
		i += size
	}
	return b.String()
}

// jsonUnsafe reports whether encoding/json leaves r raw although a terminal
// could act on it: DEL, C1 controls and Unicode format characters. C0
// controls are always escaped by encoding/json.
func jsonUnsafe(r rune) bool {
	return r == 0x7f || (r >= 0x80 && r <= 0x9f) || unicode.Is(unicode.Cf, r)
}

// safeJSON rewrites every rune jsonUnsafe reports as a \uNNNN escape (a
// surrogate pair above U+FFFF). Such a rune can only sit inside a string in
// valid JSON, so the result is still valid JSON with the same values.
func safeJSON(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			out = append(out, data[i])
		case !jsonUnsafe(r):
			out = append(out, data[i:i+size]...)
		case r > 0xFFFF:
			r -= 0x10000
			out = fmt.Appendf(out, `\u%04x\u%04x`, 0xD800+(r>>10), 0xDC00+(r&0x3FF))
		default:
			out = fmt.Appendf(out, `\u%04x`, r)
		}
		i += size
	}
	return out
}

// yamlQuoted writes s as a YAML double-quoted scalar. Go's %q escapes every
// rune that is not printable (controls, DEL, C1, format characters), in forms
// YAML reads back as the same text.
func yamlQuoted(s string) string { return fmt.Sprintf("%q", s) }

// yamlOptionalString is yamlQuoted, or a bare null for an absent value.
func yamlOptionalString(s *string) string {
	if s == nil {
		return "null"
	}
	return yamlQuoted(*s)
}

// yamlStringList is a YAML flow sequence of quoted strings; `[]` when empty.
func yamlStringList(list []string) string {
	quoted := make([]string, len(list))
	for i, s := range list {
		quoted[i] = yamlQuoted(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// writeStationYAML writes the 14 fields of a station in the schema's order.
// first prefixes the first line, rest every later one: "  - " and "    " for a
// list item, "" and "" for a bare record.
func writeStationYAML(b *strings.Builder, first, rest string, s hub.RadioStation) {
	bitrate := "null"
	if s.Bitrate != nil {
		bitrate = strconv.FormatInt(int64(*s.Bitrate), 10)
	}
	lines := []string{
		"uuid: " + yamlQuoted(s.UUID),
		"uri: " + yamlQuoted(s.URI),
		"name: " + yamlOptionalString(s.Name),
		"favicon: " + yamlOptionalString(s.Favicon),
		"homepage: " + yamlOptionalString(s.Homepage),
		"countryCode: " + yamlOptionalString(s.CountryCode),
		"languages: " + yamlStringList(s.Languages),
		"tags: " + yamlStringList(s.Tags),
		"codec: " + yamlOptionalString(s.Codec),
		"bitrate: " + bitrate,
		"hls: " + strconv.FormatBool(s.HLS),
		"lastCheckOk: " + strconv.FormatBool(s.LastCheckOK),
		"votes: " + strconv.FormatInt(s.Votes, 10),
		"clickCount: " + strconv.FormatInt(s.ClickCount, 10),
	}
	for i, line := range lines {
		if i == 0 {
			b.WriteString(first)
		} else {
			b.WriteString(rest)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
}

func yamlNextOffset(next *int32) string {
	if next == nil {
		return "nextOffset: null\n"
	}
	return "nextOffset: " + strconv.FormatInt(int64(*next), 10) + "\n"
}

// RenderStationPageYAML renders a page of stations as YAML: a `stations:` list
// with every field of every station (an absent one as null), then
// `nextOffset`. An empty page prints a `# no stations found` comment and
// `stations: []`.
func RenderStationPageYAML(p hub.RadioStationPage) string {
	var b strings.Builder
	if len(p.Stations) == 0 {
		b.WriteString("# no stations found\nstations: []\n")
	} else {
		b.WriteString("stations:\n")
		for _, s := range p.Stations {
			writeStationYAML(&b, "  - ", "    ", s)
		}
	}
	b.WriteString(yamlNextOffset(p.NextOffset))
	return b.String()
}

// stationForJSON makes the two lists non-nil, so they marshal as [] and never null.
func stationForJSON(s hub.RadioStation) hub.RadioStation {
	if s.Languages == nil {
		s.Languages = []string{}
	}
	if s.Tags == nil {
		s.Tags = []string{}
	}
	return s
}

// marshalSafeJSON marshals v and escapes what a terminal could act on.
func marshalSafeJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Every value passed here is made of strings, numbers, booleans, slices and
		// structs of those: Marshal cannot fail.
		panic(err)
	}
	return string(safeJSON(data)) + "\n"
}

// RenderStationPageJSON renders a page as one JSON object, `{"stations":[…],
// "nextOffset":…}`, with the same keys as the YAML. Stations is always an
// array, and an absent optional field is null.
func RenderStationPageJSON(p hub.RadioStationPage) string {
	stations := make([]hub.RadioStation, len(p.Stations))
	for i, s := range p.Stations {
		stations[i] = stationForJSON(s)
	}
	return marshalSafeJSON(hub.RadioStationPage{Stations: stations, NextOffset: p.NextOffset})
}

// RenderStationYAML renders one station as a bare YAML record: the same 14
// fields as in a page, without a `stations:` wrapper.
func RenderStationYAML(s hub.RadioStation) string {
	var b strings.Builder
	writeStationYAML(&b, "", "", s)
	return b.String()
}

// RenderStationJSON renders one station as a strict JSON object with the same
// keys as the YAML.
func RenderStationJSON(s hub.RadioStation) string {
	return marshalSafeJSON(stationForJSON(s))
}

// writeListYAML writes `key:` and one item per entry, or the comment and
// `key: []` for none. fields returns an item's lines, without indentation.
func writeListYAML[T any](b *strings.Builder, key, noun string, items []T, fields func(T) []string) {
	if len(items) == 0 {
		fmt.Fprintf(b, "# no %s found\n%s: []\n", noun, key)
		return
	}
	b.WriteString(key + ":\n")
	for _, it := range items {
		for i, line := range fields(it) {
			if i == 0 {
				b.WriteString("  - ")
			} else {
				b.WriteString("    ")
			}
			b.WriteString(line + "\n")
		}
	}
}

// RenderCountriesYAML renders the catalogue's countries as a `countries:` list
// of code, name and stationCount, in the order given.
func RenderCountriesYAML(list []hub.RadioCountry) string {
	var b strings.Builder
	writeListYAML(&b, "countries", "countries", list, func(c hub.RadioCountry) []string {
		return []string{"code: " + yamlQuoted(c.Code), "name: " + yamlQuoted(c.Name), "stationCount: " + strconv.FormatInt(c.StationCount, 10)}
	})
	return b.String()
}

// RenderCountriesJSON renders the countries as `{"countries":[…]}`; the list is
// always an array.
func RenderCountriesJSON(list []hub.RadioCountry) string {
	if list == nil {
		list = []hub.RadioCountry{}
	}
	return marshalSafeJSON(struct {
		Countries []hub.RadioCountry `json:"countries"`
	}{list})
}

// RenderLanguagesYAML renders the catalogue's languages as a `languages:` list
// of name, code (null when absent) and stationCount, in the order given.
func RenderLanguagesYAML(list []hub.RadioLanguage) string {
	var b strings.Builder
	writeListYAML(&b, "languages", "languages", list, func(l hub.RadioLanguage) []string {
		return []string{"name: " + yamlQuoted(l.Name), "code: " + yamlOptionalString(l.Code), "stationCount: " + strconv.FormatInt(l.StationCount, 10)}
	})
	return b.String()
}

// RenderLanguagesJSON renders the languages as `{"languages":[…]}`; the list is
// always an array.
func RenderLanguagesJSON(list []hub.RadioLanguage) string {
	if list == nil {
		list = []hub.RadioLanguage{}
	}
	return marshalSafeJSON(struct {
		Languages []hub.RadioLanguage `json:"languages"`
	}{list})
}

// RenderTagPageYAML renders a page of tags as a `tags:` list of name and
// stationCount, then `nextOffset`.
func RenderTagPageYAML(p hub.RadioTagPage) string {
	var b strings.Builder
	writeListYAML(&b, "tags", "tags", p.Tags, func(g hub.RadioTag) []string {
		return []string{"name: " + yamlQuoted(g.Name), "stationCount: " + strconv.FormatInt(g.StationCount, 10)}
	})
	b.WriteString(yamlNextOffset(p.NextOffset))
	return b.String()
}

// RenderTagPageJSON renders a page of tags as `{"tags":[…],"nextOffset":…}`;
// the list is always an array.
func RenderTagPageJSON(p hub.RadioTagPage) string {
	if p.Tags == nil {
		p.Tags = []hub.RadioTag{}
	}
	return marshalSafeJSON(p)
}
