package centreon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxErrorBodySize = 1 << 20 // 1 MB

// APIError represents an error response from the Centreon API.
type APIError struct {
	HTTPStatus int    `json:"-"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("centreon API error (HTTP %d): %s", e.HTTPStatus, e.Message)
}

// routeNotFoundPrefix is the fixed leading text of the Symfony "no matching
// route" error that Centreon returns when an endpoint is not available: the path
// is unregistered, or registered but not for the requested method (for example
// the per-id detail GET on a pre-25.10 Centreon, where the path exists only for
// DELETE and PATCH). A resource-not-found 404 uses a different message, such as
// "Host not found".
const routeNotFoundPrefix = "No route found for"

// IsRouteNotFound reports whether this error is a routing 404: the requested
// endpoint is not available on the connected Centreon (the path is unregistered,
// or not registered for this method), as opposed to a resource 404 where the
// endpoint exists but the requested resource does not. Both are HTTP 404; the
// distinguishing signal is the message, which begins "No route found for ..."
// for a routing 404 and is a resource message such as "Host not found"
// otherwise. Callers use this to gate endpoints that exist only on newer
// Centreon releases and to degrade gracefully on older ones.
//
// The body "code" field is deliberately NOT used: it is not stable across
// versions. A routing 404 carries code 0 on 25.10 but code 404 on 24.10 (both
// live-verified), while a resource 404 carries code 404 on both, so code cannot
// discriminate the two on 24.10. The message prefix is the only cross-version
// signal. An empty or non-JSON 404 body lacks the prefix and is correctly not
// treated as a routing error. The message comes from the response body, so this
// check runs inside the client (the layer that owns the response); a gateway
// consumer reads only the returned bool and never has to parse the message.
func (e *APIError) IsRouteNotFound() bool {
	return e != nil &&
		e.HTTPStatus == http.StatusNotFound &&
		strings.HasPrefix(e.Message, routeNotFoundPrefix)
}

// NotFoundError indicates a resource was not found via filtered list lookup.
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("centreon: %s with ID %d not found", e.Resource, e.ID)
}

// parseError reads an HTTP response and returns an *APIError.
func parseError(resp *http.Response) error {
	apiErr := &APIError{HTTPStatus: resp.StatusCode}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
	if err != nil || len(body) == 0 {
		apiErr.Message = http.StatusText(resp.StatusCode)
		return apiErr
	}
	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Message = string(body)
	}
	return apiErr
}

// IsRouteNotFound reports whether err is, or wraps, an *APIError for a route
// that is not registered on the connected Centreon (a routing 404 rather than a
// resource 404). See (*APIError).IsRouteNotFound. It returns false for a nil
// error or any error that is not an *APIError.
func IsRouteNotFound(err error) bool {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.IsRouteNotFound()
	}
	return false
}
