package centreon

import (
	"context"
	"fmt"
	"iter"
)

// HostGroup represents a Centreon host group configuration resource.
type HostGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// CreateHostGroupRequest is the request body for creating a host group.
type CreateHostGroupRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// UpdateHostGroupRequest is the request body for updating a host group (PUT).
type UpdateHostGroupRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
}

// HostGroupService provides host group configuration operations.
type HostGroupService struct {
	client *Client
}

// List returns a paginated list of host groups.
func (s *HostGroupService) List(ctx context.Context, opts ...ListOption) (*ListResponse[HostGroup], error) {
	var resp ListResponse[HostGroup]
	err := s.client.list(ctx, "/configuration/hosts/groups", opts, &resp)
	return &resp, err
}

// All returns an iterator over all host groups.
func (s *HostGroupService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*HostGroup, error] {
	return all(ctx, s.List, opts)
}

// Get returns the host group with the given ID using a direct GET request.
func (s *HostGroupService) Get(ctx context.Context, id int) (*HostGroup, error) {
	var hg HostGroup
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/groups/%d", id), &hg); err != nil {
		return nil, err
	}
	return &hg, nil
}

// Create creates a new host group and returns its ID.
func (s *HostGroupService) Create(ctx context.Context, req CreateHostGroupRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts/groups", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing host group using PUT.
func (s *HostGroupService) Update(ctx context.Context, id int, req UpdateHostGroupRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/hosts/groups/%d", id), req)
}

// Delete deletes a host group by ID.
func (s *HostGroupService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/groups/%d", id))
}
