package unit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"sonora-cli/internal/hub"
)

// Note: hub.NewClient's construction being deferred to command-handler time
// (not package init()) is a structural property verified by code review
// (constitution Principle I), not something expressible as a runtime
// assertion — hub.NewClient itself performs no I/O, so calling it repeatedly
// here is inexpensive and side-effect-free.

func TestNewClient_EnforcesTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := hub.NewClient()
	client.Timeout = 100 * time.Millisecond // exercise the timeout path quickly

	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	_, err = client.Do(req)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if elapsed > 250*time.Millisecond {
		t.Errorf("client did not abort at timeout: took %v", elapsed)
	}
}

func TestNewClient_SingleAttemptNoRetry(t *testing.T) {
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := hub.NewClient()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	resp.Body.Close()

	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Errorf("got %d requests, want exactly 1 (no retry)", got)
	}
}

func TestClassifyError_Network(t *testing.T) {
	class, msg := hub.ClassifyError(context.DeadlineExceeded)
	if class != hub.ClassNetwork {
		t.Errorf("got class %v, want ClassNetwork", class)
	}
	if msg == "" {
		t.Error("expected a non-empty friendly message")
	}
}

func TestClassifyError_NetOpError(t *testing.T) {
	opErr := &opErrStub{}
	class, msg := hub.ClassifyError(opErr)
	if class != hub.ClassNetwork {
		t.Errorf("got class %v, want ClassNetwork", class)
	}
	if msg == "" {
		t.Error("expected a non-empty friendly message")
	}
}

func TestClassifyError_HubStatus(t *testing.T) {
	class, msg := hub.ClassifyError(&hub.StatusError{StatusCode: 503})
	if class != hub.ClassHub {
		t.Errorf("got class %v, want ClassHub", class)
	}
	if msg == "" {
		t.Error("expected a non-empty friendly message")
	}
}

func TestClassifyError_DecodeMismatch(t *testing.T) {
	class, msg := hub.ClassifyError(&hub.DecodeError{Err: errors.New("json: cannot unmarshal string into Go value of type int")})
	if class != hub.ClassHub {
		t.Errorf("got class %v, want ClassHub", class)
	}
	if msg == "" {
		t.Error("expected a non-empty friendly message")
	}
}

func TestErrorClass_ExitCodes(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassNone:     0,
		hub.ClassUsage:    2,
		hub.ClassHub:      3,
		hub.ClassNetwork:  4,
		hub.ClassNotFound: 5,
	}
	for class, want := range cases {
		if got := class.ExitCode(); got != want {
			t.Errorf("class %v: got exit code %d, want %d", class, got, want)
		}
	}
}

func TestClassifyError_NotFound(t *testing.T) {
	class, msg := hub.ClassifyError(&hub.NotFoundError{Resource: "output", ID: "x"})
	if class != hub.ClassNotFound {
		t.Errorf("got class %v, want ClassNotFound", class)
	}
	if !strings.Contains(msg, "x") {
		t.Errorf("expected friendly message to name the identifier, got: %q", msg)
	}
	if msg != "output not found: x" {
		t.Errorf("regression: output not-found message changed, got: %q", msg)
	}

	inputClass, inputMsg := hub.ClassifyError(&hub.NotFoundError{Resource: "input", ID: "x"})
	if inputClass != hub.ClassNotFound {
		t.Errorf("got class %v, want ClassNotFound", inputClass)
	}
	if inputMsg != "input not found: x" {
		t.Errorf("expected %q, got %q", "input not found: x", inputMsg)
	}

	distinct := map[hub.ErrorClass]bool{hub.ClassUsage: true, hub.ClassHub: true, hub.ClassNetwork: true}
	if distinct[hub.ClassNotFound] {
		t.Fatalf("test setup error: ClassNotFound must not equal ClassUsage/ClassHub/ClassNetwork")
	}
	codes := map[int]bool{}
	for _, c := range []hub.ErrorClass{hub.ClassUsage, hub.ClassHub, hub.ClassNetwork, hub.ClassNotFound} {
		if codes[c.ExitCode()] {
			t.Errorf("exit code %d reused across classes", c.ExitCode())
		}
		codes[c.ExitCode()] = true
	}
}

