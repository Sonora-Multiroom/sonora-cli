package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const chimesBody = `{"chimes":[{"id":"ding","durationMillis":1800},{"id":"gong","durationMillis":3200}]}`

func TestPlay_Chime_SentVerbatim(t *testing.T) {
	for _, value := range []string{"ding", "none", "Ding"} {
		t.Run(value, func(t *testing.T) {
			srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)
			plain := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)

			res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--chime", value, "--hub-url", srv.URL)

			if res.exitCode != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
			}
			if m.playRequest["chime"] != value {
				t.Errorf("expected chime=%q in request body, got: %+v", value, m.playRequest)
			}
			if res.stdout != plain.stdout {
				t.Errorf("expected the same stdout as without --chime\nwith:\n%s\nwithout:\n%s", res.stdout, plain.stdout)
			}
		})
	}
}

func TestPlay_NoChime_FieldAbsent(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if _, ok := m.playRequest["chime"]; ok {
		t.Errorf("expected chime omitted from request body, got: %+v", m.playRequest)
	}
}

func TestPlay_BlankChime_UsageErrorNoRequest(t *testing.T) {
	for _, value := range []string{"", "  "} {
		srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

		res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--chime", value, "--hub-url", srv.URL)

		if res.exitCode != 2 {
			t.Fatalf("value %q: exit code = %d, want 2; stderr: %s", value, res.exitCode, res.stderr)
		}
		if !strings.Contains(res.stderr, "--chime") {
			t.Errorf("value %q: expected stderr to name --chime, got: %s", value, res.stderr)
		}
		if m.playRequest != nil {
			t.Errorf("value %q: expected no request to the play endpoint, got: %+v", value, m.playRequest)
		}
	}
}

func TestPlay_UnknownChime_HubDetailReported(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)
	m.playStatus = http.StatusBadRequest
	m.playBody = map[string]any{"title": "Bad Request", "detail": "Unknown chime: nope", "status": 400}

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--chime", "nope", "--hub-url", srv.URL)

	if res.exitCode != 6 {
		t.Fatalf("exit code = %d, want 6; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "Unknown chime: nope") {
		t.Errorf("expected the hub's detail in stderr, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestPlay_HelpDescribesChime(t *testing.T) {
	res := runCLI(t, "play", "--help")
	if !strings.Contains(res.stdout, "--chime") || !strings.Contains(res.stdout, "none") {
		t.Errorf("expected --help to describe --chime and none, got:\n%s", res.stdout)
	}
}

func TestRoute_Chime_SentVerbatim(t *testing.T) {
	for _, value := range []string{"gong", "none"} {
		t.Run(value, func(t *testing.T) {
			srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"whole-house"})
			plain := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--hub-url", srv.URL)

			res := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--chime", value, "--hub-url", srv.URL)

			if res.exitCode != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
			}
			if m.routesRequest["chime"] != value {
				t.Errorf("expected chime=%q in request body, got: %+v", value, m.routesRequest)
			}
			if res.stdout != plain.stdout {
				t.Errorf("expected the same stdout as without --chime\nwith:\n%s\nwithout:\n%s", res.stdout, plain.stdout)
			}
		})
	}
}

func TestRoute_NoChime_FieldAbsent(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"whole-house"})

	res := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if _, ok := m.routesRequest["chime"]; ok {
		t.Errorf("expected chime omitted from request body, got: %+v", m.routesRequest)
	}
}

func TestRoute_BlankChime_UsageErrorNoRequest(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"whole-house"})

	res := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--chime", " ", "--hub-url", srv.URL)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "--chime") {
		t.Errorf("expected stderr to name --chime, got: %s", res.stderr)
	}
	if m.routesRequest != nil {
		t.Errorf("expected no request to the routes endpoint, got: %+v", m.routesRequest)
	}
}

func TestRoute_UnknownChime_HubDetailReported(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"whole-house"})
	m.routesStatus = http.StatusBadRequest
	m.routesBody = map[string]any{"title": "Bad Request", "detail": "Unknown chime: nope", "status": 400}

	res := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--chime", "nope", "--hub-url", srv.URL)

	if res.exitCode != 6 {
		t.Fatalf("exit code = %d, want 6; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "Unknown chime: nope") {
		t.Errorf("expected the hub's detail in stderr, got:\n%s", res.stderr)
	}
}

func TestGetChimes_YAML(t *testing.T) {
	srv := mockOutputsServer(t, chimesBody)

	res := runCLI(t, "get", "chimes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	want := "chimes:\n  - id: \"ding\"\n    durationMillis: 1800\n  - id: \"gong\"\n    durationMillis: 3200\n"
	if res.stdout != want {
		t.Errorf("stdout:\n%s\nwant:\n%s", res.stdout, want)
	}
}

func TestGetChimes_JSON(t *testing.T) {
	srv := mockOutputsServer(t, chimesBody)

	res := runCLI(t, "get", "chimes", "--json", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	want := `{"chimes":[{"id":"ding","durationMillis":1800},{"id":"gong","durationMillis":3200}]}` + "\n"
	if res.stdout != want {
		t.Errorf("stdout: %s want: %s", res.stdout, want)
	}
}

func TestGetChimes_Empty(t *testing.T) {
	srv := mockOutputsServer(t, `{"chimes":[]}`)

	res := runCLI(t, "get", "chimes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if want := "# no chimes found\nchimes: []\n"; res.stdout != want {
		t.Errorf("stdout:\n%s\nwant:\n%s", res.stdout, want)
	}
}

func TestGetChimes_OlderHub404_ExitsHubError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "get", "chimes", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" || res.stderr == "" {
		t.Errorf("expected a message on stderr only, got stdout %q stderr %q", res.stdout, res.stderr)
	}
}

func TestGetChimes_UsageErrors(t *testing.T) {
	srv := mockOutputsServer(t, chimesBody)
	for name, args := range map[string][]string{
		"list verb":       {"list", "chimes"},
		"with an id":      {"get", "chimes/ding"},
		"extra argument":  {"get", "chimes", "ding"},
		"unknown flag":    {"get", "chimes", "--nope"},
		"list with an id": {"list", "chimes/ding"},
	} {
		t.Run(name, func(t *testing.T) {
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("expected empty stdout, got:\n%s", res.stdout)
			}
		})
	}
}

func TestGetOutputs_ShowsChimeOnlyWhenSet(t *testing.T) {
	srv := mockOutputsServer(t, `[{"outputId":"office","displayName":"Office","volume":30,"muted":false,"available":true,"enabled":true,"chime":"ding"},`+
		`{"outputId":"kitchen","displayName":"Kitchen","volume":30,"muted":false,"available":true,"enabled":true}]`)

	res := runCLI(t, "get", "outputs", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if got := strings.Count(res.stdout, "chime:"); got != 1 {
		t.Errorf("expected exactly one chime line (for office), got %d in:\n%s", got, res.stdout)
	}
	if !strings.Contains(res.stdout, "    chime: \"ding\"\n  - outputId: \"kitchen\"") {
		t.Errorf("expected office's chime right before kitchen, got:\n%s", res.stdout)
	}
}
