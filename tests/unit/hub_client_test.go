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

func TestClassifyError_AmbiguousTarget(t *testing.T) {
	class, msg := hub.ClassifyError(&hub.AmbiguousTargetError{ID: "shared-id"})
	if class != hub.ClassAmbiguous {
		t.Errorf("got class %v, want ClassAmbiguous", class)
	}
	if !strings.Contains(msg, "shared-id") {
		t.Errorf("expected friendly message to name the identifier, got: %q", msg)
	}
}

func TestErrorClass_NewExitCodes(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassValidation:         6,
		hub.ClassAmbiguous:          7,
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
		hub.ClassValidation, hub.ClassAmbiguous, hub.ClassRouteFailed,
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

// opErrStub satisfies net.Error the way a *net.OpError does for the
// purposes of classification testing, without depending on constructing a
// real *net.OpError.
type opErrStub struct{}

func (e *opErrStub) Error() string   { return "dial tcp: connection refused" }
func (e *opErrStub) Timeout() bool   { return false }
func (e *opErrStub) Temporary() bool { return false }
