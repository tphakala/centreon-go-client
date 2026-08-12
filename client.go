package centreon

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout  = 30 * time.Second
	httpStatusError = 400
)

// Client is a Centreon Web REST API client.
type Client struct {
	baseURL    string
	apiVersion string
	httpClient *http.Client

	mu       sync.Mutex
	token    string
	username string
	password string
	logger   *slog.Logger

	// loginSem is a capacity-1 semaphore serializing concurrent
	// re-authentication so only one goroutine calls login() at a time
	// (prevents a thundering herd on 401). Unlike a sync.Mutex it is acquired
	// via a select against ctx.Done(), so a waiter honors its context's
	// cancellation or deadline instead of blocking until login() returns.
	// Initialized in NewClient; a zero-value Client cannot be used directly
	// because its other unexported fields are set only by NewClient.
	loginSem chan struct{}

	MonitoringServers    *MonitoringServerService
	Commands             *CommandService
	Hosts                *HostService
	HostGroups           *HostGroupService
	HostCategories       *HostCategoryService
	HostSeverities       *HostSeverityService
	HostTemplates        *HostTemplateService
	Services             *ServiceService
	ServiceGroups        *ServiceGroupService
	ServiceCategories    *ServiceCategoryService
	ServiceSeverities    *ServiceSeverityService
	ServiceTemplates     *ServiceTemplateService
	Monitoring           *MonitoringResourceService
	MonitoringHosts      *MonitoringHostService
	MonitoringServices   *MonitoringServiceService
	Operations           *OperationsService
	Users                *UserService
	ContactGroups        *ContactGroupService
	ContactTemplates     *ContactTemplateService
	UserFilters          *UserFilterService
	TimePeriods          *TimePeriodService
	NotificationPolicies *NotificationPolicyService
	Downtimes            *DowntimeService
	Acknowledgements     *AcknowledgementService
}

// Option configures a Client.
type Option func(*Client)

// NewClient creates a new Centreon API client.
// It defaults to API version "latest" and a 30-second HTTP timeout.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		// url.Parse returns a *url.Error whose Error() echoes the raw URL,
		// which may carry userinfo credentials (CWE-532). Redact the URL while
		// keeping the *url.Error type so callers can still inspect it.
		return nil, fmt.Errorf("centreon: invalid base URL: %w", redactURLError(err))
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("centreon: invalid base URL %q: missing scheme or host", redactURL(baseURL))
	}

	c := &Client{
		baseURL:    strings.TrimRight(u.String(), "/"),
		apiVersion: "latest",
		httpClient: &http.Client{Timeout: defaultTimeout},
		loginSem:   make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.MonitoringServers = &MonitoringServerService{client: c}
	c.Commands = &CommandService{client: c}
	c.Hosts = &HostService{client: c}
	c.HostGroups = &HostGroupService{client: c}
	c.HostCategories = &HostCategoryService{client: c}
	c.HostSeverities = &HostSeverityService{client: c}
	c.HostTemplates = &HostTemplateService{client: c}
	c.Services = &ServiceService{client: c}
	c.ServiceGroups = &ServiceGroupService{client: c}
	c.ServiceCategories = &ServiceCategoryService{client: c}
	c.ServiceSeverities = &ServiceSeverityService{client: c}
	c.ServiceTemplates = &ServiceTemplateService{client: c}
	c.Monitoring = &MonitoringResourceService{client: c}
	c.MonitoringHosts = &MonitoringHostService{client: c}
	c.MonitoringServices = &MonitoringServiceService{client: c}
	c.Operations = &OperationsService{client: c}
	c.Users = &UserService{client: c}
	c.ContactGroups = &ContactGroupService{client: c}
	c.ContactTemplates = &ContactTemplateService{client: c}
	c.UserFilters = &UserFilterService{client: c}
	c.TimePeriods = &TimePeriodService{client: c}
	c.NotificationPolicies = &NotificationPolicyService{client: c}
	c.Downtimes = &DowntimeService{client: c}
	c.Acknowledgements = &AcknowledgementService{client: c}
	return c, nil
}

// WithVersion sets the API version (e.g. "v2", "latest").
func WithVersion(v string) Option {
	return func(c *Client) { c.apiVersion = v }
}

// WithCredentials sets the username and password for authentication.
func WithCredentials(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithAPIToken sets a pre-existing API token.
func WithAPIToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP client timeout. Defaults to 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithInsecureTLS disables TLS certificate verification.
// Use only for testing against instances with self-signed certificates.
func WithInsecureTLS() Option {
	return func(c *Client) {
		transport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return
		}
		transport = transport.Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // intentional for test instances
		c.httpClient.Transport = transport
	}
}

