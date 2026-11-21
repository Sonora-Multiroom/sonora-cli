package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
)

// Request/response shapes here mirror #/components/schemas/ExtensionInventory
// and #/components/schemas/Extension, and the listExtensions operation, in
// api/openapi.json (constitution Principle II). ListExtensions is used
// internally only, by the TTS "not available" diagnosis (research.md §5).

func TestListExtensions_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/extensions" {
			t.Errorf("got path %q, want /api/v2/extensions", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("got method %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		reason := "bad jar"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"loadingEnabled": true,
			"extensions": []map[string]any{
				{"id": "tts", "name": "Text to Speech", "status": "ACTIVE", "rejectionReason": nil},
				{"id": "other", "name": "Other", "status": "REJECTED", "rejectionReason": reason},
			},
		})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inv, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.LoadingEnabled == nil || !*inv.LoadingEnabled {
		t.Errorf("expected LoadingEnabled true, got %+v", inv.LoadingEnabled)
	}
	if len(inv.Extensions) != 2 {
		t.Fatalf("expected 2 extensions, got %d: %+v", len(inv.Extensions), inv.Extensions)
	}
	if inv.Extensions[0].ID != "tts" || inv.Extensions[0].Status != "ACTIVE" {
		t.Errorf("unexpected first extension: %+v", inv.Extensions[0])
	}
	if inv.Extensions[0].RejectionReason != nil {
		t.Errorf("expected nil rejection reason for tts, got %v", *inv.Extensions[0].RejectionReason)
	}
	if inv.Extensions[1].RejectionReason == nil || *inv.Extensions[1].RejectionReason != "bad jar" {
		t.Errorf("expected rejection reason %q, got %+v", "bad jar", inv.Extensions[1].RejectionReason)
	}
}

func TestListExtensions_MissingExtensionsArrayDecodesAsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"loadingEnabled": true})
	}))
	defer srv.Close()

	client := hub.NewClient()
	inv, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inv.Extensions) != 0 {
		t.Errorf("expected an empty Extensions slice, got %+v", inv.Extensions)
	}
}

func TestListExtensions_404_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a 404 response, got nil")
	}
	var statusErr *hub.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *hub.StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusNotFound {
		t.Errorf("got StatusCode %d, want 404", statusErr.StatusCode)
	}
}

func TestListExtensions_NonJSON200_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := hub.NewClient()
	_, err := hub.ListExtensions(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an error for a non-JSON 200 body, got nil")
	}
	var decodeErr *hub.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *hub.DecodeError, got %T: %v", err, err)
	}
}

