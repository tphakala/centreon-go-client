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
)

func TestProxyService_Get_Unset(t *testing.T) {
	mux, c := newTestMux(t)
	// Unconfigured proxy: url/port/user/password are null, protocol defaults to
	// "http://". The nullable fields must decode to nil pointers (distinct from an
	// empty string or 0); retyping any of them to a non-pointer would fail here.
	mux.HandleFunc("GET /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"url": nil, "port": nil, "user": nil, "password": nil, "protocol": "http://",
		})
	})

	got, err := c.Proxy.Get(t.Context())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantNilStrPtr(t, "URL", got.URL)
	wantNilIntPtr(t, "Port", got.Port)
	wantNilStrPtr(t, "User", got.User)
	wantNilStrPtr(t, "Password", got.Password)
	wantStr(t, "Protocol", got.Protocol, "http://")
}

func TestProxyService_Get_Configured(t *testing.T) {
	mux, c := newTestMux(t)
	// A configured proxy returns the password in cleartext (live-verified); the
	// pointer fields must carry the values.
	mux.HandleFunc("GET /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"url": "proxy.example.com", "port": 8080, "user": "puser",
			"password": "s3cr3t-pw", "protocol": "http://",
		})
	})

	got, err := c.Proxy.Get(t.Context())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantStrPtr(t, "URL", got.URL, "proxy.example.com")
	wantIntPtr(t, "Port", got.Port, 8080)
	wantStrPtr(t, "User", got.User, "puser")
	wantStrPtr(t, "Password", got.Password, "s3cr3t-pw")
	wantStr(t, "Protocol", got.Protocol, "http://")
}

// TestProxyConfiguration_RedactsSecret pins that a ProxyConfiguration never emits
// its Password through fmt formatting (String/GoString) or structured logging
// (LogValue), while json.Marshal still writes it. Removing any of the three
// methods, or making one echo the password, fails this test.
func TestProxyConfiguration_RedactsSecret(t *testing.T) {
	const secret = "s3cr3t-proxy-pw-abcdef"
	url, user, pw := "proxy.example.com", "puser", secret
	port := 8080
	p := ProxyConfiguration{URL: &url, Port: &port, User: &user, Password: &pw, Protocol: "http://"}

	// fmt path: String, GoString ("%#v"), "%v", "%s" must all redact. The verbs
	// carry a prefix so they exercise the formatter without tripping gocritic's
	// redundantSprint.
	for _, s := range []string{p.String(), p.GoString(), fmt.Sprintf("p=%v", p), fmt.Sprintf("p=%s", p), fmt.Sprintf("p=%#v", p)} {
		if strings.Contains(s, secret) {
			t.Errorf("formatted ProxyConfiguration leaked the password: %q", s)
		}
		if !strings.Contains(s, "<redacted>") {
			t.Errorf("formatted ProxyConfiguration = %q, want it to contain <redacted>", s)
		}
		if !strings.Contains(s, "proxy.example.com") {
			t.Errorf("formatted ProxyConfiguration = %q, want it to show the url", s)
		}
	}

	// Real slog path: slog.JSONHandler resolves slog.LogValuer, exercising
	// LogValue end to end. The emitted JSON must redact the password.
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("proxy", "cfg", p)
	if out := buf.String(); strings.Contains(out, secret) {
		t.Errorf("slog JSON output leaked the password: %s", out)
	} else if !strings.Contains(out, "<redacted>") {
		t.Errorf("slog JSON output = %s, want it to contain <redacted>", out)
	}

	// json.Marshal MUST still write the password so a caller can persist config.
	//nolint:gosec // G117: this assertion deliberately marshals the password; it pins the persist contract (redact in logs, but keep it in JSON), mirroring Token.Value.
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(data), secret) {
		t.Errorf("json.Marshal output = %s, want it to contain the password", data)
	}

	// A nil Password renders as <empty>, not <redacted>, and must not panic.
	if s := (ProxyConfiguration{Protocol: "http://"}).String(); !strings.Contains(s, "<empty>") {
		t.Errorf("String() for nil Password = %q, want it to contain <empty>", s)
	}
}

func TestProxyService_Get_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.Proxy.Get(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}