func TestClassifyError_APIError_StatusMappings(t *testing.T) {
	cases := []struct {
		status int
		class  hub.ErrorClass
	}{
		{400, hub.ClassValidation},
		{422, hub.ClassRouteFailed},
		{502, hub.ClassSourceUnreachable},
		{503, hub.ClassServiceUnavailable},
		{418, hub.ClassHub}, // any other status code falls back to ClassHub
	}
	for _, c := range cases {
		class, msg := hub.ClassifyError(&hub.APIError{StatusCode: c.status, Title: "Error", Detail: "detail"})
		if class != c.class {
			t.Errorf("status %d: got class %v, want %v", c.status, class, c.class)
		}
		if msg == "" {
			t.Errorf("status %d: expected a non-empty friendly message", c.status)
		}
	}
}

func TestErrorClass_NewExitCodes(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassValidation:         6,
		hub.ClassRouteFailed:        8,
		hub.ClassSourceUnreachable:  9,
		hub.ClassServiceUnavailable: 10,
	}
	for class, want := range cases {
		if got := class.ExitCode(); got != want {
			t.Errorf("class %v: got exit code %d, want %d", class, got, want)
		}
	}
}

func TestErrorClass_AllExitCodesDistinct(t *testing.T) {
	all := []hub.ErrorClass{
		hub.ClassUsage, hub.ClassHub, hub.ClassNetwork, hub.ClassNotFound,
		hub.ClassValidation, hub.ClassRouteFailed,
		hub.ClassSourceUnreachable, hub.ClassServiceUnavailable,
	}
	codes := map[int]hub.ErrorClass{}
	for _, c := range all {
		if prev, ok := codes[c.ExitCode()]; ok {
			t.Errorf("exit code %d reused: %v and %v", c.ExitCode(), prev, c)
		}
		codes[c.ExitCode()] = c
	}
}

func TestErrorClass_RouteExitCodes(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassInputNotFound:  11,
		hub.ClassTargetNotFound: 12,
	}
	for class, want := range cases {
		if got := class.ExitCode(); got != want {
			t.Errorf("class %v: got exit code %d, want %d", class, got, want)
		}
	}
}

func TestErrorClass_AllExitCodesDistinct_IncludingRoute(t *testing.T) {
	all := []hub.ErrorClass{
		hub.ClassUsage, hub.ClassHub, hub.ClassNetwork, hub.ClassNotFound,
		hub.ClassValidation, hub.ClassRouteFailed,
		hub.ClassSourceUnreachable, hub.ClassServiceUnavailable,
		hub.ClassInputNotFound, hub.ClassTargetNotFound,
	}
	codes := map[int]hub.ErrorClass{}
	for _, c := range all {
		if prev, ok := codes[c.ExitCode()]; ok {
			t.Errorf("exit code %d reused: %v and %v", c.ExitCode(), prev, c)
		}
		codes[c.ExitCode()] = c
	}
}

// TestClassifyError_NotFound_RegressionInputResourceStillClassNotFound
// guards research.md §3's decision: hub.ClassifyError itself is NOT modified
// to split on NotFoundError.Resource — a *hub.NotFoundError{Resource:
// "input"} routed directly through ClassifyError (as every command other
// than route does) must still classify as the existing generic
// ClassNotFound (exit 5), not the new route-only ClassInputNotFound (11).
// The input/target split is applied locally inside route.Run instead.
func TestClassifyError_NotFound_RegressionInputResourceStillClassNotFound(t *testing.T) {
	class, _ := hub.ClassifyError(&hub.NotFoundError{Resource: "input", ID: "x"})
	if class != hub.ClassNotFound {
		t.Errorf("got class %v, want ClassNotFound (regression: ClassifyError must not special-case Resource==\"input\")", class)
	}
}

