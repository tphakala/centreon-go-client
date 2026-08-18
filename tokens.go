package centreon

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"time"
)

// Token represents an authentication token from the Centreon administration API
// (GET/POST /administration/tokens on Centreon Web 25.10.x). Tokens have no
// numeric id; the token name is the identity used for deletion.
type Token struct {
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	IsRevoked      bool       `json:"is_revoked"`
	CreationDate   time.Time  `json:"creation_date"`
	ExpirationDate *time.Time `json:"expiration_date"`
	User           NamedRef   `json:"user"`
	Creator        NamedRef   `json:"creator"`

	// Value is the token secret. The API returns it exactly once, in the Create
	// response; List and All always leave it empty because the list endpoint
	// omits the "token" key. Treat it as a credential: never log it and never
	// store it in cleartext. The client itself never logs request or response
	// bodies (see client.do and logDebug/logError, which log only method,
	// redacted URL, status, and duration), so the secret cannot leak through
	// client logging.
	//
	// Token redacts Value from its String and slog output (see String and
	// LogValue), so "%v"/"%s" formatting and structured logging are safe. Note
	// that json.Marshal of a Token still writes Value (tagged "token") so the
	// caller can persist the created secret; do not marshal a Token into logs.
	Value string `json:"token,omitzero"`
}

// String implements fmt.Stringer and deliberately redacts the secret Value so
// that formatting a Token with "%v" or "%s" cannot leak it. Access the Value
// field directly when the secret is genuinely needed.
func (t Token) String() string {
	return fmt.Sprintf("centreon.Token{Name:%q, Type:%q, IsRevoked:%t, Value:%s}",
		t.Name, t.Type, t.IsRevoked, redactedSecret(t.Value))
}

// GoString implements fmt.GoStringer so the Go-syntax verb "%#v" redacts the
// secret Value as well. Without it, "%#v" reflects the struct directly and
// prints the cleartext secret, bypassing String.
//
//nolint:gocritic // hugeParam: value receiver mirrors String so "%#v" redacts Token values, not only *Token.
func (t Token) GoString() string {
	return t.String()
}

// LogValue implements slog.LogValuer so structured logging of a Token redacts
// the secret Value rather than emitting it. The receiver is a value, not a
// pointer, so that logging a Token value (not only a *Token) still redacts;
// a pointer receiver would leave value tokens unredacted.
//
//nolint:gocritic // hugeParam: the value receiver is required for slog.LogValuer to redact Token values.
func (t Token) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("name", t.Name),
		slog.String("type", t.Type),
		slog.Bool("is_revoked", t.IsRevoked),
		slog.String("value", redactedSecret(t.Value)),
	)
}

// redactedSecret returns a placeholder for a token secret: "<redacted>" when a
// secret is present and "<empty>" when it is not, never the secret itself.
func redactedSecret(v string) string {
	if v == "" {
		return "<empty>"
	}
	return "<redacted>"
}

// CreateTokenRequest is the request body for POST /administration/tokens on
// Centreon Web 25.10.x. Name, Type, and UserID are required (verified against a
// live 25.10.16 instance). Type is the token kind ("api" or "cma"). UserID is
// the id of the user the token authenticates as. ExpirationDate is optional:
// when nil it is omitted from the body and the server creates a token with no
// expiry.
type CreateTokenRequest struct {
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	UserID         int        `json:"user_id"`
	ExpirationDate *time.Time `json:"expiration_date,omitzero"`
}

// TokenService provides API token operations (issue #65). It wraps
// /administration/tokens on Centreon Web 25.10+.
type TokenService struct {
	client *Client
}

// List returns a paginated list of tokens. The returned Token values never
// carry the secret in Value; only Create returns it.
func (s *TokenService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Token], error) {
	var resp ListResponse[Token]
	err := s.client.list(ctx, "/administration/tokens", opts, &resp)
	return &resp, err
}

// All returns an iterator over all tokens.
func (s *TokenService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Token, error] {
	return all(ctx, s.List, opts)
}

// Create creates a new token and returns the full created Token, including the
// secret in Value.
//
// Unlike the other Create methods in this package, which return only the new
// integer id, this returns the whole *Token. A token has no numeric id, and the
// secret in Value is returned exactly once, by this call; returning it here is
// the only way the caller can ever obtain it, so an id-only return would make
// the endpoint useless. The caller owns the secret from here: store it securely
// and never log it.
func (s *TokenService) Create(ctx context.Context, req *CreateTokenRequest) (*Token, error) {
	var result Token
	if err := s.client.post(ctx, "/administration/tokens", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes the token with the given name. The token name is the delete
// key because tokens have no numeric id; it is URL-escaped into the path.
// Returns an *APIError with HTTP status 404 when no token has that name.
func (s *TokenService) Delete(ctx context.Context, name string) error {
	return s.client.delete(ctx, "/administration/tokens/"+url.PathEscape(name))
}
