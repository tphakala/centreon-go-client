package centreon

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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

	// loginMu serializes concurrent re-authentication attempts so that only
	// one goroutine calls login() at a time (prevents thundering herd on 401).
	loginMu sync.Mutex

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
		return nil, fmt.Errorf("centreon: invalid base URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("centreon: invalid base URL %q: missing scheme or host", baseURL)
	}

	c := &Client{
		baseURL:    strings.TrimRight(u.String(), "/"),
		apiVersion: "latest",
		httpClient: &http.Client{Timeout: defaultTimeout},
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
		return nil, fmt.Errorf("centreon: create request: %w", err)
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
		c.logError(ctx, "request failed", method, reqURL, err, duration)
		return nil, err
	}
	c.logDebug(ctx, "request completed", method, reqURL, resp.StatusCode, duration)
	return resp, nil
}

// do is the core request method. It sends a request and decodes the response.
// On 401, it attempts to re-authenticate via login() and retries once.
// To avoid a thundering herd when many goroutines get 401 simultaneously,
// it compares the token before the request to the current token: if another
// goroutine already refreshed the token, it skips login and just retries.
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

		// loginMu serializes concurrent re-authentication so only one goroutine
		// calls login(); others wait and then retry with the refreshed token.
		c.loginMu.Lock()
		c.mu.Lock()
		tokenNow := c.token
		c.mu.Unlock()

		if tokenNow == tokenBefore {
			// Token has not been refreshed by another goroutine; do it now.
			c.logInfo(ctx, "token expired, re-authenticating")
			loginErr := c.login(ctx)
			c.loginMu.Unlock()
			if loginErr != nil {
				return loginErr
			}
		} else {
			// Another goroutine already refreshed; just retry with the new token.
			c.loginMu.Unlock()
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

	// 204 No Content — nothing to decode
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

// Logging helpers — no-ops when logger is nil.
// Include tool name from context and request duration when available.

func (c *Client) logDebug(ctx context.Context, msg, method, reqURL string, status int, duration time.Duration) {
	if c.logger != nil {
		attrs := []any{"method", method, "url", reqURL, "status", status, "duration", duration}
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
		attrs := []any{"method", method, "url", reqURL, "error", err, "duration", duration}
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
