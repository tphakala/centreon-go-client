package centreon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error response from the Centreon API.
type APIError struct {
	HTTPStatus int    `json:"-"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("centreon API error (HTTP %d): %s", e.HTTPStatus, e.Message)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		apiErr.Message = http.StatusText(resp.StatusCode)
		return apiErr
	}
	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Message = string(body)
	}
	return apiErr
}
