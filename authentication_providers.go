package centreon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// AuthenticationProviderService provides read and full-replace update access to
// the four authentication providers exposed as singletons under
// /administration/authentication/providers: the built-in local password policy,
// OpenID Connect, SAML, and Web-SSO.
//
// Each provider has a GET (read) and a PUT (full-object replace). An empty-body
// PUT is rejected with the provider's required-field set, so the write body is
// the whole object: a caller reads the provider, mutates it, and writes it back.
// POST and DELETE return HTTP 405; the /ldap provider and the bare providers
// root are absent (404). OpenID additionally accepts PATCH for a partial update,
// but the writable partial subset is not modeled here, so UpdateOpenID uses the
// full-replace PUT. All routes and verbs verified live against Centreon Web
// 25.10.16.
type AuthenticationProviderService struct {
	client *Client
}

const authProvidersBase = "/administration/authentication/providers"

// ---- local ----

// LocalProvider is the built-in local authentication provider
// (GET/PUT /administration/authentication/providers/local). Unlike the other
// three providers it has no is_active/is_forced flags: local is always the
// fallback. The fields below are modeled from the live 25.10.16 shape.
type LocalProvider struct {
	PasswordSecurityPolicy LocalPasswordSecurityPolicy `json:"password_security_policy"`
}

// LocalPasswordSecurityPolicy is the local provider's password policy.
type LocalPasswordSecurityPolicy struct {
	PasswordMinLength      int                     `json:"password_min_length"`
	HasUppercase           bool                    `json:"has_uppercase"`
	HasLowercase           bool                    `json:"has_lowercase"`
	HasNumber              bool                    `json:"has_number"`
	HasSpecialCharacter    bool                    `json:"has_special_character"`
	Attempts               int                     `json:"attempts"`
	BlockingDuration       int                     `json:"blocking_duration"`
	PasswordExpiration     LocalPasswordExpiration `json:"password_expiration"`
	CanReusePasswords      bool                    `json:"can_reuse_passwords"`
	DelayBeforeNewPassword int                     `json:"delay_before_new_password"`
}

// LocalPasswordExpiration is the password-expiration sub-policy of the local
// provider.
type LocalPasswordExpiration struct {
	ExpirationDelay int      `json:"expiration_delay"`
	ExcludedUsers   []string `json:"excluded_users"`
}

