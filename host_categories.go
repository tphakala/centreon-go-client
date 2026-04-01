package centreon

import (
	"context"
	"fmt"
	"iter"
)

// HostCategory represents a Centreon host category configuration resource.
type HostCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// CreateHostCategoryRequest is the request body for creating a host category.
type CreateHostCategoryRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// UpdateHostCategoryRequest is the request body for updating a host category (PUT).
type UpdateHostCategoryRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// HostCategoryService provides host category configuration operations.
type HostCategoryService struct {
	client *Client
}

// List returns a paginated list of host categories.
func (s *HostCategoryService) List(ctx context.Context, opts ...ListOption) (*ListResponse[HostCategory], error) {
	var resp ListResponse[HostCategory]
	err := s.client.list(ctx, "/configuration/hosts/categories", opts, &resp)
	return &resp, err
}

// All returns an iterator over all host categories.
func (s *HostCategoryService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*HostCategory, error] {
	return all(ctx, s.List, opts)
}

// Get returns the host category with the given ID using a direct GET request.
func (s *HostCategoryService) Get(ctx context.Context, id int) (*HostCategory, error) {
	var cat HostCategory
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/categories/%d", id), &cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

// Create creates a new host category and returns its ID.
func (s *HostCategoryService) Create(ctx context.Context, req CreateHostCategoryRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts/categories", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing host category using PUT.
func (s *HostCategoryService) Update(ctx context.Context, id int, req UpdateHostCategoryRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/hosts/categories/%d", id), req, nil)
}

// Delete deletes a host category by ID.
func (s *HostCategoryService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/categories/%d", id))
}