// opErrStub satisfies net.Error the way a *net.OpError does for the
// purposes of classification testing, without depending on constructing a
// real *net.OpError.
type opErrStub struct{}

func (e *opErrStub) Error() string   { return "dial tcp: connection refused" }
func (e *opErrStub) Timeout() bool   { return false }
func (e *opErrStub) Temporary() bool { return false }

// --- 009-tts-commands: ClassTTSUnavailable, TTSError/TTSUnavailableError
// classification, and the configurable client timeout (T002). ---

func TestErrorClass_TTSUnavailableExitCode(t *testing.T) {
	if got := hub.ClassTTSUnavailable.ExitCode(); got != 13 {
		t.Errorf("ClassTTSUnavailable.ExitCode() = %d, want 13", got)
	}
}

func TestErrorClass_AllExitCodesDistinct_IncludingTTS(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassUsage:              2,
		hub.ClassHub:                3,
		hub.ClassNetwork:            4,
		hub.ClassNotFound:           5,
		hub.ClassValidation:         6,
		hub.ClassRouteFailed:        8,
		hub.ClassSourceUnreachable:  9,
		hub.ClassServiceUnavailable: 10,
		hub.ClassInputNotFound:      11,
		hub.ClassTargetNotFound:     12,
		hub.ClassTTSUnavailable:     13,
	}
	seen := map[int]hub.ErrorClass{}
	for class, want := range cases {
		got := class.ExitCode()
		if got != want {
			t.Errorf("class %v: got exit code %d, want %d", class, got, want)
		}
		if prev, ok := seen[got]; ok {
			t.Errorf("exit code %d reused: %v and %v", got, prev, class)
		}
		seen[got] = class
	}
}

func TestClassifyError_TTSError_ByCode(t *testing.T) {
	cases := []struct {
		code  string
		class hub.ErrorClass
	}{
		{"TARGET_NOT_FOUND", hub.ClassTargetNotFound},
		{"INVALID_REQUEST", hub.ClassValidation},
		{"PROVIDER_NOT_FOUND", hub.ClassNotFound},
		{"PROVIDER_TIMEOUT", hub.ClassServiceUnavailable},
		{"PROVIDER_RATE_LIMITED", hub.ClassServiceUnavailable},
		{"PROVIDER_ERROR", hub.ClassServiceUnavailable},
		{"FORMAT_NORMALIZATION_FAILED", hub.ClassServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			err := &hub.TTSError{StatusCode: 400, Code: c.code, Message: "boom"}
			class, msg := hub.ClassifyError(err)
			if class != c.class {
				t.Errorf("code %s: got class %v, want %v", c.code, class, c.class)
			}
			want := c.code + ": boom"
			if msg != want {
				t.Errorf("code %s: got message %q, want %q", c.code, msg, want)
			}
		})
	}
}