// GetLocal returns the local password-policy provider.
func (s *AuthenticationProviderService) GetLocal(ctx context.Context) (*LocalProvider, error) {
	var r LocalProvider
	if err := s.client.get(ctx, authProvidersBase+"/local", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateLocal replaces the local password-policy provider (PUT, full replace).
func (s *AuthenticationProviderService) UpdateLocal(ctx context.Context, req *LocalProvider) error {
	return s.client.put(ctx, authProvidersBase+"/local", req)
}

// ---- openid ----

// OpenIDProvider is the OpenID Connect provider
// (GET/PUT/PATCH /administration/authentication/providers/openid).
//
// ClientSecret is a credential and the GET endpoint returns it in cleartext, so
// OpenIDProvider redacts it from its String, GoString, and slog output (see
// String and LogValue). json.Marshal of an OpenIDProvider still writes the
// secret, which the full-replace PUT requires; do not marshal one into logs.
//
// The nullable scalars (base_url, the endpoint URLs, login_claim, client_id,
// client_secret, the bind attributes, redirect_url) are pointers so a null on
// the wire is distinct from an empty string.
type OpenIDProvider struct {
	IsActive                   bool                           `json:"is_active"`
	IsForced                   bool                           `json:"is_forced"`
	BaseURL                    *string                        `json:"base_url"`
	AuthorizationEndpoint      *string                        `json:"authorization_endpoint"`
	TokenEndpoint              *string                        `json:"token_endpoint"`
	IntrospectionTokenEndpoint *string                        `json:"introspection_token_endpoint"`
	UserinfoEndpoint           *string                        `json:"userinfo_endpoint"`
	EndsessionEndpoint         *string                        `json:"endsession_endpoint"`
	ConnectionScopes           []string                       `json:"connection_scopes"`
	LoginClaim                 *string                        `json:"login_claim"`
	ClientID                   *string                        `json:"client_id"`
	ClientSecret               *string                        `json:"client_secret"`
	AuthenticationType         string                         `json:"authentication_type"`
	VerifyPeer                 bool                           `json:"verify_peer"`
	AutoImport                 bool                           `json:"auto_import"`
	ContactTemplate            *NamedRef                      `json:"contact_template"`
	EmailBindAttribute         *string                        `json:"email_bind_attribute"`
	FullnameBindAttribute      *string                        `json:"fullname_bind_attribute"`
	RolesMapping               OpenIDRolesMapping             `json:"roles_mapping"`
	AuthenticationConditions   OpenIDAuthenticationConditions `json:"authentication_conditions"`
	GroupsMapping              OpenIDGroupsMapping            `json:"groups_mapping"`
	RedirectURL                *string                        `json:"redirect_url"`
}

// OpenIDEndpoint selects where a mapping resolves its claim. It is shared by the
// three OpenID mappings. CustomEndpoint is a pointer because it is "" inside
// roles_mapping.endpoint but null inside authentication_conditions.endpoint and
// groups_mapping.endpoint; the pointer preserves that "" vs null distinction.
type OpenIDEndpoint struct {
	Type           string  `json:"type"`
	CustomEndpoint *string `json:"custom_endpoint"`
}

// OpenIDRolesMapping maps OpenID claims to Centreon ACL roles. Relations is the
// list of claim-to-role rules; its element shape is unverified (empty on a fresh
// box), so it is kept as raw JSON to round-trip losslessly until a configured
// instance pins the shape.
type OpenIDRolesMapping struct {
	IsEnabled          bool              `json:"is_enabled"`
	ApplyOnlyFirstRole bool              `json:"apply_only_first_role"`
	AttributePath      string            `json:"attribute_path"`
	Endpoint           OpenIDEndpoint    `json:"endpoint"`
	Relations          []json.RawMessage `json:"relations"`
}

// OpenIDAuthenticationConditions gates login on a claim value and client
// address. Unlike the SAML counterpart it carries an endpoint object plus the
// trusted/blacklisted client-address lists.
type OpenIDAuthenticationConditions struct {
	IsEnabled                bool           `json:"is_enabled"`
	AttributePath            string         `json:"attribute_path"`
	Endpoint                 OpenIDEndpoint `json:"endpoint"`
	AuthorizedValues         []string       `json:"authorized_values"`
	TrustedClientAddresses   []string       `json:"trusted_client_addresses"`
	BlacklistClientAddresses []string       `json:"blacklist_client_addresses"`
}

// OpenIDGroupsMapping maps OpenID claims to Centreon contact groups. Relations
// is kept as raw JSON for the same reason as OpenIDRolesMapping.Relations.
type OpenIDGroupsMapping struct {
	IsEnabled     bool              `json:"is_enabled"`
	AttributePath string            `json:"attribute_path"`
	Endpoint      OpenIDEndpoint    `json:"endpoint"`
	Relations     []json.RawMessage `json:"relations"`
}

// String implements fmt.Stringer and redacts ClientSecret so that formatting an
// OpenIDProvider with "%v" or "%s" cannot leak the credential. Access the
// ClientSecret field directly when the secret is genuinely needed.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts OpenIDProvider values, not only pointers.
func (p OpenIDProvider) String() string {
	return fmt.Sprintf("centreon.OpenIDProvider{IsActive:%t, BaseURL:%s, ClientID:%s, ClientSecret:%s, AuthenticationType:%q}",
		p.IsActive, quoteStrPtr(p.BaseURL), quoteStrPtr(p.ClientID),
		redactedSecret(derefStr(p.ClientSecret)), p.AuthenticationType)
}

// GoString implements fmt.GoStringer so the Go-syntax verb "%#v" redacts
// ClientSecret as well. Without it, "%#v" reflects the struct directly and
// prints the cleartext secret, bypassing String.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts OpenIDProvider values, not only pointers.
func (p OpenIDProvider) GoString() string { return p.String() }

// LogValue implements slog.LogValuer so structured logging of an OpenIDProvider
// redacts ClientSecret rather than emitting it. The receiver is a value so
// logging an OpenIDProvider value (not only a pointer) still redacts.
//
//nolint:gocritic // hugeParam: the value receiver is required for slog.LogValuer to redact OpenIDProvider values.
func (p OpenIDProvider) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("is_active", p.IsActive),
		slog.String("base_url", derefStr(p.BaseURL)),
		slog.String("client_id", derefStr(p.ClientID)),
		slog.String("client_secret", redactedSecret(derefStr(p.ClientSecret))),
		slog.String("authentication_type", p.AuthenticationType),
	)
}

