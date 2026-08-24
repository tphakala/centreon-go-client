package centreon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// wantJSONKeys marshals v and asserts its exact top-level JSON key set. It
// catches a mistyped, missing, or extra json tag on any struct (the repo's
// dominant wire-format defect class); a swapped-tag pair keeps the same key set
// and is caught instead by the value assertions in the decode tests.
func wantJSONKeys(t *testing.T, name string, v any, want ...string) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%s): %v", name, err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal(%s): %v", name, err)
	}
	got := slices.Sorted(maps.Keys(m))
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("%s json keys = %v, want %v", name, got, want)
	}
}

func TestAuthProvider_GetLocal(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/authentication/providers/local", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"password_security_policy": map[string]any{
				"password_min_length": 12, "has_uppercase": true, "has_lowercase": true,
				"has_number": true, "has_special_character": true, "attempts": 5,
				"blocking_duration": 900,
				"password_expiration": map[string]any{
					"expiration_delay": 15552000, "excluded_users": []string{"centreon-gorgone"},
				},
				"can_reuse_passwords": false, "delay_before_new_password": 3600,
			},
		})
	})

	got, err := c.Authentication.GetLocal(t.Context())
	if err != nil {
		t.Fatalf("GetLocal: %v", err)
	}
	p := got.PasswordSecurityPolicy
	wantInt(t, "PasswordMinLength", p.PasswordMinLength, 12)
	wantBool(t, "HasUppercase", p.HasUppercase, true)
	wantBool(t, "HasSpecialCharacter", p.HasSpecialCharacter, true)
	wantInt(t, "Attempts", p.Attempts, 5)
	wantInt(t, "BlockingDuration", p.BlockingDuration, 900)
	wantInt(t, "PasswordExpiration.ExpirationDelay", p.PasswordExpiration.ExpirationDelay, 15552000)
	wantStrSlice(t, "PasswordExpiration.ExcludedUsers", p.PasswordExpiration.ExcludedUsers, []string{"centreon-gorgone"})
	wantBool(t, "CanReusePasswords", p.CanReusePasswords, false)
	wantInt(t, "DelayBeforeNewPassword", p.DelayBeforeNewPassword, 3600)
}

func TestAuthProvider_GetOpenID(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/authentication/providers/openid", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, openIDFreshBox())
	})

	got, err := c.Authentication.GetOpenID(t.Context())
	if err != nil {
		t.Fatalf("GetOpenID: %v", err)
	}
	wantBool(t, "IsActive", got.IsActive, false)
	// Nullable scalars decode to nil pointers, distinct from an empty string.
	wantNilStrPtr(t, "BaseURL", got.BaseURL)
	wantNilStrPtr(t, "ClientID", got.ClientID)
	wantNilStrPtr(t, "ClientSecret", got.ClientSecret)
	wantNilStrPtr(t, "RedirectURL", got.RedirectURL)
	wantStr(t, "AuthenticationType", got.AuthenticationType, "client_secret_post")
	wantBool(t, "VerifyPeer", got.VerifyPeer, true)
	// connection_scopes = [] must decode to a non-nil empty slice.
	if got.ConnectionScopes == nil {
		t.Error("ConnectionScopes = nil, want non-nil empty slice")
	}
	wantInt(t, "len(ConnectionScopes)", len(got.ConnectionScopes), 0)
	// The endpoint custom_endpoint distinction: "" in roles_mapping (pointer to
	// ""), null in authentication_conditions and groups_mapping (nil). Retyping
	// CustomEndpoint to string would collapse this distinction.
	wantStrPtr(t, "RolesMapping.Endpoint.CustomEndpoint", got.RolesMapping.Endpoint.CustomEndpoint, "")
	wantStr(t, "RolesMapping.Endpoint.Type", got.RolesMapping.Endpoint.Type, "introspection_endpoint")
	wantNilStrPtr(t, "AuthenticationConditions.Endpoint.CustomEndpoint", got.AuthenticationConditions.Endpoint.CustomEndpoint)
	wantNilStrPtr(t, "GroupsMapping.Endpoint.CustomEndpoint", got.GroupsMapping.Endpoint.CustomEndpoint)
}

