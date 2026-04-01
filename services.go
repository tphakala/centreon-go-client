package centreon

import (
	"context"
	"fmt"
	"iter"
)

// Service represents a Centreon service configuration resource.
type Service struct {
	ID                  int    `json:"id"`
	HostID              int    `json:"host_id"`
	Name                string `json:"name"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      int    `json:"check_command_id,omitzero"`
	MaxCheckAttempts    int    `json:"max_check_attempts,omitzero"`
	NormalCheckInterval int    `json:"normal_check_interval,omitzero"`
	RetryCheckInterval  int    `json:"retry_check_interval,omitzero"`
	ActiveChecksEnabled *bool  `json:"active_checks_enabled"`
	IsActivated         bool   `json:"is_activated"`
}

// CreateServiceRequest is the request body for creating a service.
type CreateServiceRequest struct {
	HostID         int    `json:"host_id"`
	Name           string `json:"name"`
	Alias          string `json:"alias,omitzero"`
	CheckCommandID int    `json:"check_command_id,omitzero"`
}

// UpdateServiceRequest is the request body for updating a service (PATCH).
type UpdateServiceRequest struct {
	Name                *string `json:"name,omitempty"`
	Alias               *string `json:"alias,omitempty"`
	CheckCommandID      *int    `json:"check_command_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	ActiveChecksEnabled *bool   `json:"active_checks_enabled,omitempty"`
	IsActivated         *bool   `json:"is_activated,omitempty"`
}

// ServiceService provides service configuration operations.
type ServiceService struct {
	client *Client
}

// List returns a paginated list of services.
func (s *ServiceService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Service], error) {
	var resp ListResponse[Service]
	err := s.client.list(ctx, "/configuration/services", opts, &resp)
	return &resp, err
}

// All returns an iterator over all services.
func (s *ServiceService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Service, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the service with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *ServiceService) GetByID(ctx context.Context, id int) (*Service, error) {
	return getByID(ctx, s.List, "service", id)
}

// Create creates a new service and returns its ID.
func (s *ServiceService) Create(ctx context.Context, req CreateServiceRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/services", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing service using PATCH.
func (s *ServiceService) Update(ctx context.Context, id int, req UpdateServiceRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/services/%d", id), req, nil)
}

// Delete deletes a service by ID.
func (s *ServiceService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/%d", id))
}
