package centreon

import (
	"context"
	"fmt"
	"iter"
)

// User represents a Centreon user (contact).
type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	Email       string `json:"email,omitzero"`
	IsAdmin     bool   `json:"is_admin"`
	IsActivated bool   `json:"is_activated,omitzero"`
}

// UpdateUserRequest is the request body for updating a user (PATCH).
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Alias *string `json:"alias,omitempty"`
	Email *string `json:"email,omitempty"`
}

// UserService provides user/contact operations.
type UserService struct {
	client *Client
}

// List returns a paginated list of users.
func (s *UserService) List(ctx context.Context, opts ...ListOption) (*ListResponse[User], error) {
	var resp ListResponse[User]
	err := s.client.list(ctx, "/configuration/users", opts, &resp)
	return &resp, err
}

// All returns an iterator over all users.
func (s *UserService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*User, error] {
	return all(ctx, s.List, opts)
}

// Update updates an existing user using PATCH.
//
// On Centreon Web 25.10.x this returns an *APIError with HTTP 404
// ("No route found"): the PATCH /configuration/users/{id} route is not
// registered, and PUT, POST, and DELETE for users and contacts are likewise
// unregistered, so user and contact configuration is effectively read-only
// through the v2 REST API on that version. The method is retained for older
// and future versions that register the route.
func (s *UserService) Update(ctx context.Context, id int, req UpdateUserRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/users/%d", id), req)
}