func TestAuthProvider_GetSAML(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/authentication/providers/saml", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, samlFreshBox("cert-data"))
	})

	got, err := c.Authentication.GetSAML(t.Context())
	if err != nil {
		t.Fatalf("GetSAML: %v", err)
	}
	wantStr(t, "Certificate", got.Certificate, "cert-data")
	wantStr(t, "RequestedAuthnContextComparison", got.RequestedAuthnContextComparison, "exact")
	wantNilStrPtr(t, "LogoutFromURL", got.LogoutFromURL)
	// SAML authentication_conditions has only these three keys (no endpoint, no
	// client-address lists), unlike OpenID.
	wantJSONKeys(t, "SAMLAuthenticationConditions", got.AuthenticationConditions,
		"is_enabled", "attribute_path", "authorized_values")
	// SAML mappings have no endpoint object.
	wantJSONKeys(t, "SAMLRolesMapping", got.RolesMapping,
		"is_enabled", "apply_only_first_role", "attribute_path", "relations")
	wantJSONKeys(t, "SAMLGroupsMapping", got.GroupsMapping,
		"is_enabled", "attribute_path", "relations")
}

func TestAuthProvider_GetWebSSO(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/authentication/providers/web-sso", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_active": false, "is_forced": false,
			"trusted_client_addresses": []string{}, "blacklist_client_addresses": []string{},
			"login_header_attribute": "HTTP_AUTH_USER",
			"pattern_matching_login": nil, "pattern_replace_login": nil,
		})
	})

	got, err := c.Authentication.GetWebSSO(t.Context())
	if err != nil {
		t.Fatalf("GetWebSSO: %v", err)
	}
	wantStr(t, "LoginHeaderAttribute", got.LoginHeaderAttribute, "HTTP_AUTH_USER")
	wantNilStrPtr(t, "PatternMatchingLogin", got.PatternMatchingLogin)
	wantNilStrPtr(t, "PatternReplaceLogin", got.PatternReplaceLogin)
	if got.TrustedClientAddresses == nil {
		t.Error("TrustedClientAddresses = nil, want non-nil empty slice")
	}
}

func TestAuthProvider_Update_SendsBody(t *testing.T) {
	// Each Update* must PUT (the route is registered for PUT only, so a wrong verb
	// 404s) and marshal the object as the body. For openid/saml the marshaled body
	// must still contain the secret (the PUT needs it), proving redaction does not
	// touch json.Marshal.
	tests := []struct {
		name       string
		route      string
		call       func(c *Client) error
		wantSecret string
	}{
		{
			"local", "PUT /centreon/api/latest/administration/authentication/providers/local",
			func(c *Client) error {
				return c.Authentication.UpdateLocal(t.Context(), &LocalProvider{
					PasswordSecurityPolicy: LocalPasswordSecurityPolicy{PasswordMinLength: 8},
				})
			}, "",
		},
		{
			"openid", "PUT /centreon/api/latest/administration/authentication/providers/openid",
			func(c *Client) error {
				secret := "oidc-secret-xyz"
				return c.Authentication.UpdateOpenID(t.Context(), &OpenIDProvider{ClientSecret: &secret})
			}, "oidc-secret-xyz",
		},
		{
			"saml", "PUT /centreon/api/latest/administration/authentication/providers/saml",
			func(c *Client) error {
				return c.Authentication.UpdateSAML(t.Context(), &SAMLProvider{Certificate: "saml-cert-xyz"})
			}, "saml-cert-xyz",
		},
		{
			"web-sso", "PUT /centreon/api/latest/administration/authentication/providers/web-sso",
			func(c *Client) error {
				return c.Authentication.UpdateWebSSO(t.Context(), &WebSSOProvider{LoginHeaderAttribute: "HTTP_X"})
			}, "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, c := newTestMux(t)
			var gotBody string
			mux.HandleFunc(tt.route, func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				gotBody = string(b)
				w.WriteHeader(http.StatusNoContent)
			})
			if err := tt.call(c); err != nil {
				t.Fatalf("Update: %v", err)
			}
			if gotBody == "" {
				t.Fatal("expected a request body, got empty")
			}
			if tt.wantSecret != "" && !strings.Contains(gotBody, tt.wantSecret) {
				t.Errorf("PUT body = %s, want it to contain the secret %q (json.Marshal must keep it)", gotBody, tt.wantSecret)
			}
		})
	}
}

