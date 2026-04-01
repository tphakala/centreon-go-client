package centreon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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

	MonitoringServers *MonitoringServerService
	Commands          *CommandService
	Hosts             *HostService
	HostGroups        *HostGroupService
	HostCategories    *HostCategoryService
	HostSeverities    *HostSeverityService
	HostTemplates     *HostTemplateService
	Services          *ServiceService
	ServiceGroups      *ServiceGroupService
	ServiceCategories  *ServiceCategoryService
	ServiceSeverities  *ServiceSeverityService
	ServiceTemplates   *ServiceTemplateService
	Monitoring         *MonitoringResourceService
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
		httpClient: &http.Client{Timeout: 30 * time.Second},
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

// buildURL constructs the full API URL for the given path.
func (c *Client) buildURL(path string) string {
	return fmt.Sprintf("%s/centreon/api/%s%s", c.baseURL, c.apiVersion, path)
}

// sendRequest builds and executes an HTTP request. It marshals body to JSON
// if non-nil and sets the appropriate headers.
func (c *Client) sendRequest(ctx context.Context, method, url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("centreon: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("centreon: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.mu.Lock()
	token := c.token
	c.mu.Unlock()
	if token != "" {
		req.Header.Set("X-AUTH-TOKEN", token)
	}

	return c.httpClient.Do(req)
}

// do is the core request method. It sends a request and decodes the response.
// On 401, it attempts to re-authenticate via login() and retries once.
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	fullURL := c.buildURL(path)

	resp, err := c.sendRequest(ctx, method, fullURL, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Auto-renew on 401
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		if loginErr := c.login(ctx); loginErr != nil {
			return loginErr
		}
		resp, err = c.sendRequest(ctx, method, fullURL, body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode >= 400 {
		return parseError(resp)
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

// Convenience methods.

func (c *Client) get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) post(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) put(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

func (c *Client) patch(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPatch, path, body, result)
}

func (c *Client) delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
