package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/api"
	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Shapes here mirror #/components/schemas/HubInfoResponse and the
// GET /api/v2/info operation in api/openapi.json (constitution Principle II).

const infoFixture = `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33","startedAt":"2026-11-06T10:06:59.512Z","uptimeSeconds":16981}`

// specSchemaProperties returns the property names of a component schema in
// api.Spec, so a test compares a Go type with the spec, not with a copy.
func specSchemaProperties(t *testing.T, schema string) []string {
	t.Helper()
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(api.Spec, &doc); err != nil {
		t.Fatalf("api.Spec is not valid JSON: %v", err)
	}
	s, ok := doc.Components.Schemas[schema]
	if !ok {
		t.Fatalf("api.Spec has no schema %s", schema)
	}
	names := make([]string, 0, len(s.Properties))
	for name := range s.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// keysOf returns the sorted keys of a JSON object.
func keysOf(t *testing.T, data []byte) []string {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("not a JSON object: %v: %s", err, data)
	}
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func infoServer(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	got := &http.Request{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = *r.Clone(context.Background())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

func TestGetInfo_RequestAndDecode(t *testing.T) {
	srv, got := infoServer(t, 200, infoFixture)

	info, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Method != http.MethodGet || got.URL.Path != "/api/v2/info" {
		t.Errorf("got %s %s, want GET /api/v2/info", got.Method, got.URL.Path)
	}
	if info.Product != "multiroom" || info.Version != "0.1.33" || info.APIVersion != "0.1.33" ||
		info.StartedAt != "2026-11-06T10:06:59.512Z" || info.UptimeSeconds == nil || *info.UptimeSeconds != 16981 {
		t.Errorf("unexpected info: %+v", info)
	}
}

func TestInfo_MarshalsWithSpecFieldNames(t *testing.T) {
	srv, _ := infoServer(t, 200, infoFixture)
	info, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := specSchemaProperties(t, "HubInfoResponse")
	if got := keysOf(t, data); !equalStrings(got, want) {
		t.Errorf("marshalled keys %v, want the spec's %v", got, want)
	}
	// The fixture itself uses only properties the spec declares.
	if got := keysOf(t, []byte(infoFixture)); !equalStrings(got, want) {
		t.Errorf("fixture keys %v differ from the spec's %v", got, want)
	}
}

func TestGetInfo_OptionalFieldsAbsent(t *testing.T) {
	srv, _ := infoServer(t, 200, `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33"}`)
	info, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UptimeSeconds != nil {
		t.Errorf("UptimeSeconds = %d, want nil", *info.UptimeSeconds)
	}
	if info.StartedAt != "" {
		t.Errorf("StartedAt = %q, want empty", info.StartedAt)
	}
	data, _ := json.Marshal(info)
	if got := keysOf(t, data); len(got) != 3 {
		t.Errorf("marshalled keys %v, want only product, version, apiVersion", got)
	}
}

func TestGetInfo_PassesValuesThrough(t *testing.T) {
	srv, _ := infoServer(t, 200, `{"product":"other","version":"unknown","apiVersion":"unknown","uptimeSeconds":0}`)
	info, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Product != "other" || info.Version != "unknown" || info.APIVersion != "unknown" {
		t.Errorf("unexpected info: %+v", info)
	}
	if info.UptimeSeconds == nil || *info.UptimeSeconds != 0 {
		t.Errorf("a zero uptime the hub sent must stay non-nil 0, got %v", info.UptimeSeconds)
	}
}

func TestGetInfo_404_InfoNotOffered(t *testing.T) {
	srv, _ := infoServer(t, 404, ``)
	_, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	var notOffered *hub.InfoNotOfferedError
	if !errors.As(err, &notOffered) {
		t.Fatalf("expected *hub.InfoNotOfferedError, got %T: %v", err, err)
	}
	if class, _ := hub.ClassifyError(err); class != hub.ClassHub {
		t.Errorf("class = %v, want ClassHub", class)
	}
}

func TestGetInfo_500_StatusError(t *testing.T) {
	srv, _ := infoServer(t, 500, ``)
	_, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *hub.StatusError, got %T: %v", err, err)
	}
	if class, _ := hub.ClassifyError(err); class != hub.ClassHub {
		t.Errorf("class = %v, want ClassHub", class)
	}
}

func TestGetInfo_BadBodyIsDecodeError(t *testing.T) {
	for name, body := range map[string]string{
		"not json":         `not json`,
		"not an object":    `[1,2]`,
		"null":             `null`,
		"no product":       `{"version":"1","apiVersion":"1"}`,
		"no version":       `{"product":"multiroom","apiVersion":"1"}`,
		"no apiVersion":    `{"product":"multiroom","version":"1"}`,
		"empty product":    `{"product":"","version":"1","apiVersion":"1"}`,
		"empty version":    `{"product":"multiroom","version":"","apiVersion":"1"}`,
		"empty apiVersion": `{"product":"multiroom","version":"1","apiVersion":""}`,
		"wrong type":       `{"product":"multiroom","version":1,"apiVersion":"1"}`,
		"empty body":       ``,
		"bad uptime type":  `{"product":"multiroom","version":"1","apiVersion":"1","uptimeSeconds":"x"}`,
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := infoServer(t, 200, body)
			_, err := hub.GetInfo(context.Background(), hub.NewClient(), srv.URL)
			var decodeErr *hub.DecodeError
			if !errors.As(err, &decodeErr) {
				t.Fatalf("expected *hub.DecodeError, got %T: %v", err, err)
			}
			if class, _ := hub.ClassifyError(err); class != hub.ClassHub {
				t.Errorf("class = %v, want ClassHub", class)
			}
		})
	}
}