func TestOpenIDProvider_RedactsSecret(t *testing.T) {
	const secret = "oidc-client-secret-abcdef"
	sec := secret
	base := "https://idp.example.com"
	cid := "client-123"
	p := OpenIDProvider{IsActive: true, BaseURL: &base, ClientID: &cid, ClientSecret: &sec, AuthenticationType: "client_secret_post"}

	// fmt path: String, GoString ("%#v"), "%v", "%s" must all redact.
	for _, s := range []string{p.String(), p.GoString(), fmt.Sprintf("p=%v", p), fmt.Sprintf("p=%s", p), fmt.Sprintf("p=%#v", p)} {
		if strings.Contains(s, secret) {
			t.Errorf("formatted OpenIDProvider leaked the secret: %q", s)
		}
		if !strings.Contains(s, "<redacted>") {
			t.Errorf("formatted OpenIDProvider = %q, want it to contain <redacted>", s)
		}
		if !strings.Contains(s, "client-123") {
			t.Errorf("formatted OpenIDProvider = %q, want it to show the client_id", s)
		}
	}

	// Real slog path.
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("openid", "cfg", p)
	if out := buf.String(); strings.Contains(out, secret) {
		t.Errorf("slog output leaked the secret: %s", out)
	} else if !strings.Contains(out, "<redacted>") {
		t.Errorf("slog output = %s, want it to contain <redacted>", out)
	}

	// json.Marshal MUST still write the secret so a caller can PUT it back.
	//nolint:gosec // G117: this assertion deliberately marshals the secret; it pins the persist contract (redact in logs, but keep it in JSON so a full-replace PUT carries it), mirroring Token.Value and ProxyConfiguration.Password.
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(data), secret) {
		t.Errorf("json.Marshal output = %s, want it to contain the secret", data)
	}

	// A nil ClientSecret renders as <empty>, nil-safe on both paths.
	nilP := OpenIDProvider{AuthenticationType: "client_secret_post"}
	if s := nilP.String(); !strings.Contains(s, "<empty>") {
		t.Errorf("String() for nil ClientSecret = %q, want <empty>", s)
	}
	var nbuf bytes.Buffer
	slog.New(slog.NewJSONHandler(&nbuf, nil)).Info("openid", "cfg", nilP) // must not panic
}

