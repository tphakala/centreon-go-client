package centreon

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestLogin_Success(t *testing.T) {
	mux, c := newTestMux(t)
	// Override: use credentials instead of token
	c.token = ""
	c.username = "admin"
	c.password = "secret"

	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, r *http.Request) {
		// Verify request body structure
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode login request: %v", err)
		}
		if req.Security.Credentials.Login != "admin" {
			t.Errorf("login = %q, want %q", req.Security.Credentials.Login, "admin")
		}
		if req.Security.Credentials.Password != "secret" {
			t.Errorf("password = %q, want %q", req.Security.Credentials.Password, "secret")
		}

		writeJSON(w, 200, loginResponse{
			Security: loginSecurityResponse{Token: "new-token-123"},
		})
	})

	err := c.Login(t.Context())
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	c.mu.Lock()
	token := c.token
	c.mu.Unlock()
	if token != "new-token-123" {
		t.Errorf("token = %q, want %q", token, "new-token-123")
	}
}

func TestLogin_NoCredentials(t *testing.T) {
	c, _ := NewClient("http://example.com")
	err := c.Login(t.Context())
	if err == nil {
		t.Fatal("expected error when no credentials are set")
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	mux, c := newTestMux(t)
	c.token = ""
	c.username = "admin"
	c.password = "wrong"

	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 401, map[string]any{"code": 1, "message": "bad credentials"})
	})

	err := c.Login(t.Context())
	if err == nil {
		t.Fatal("expected error for bad credentials")
	}
}

func TestAutoRenew_On401(t *testing.T) {
	mux, c := newTestMux(t)
	c.username = "admin"
	c.password = "secret"
	c.token = "expired-token"

	var hostCalls atomic.Int32

	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, loginResponse{
			Security: loginSecurityResponse{Token: "fresh-token"},
		})
	})

	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		call := hostCalls.Add(1)
		token := r.Header.Get("X-AUTH-TOKEN")

		if call == 1 {
			// First call: reject with 401
			if token != "expired-token" {
				t.Errorf("first call token = %q, want %q", token, "expired-token")
			}
			w.WriteHeader(401)
			return
		}
		// Second call: should have fresh token
		if token != "fresh-token" {
			t.Errorf("second call token = %q, want %q", token, "fresh-token")
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})

	var result map[string]string
	err := c.get(t.Context(), "/hosts", &result)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want %q", result["status"], "ok")
	}
	if hostCalls.Load() != 2 {
		t.Errorf("host handler called %d times, want 2", hostCalls.Load())
	}
}

func TestLogout_ClearsToken(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/logout", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	})

	err := c.Logout(t.Context())
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}

	c.mu.Lock()
	token := c.token
	c.mu.Unlock()
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
}
