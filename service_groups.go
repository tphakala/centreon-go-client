package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ServiceGroup represents a Centreon service group configuration resource.
type ServiceGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// CreateServiceGroupRequest is the request body for creating a service group.
type CreateServiceGroupRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// UpdateServiceGroupRequest is the request body for updating a service group (PUT).
type UpdateServiceGroupRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// ServiceGroupService provides service group configuration operations.
type ServiceGroupService struct {
	client *Client
}

// List returns a paginated list of service groups.
func (s *ServiceGroupService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ServiceGroup], error) {
	var resp ListResponse[ServiceGroup]
	err := s.client.list(ctx, "/configuration/services/groups", opts, &resp)
	return &resp, err
}

// All returns an iterator over all service groups.
func (s *ServiceGroupService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ServiceGroup, error] {
	return all(ctx, s.List, opts)
}

// Get returns the service group with the given ID using a direct GET request.
func (s *ServiceGroupService) Get(ctx context.Context, id int) (*ServiceGroup, error) {
	var sg ServiceGroup
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/services/groups/%d", id), &sg); err != nil {
		return nil, err
	}
	return &sg, nil
}

// Create creates a new service group and returns its ID.
func (s *ServiceGroupService) Create(ctx context.Context, req CreateServiceGroupRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/services/groups", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing service group using PUT.
func (s *ServiceGroupService) Update(ctx context.Context, id int, req UpdateServiceGroupRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/services/groups/%d", id), req)
}

// Delete deletes a service group by ID.
func (s *ServiceGroupService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/groups/%d", id))
}