func TestSAMLProvider_RedactsSecret(t *testing.T) {
	const secret = "-----BEGIN CERT-----saml-secret-material-----END CERT-----"
	p := SAMLProvider{IsActive: true, EntityIDURL: "https://sp.example.com", RemoteLoginURL: "https://idp.example.com/sso", Certificate: secret}

	for _, s := range []string{p.String(), p.GoString(), fmt.Sprintf("p=%v", p), fmt.Sprintf("p=%s", p), fmt.Sprintf("p=%#v", p)} {
		if strings.Contains(s, secret) {
			t.Errorf("formatted SAMLProvider leaked the certificate: %q", s)
		}
		if !strings.Contains(s, "<redacted>") {
			t.Errorf("formatted SAMLProvider = %q, want it to contain <redacted>", s)
		}
		if !strings.Contains(s, "https://sp.example.com") {
			t.Errorf("formatted SAMLProvider = %q, want it to show the entity_id_url", s)
		}
	}

	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("saml", "cfg", p)
	if out := buf.String(); strings.Contains(out, secret) {
		t.Errorf("slog output leaked the certificate: %s", out)
	} else if !strings.Contains(out, "<redacted>") {
		t.Errorf("slog output = %s, want it to contain <redacted>", out)
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if !strings.Contains(string(data), secret) {
		t.Errorf("json.Marshal output = %s, want it to contain the certificate", data)
	}

	// Empty Certificate renders as <empty>.
	empty := SAMLProvider{EntityIDURL: "x"}
	if s := empty.String(); !strings.Contains(s, "<empty>") {
		t.Errorf("String() for empty Certificate = %q, want <empty>", s)
	}
}

func TestAuthProvider_Get_Error(t *testing.T) {
	tests := []struct {
		name  string
		route string
		call  func(c *Client) (any, error)
	}{
		{"GetLocal", "GET /centreon/api/latest/administration/authentication/providers/local",
			func(c *Client) (any, error) { return c.Authentication.GetLocal(t.Context()) }},
		{"GetOpenID", "GET /centreon/api/latest/administration/authentication/providers/openid",
			func(c *Client) (any, error) { return c.Authentication.GetOpenID(t.Context()) }},
		{"GetSAML", "GET /centreon/api/latest/administration/authentication/providers/saml",
			func(c *Client) (any, error) { return c.Authentication.GetSAML(t.Context()) }},
		{"GetWebSSO", "GET /centreon/api/latest/administration/authentication/providers/web-sso",
			func(c *Client) (any, error) { return c.Authentication.GetWebSSO(t.Context()) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, c := newTestMux(t)
			mux.HandleFunc(tt.route, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
			})
			_, err := tt.call(c)
			if err == nil {
				t.Fatal("expected an error on HTTP 500, got nil")
			}
			if _, ok := errors.AsType[*APIError](err); !ok {
				t.Errorf("expected *APIError, got %T", err)
			}
		})
	}
}

// TestAuthProvider_WireTags pins the exact top-level json key set of every
// provider and nested type, so a mistyped/dropped/extra tag fails even when no
// decode test happens to read that field.
func TestAuthProvider_WireTags(t *testing.T) {
	wantJSONKeys(t, "LocalProvider", LocalProvider{}, "password_security_policy")
	wantJSONKeys(t, "LocalPasswordSecurityPolicy", LocalPasswordSecurityPolicy{},
		"password_min_length", "has_uppercase", "has_lowercase", "has_number",
		"has_special_character", "attempts", "blocking_duration", "password_expiration",
		"can_reuse_passwords", "delay_before_new_password")
	wantJSONKeys(t, "LocalPasswordExpiration", LocalPasswordExpiration{},
		"expiration_delay", "excluded_users")

	wantJSONKeys(t, "OpenIDProvider", OpenIDProvider{},
		"is_active", "is_forced", "base_url", "authorization_endpoint", "token_endpoint",
		"introspection_token_endpoint", "userinfo_endpoint", "endsession_endpoint",
		"connection_scopes", "login_claim", "client_id", "client_secret",
		"authentication_type", "verify_peer", "auto_import", "contact_template",
		"email_bind_attribute", "fullname_bind_attribute", "roles_mapping",
		"authentication_conditions", "groups_mapping", "redirect_url")
	wantJSONKeys(t, "OpenIDEndpoint", OpenIDEndpoint{}, "type", "custom_endpoint")
	wantJSONKeys(t, "OpenIDRolesMapping", OpenIDRolesMapping{},
		"is_enabled", "apply_only_first_role", "attribute_path", "endpoint", "relations")
	wantJSONKeys(t, "OpenIDAuthenticationConditions", OpenIDAuthenticationConditions{},
		"is_enabled", "attribute_path", "endpoint", "authorized_values",
		"trusted_client_addresses", "blacklist_client_addresses")
	wantJSONKeys(t, "OpenIDGroupsMapping", OpenIDGroupsMapping{},
		"is_enabled", "attribute_path", "endpoint", "relations")

	wantJSONKeys(t, "SAMLProvider", SAMLProvider{},
		"is_active", "is_forced", "entity_id_url", "remote_login_url", "certificate",
		"user_id_attribute", "requested_authn_context", "requested_authn_context_comparison",
		"logout_from", "logout_from_url", "auto_import", "contact_template",
		"email_bind_attribute", "fullname_bind_attribute", "roles_mapping",
		"authentication_conditions", "groups_mapping")
	wantJSONKeys(t, "SAMLRolesMapping", SAMLRolesMapping{},
		"is_enabled", "apply_only_first_role", "attribute_path", "relations")
	wantJSONKeys(t, "SAMLAuthenticationConditions", SAMLAuthenticationConditions{},
		"is_enabled", "attribute_path", "authorized_values")
	wantJSONKeys(t, "SAMLGroupsMapping", SAMLGroupsMapping{},
		"is_enabled", "attribute_path", "relations")

	wantJSONKeys(t, "WebSSOProvider", WebSSOProvider{},
		"is_active", "is_forced", "trusted_client_addresses", "blacklist_client_addresses",
		"login_header_attribute", "pattern_matching_login", "pattern_replace_login")
}