// WithLogger enables structured logging for API requests and errors.
// If nil, logging is disabled (the default).
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) { c.logger = l }
}

// Token returns the current authentication token.
// This can be used to cache the token for external use.
func (c *Client) Token() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.token
}

type contextKey string

const toolNameKey contextKey = "centreon.tool"

// WithToolName returns a context annotated with a tool/caller name
// that will be included in log output. This helps correlate API calls
// to the higher-level operation that triggered them.
func WithToolName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, toolNameKey, name)
}

// toolName extracts the tool name from the context, or returns "" if not set.
func toolName(ctx context.Context) string {
	if v, ok := ctx.Value(toolNameKey).(string); ok {
		return v
	}
	return ""
}

// buildURL constructs the full API URL for the given path.
func (c *Client) buildURL(path string) string {
	return fmt.Sprintf("%s/centreon/api/%s%s", c.baseURL, c.apiVersion, path)
}

// sendRequest builds and executes an HTTP request. It marshals body to JSON
// if non-nil and sets the appropriate headers.
func (c *Client) sendRequest(ctx context.Context, method, reqURL string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("centreon: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		// A malformed reqURL yields a *url.Error that echoes the URL, which
		// carries the userinfo credential. Redact before returning (CWE-532).
		return nil, fmt.Errorf("centreon: create request: %w", redactURLError(err))
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "centreon-go-client")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.mu.Lock()
	token := c.token
	c.mu.Unlock()
	if token != "" {
		req.Header.Set("X-AUTH-TOKEN", token)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		// A sibling request cancelled via a shared context (e.g. errgroup) is
		// expected noise, not a failure, so log it at Debug. DeadlineExceeded
		// and every other transport error stay at Error. The error is returned
		// unchanged either way, so callers that care still see the cancellation.
		if errors.Is(err, context.Canceled) {
			c.logDebug(ctx, "request cancelled", method, reqURL, 0, duration)
		} else {
			c.logError(ctx, "request failed", method, reqURL, err, duration)
		}
		return nil, err
	}
	c.logDebug(ctx, "request completed", method, reqURL, resp.StatusCode, duration)
	return resp, nil
}

// acquireLogin takes the re-auth semaphore, honoring ctx. It returns
// ctx.Err() if the context is cancelled or times out while waiting.
func (c *Client) acquireLogin(ctx context.Context) error {
	select {
	case c.loginSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releaseLogin releases the re-auth semaphore taken by acquireLogin.
func (c *Client) releaseLogin() { <-c.loginSem }

// reauthenticate serializes a 401-triggered re-login through the loginSem
// semaphore. It first acquires the semaphore honoring ctx, returning ctx.Err()
// if the context is cancelled or times out while waiting. It then compares the
// current token to tokenBefore: if another goroutine already refreshed it, it
// returns nil so the caller retries with the new token. Otherwise it calls
// login() once, and on failure logs "re-authentication failed" and returns the
// error so the caller aborts instead of re-sending with a stale token. The
// semaphore is released via defer, so it is freed even if login() panics.
func (c *Client) reauthenticate(ctx context.Context, method, fullURL, tokenBefore string) error {
	if err := c.acquireLogin(ctx); err != nil {
		return err
	}
	defer c.releaseLogin()

	c.mu.Lock()
	tokenNow := c.token
	c.mu.Unlock()

	// Another goroutine already refreshed the token; retry with the new one.
	if tokenNow != tokenBefore {
		return nil
	}

	// Token has not been refreshed by another goroutine; do it now.
	c.logInfo(ctx, "token expired, re-authenticating")
	if loginErr := c.login(ctx); loginErr != nil {
		c.logError(ctx, "re-authentication failed", method, fullURL, loginErr, 0)
		return loginErr
	}
	return nil
}

// do is the core request method. It sends a request and decodes the response.
// On 401, it re-authenticates via reauthenticate() and retries once. Concurrent
// 401s are serialized through a context-aware semaphore: only one goroutine
// calls login() while the others wait, honoring their context's cancellation or
// deadline, and then retry with the refreshed token. To avoid a thundering herd,
// reauthenticate compares the token captured before the request to the current
// token and skips login() if another goroutine already refreshed it.
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	fullURL := c.buildURL(path)

	// Capture token before sending so we can detect concurrent refreshes.
	c.mu.Lock()
	tokenBefore := c.token
	c.mu.Unlock()

	resp, err := c.sendRequest(ctx, method, fullURL, body)
	if err != nil {
		return err
	}

	// Auto-renew on 401 if credentials are available
	if resp.StatusCode == http.StatusUnauthorized && c.username != "" {
		resp.Body.Close() //nolint:errcheck // best-effort cleanup before retry

		if err := c.reauthenticate(ctx, method, fullURL, tokenBefore); err != nil {
			return err
		}

		resp, err = c.sendRequest(ctx, method, fullURL, body)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort cleanup

	if resp.StatusCode >= httpStatusError {
		apiErr := parseError(resp)
		c.logError(ctx, "API error", method, fullURL, apiErr, 0)
		return apiErr
	}

	// 204 No Content: nothing to decode
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("centreon: decode response: %w", err)
		}
	}
	return nil
}

