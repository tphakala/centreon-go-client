package centreon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTokenService_List(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/tokens", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"name": "with-expiry", "type": "api", "is_revoked": false,
					"creation_date":   "2026-08-17T09:47:22+00:00",
					"expiration_date": "2036-08-14T09:47:22+00:00",
					"user":            map[string]any{"id": 1, "name": "Admin Centreon"},
					"creator":         map[string]any{"id": 2, "name": "admin"},
				},
				{
					"name": "null-expiry", "type": "api", "is_revoked": true,
					"creation_date":   "2026-08-18T08:37:49+00:00",
					"expiration_date": nil,
					"user":            map[string]any{"id": 1, "name": "Admin Centreon"},
					"creator":         map[string]any{"id": 1, "name": "Admin Centreon"},
				},
				{
					// expiration_date key entirely absent.
					"name": "absent-expiry", "type": "cma", "is_revoked": false,
					"creation_date": "2026-08-18T08:38:35+00:00",
					"user":          map[string]any{"id": 3, "name": "Op"},
					"creator":       map[string]any{"id": 1, "name": "Admin Centreon"},
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 3},
		})
	})
	resp, err := c.Tokens.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 3 {
		t.Fatalf("len(Result) = %d, want 3", len(resp.Result))
	}

	a := resp.Result[0]
	wantStr(t, "[0].Name", a.Name, "with-expiry")
	wantStr(t, "[0].Type", a.Type, "api")
	wantBool(t, "[0].IsRevoked", a.IsRevoked, false)
	wantInt(t, "[0].User.ID", a.User.ID, 1)
	wantStr(t, "[0].Creator.Name", a.Creator.Name, "admin")
	wantCreation := time.Date(2026, 8, 17, 9, 47, 22, 0, time.UTC)
	if !a.CreationDate.Equal(wantCreation) {
		t.Errorf("[0].CreationDate = %s, want %s", a.CreationDate, wantCreation)
	}
	if a.ExpirationDate == nil {
		t.Errorf("[0].ExpirationDate = nil, want non-nil")
	} else {
		wantExp := time.Date(2036, 8, 14, 9, 47, 22, 0, time.UTC)
		if !a.ExpirationDate.Equal(wantExp) {
			t.Errorf("[0].ExpirationDate = %s, want %s", a.ExpirationDate, wantExp)
		}
	}

	// A JSON null expiration_date must decode to a nil pointer, not the zero
	// time. This is what a non-pointer time.Time field would get wrong.
	if resp.Result[1].ExpirationDate != nil {
		t.Errorf("[1].ExpirationDate = %v, want nil (null in JSON)", resp.Result[1].ExpirationDate)
	}
	// An absent expiration_date key also decodes to nil.
	if resp.Result[2].ExpirationDate != nil {
		t.Errorf("[2].ExpirationDate = %v, want nil (key absent)", resp.Result[2].ExpirationDate)
	}
	wantBool(t, "[1].IsRevoked", resp.Result[1].IsRevoked, true)
	wantStr(t, "[2].Type", resp.Result[2].Type, "cma")

	// The list endpoint omits the secret, so Value must be empty for every token.
	for i, tok := range resp.Result {
		if tok.Value != "" {
			t.Errorf("Result[%d].Value = %q, want empty (list omits the secret)", i, tok.Value)
		}
	}
}

func TestTokenService_Create(t *testing.T) {
	const secret = "abcdef0123456789-secret-token-value-do-not-log"
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/administration/tokens", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["name"]); got != `"ci-token"` {
			t.Errorf("name = %s, want \"ci-token\"", got)
		}
		if got := string(raw["type"]); got != `"api"` {
			t.Errorf("type = %s, want \"api\"", got)
		}
		if got := string(raw["user_id"]); got != "1" {
			t.Errorf("user_id = %s, want 1", got)
		}
		// A nil ExpirationDate must be omitted entirely (omitzero), not sent as null.
		if _, ok := raw["expiration_date"]; ok {
			t.Errorf("expiration_date present = %s, want absent when nil", raw["expiration_date"])
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"name": "ci-token", "type": "api", "is_revoked": false,
			"creation_date":   "2026-08-18T08:38:35+00:00",
			"expiration_date": nil,
			"user":            map[string]any{"id": 1, "name": "Admin Centreon"},
			"creator":         map[string]any{"id": 1, "name": "Admin Centreon"},
			"token":           secret,
		})
	})
	tok, err := c.Tokens.Create(t.Context(), &CreateTokenRequest{Name: "ci-token", Type: "api", UserID: 1})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	wantStr(t, "Name", tok.Name, "ci-token")
	wantStr(t, "Type", tok.Type, "api")
	// The secret is returned only on Create, in the "token" key.
	wantStr(t, "Value", tok.Value, secret)
}

func TestTokenService_CreateWithExpiration(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/administration/tokens", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		// A non-nil ExpirationDate must serialize as an RFC3339 string.
		if got := string(raw["expiration_date"]); got != `"2027-01-01T00:00:00Z"` {
			t.Errorf("expiration_date = %s, want \"2027-01-01T00:00:00Z\"", got)
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"name": "exp-token", "type": "api",
			"creation_date": "2026-08-18T08:38:35+00:00",
			"user":          map[string]any{"id": 1, "name": "Admin Centreon"},
			"creator":       map[string]any{"id": 1, "name": "Admin Centreon"},
			"token":         "x",
		})
	})
	exp := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := c.Tokens.Create(t.Context(), &CreateTokenRequest{
		Name: "exp-token", Type: "api", UserID: 1, ExpirationDate: &exp,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestTokenService_CreateError(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{"validation", http.StatusUnprocessableEntity},
		{"conflict", http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux, c := newTestMux(t)
			mux.HandleFunc("POST /centreon/api/latest/administration/tokens", func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tc.status, map[string]any{"code": 0, "message": "bad"})
			})
			tok, err := c.Tokens.Create(t.Context(), &CreateTokenRequest{Name: "x", Type: "api", UserID: 1})
			if err == nil {
				t.Fatal("Create: expected error, got nil")
			}
			if tok != nil {
				t.Errorf("Create token = %+v, want nil on error", tok)
			}
			apiErr, ok := errors.AsType[*APIError](err)
			if !ok || apiErr.HTTPStatus != tc.status {
				t.Errorf("err = %v, want *APIError with HTTP %d", err, tc.status)
			}
		})
	}
}