// openIDFreshBox returns the live 25.10.16 openid JSON for an unconfigured box,
// including the "" vs null custom_endpoint asymmetry across the three mappings.
func openIDFreshBox() map[string]any {
	return map[string]any{
		"is_active": false, "is_forced": false, "base_url": nil,
		"authorization_endpoint": nil, "token_endpoint": nil, "introspection_token_endpoint": nil,
		"userinfo_endpoint": nil, "endsession_endpoint": nil, "connection_scopes": []string{},
		"login_claim": nil, "client_id": nil, "client_secret": nil,
		"authentication_type": "client_secret_post", "verify_peer": true, "auto_import": false,
		"contact_template": nil, "email_bind_attribute": nil, "fullname_bind_attribute": nil,
		"roles_mapping": map[string]any{
			"is_enabled": false, "apply_only_first_role": false, "attribute_path": "",
			"endpoint":  map[string]any{"type": "introspection_endpoint", "custom_endpoint": ""},
			"relations": []any{},
		},
		"authentication_conditions": map[string]any{
			"is_enabled": false, "attribute_path": "",
			"endpoint":                   map[string]any{"type": "introspection_endpoint", "custom_endpoint": nil},
			"authorized_values":          []any{},
			"trusted_client_addresses":   []any{},
			"blacklist_client_addresses": []any{},
		},
		"groups_mapping": map[string]any{
			"is_enabled": false, "attribute_path": "",
			"endpoint":  map[string]any{"type": "introspection_endpoint", "custom_endpoint": nil},
			"relations": []any{},
		},
		"redirect_url": nil,
	}
}

func samlFreshBox(cert string) map[string]any {
	return map[string]any{
		"is_active": false, "is_forced": false, "entity_id_url": "", "remote_login_url": "",
		"certificate": cert, "user_id_attribute": "", "requested_authn_context": false,
		"requested_authn_context_comparison": "exact", "logout_from": false, "logout_from_url": nil,
		"auto_import": false, "contact_template": nil, "email_bind_attribute": nil, "fullname_bind_attribute": nil,
		"roles_mapping": map[string]any{
			"is_enabled": false, "apply_only_first_role": false, "attribute_path": "", "relations": []any{},
		},
		"authentication_conditions": map[string]any{
			"is_enabled": false, "attribute_path": "", "authorized_values": []any{},
		},
		"groups_mapping": map[string]any{
			"is_enabled": false, "attribute_path": "", "relations": []any{},
		},
	}
}
