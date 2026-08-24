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
	class, msg := hub.ClassifyError(&hub.NotFoundError{OutputID: "x"})
	if class != hub.ClassNotFound {
		t.Errorf("got class %v, want ClassNotFound", class)
	}
	if !strings.Contains(msg, "x") {
		t.Errorf("expected friendly message to name the identifier, got: %q", msg)
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

// opErrStub satisfies net.Error the way a *net.OpError does for the
// purposes of classification testing, without depending on constructing a
// real *net.OpError.
type opErrStub struct{}

func (e *opErrStub) Error() string   { return "dial tcp: connection refused" }
func (e *opErrStub) Timeout() bool   { return false }
func (e *opErrStub) Temporary() bool { return false }
