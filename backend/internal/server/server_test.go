package server

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ziad-azm/ZCRM/backend/internal/middleware"
)

// lockedBuffer guards the log buffer: the server writes the log line from its
// own goroutine while the test reads it.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// TestRoutes covers the happy path plus the routing negatives that the trailing
// slash and method behaviour depend on.
func TestRoutes(t *testing.T) {
	ts := httptest.NewServer(New(discardLogger(), "test"))
	defer ts.Close()

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health ok", http.MethodGet, "/health", http.StatusOK},
		{"unknown route", http.MethodGet, "/nope", http.StatusNotFound},
		{"trailing slash is a distinct route", http.MethodGet, "/health/", http.StatusNotFound},
		{"wrong method", http.MethodPost, "/health", http.StatusMethodNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, ts.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("%s %s = %d, want %d", tc.method, tc.path, resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

// TestHealthBodyAndLog asserts the response body reaches the client and that the
// request produced exactly one structured log line.
func TestHealthBodyAndLog(t *testing.T) {
	logBuf := &lockedBuffer{}
	log := slog.New(slog.NewJSONHandler(logBuf, nil))

	ts := httptest.NewServer(New(log, "test"))
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("get /health: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Errorf("body = %s, want it to contain \"status\":\"ok\"", body)
	}

	logged := logBuf.String()
	if got := strings.Count(logged, `"msg":"http request"`); got != 1 {
		t.Errorf("http request log lines = %d, want 1; log was:\n%s", got, logged)
	}
	for _, field := range []string{`"request_id":`, `"method":"GET"`, `"path":"/health"`, `"status":200`, `"bytes":`, `"duration":`, `"remote_addr":`} {
		if !strings.Contains(logged, field) {
			t.Errorf("log line missing %s; log was:\n%s", field, logged)
		}
	}
	// RequestID must run before the logger, or this field is empty.
	if strings.Contains(logged, `"request_id":""`) {
		t.Error(`request_id is empty: chimw.RequestID must be mounted before RequestLogger`)
	}
}

// TestPanicIsLoggedAndRecovered is the regression test for middleware ordering:
// RequestLogger must sit above Recoverer so a panicking route still logs a line.
func TestPanicIsLoggedAndRecovered(t *testing.T) {
	logBuf := &lockedBuffer{}
	log := slog.New(slog.NewJSONHandler(logBuf, nil))

	// Same chain and order as New, plus a route that panics.
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimw.Recoverer)
	r.Get("/boom", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/boom")
	if err != nil {
		t.Fatalf("get /boom: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	logged := logBuf.String()
	if got := strings.Count(logged, `"msg":"http request"`); got != 1 {
		t.Errorf("http request log lines = %d, want 1; log was:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"status":500`) {
		t.Errorf("log line missing \"status\":500; log was:\n%s", logged)
	}
}