func TestTokenService_Delete(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/administration/tokens/tok-1", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Tokens.Delete(t.Context(), "tok-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

// TestTokenService_DeleteEscapesName pins that the token name is URL-escaped
// into the path. A bare handler is used (not a ServeMux pattern) so the
// assertion reads the raw escaped path directly and does not depend on mux
// pattern-escaping semantics.
func TestTokenService_DeleteEscapesName(t *testing.T) {
	// The name carries a space and a slash. url.PathEscape encodes both (space to
	// %20, slash to %2F). Without escaping the slash would stay a path separator
	// and change the request path, so this pins the escaping rather than relying
	// on net/http to encode a lone space.
	const wantPath = "/centreon/api/latest/administration/tokens/a%20b%2Fc"
	var gotPath string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	if err := c.Tokens.Delete(t.Context(), "a b/c"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotPath != wantPath {
		t.Errorf("escaped path = %q, want %q", gotPath, wantPath)
	}
}

func TestTokenService_DeleteNotFound(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("DELETE /centreon/api/latest/administration/tokens/missing", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "Token not found"})
	})
	err := c.Tokens.Delete(t.Context(), "missing")
	if err == nil {
		t.Fatal("Delete: expected error, got nil")
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok || apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("err = %v, want *APIError with HTTP 404", err)
	}
}

// TestTokenService_CreateDoesNotLogSecret is a security regression guard: the
// token secret returned by Create must never appear in any log record. The
// client logs only method, redacted URL, status, and duration (no request or
// response bodies), so this holds today; the test fails if a future change
// starts logging bodies.
func TestTokenService_CreateDoesNotLogSecret(t *testing.T) {
	const secret = "TOP-SECRET-TOKEN-VALUE-1234567890"
	c, h := newLoggedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, map[string]any{
			"name": "t", "type": "api",
			"creation_date": "2026-08-18T08:38:35+00:00",
			"user":          map[string]any{"id": 1, "name": "Admin"},
			"creator":       map[string]any{"id": 1, "name": "Admin"},
			"token":         secret,
		})
	}))
	tok, err := c.Tokens.Create(t.Context(), &CreateTokenRequest{Name: "t", Type: "api", UserID: 1})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tok.Value != secret {
		t.Fatalf("Value = %q, want the secret (test setup sanity)", tok.Value)
	}
	// At least one line was logged (the completed request), so the guard is live.
	lines := h.snapshot()
	if len(lines) == 0 {
		t.Fatal("no log lines captured; the logging path did not run")
	}
	for _, l := range lines {
		if strings.Contains(l.msg, secret) {
			t.Errorf("log message %q contains the token secret", l.msg)
		}
		for k, v := range l.attrs {
			if strings.Contains(v, secret) {
				t.Errorf("log attr %s=%q contains the token secret", k, v)
			}
		}
	}
}

// TestToken_RedactsSecret pins that a Token never emits its secret Value through
// fmt formatting (String) or structured logging (LogValue), while still showing
// the non-sensitive fields. Removing either method, or making it echo Value,
// fails this test.
func TestToken_RedactsSecret(t *testing.T) {
	const secret = "SUPER-SECRET-TOKEN-VALUE-abcdef0123456789"
	tok := Token{Name: "ci-token", Type: "api", Value: secret}

	// fmt path: String, "%v", "%s", and the Go-syntax "%#v" must all redact the
	// secret. "%#v" goes through GoString; the others through String. The verbs
	// carry a prefix so they exercise the formatter without tripping gocritic's
	// redundantSprint (which only fires on a bare "%v"/"%s").
	for _, s := range []string{tok.String(), tok.GoString(), fmt.Sprintf("tok=%v", tok), fmt.Sprintf("tok=%s", tok), fmt.Sprintf("tok=%#v", tok)} {
		if strings.Contains(s, secret) {
			t.Errorf("formatted Token leaked the secret: %q", s)
		}
		if !strings.Contains(s, "<redacted>") {
			t.Errorf("formatted Token = %q, want it to contain <redacted>", s)
		}
		if !strings.Contains(s, "ci-token") {
			t.Errorf("formatted Token = %q, want it to show the name", s)
		}
	}

	// Real slog path through a production handler: slog.JSONHandler resolves
	// slog.LogValuer, so this exercises LogValue end to end (a handler that
	// rendered the value via String would not). The emitted JSON must redact the
	// secret and must not contain it.
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("token issued", "tok", tok)
	out := buf.String()
	if strings.Contains(out, secret) {
		t.Errorf("slog JSON output leaked the secret: %s", out)
	}
	if !strings.Contains(out, "<redacted>") {
		t.Errorf("slog JSON output = %s, want it to contain <redacted>", out)
	}

	// An empty Value renders as <empty>, not <redacted>.
	if s := (Token{Name: "n", Type: "api"}).String(); !strings.Contains(s, "<empty>") {
		t.Errorf("String() for empty Value = %q, want it to contain <empty>", s)
	}
}