// redactURL returns rawURL in a form safe for logs and error messages, with the
// userinfo password masked. When the URL parses with a host the username is kept
// (via url.URL.Redacted); for opaque or malformed input the whole userinfo, and
// any scheme prefix, are masked. It is applied only at log/error-formatting time;
// the stored baseURL and the URL sent on the wire are never modified, so real
// requests still carry the credential.
func redactURL(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		return u.Redacted()
	}
	// Fall back for opaque or malformed input that url.Parse does not expose as
	// userinfo (e.g. "admin:secret@host/path" parses with an empty Host, so
	// (*url.URL).Redacted would return the secret verbatim).
	return redactUserinfoString(rawURL)
}

// redactUserinfoString masks the userinfo of a string that url.Parse could not
// expose as userinfo (opaque or malformed input). Everything up to the LAST '@'
// is replaced with "xxxxx": the userinfo terminates at a '@' at or before the
// last one, so no credential can survive whatever else the string contains. It
// deliberately does not preserve a leading scheme, because a "://" inside the
// userinfo cannot be told apart from a real scheme prefix; masking the scheme too
// is safe in this error-message-only fallback. Input with no '@' is returned
// unchanged.
func redactUserinfoString(raw string) string {
	at := strings.LastIndexByte(raw, '@')
	if at < 0 {
		return raw
	}
	return "xxxxx@" + raw[at+1:]
}

// redactURLError redacts userinfo credentials from a *url.Error's URL field,
// which its Error() method echoes verbatim (CWE-532). It redacts in place and
// returns the original error, preserving the errors.Is/As chain. Callers must
// redact BEFORE wrapping with fmt.Errorf: a wrapper caches its message at
// construction, so redacting a *url.Error already inside a wrapper would not
// update the frozen message. An error with no *url.Error is unchanged.
func redactURLError(err error) error {
	if uerr, ok := errors.AsType[*url.Error](err); ok {
		uerr.URL = redactURL(uerr.URL)
	}
	return err
}

// Logging helpers: no-ops when logger is nil.
// Include tool name from context and request duration when available.

func (c *Client) logDebug(ctx context.Context, msg, method, reqURL string, status int, duration time.Duration) {
	if c.logger != nil {
		attrs := []any{"method", method, "url", redactURL(reqURL), "status", status, "duration", duration}
		if tool := toolName(ctx); tool != "" {
			attrs = append(attrs, "tool", tool)
		}
		c.logger.Debug(msg, attrs...)
	}
}

func (c *Client) logInfo(ctx context.Context, msg string) {
	if c.logger != nil {
		attrs := []any{}
		if tool := toolName(ctx); tool != "" {
			attrs = append(attrs, "tool", tool)
		}
		c.logger.Info(msg, attrs...)
	}
}

func (c *Client) logError(ctx context.Context, msg, method, reqURL string, err error, duration time.Duration) {
	if c.logger != nil {
		attrs := []any{"method", method, "url", redactURL(reqURL), "error", err, "duration", duration}
		if tool := toolName(ctx); tool != "" {
			attrs = append(attrs, "tool", tool)
		}
		c.logger.Error(msg, attrs...)
	}
}

// Convenience methods.

func (c *Client) get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) post(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) put(ctx context.Context, path string, body any) error {
	return c.do(ctx, http.MethodPut, path, body, nil)
}

func (c *Client) patch(ctx context.Context, path string, body any) error {
	return c.do(ctx, http.MethodPatch, path, body, nil)
}

func (c *Client) delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
