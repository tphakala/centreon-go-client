package centreon

import (
	"context"
	"fmt"
	"log/slog"
)

// ProxyConfiguration is the central proxy configuration returned by
// GET /configuration/proxy (used for Centreon's outbound connections, for
// example to centreon.com for plugin packs). URL, User, Password, and Port are
// nullable (JSON null when unset), so they are pointers; Protocol defaults to
// "http://".
//
// Password is a credential and the GET endpoint returns it in cleartext
// (live-verified against 25.10.16), so ProxyConfiguration redacts it from its
// String, GoString, and slog output (see String and LogValue). json.Marshal of
// a ProxyConfiguration still writes the password so a value can be persisted; do
// not marshal one into logs.
type ProxyConfiguration struct {
	URL      *string `json:"url"`
	Port     *int    `json:"port"`
	User     *string `json:"user"`
	Password *string `json:"password"`
	Protocol string  `json:"protocol"`
}

// String implements fmt.Stringer and redacts the Password so that formatting a
// ProxyConfiguration with "%v" or "%s" cannot leak the proxy credential. Access
// the Password field directly when the secret is genuinely needed.
func (p ProxyConfiguration) String() string {
	return fmt.Sprintf("centreon.ProxyConfiguration{URL:%s, Port:%s, User:%s, Password:%s, Protocol:%q}",
		quoteStrPtr(p.URL), intPtrString(p.Port), quoteStrPtr(p.User), redactedSecret(derefStr(p.Password)), p.Protocol)
}

// GoString implements fmt.GoStringer so the Go-syntax verb "%#v" redacts the
// Password as well. Without it, "%#v" reflects the struct directly and prints the
// cleartext credential, bypassing String.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts ProxyConfiguration values, not only pointers.
func (p ProxyConfiguration) GoString() string {
	return p.String()
}

// LogValue implements slog.LogValuer so structured logging of a
// ProxyConfiguration redacts the Password rather than emitting it. The receiver
// is a value, not a pointer, so logging a ProxyConfiguration value (not only a
// pointer) still redacts.
//
//nolint:gocritic // hugeParam: the value receiver is required for slog.LogValuer to redact ProxyConfiguration values.
func (p ProxyConfiguration) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("url", derefStr(p.URL)),
		slog.String("user", derefStr(p.User)),
		slog.String("password", redactedSecret(derefStr(p.Password))),
		slog.String("protocol", p.Protocol),
	)
}

// derefStr returns the pointed-to string, or "" when the pointer is nil. It keeps
// the redaction and formatting paths nil-safe: Password (and the other fields) are
// null on the wire when unset.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// quoteStrPtr renders a *string for String(): a quoted value, or the literal
// <nil> for a nil pointer (distinct from an empty string "").
func quoteStrPtr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q", *s)
}

// intPtrString renders a *int for String(): the number, or <nil> for a nil
// pointer.
func intPtrString(i *int) string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *i)
}

// ProxyService provides access to the central proxy configuration
// (/configuration/proxy).
//
// Only Get is implemented. The v2 REST update path (PUT) is unusable on Centreon
// Web 25.10.16: any PUT body carrying the "protocol" key returns HTTP 500
// ("You must define a type for ...Proxy::$protocol") for every value type, and
// the GET response always includes protocol, so a Get-modify-Update round-trip
// cannot succeed. Update is therefore deferred until the server defect is fixed;
// see issue #98.
type ProxyService struct {
	client *Client
}

// Get returns the central proxy configuration. The returned Password is in
// cleartext; ProxyConfiguration redacts it from String/slog output, but treat the
// field itself as a secret.
func (s *ProxyService) Get(ctx context.Context) (*ProxyConfiguration, error) {
	var result ProxyConfiguration
	if err := s.client.get(ctx, "/configuration/proxy", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
