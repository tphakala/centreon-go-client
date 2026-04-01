package centreon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type loginRequest struct {
	Security loginSecurity `json:"security"`
}

type loginSecurity struct {
	Credentials loginCredentials `json:"credentials"`
}

type loginCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Security loginSecurityResponse `json:"security"`
}

type loginSecurityResponse struct {
	Token string `json:"token"`
}

// Login authenticates with the Centreon API using the configured credentials.
// It stores the returned token for subsequent requests.
func (c *Client) Login(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("centreon: credentials not configured; use WithCredentials")
	}
	return c.login(ctx)
}

// login sends a POST /login directly via sendRequest (not do()) to avoid
// an infinite 401 retry loop. It parses the token from the response and
// stores it under the mutex.
func (c *Client) login(ctx context.Context) error {
	reqBody := loginRequest{
		Security: loginSecurity{
			Credentials: loginCredentials{
				Login:    c.username,
				Password: c.password,
			},
		},
	}

	fullURL := c.buildURL("/login")
	resp, err := c.sendRequest(ctx, http.MethodPost, fullURL, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseError(resp)
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("centreon: decode login response: %w", err)
	}

	c.mu.Lock()
	c.token = loginResp.Security.Token
	c.mu.Unlock()

	return nil
}

// Logout sends a logout request and clears the stored token.
func (c *Client) Logout(ctx context.Context) error {
	err := c.get(ctx, "/logout", nil)
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
	return err
}
