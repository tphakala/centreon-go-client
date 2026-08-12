package centreon

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
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

// TestReauthCancelled_HonorsContext verifies that a goroutine waiting on the
// re-auth semaphore returns its context's error promptly instead of blocking
// until login() finishes. The semaphore is occupied by the test, so acquireLogin
// can only proceed via the ctx.Done() arm.
func TestReauthCancelled_HonorsContext(t *testing.T) {
	mux, c := newTestMux(t)
	c.username = "admin"
	c.password = "secret"
	c.token = "expired-token"

	// Occupy the re-auth semaphore and never release it: reauthenticate must fall
	// back to the context, not block forever.
	c.loginSem <- struct{}{}

	var hostHits atomic.Int32
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, _ *http.Request) {
		hostHits.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	})

	// A short deadline, not a pre-cancelled context: the first request must
	// succeed (return 401) so execution reaches the semaphore-guarded re-auth
	// path. A pre-cancelled context would instead short-circuit the very first
	// sendRequest in do() and never exercise acquireLogin's ctx.Done() arm. The
	// deadline is generous relative to a local httptest round trip, so the block
	// happens in acquireLogin; the 2s bound below fails fast if that arm is gone.
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		var result map[string]string
		done <- c.get(ctx, "/hosts", &result)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.DeadlineExceeded or context.Canceled", err)
		}
		// Positive control: the /hosts request must have been served (401) so
		// execution reached the semaphore-guarded re-auth path and blocked in
		// acquireLogin, rather than the first sendRequest timing out earlier and
		// verifying nothing. The retry never fires (reauthenticate returns the
		// ctx error), so the handler is hit exactly once.
		if got := hostHits.Load(); got != 1 {
			t.Errorf("hosts handler hit %d times, want 1 (request must reach acquireLogin)", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("do() did not return after the context deadline; acquireLogin ignored ctx")
	}
}

// TestReauthLoginFailure_Logged verifies that a failed re-authentication (login
// rejected with a 4xx) emits a distinct error line so the operator can see the
// failure originated in the re-auth path.
func TestReauthLoginFailure_Logged(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusForbidden, map[string]any{"code": 1, "message": "login rejected"})
	})
	c, h := newLoggedClient(t, mux)
	c.username = "admin"
	c.password = "secret"
	c.token = "expired-token"

	var result map[string]string
	err := c.get(t.Context(), "/hosts", &result)
	if err == nil {
		t.Fatal("expected error from failed re-authentication")
	}
	findLine(t, h, slog.LevelError, "re-authentication failed")
}

// #42: the request URL logged on the success (debug) path must have its
// userinfo password redacted, while the real request still carries the
// credential (net/http promotes URL userinfo to a Basic auth header).
func TestSendRequest_DebugLogRedactsCredentials(t *testing.T) {
	var gotAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	c, h := newCredClient(t, mux)

	var result map[string]string
	if err := c.get(t.Context(), "/hosts", &result); err != nil {
		t.Fatalf("get: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	if gotAuth != wantAuth {
		t.Errorf("server Authorization = %q, want %q (credential must reach the wire)", gotAuth, wantAuth)
	}

	line := findLine(t, h, slog.LevelDebug, "request completed")
	got, ok := line.attrs["url"]
	if !ok {
		t.Fatal("debug record missing url attr")
	}
	if strings.Contains(got, "secret") {
		t.Errorf("logged url leaked password: %q", got)
	}
	if !strings.Contains(got, "admin:xxxxx@") {
		t.Errorf("logged url = %q, want redacted admin:xxxxx@", got)
	}
}

// #42: the URL logged on the API-error (>=400) path must be redacted too.
func TestDo_APIErrorLogRedactsCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 1, "message": "boom"})
	})
	c, h := newCredClient(t, mux)

	var result any
	if err := c.get(t.Context(), "/hosts", &result); err == nil {
		t.Fatal("expected API error")
	}

	line := findLine(t, h, slog.LevelError, "API error")
	got := line.attrs["url"]
	if strings.Contains(got, "secret") {
		t.Errorf("API-error log leaked password: %q", got)
	}
	if !strings.Contains(got, "admin:xxxxx@") {
		t.Errorf("logged url = %q, want redacted admin:xxxxx@", got)
	}
}

// #42: the URL logged on the transport-error path must be redacted.
func TestSendRequest_TransportErrorLogRedactsCredentials(t *testing.T) {
	c, h := newCredClientClosedServer(t)

	var result any
	err := c.get(t.Context(), "/hosts", &result)
	if err == nil {
		t.Fatal("expected transport error")
	}
	// The returned error must not carry the clear password either. net/http masks
	// it to admin:***@host via stripPassword; this pins that we do not regress it.
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("returned transport error leaked password: %v", err)
	}

	line := findLine(t, h, slog.LevelError, "request failed")
	got := line.attrs["url"]
	if strings.Contains(got, "secret") {
		t.Errorf("transport-error log leaked password in url attr: %q", got)
	}
	if !strings.Contains(got, "admin:xxxxx@") {
		t.Errorf("logged url = %q, want redacted admin:xxxxx@", got)
	}
	if errAttr := line.attrs["error"]; strings.Contains(errAttr, "secret") {
		t.Errorf("transport-error log leaked password in error attr: %q", errAttr)
	}
}

