package centreon

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"
)

// newTestClient creates an httptest.Server wired to handler and returns a Client
// pre-configured with WithAPIToken("test-token"). The server is closed when the
// test finishes.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, WithAPIToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// newTestMux creates a ServeMux and a Client wired to it.
func newTestMux(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	c := newTestClient(t, mux)
	return mux, c
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) //nolint:errchkjson // test helper
}

// The want* helpers assert a single decoded field, keeping field-by-field
// round-trip tests flat (one call per field) so a mistyped or mistagged struct
// field still fails, without inflating each test's cognitive complexity.

func wantInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

func wantStr(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}

func wantBool(t *testing.T, name string, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func wantIntPtr(t *testing.T, name string, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %d", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %d, want %d", name, *got, want)
	}
}

func wantNilIntPtr(t *testing.T, name string, got *int) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %v, want nil", name, got)
	}
}

func wantStrPtr(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %q", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %q, want %q", name, *got, want)
	}
}

func wantNilStrPtr(t *testing.T, name string, got *string) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %q, want nil", name, *got)
	}
}

func wantStrSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// logLine is a captured log record reduced to the fields tests assert on, so
// the heavy slog.Record is not copied around after capture.
type logLine struct {
	level slog.Level
	msg   string
	attrs map[string]string
}

// recordingHandler is a minimal slog.Handler that captures emitted records so
// tests can assert on level, message, and individual attributes.
type recordingHandler struct {
	mu    sync.Mutex
	lines []logLine
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

//nolint:gocritic // slog.Handler.Handle requires slog.Record by value.
func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	line := logLine{level: r.Level, msg: r.Message, attrs: make(map[string]string, r.NumAttrs())}
	r.Attrs(func(a slog.Attr) bool {
		line.attrs[a.Key] = a.Value.String()
		return true
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lines = append(h.lines, line)
	return nil
}

// WithAttrs and WithGroup return the same handler unchanged: the client logs
// via Debug/Error directly and never derives a sub-logger, so there is no
// attr/group state to thread through here.
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

// snapshot returns a copy of the log lines captured so far.
func (h *recordingHandler) snapshot() []logLine {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.lines)
}

// findLine returns the single captured line at the given level with the given
// message, failing if there is not exactly one.
func findLine(t *testing.T, h *recordingHandler, level slog.Level, msg string) logLine {
	t.Helper()
	var matches []logLine
	for _, l := range h.snapshot() {
		if l.level == level && l.msg == msg {
			matches = append(matches, l)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("found %d records at %v with msg %q, want exactly 1", len(matches), level, msg)
	}
	return matches[0]
}

// countAtLevel returns how many captured lines are at exactly the given level.
func countAtLevel(h *recordingHandler, level slog.Level) int {
	n := 0
	for _, l := range h.snapshot() {
		if l.level == level {
			n++
		}
	}
	return n
}

// newCredClient starts a test server wired to handler and returns a Client whose
// base URL embeds userinfo credentials, together with the recording handler
// capturing its logs. Real requests therefore carry the credential (net/http
// promotes URL userinfo to a Basic auth header) while the logs must not.
func newCredClient(t *testing.T, handler http.Handler) (*Client, *recordingHandler) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse srv.URL: %v", err)
	}
	h := &recordingHandler{}
	c, err := NewClient("http://admin:secret@"+u.Host, WithAPIToken("test-token"), WithLogger(slog.New(h)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, h
}

// newCredClientClosedServer returns a credentialed, logging Client pointed at a
// server that has already been closed, so the next request fails with a
// transport error (connection refused).
func newCredClientClosedServer(t *testing.T) (*Client, *recordingHandler) {
	t.Helper()
	srv := httptest.NewServer(http.NewServeMux())
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse srv.URL: %v", err)
	}
	srv.Close()
	h := &recordingHandler{}
	c, err := NewClient("http://admin:secret@"+u.Host, WithAPIToken("test-token"), WithLogger(slog.New(h)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, h
}

// newLoggedClient starts a test server wired to handler and returns a Client
// with a recording logger and a plain (credential-free) base URL.
func newLoggedClient(t *testing.T, handler http.Handler) (*Client, *recordingHandler) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	h := &recordingHandler{}
	c, err := NewClient(srv.URL, WithAPIToken("test-token"), WithLogger(slog.New(h)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, h
}