// GetOpenID returns the OpenID Connect provider. The returned ClientSecret is in
// cleartext; OpenIDProvider redacts it from String/slog output, but treat the
// field itself as a secret.
func (s *AuthenticationProviderService) GetOpenID(ctx context.Context) (*OpenIDProvider, error) {
	var r OpenIDProvider
	if err := s.client.get(ctx, authProvidersBase+"/openid", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateOpenID replaces the OpenID Connect provider (PUT, full replace). The
// endpoint also accepts PATCH for a partial update, but that subset is not
// modeled here.
func (s *AuthenticationProviderService) UpdateOpenID(ctx context.Context, req *OpenIDProvider) error {
	return s.client.put(ctx, authProvidersBase+"/openid", req)
}

// ---- saml ----

// SAMLProvider is the SAML provider
// (GET/PUT /administration/authentication/providers/saml).
//
// Certificate is a credential; SAMLProvider redacts it from its String,
// GoString, and slog output while json.Marshal still writes it (the full-replace
// PUT needs it). The SAML mappings have NO endpoint object (unlike OpenID) and
// authentication_conditions omits the client-address lists, so the nested types
// are distinct from their OpenID counterparts and must not be shared.
type SAMLProvider struct {
	IsActive                        bool                         `json:"is_active"`
	IsForced                        bool                         `json:"is_forced"`
	EntityIDURL                     string                       `json:"entity_id_url"`
	RemoteLoginURL                  string                       `json:"remote_login_url"`
	Certificate                     string                       `json:"certificate"`
	UserIDAttribute                 string                       `json:"user_id_attribute"`
	RequestedAuthnContext           bool                         `json:"requested_authn_context"`
	RequestedAuthnContextComparison string                       `json:"requested_authn_context_comparison"`
	LogoutFrom                      bool                         `json:"logout_from"`
	LogoutFromURL                   *string                      `json:"logout_from_url"`
	AutoImport                      bool                         `json:"auto_import"`
	ContactTemplate                 *NamedRef                    `json:"contact_template"`
	EmailBindAttribute              *string                      `json:"email_bind_attribute"`
	FullnameBindAttribute           *string                      `json:"fullname_bind_attribute"`
	RolesMapping                    SAMLRolesMapping             `json:"roles_mapping"`
	AuthenticationConditions        SAMLAuthenticationConditions `json:"authentication_conditions"`
	GroupsMapping                   SAMLGroupsMapping            `json:"groups_mapping"`
}

// SAMLRolesMapping maps SAML attributes to Centreon ACL roles. Relations is kept
// as raw JSON (empty on a fresh box, element shape unverified).
type SAMLRolesMapping struct {
	IsEnabled          bool              `json:"is_enabled"`
	ApplyOnlyFirstRole bool              `json:"apply_only_first_role"`
	AttributePath      string            `json:"attribute_path"`
	Relations          []json.RawMessage `json:"relations"`
}

// SAMLAuthenticationConditions gates login on a SAML attribute value. Unlike the
// OpenID counterpart it has no endpoint object and no client-address lists.
type SAMLAuthenticationConditions struct {
	IsEnabled        bool     `json:"is_enabled"`
	AttributePath    string   `json:"attribute_path"`
	AuthorizedValues []string `json:"authorized_values"`
}

// SAMLGroupsMapping maps SAML attributes to Centreon contact groups. Relations
// is kept as raw JSON for the same reason as SAMLRolesMapping.Relations.
type SAMLGroupsMapping struct {
	IsEnabled     bool              `json:"is_enabled"`
	AttributePath string            `json:"attribute_path"`
	Relations     []json.RawMessage `json:"relations"`
}

// String implements fmt.Stringer and redacts Certificate so that formatting a
// SAMLProvider with "%v" or "%s" cannot leak the credential.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts SAMLProvider values, not only pointers.
func (p SAMLProvider) String() string {
	return fmt.Sprintf("centreon.SAMLProvider{IsActive:%t, EntityIDURL:%q, RemoteLoginURL:%q, Certificate:%s}",
		p.IsActive, p.EntityIDURL, p.RemoteLoginURL, redactedSecret(p.Certificate))
}

// GoString implements fmt.GoStringer so "%#v" redacts Certificate as well.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts SAMLProvider values, not only pointers.
func (p SAMLProvider) GoString() string { return p.String() }

// LogValue implements slog.LogValuer so structured logging of a SAMLProvider
// redacts Certificate rather than emitting it.
//
//nolint:gocritic // hugeParam: the value receiver is required for slog.LogValuer to redact SAMLProvider values.
func (p SAMLProvider) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("is_active", p.IsActive),
		slog.String("entity_id_url", p.EntityIDURL),
		slog.String("remote_login_url", p.RemoteLoginURL),
		slog.String("certificate", redactedSecret(p.Certificate)),
	)
}

