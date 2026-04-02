package centreon

import (
	"context"
	"fmt"
	"iter"
)

// UserFilter represents a Centreon user filter.
type UserFilter struct {
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	Criteria []FilterCriteria `json:"criteria,omitzero"`
}

// FilterCriteria represents a single criterion in a user filter.
type FilterCriteria struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Value      any    `json:"value"`
	ObjectType string `json:"object_type,omitzero"`
}

// CreateUserFilterRequest is the request body for creating a user filter.
type CreateUserFilterRequest struct {
	Name     string           `json:"name"`
	Criteria []FilterCriteria `json:"criteria,omitzero"`
}

// UpdateUserFilterRequest is the request body for replacing a user filter (PUT).
type UpdateUserFilterRequest struct {
	Name     string           `json:"name"`
	Criteria []FilterCriteria `json:"criteria,omitzero"`
}

// PatchUserFilterRequest is the request body for partially updating a user filter (PATCH).
type PatchUserFilterRequest struct {
	Name *string `json:"name,omitempty"`
}

// UserFilterService provides user filter operations.
type UserFilterService struct {
	client *Client
}

// List returns a paginated list of user filters.
func (s *UserFilterService) List(ctx context.Context, opts ...ListOption) (*ListResponse[UserFilter], error) {
	var resp ListResponse[UserFilter]
	err := s.client.list(ctx, "/users/filters/events-view", opts, &resp)
	return &resp, err
}

// All returns an iterator over all user filters.
func (s *UserFilterService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*UserFilter, error] {
	return all(ctx, s.List, opts)
}

// Get returns the user filter with the given ID.
func (s *UserFilterService) Get(ctx context.Context, id int) (*UserFilter, error) {
	var uf UserFilter
	if err := s.client.get(ctx, fmt.Sprintf("/users/filters/events-view/%d", id), &uf); err != nil {
		return nil, err
	}
	return &uf, nil
}

// Create creates a new user filter and returns its ID.
func (s *UserFilterService) Create(ctx context.Context, req CreateUserFilterRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/users/filters/events-view", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing user filter using PUT.
func (s *UserFilterService) Update(ctx context.Context, id int, req UpdateUserFilterRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/users/filters/events-view/%d", id), req)
}

// Patch partially updates an existing user filter using PATCH.
func (s *UserFilterService) Patch(ctx context.Context, id int, req PatchUserFilterRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/users/filters/events-view/%d", id), req)
}

// Delete deletes a user filter by ID.
func (s *UserFilterService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/users/filters/events-view/%d", id))
}