func TestListExtensions_DecodesVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"extensions":[
			{"id":"radio","status":"ACTIVE","version":"0.1.0"},
			{"id":"a","status":"ACTIVE","version":null},
			{"id":"b","status":"ACTIVE"}]}`))
	}))
	defer srv.Close()

	inv, err := hub.ListExtensions(context.Background(), hub.NewClient(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := inv.Extensions[0].Version; v == nil || *v != "0.1.0" {
		t.Errorf("radio version = %v, want 0.1.0", v)
	}
	if inv.Extensions[1].Version != nil || inv.Extensions[2].Version != nil {
		t.Errorf("null and absent version must decode to nil: %+v %+v", inv.Extensions[1].Version, inv.Extensions[2].Version)
	}
}

// inventoryServer fakes GET /api/v2/extensions and counts requests.
func inventoryServer(t *testing.T, status int, body string) (*httptest.Server, *int) {
	t.Helper()
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

const radioHead = "radio is not available on this hub"

func TestDiagnoseRadioUnavailable(t *testing.T) {
	cases := []struct {
		name          string
		status        int
		body          string
		wantDiagnosis hub.ExtensionDiagnosis
		wantMsg       string
		check         func(t *testing.T, e *hub.RadioUnavailableError)
	}{
		{
			name: "not installed", status: 200, body: `{"loadingEnabled":true,"extensions":[]}`,
			wantDiagnosis: hub.DiagnosisNotInstalled,
			wantMsg:       radioHead + ": the radio extension is not installed on this hub",
		},
		{
			name: "not installed, loading switched off", status: 200, body: `{"loadingEnabled":false,"extensions":[]}`,
			wantDiagnosis: hub.DiagnosisNotInstalled,
			wantMsg:       radioHead + ": the radio extension is not installed on this hub (extension loading is switched off in the hub's configuration)",
			check: func(t *testing.T, e *hub.RadioUnavailableError) {
				if !e.LoadingDisabled {
					t.Error("LoadingDisabled = false, want true")
				}
			},
		},
		{
			name: "decoy name, matched by id only", status: 200, body: `{"extensions":[{"id":"other","name":"radio","status":"ACTIVE"}]}`,
			wantDiagnosis: hub.DiagnosisNotInstalled,
			wantMsg:       radioHead + ": the radio extension is not installed on this hub",
		},
		{
			name: "disabled", status: 200, body: `{"extensions":[{"id":"radio","status":"DISABLED"}]}`,
			wantDiagnosis: hub.DiagnosisDisabled,
			wantMsg:       radioHead + ": the radio extension is installed but disabled in the hub's configuration",
		},
		{
			name: "rejected with multi-line reason", status: 200,
			body:          `{"extensions":[{"id":"radio","status":"REJECTED","rejectionReason":"bad jar:\nline 2"}]}`,
			wantDiagnosis: hub.DiagnosisRejected,
			wantMsg:       radioHead + ": the radio extension failed to load: bad jar: line 2",
			check: func(t *testing.T, e *hub.RadioUnavailableError) {
				if e.Reason != "bad jar: line 2" {
					t.Errorf("Reason = %q", e.Reason)
				}
			},
		},
		{
			name: "rejected without reason", status: 200, body: `{"extensions":[{"id":"radio","status":"REJECTED","rejectionReason":null}]}`,
			wantDiagnosis: hub.DiagnosisRejected,
			wantMsg:       radioHead + ": the radio extension failed to load",
		},
		{
			name: "inert", status: 200, body: `{"extensions":[{"id":"radio","status":"INERT"}]}`,
			wantDiagnosis: hub.DiagnosisInert,
			wantMsg:       radioHead + ": the radio extension is loaded but inactive",
		},
		{
			name: "active, old version", status: 200, body: `{"extensions":[{"id":"radio","status":"ACTIVE","version":"0.1.0"}]}`,
			wantDiagnosis: hub.DiagnosisVersionMismatch,
			wantMsg:       radioHead + ": the radio extension on this hub (version 0.1.0) cannot search stations; it needs 0.2.0 or later",
			check: func(t *testing.T, e *hub.RadioUnavailableError) {
				if e.Version != "0.1.0" {
					t.Errorf("Version = %q", e.Version)
				}
			},
		},
		{
			name: "active, version unknown", status: 200, body: `{"extensions":[{"id":"radio","status":"ACTIVE"}]}`,
			wantDiagnosis: hub.DiagnosisVersionMismatch,
			wantMsg:       radioHead + ": the radio extension on this hub (version unknown) cannot search stations; it needs 0.2.0 or later",
		},
		{
			name: "unknown status", status: 200, body: `{"extensions":[{"id":"radio","status":"WHATEVER"}]}`,
			wantDiagnosis: hub.DiagnosisUnknown, wantMsg: radioHead,
		},
		{
			name: "inventory 500", status: 500, body: ``,
			wantDiagnosis: hub.DiagnosisUnknown, wantMsg: radioHead,
			check: func(t *testing.T, e *hub.RadioUnavailableError) {
				var se *hub.StatusError
				if !errors.As(e.Cause, &se) || se.StatusCode != 500 {
					t.Errorf("Cause = %v, want *StatusError 500", e.Cause)
				}
			},
		},
		{
			name: "malformed inventory", status: 200, body: `not json`,
			wantDiagnosis: hub.DiagnosisUnknown, wantMsg: radioHead,
			check: func(t *testing.T, e *hub.RadioUnavailableError) {
				var de *hub.DecodeError
				if !errors.As(e.Cause, &de) {
					t.Errorf("Cause = %v, want *DecodeError", e.Cause)
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, count := inventoryServer(t, c.status, c.body)
			e := hub.DiagnoseRadioUnavailable(context.Background(), hub.NewClient(), srv.URL)
			if e == nil {
				t.Fatal("DiagnoseRadioUnavailable returned nil")
			}
			if e.Diagnosis != c.wantDiagnosis {
				t.Errorf("Diagnosis = %v, want %v", e.Diagnosis, c.wantDiagnosis)
			}
			if e.Error() != c.wantMsg {
				t.Errorf("Error() = %q, want %q", e.Error(), c.wantMsg)
			}
			if *count != 1 {
				t.Errorf("requests = %d, want 1", *count)
			}
			if c.check != nil {
				c.check(t, e)
			}
		})
	}
}

func TestDiagnoseRadioUnavailable_Inventory404_HubAddress(t *testing.T) {
	srv, count := inventoryServer(t, 404, ``)
	e := hub.DiagnoseRadioUnavailable(context.Background(), hub.NewClient(), srv.URL)
	if e.Diagnosis != hub.DiagnosisHubAddress || e.BaseURL != srv.URL {
		t.Fatalf("got %+v, want hub-address diagnosis with BaseURL %q", e, srv.URL)
	}
	want := srv.URL + " is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded"
	if e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
	if strings.HasPrefix(e.Error(), radioHead) {
		t.Error("hub-address text must not start with the radio head")
	}
	if *count != 1 {
		t.Errorf("requests = %d, want 1", *count)
	}
}

func TestRadioUnavailableError_Unwrap(t *testing.T) {
	cause := errors.New("boom")
	e := &hub.RadioUnavailableError{Diagnosis: hub.DiagnosisUnknown, Cause: cause}
	if !errors.Is(e, cause) {
		t.Error("errors.Is(e, cause) = false")
	}
}

// TTSDiagnosis stays as an alias of ExtensionDiagnosis so existing callers compile.
func TestTTSDiagnosis_IsAliasOfExtensionDiagnosis(t *testing.T) {
	var d hub.TTSDiagnosis = hub.DiagnosisDisabled //nolint:staticcheck // deliberate use of the deprecated alias
	var e hub.ExtensionDiagnosis = d
	if e != hub.DiagnosisDisabled {
		t.Errorf("got %v", e)
	}
	u := &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisDisabled}
	if u.Diagnosis != e {
		t.Errorf("TTSUnavailableError.Diagnosis = %v", u.Diagnosis)
	}
}
