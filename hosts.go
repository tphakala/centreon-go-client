package centreon

import (
	"context"
	"fmt"
	"iter"
)

// Host represents a Centreon host configuration resource.
type Host struct {
	ID                  int    `json:"id"`
	MonitoringServerID  int    `json:"monitoring_server_id"`
	Name                string `json:"name"`
	Address             string `json:"address"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      int    `json:"check_command_id,omitzero"`
	MaxCheckAttempts    int    `json:"max_check_attempts,omitzero"`
	NormalCheckInterval int    `json:"normal_check_interval,omitzero"`
	RetryCheckInterval  int    `json:"retry_check_interval,omitzero"`
	ActiveChecksEnabled *bool  `json:"active_checks_enabled"`
	IsActivated         bool   `json:"is_activated"`
}

// CreateHostRequest is the request body for creating a host.
type CreateHostRequest struct {
	MonitoringServerID int    `json:"monitoring_server_id"`
	Name               string `json:"name"`
	Address            string `json:"address"`
	Alias              string `json:"alias,omitzero"`
	CheckCommandID     int    `json:"check_command_id,omitzero"`
}

// UpdateHostRequest is the request body for updating a host (PATCH).
type UpdateHostRequest struct {
	Name                *string `json:"name,omitempty"`
	Alias               *string `json:"alias,omitempty"`
	Address             *string `json:"address,omitempty"`
	CheckCommandID      *int    `json:"check_command_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	ActiveChecksEnabled *bool   `json:"active_checks_enabled,omitempty"`
	IsActivated         *bool   `json:"is_activated,omitempty"`
}

// HostService provides host configuration operations.
type HostService struct {
	client *Client
}

// List returns a paginated list of hosts.
func (s *HostService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Host], error) {
	var resp ListResponse[Host]
	err := s.client.list(ctx, "/configuration/hosts", opts, &resp)
	return &resp, err
}

// All returns an iterator over all hosts.
func (s *HostService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Host, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the host with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *HostService) GetByID(ctx context.Context, id int) (*Host, error) {
	return getByID(ctx, s.List, "host", id)
}

// Create creates a new host and returns its ID.
func (s *HostService) Create(ctx context.Context, req CreateHostRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing host using PATCH.
func (s *HostService) Update(ctx context.Context, id int, req UpdateHostRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/hosts/%d", id), req, nil)
}

// Delete deletes a host by ID.
func (s *HostService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/%d", id))
}
