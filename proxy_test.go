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

	// Port must appear in BOTH String and slog output (parity: a non-default port
	// must be identifiable in structured logs, not only in String).
	if s := p.String(); !strings.Contains(s, "8080") {
		t.Errorf("String() = %q, want it to contain the port 8080", s)
	}
	var pbuf bytes.Buffer
	slog.New(slog.NewJSONHandler(&pbuf, nil)).Info("proxy", "cfg", p)
	if out := pbuf.String(); !strings.Contains(out, "8080") {
		t.Errorf("slog output = %s, want it to contain the port 8080", out)
	}

	// A nil Password renders as <empty>, not <redacted>, and a nil Port must be
	// rendered nil-safe (no panic) on both the String and slog paths.
	nilCfg := ProxyConfiguration{Protocol: "http://"}
	if s := nilCfg.String(); !strings.Contains(s, "<empty>") {
		t.Errorf("String() for nil Password = %q, want it to contain <empty>", s)
	}
	var nbuf bytes.Buffer
	slog.New(slog.NewJSONHandler(&nbuf, nil)).Info("proxy", "cfg", nilCfg) // must not panic on nil Port/Password
}

// TestProxyService_Update pins the PUT body: it must carry only url/port/user/
// password and MUST NOT carry "protocol". Sending protocol returns HTTP 500 on
// 25.10.16 (server defect), and Update deliberately omits it so a
// Get-modify-Update round-trip works. The password is written in cleartext (the
// wire needs the real value; redaction lives only in String/slog).
func TestProxyService_Update(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("PUT /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})

	url, user, pw := "proxy.example.com", "puser", "s3cr3t-pw"
	port := 3128
	// Protocol is set on the input to prove Update strips it from the wire.
	req := &ProxyConfiguration{URL: &url, Port: &port, User: &user, Password: &pw, Protocol: "http://"}
	if err := c.Proxy.Update(t.Context(), req); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// The emitted body must not include protocol (the poisoned key).
	wantJSONKeys(t, "proxy update body", gotBody, "url", "port", "user", "password")
	if _, ok := gotBody["protocol"]; ok {
		t.Errorf("proxy update body carries protocol %v, want it omitted", gotBody["protocol"])
	}
	wantStr(t, "body.url", fmt.Sprintf("%v", gotBody["url"]), "proxy.example.com")
	wantStr(t, "body.user", fmt.Sprintf("%v", gotBody["user"]), "puser")
	// password must be the cleartext value, not redacted, on the wire.
	wantStr(t, "body.password", fmt.Sprintf("%v", gotBody["password"]), "s3cr3t-pw")
	if got, ok := gotBody["port"].(float64); !ok || int(got) != 3128 {
		t.Errorf("body.port = %v, want 3128", gotBody["port"])
	}
}

// TestProxyService_Update_ClearsFields pins that nil fields serialize as JSON null
// (a full-replace PUT clears them), still without a protocol key.
func TestProxyService_Update_ClearsFields(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("PUT /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Proxy.Update(t.Context(), &ProxyConfiguration{}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	wantJSONKeys(t, "proxy update body", gotBody, "url", "port", "user", "password")
	for _, k := range []string{"url", "port", "user", "password"} {
		if v, ok := gotBody[k]; !ok || v != nil {
			t.Errorf("body.%s = %v (present=%v), want explicit null", k, v, ok)
		}
	}
}

func TestProxyService_Update_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("PUT /centreon/api/latest/configuration/proxy", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	err := c.Proxy.Update(t.Context(), &ProxyConfiguration{})
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
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