// GetSAML returns the SAML provider. The returned Certificate is in cleartext;
// SAMLProvider redacts it from String/slog output, but treat the field itself as
// a secret.
func (s *AuthenticationProviderService) GetSAML(ctx context.Context) (*SAMLProvider, error) {
	var r SAMLProvider
	if err := s.client.get(ctx, authProvidersBase+"/saml", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateSAML replaces the SAML provider (PUT, full replace).
func (s *AuthenticationProviderService) UpdateSAML(ctx context.Context, req *SAMLProvider) error {
	return s.client.put(ctx, authProvidersBase+"/saml", req)
}

// ---- web-sso ----

// WebSSOProvider is the Web SSO (reverse-proxy header) provider
// (GET/PUT /administration/authentication/providers/web-sso). It carries no
// credential, so it needs no redaction methods.
type WebSSOProvider struct {
	IsActive                 bool     `json:"is_active"`
	IsForced                 bool     `json:"is_forced"`
	TrustedClientAddresses   []string `json:"trusted_client_addresses"`
	BlacklistClientAddresses []string `json:"blacklist_client_addresses"`
	LoginHeaderAttribute     string   `json:"login_header_attribute"`
	PatternMatchingLogin     *string  `json:"pattern_matching_login"`
	PatternReplaceLogin      *string  `json:"pattern_replace_login"`
}

// GetWebSSO returns the Web SSO provider.
func (s *AuthenticationProviderService) GetWebSSO(ctx context.Context) (*WebSSOProvider, error) {
	var r WebSSOProvider
	if err := s.client.get(ctx, authProvidersBase+"/web-sso", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateWebSSO replaces the Web SSO provider (PUT, full replace).
func (s *AuthenticationProviderService) UpdateWebSSO(ctx context.Context, req *WebSSOProvider) error {
	return s.client.put(ctx, authProvidersBase+"/web-sso", req)
}