func TestClassifyError_TTSError_UnknownOrEmptyCodeFallsBackToStatus(t *testing.T) {
	cases := []struct {
		name   string
		status int
		class  hub.ErrorClass
	}{
		{"unknown code 400", 400, hub.ClassValidation},
		{"unknown code 503", 503, hub.ClassServiceUnavailable},
		{"unknown code 500", 500, hub.ClassHub},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := &hub.TTSError{StatusCode: c.status, Code: "SOMETHING_NEW", Message: "boom"}
			class, msg := hub.ClassifyError(err)
			if class != c.class {
				t.Errorf("got class %v, want %v", class, c.class)
			}
			if msg != "SOMETHING_NEW: boom" {
				t.Errorf("got message %q, want %q", msg, "SOMETHING_NEW: boom")
			}
		})
	}

	emptyCases := []struct {
		name   string
		status int
		class  hub.ErrorClass
		msg    string
	}{
		{"empty code, message present, 400", 400, hub.ClassValidation, "hub rejected the request (HTTP 400): boom"},
		{"empty code, message present, 503", 503, hub.ClassServiceUnavailable, "hub rejected the request (HTTP 503): boom"},
		{"empty code, message present, 500", 500, hub.ClassHub, "hub rejected the request (HTTP 500): boom"},
		{"empty code, no message, 500", 500, hub.ClassHub, "hub reported an error (HTTP 500)"},
	}
	for _, c := range emptyCases {
		t.Run(c.name, func(t *testing.T) {
			message := "boom"
			if c.msg == "hub reported an error (HTTP 500)" {
				message = ""
			}
			err := &hub.TTSError{StatusCode: c.status, Code: "", Message: message}
			class, msg := hub.ClassifyError(err)
			if class != c.class {
				t.Errorf("got class %v, want %v", class, c.class)
			}
			if msg != c.msg {
				t.Errorf("got message %q, want %q", msg, c.msg)
			}
		})
	}
}

func TestClassifyError_TTSUnavailableError_ByDiagnosis(t *testing.T) {
	const head = "text-to-speech is not available on this hub"
	cases := []struct {
		name      string
		err       *hub.TTSUnavailableError
		wantClass hub.ErrorClass
		wantMsg   string
	}{
		{
			"not installed",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisNotInstalled},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension is not installed on this hub",
		},
		{
			"not installed, loading disabled",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisNotInstalled, LoadingDisabled: true},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension is not installed on this hub (extension loading is switched off in the hub's configuration)",
		},
		{
			"disabled",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisDisabled},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension is installed but disabled in the hub's configuration",
		},
		{
			"rejected with reason",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisRejected, Reason: "bad jar"},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension failed to load: bad jar",
		},
		{
			"rejected without reason",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisRejected},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension failed to load",
		},
		{
			"inert",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisInert},
			hub.ClassTTSUnavailable,
			head + ": the TTS extension is loaded but inactive",
		},
		{
			"version mismatch",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisVersionMismatch},
			hub.ClassTTSUnavailable,
			head + ": the hub's TTS API does not match this CLI version",
		},
		{
			"unknown",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisUnknown},
			hub.ClassTTSUnavailable,
			head,
		},
		{
			"hub address",
			&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisHubAddress, BaseURL: "http://example.invalid"},
			hub.ClassNetwork,
			"http://example.invalid is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded; set the correct address with --hub-url, MULTIROOM_URL, or the config file",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			class, msg := hub.ClassifyError(c.err)
			if class != c.wantClass {
				t.Errorf("got class %v, want %v", class, c.wantClass)
			}
			if msg != c.wantMsg {
				t.Errorf("got message %q, want %q", msg, c.wantMsg)
			}
			if c.err.Diagnosis != hub.DiagnosisHubAddress && class != hub.ClassTTSUnavailable {
				t.Errorf("expected every non-HubAddress diagnosis to classify as ClassTTSUnavailable")
			}
		})
	}
}

func TestTTSUnavailableError_Unwrap(t *testing.T) {
	cause := errors.New("dial failed")
	err := &hub.TTSUnavailableError{Diagnosis: hub.DiagnosisUnknown, Cause: cause}
	if !errors.Is(err, cause) {
		t.Errorf("expected errors.Is to find Cause via Unwrap")
	}
}

func TestNewClientWithTimeout_EnforcesGivenTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := hub.NewClientWithTimeout(50 * time.Millisecond)

	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	_, err = client.Do(req)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if elapsed > 250*time.Millisecond {
		t.Errorf("client did not abort at timeout: took %v", elapsed)
	}
}

func TestNewClient_DefaultTimeoutIsFiveSeconds(t *testing.T) {
	client := hub.NewClient()
	if client.Timeout != 5*time.Second {
		t.Errorf("NewClient().Timeout = %v, want 5s", client.Timeout)
	}
}
