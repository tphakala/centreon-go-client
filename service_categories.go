package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ServiceCategory represents a Centreon service category configuration resource.
type ServiceCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// CreateServiceCategoryRequest is the request body for creating a service category.
type CreateServiceCategoryRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// ServiceCategoryService provides service category configuration operations.
type ServiceCategoryService struct {
	client *Client
}

// List returns a paginated list of service categories.
func (s *ServiceCategoryService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ServiceCategory], error) {
	var resp ListResponse[ServiceCategory]
	err := s.client.list(ctx, "/configuration/services/categories", opts, &resp)
	return &resp, err
}

// All returns an iterator over all service categories.
func (s *ServiceCategoryService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ServiceCategory, error] {
	return all(ctx, s.List, opts)
}

// Create creates a new service category and returns its ID.
func (s *ServiceCategoryService) Create(ctx context.Context, req CreateServiceCategoryRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/services/categories", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Delete deletes a service category by ID.
func (s *ServiceCategoryService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/categories/%d", id))
}
