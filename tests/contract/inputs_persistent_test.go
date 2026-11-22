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

// Shapes mirror the createInput and persistInput operations and the 409
// reasons of ErrorResponse in api/openapi.json (hub 0.1.35, HUB-041).

func storedInputBody() map[string]any {
	return map[string]any{
		"inputId": "kiss", "displayName": "Kiss", "uri": "http://x/s", "enabled": true,
		"autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": false,
		"defaultJoinMode": nil, "defaultRouteVolume": nil, "defaultDuckingLevelDb": nil,
		"origin": "STORED", "playable": true, "unplayableReason": nil,
	}
}

func TestCreateInput_PersistentIsSentOnlyWhenSet(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = nil
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(storedInputBody())
	}))
	defer srv.Close()

	yes := true
	req := hub.CreateInputRequest{InputID: "kiss", DisplayName: "Kiss", URI: "http://x/s", Persistent: &yes}
	if _, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["persistent"] != true {
		t.Errorf("persistent not sent: %+v", got)
	}

	req.Persistent = nil
	if _, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := got["persistent"]; present {
		t.Errorf("persistent sent although unset: %+v", got)
	}
}

func TestPersistInput_Success(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(storedInputBody())
	}))
	defer srv.Close()

	in, err := hub.PersistInput(context.Background(), hub.NewClient(), srv.URL, "kiss")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != http.MethodPut || path != "/api/v2/inputs/kiss/persistent" {
		t.Errorf("got %s %s", method, path)
	}
	if in.Origin != "STORED" {
		t.Errorf("origin = %q, want STORED", in.Origin)
	}
}

func TestPersistInput_Errors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		check  func(t *testing.T, err error)
	}{
		{"404", 404, `{}`, func(t *testing.T, err error) {
			var nf *hub.NotFoundError
			if !errors.As(err, &nf) || nf.ID != "kiss" {
				t.Errorf("want NotFoundError for kiss, got %v", err)
			}
		}},
		{"400", 400, `{"title":"Bad Request","detail":"Input is configured in multiroom.yml"}`, func(t *testing.T, err error) {
			var ae *hub.APIError
			if !errors.As(err, &ae) || ae.StatusCode != 400 || ae.Detail == "" {
				t.Errorf("want 400 APIError with detail, got %v", err)
			}
		}},
		{"409", 409, `{"title":"Conflict","detail":"full","reason":"INPUT_STORE_FULL"}`, func(t *testing.T, err error) {
			var ae *hub.APIError
			if !errors.As(err, &ae) || ae.Reason != hub.ReasonInputStoreFull {
				t.Errorf("want reason INPUT_STORE_FULL, got %v", err)
			}
		}},
		{"500", 500, `{"title":"Internal Server Error","detail":"read-only: the store could not be loaded"}`, func(t *testing.T, err error) {
			_, msg := hub.ClassifyError(err)
			if !strings.Contains(msg, "read-only: the store could not be loaded") {
				t.Errorf("500 message lost the store detail: %q", msg)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			_, err := hub.PersistInput(context.Background(), hub.NewClient(), srv.URL, "kiss")
			if err == nil {
				t.Fatal("expected an error")
			}
			tc.check(t, err)
		})
	}
}

func TestCreateInput_StoreFailureKeepsDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"title":"Internal Server Error","detail":"timed out after 5 s"}`))
	}))
	defer srv.Close()
	yes := true
	_, err := hub.CreateInput(context.Background(), hub.NewClient(), srv.URL,
		hub.CreateInputRequest{InputID: "a", DisplayName: "A", URI: "u", Persistent: &yes})
	class, msg := hub.ClassifyError(err)
	if class != hub.ClassHub || !strings.Contains(msg, "timed out after 5 s") {
		t.Errorf("got %v %q", class, msg)
	}
}

func TestClassifyError_InputConflictReasons(t *testing.T) {
	cases := map[string]string{
		hub.ReasonInputUnplayable: "extension",
		hub.ReasonInputIDTaken:    "already",
		hub.ReasonInputStoreFull:  "full",
		"SOMETHING_NEW":           "",
	}
	for reason, want := range cases {
		err := &hub.APIError{StatusCode: 409, Reason: reason}
		class, msg := hub.ClassifyError(err)
		if class != hub.ClassConflict {
			t.Errorf("%s: class = %v, want conflict", reason, class)
		}
		if want != "" && !strings.Contains(msg, want) {
			t.Errorf("%s: message %q lacks %q", reason, msg, want)
		}
		if want == "" && msg == "" {
			t.Errorf("%s: empty message", reason)
		}
	}
}
