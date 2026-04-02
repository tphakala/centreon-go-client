package centreon

import (
	"context"
	"fmt"
	"iter"
)

// HostTemplate represents a Centreon host template configuration resource.
type HostTemplate struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      *int   `json:"check_command_id"`
	CheckTimeperiodID   *int   `json:"check_timeperiod_id"`
	MaxCheckAttempts    *int   `json:"max_check_attempts"`
	NormalCheckInterval *int   `json:"normal_check_interval"`
	RetryCheckInterval  *int   `json:"retry_check_interval"`
	IsLocked            bool   `json:"is_locked"`
}

// CreateHostTemplateRequest is the request body for creating a host template.
type CreateHostTemplateRequest struct {
	Name           string `json:"name"`
	Alias          string `json:"alias,omitzero"`
	CheckCommandID int    `json:"check_command_id,omitzero"`
}

// UpdateHostTemplateRequest is the request body for updating a host template (PATCH).
type UpdateHostTemplateRequest struct {
	Name                *string `json:"name,omitempty"`
	Alias               *string `json:"alias,omitempty"`
	CheckCommandID      *int    `json:"check_command_id,omitempty"`
	CheckTimeperiodID   *int    `json:"check_timeperiod_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	IsLocked            *bool   `json:"is_locked,omitempty"`
}

// HostTemplateService provides host template configuration operations.
type HostTemplateService struct {
	client *Client
}

// List returns a paginated list of host templates.
func (s *HostTemplateService) List(ctx context.Context, opts ...ListOption) (*ListResponse[HostTemplate], error) {
	var resp ListResponse[HostTemplate]
	err := s.client.list(ctx, "/configuration/hosts/templates", opts, &resp)
	return &resp, err
}

// All returns an iterator over all host templates.
func (s *HostTemplateService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*HostTemplate, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the host template with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *HostTemplateService) GetByID(ctx context.Context, id int) (*HostTemplate, error) {
	return getByID(ctx, s.List, "host template", id)
}

// Create creates a new host template and returns its ID.
func (s *HostTemplateService) Create(ctx context.Context, req CreateHostTemplateRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts/templates", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing host template using PATCH.
func (s *HostTemplateService) Update(ctx context.Context, id int, req UpdateHostTemplateRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/hosts/templates/%d", id), req)
}

// Delete deletes a host template by ID.
func (s *HostTemplateService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/templates/%d", id))
}
