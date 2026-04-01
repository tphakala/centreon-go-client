package centreon

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c, err := NewClient("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiVersion != "latest" {
		t.Errorf("apiVersion = %q, want %q", c.apiVersion, "latest")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

func TestNewClient_WithVersion(t *testing.T) {
	c, err := NewClient("http://example.com", WithVersion("v2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiVersion != "v2" {
		t.Errorf("apiVersion = %q, want %q", c.apiVersion, "v2")
	}
}

func TestNewClient_InvalidURL(t *testing.T) {
	_, err := NewClient("://bad-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestBuildURL_NoTrailingSlash(t *testing.T) {
	c, _ := NewClient("http://example.com")
	got := c.buildURL("/hosts")
	want := "http://example.com/centreon/api/latest/hosts"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildURL_TrailingSlash(t *testing.T) {
	c, _ := NewClient("http://example.com/")
	got := c.buildURL("/hosts")
	want := "http://example.com/centreon/api/latest/hosts"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGet_TokenHeader(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-AUTH-TOKEN")
		if token != "test-token" {
			t.Errorf("X-AUTH-TOKEN = %q, want %q", token, "test-token")
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	var result map[string]string
	err := c.get(t.Context(), "/hosts", &result)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want %q", result["status"], "ok")
	}
}

func TestPost_ContentType(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 42})
	})

	body := map[string]string{"name": "host1"}
	var result map[string]int
	err := c.post(t.Context(), "/hosts", body, &result)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if result["id"] != 42 {
		t.Errorf("id = %d, want 42", result["id"])
	}
}

func TestDelete_204(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("DELETE /centreon/api/latest/hosts/1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.delete(t.Context(), "/hosts/1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestErrorResponse_ParsedAsAPIError(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusForbidden, map[string]any{"code": 42, "message": "forbidden"})
	})

	var result any
	err := c.get(t.Context(), "/hosts", &result)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != 403 {
		t.Errorf("HTTPStatus = %d, want 403", apiErr.HTTPStatus)
	}
	if apiErr.Code != 42 {
		t.Errorf("Code = %d, want 42", apiErr.Code)
	}
	if !strings.Contains(apiErr.Message, "forbidden") {
		t.Errorf("Message = %q, want to contain %q", apiErr.Message, "forbidden")
	}
}

func TestNewClient_WithCredentials(t *testing.T) {
	c, _ := NewClient("http://example.com", WithCredentials("admin", "secret"))
	if c.username != "admin" || c.password != "secret" {
		t.Errorf("credentials = %q/%q, want admin/secret", c.username, c.password)
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c, _ := NewClient("http://example.com", WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Error("expected custom HTTP client to be used")
	}
}

func TestToken_ReturnsPreConfiguredToken(t *testing.T) {
	c, err := NewClient("http://example.com", WithAPIToken("my-static-token"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.Token(); got != "my-static-token" {
		t.Errorf("Token() = %q, want %q", got, "my-static-token")
	}
}

func TestConcurrent401_LoginCalledOnce(t *testing.T) {
	mux, c := newTestMux(t)
	c.username = "admin"
	c.password = "secret"
	c.token = "expired-token"

	var loginCalls atomic.Int32

	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, _ *http.Request) {
		loginCalls.Add(1)
		writeJSON(w, http.StatusOK, loginResponse{
			Security: loginSecurityResponse{Token: "fresh-token"},
		})
	})

	// Return 401 for expired-token, 200 for fresh-token.
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-AUTH-TOKEN")
		if token == "expired-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Fire two concurrent requests that will both encounter the 401.
	const goroutines = 2
	errs := make([]error, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			var result map[string]string
			errs[i] = c.get(t.Context(), "/hosts", &result)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, err)
		}
	}

	if n := loginCalls.Load(); n != 1 {
		t.Errorf("login called %d times, want 1", n)
	}
}
