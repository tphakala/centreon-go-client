package centreon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates an httptest.Server wired to handler, and a Client
// pre-configured with WithAPIToken("test-token"). The server is closed when
// the test finishes.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, WithAPIToken("test-token"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// newTestMux creates a ServeMux and a Client wired to it.
func newTestMux(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	c, _ := newTestClient(t, mux)
	return mux, c
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