// #42: a url.Parse failure must not echo the raw credential-bearing input. A NUL
// byte is a control character that makes url.Parse fail; its *url.Error text
// would otherwise include the full URL.
func TestNewClient_ParseErrorDoesNotLeakCredentials(t *testing.T) {
	_, err := NewClient("https://admin:secret@\x00host/centreon")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("constructor error leaked password: %v", err)
	}
	// Redaction must preserve the *url.Error type so callers can still inspect it.
	if _, ok := errors.AsType[*url.Error](err); !ok {
		t.Errorf("expected *url.Error to be preserved in the chain, got %T", err)
	}
}

// #42: a reqURL that fails to parse inside http.NewRequestWithContext yields a
// *url.Error echoing the credentialed URL. That error is returned to the caller
// (bypassing the loggers entirely), so it must be redacted too.
func TestSendRequest_RequestBuildErrorDoesNotLeakCredentials(t *testing.T) {
	c, _ := newCredClient(t, http.NewServeMux())

	// A control character in the path makes the composed request URL invalid.
	err := c.get(t.Context(), "/hosts/\x00bad", nil)
	if err == nil {
		t.Fatal("expected request-build error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("returned error leaked password: %v", err)
	}
	if _, ok := errors.AsType[*url.Error](err); !ok {
		t.Errorf("expected *url.Error to be preserved in the chain, got %T", err)
	}
}

// #42: the "missing scheme or host" error must redact credentials for both the
// opaque form (parses with an empty Host, so url.Redacted alone would leak) and
// the scheme-relative form.
func TestNewClient_MissingSchemeDoesNotLeakCredentials(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		wantSubstr string
	}{
		{"opaque", "admin:secret@host/centreon", "xxxxx@host"},
		{"authority", "//admin:secret@host", "admin:xxxxx@host"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.baseURL)
			if err == nil {
				t.Fatalf("expected error for %q", tt.baseURL)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Errorf("constructor error leaked password: %v", err)
			}
			if !strings.Contains(err.Error(), "missing scheme or host") {
				t.Errorf("error = %v, want to mention 'missing scheme or host'", err)
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Errorf("error = %v, want to contain %q", err, tt.wantSubstr)
			}
		})
	}
}

// #44: a request failing because the caller's context was cancelled must log at
// Debug ("request cancelled"), never at Error.
func TestSendRequest_ContextCanceledLogsDebug(t *testing.T) {
	entered := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done() // block until the client cancels
	})
	c, h := newLoggedClient(t, mux)

	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		var result any
		errCh <- c.get(ctx, "/hosts", &result)
	}()

	<-entered
	cancel()

	err := <-errCh
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if n := countAtLevel(h, slog.LevelError); n != 0 {
		t.Errorf("got %d Error records, want 0 (cancellation must not log at Error)", n)
	}
	findLine(t, h, slog.LevelDebug, "request cancelled")
}

// #44: a request timing out (DeadlineExceeded) is a real failure and must stay
// at Error, not be demoted to Debug like a cancellation.
func TestSendRequest_DeadlineExceededLogsError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /centreon/api/latest/hosts", func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the deadline fires
	})
	c, h := newLoggedClient(t, mux)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	var result any
	err := c.get(ctx, "/hosts", &result)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	findLine(t, h, slog.LevelError, "request failed")
	if n := countAtLevel(h, slog.LevelDebug); n != 0 {
		t.Errorf("got %d Debug records, want 0 (a timeout is a real failure)", n)
	}
}

// #42: redactURL is the security-critical helper. Pin its full accepted-input
// domain directly, including inputs whose userinfo contains '/', '?', '#', or a
// second '@', and scheme-relative / opaque forms that url.Parse does not expose
// as userinfo. Every case must both match the exact expected output and never
// contain the clear password.
func TestRedactURL(t *testing.T) {
	const secret = "secret"
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain credential url keeps username", "https://admin:secret@host/centreon/api", "https://admin:xxxxx@host/centreon/api"},
		{"no credentials unchanged", "https://host:8443/centreon/api/latest/hosts", "https://host:8443/centreon/api/latest/hosts"},
		{"scheme-relative parseable keeps username", "//admin:secret@host", "//admin:xxxxx@host"},
		{"opaque masks whole userinfo", "admin:secret@host/centreon", "xxxxx@host/centreon"},
		{"scheme-relative parse-fail", "//admin:secret@\x00host", "xxxxx@\x00host"},
		{"full url parse-fail", "https://admin:secret@\x00host/centreon", "xxxxx@\x00host/centreon"},
		{"slash in password", "http://admin:secret/x@host", "xxxxx@host"},
		{"question mark in password", "http://admin:pw?secret@host", "xxxxx@host"},
		{"second at in userinfo", "user:p@secret@host", "xxxxx@host"},
		{"scheme marker after userinfo", "admin:secret@host/a://b", "xxxxx@host/a://b"},
		{"scheme marker inside userinfo", "admin:secret://host@x", "xxxxx@x"},
		{"host colon port not userinfo", "host:8080", "host:8080"},
		{"at in path no creds", "///a@b", "xxxxx@b"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactURL(tt.in)
			if got != tt.want {
				t.Errorf("redactURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if strings.Contains(got, secret) {
				t.Errorf("redactURL(%q) leaked the password: %q", tt.in, got)
			}
		})
	}
}
